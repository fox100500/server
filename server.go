package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"

	"server/api"
	"server/biznes"
	"server/config"
	"server/controller"
	"server/mydb"

	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	isRunAPI     = false
	accountID    string
	clientWebSoc *api.WebSocClient
	loggerWebSoc *log.Logger
	loggerHTTP   *log.Logger

	register        *controller.Register
	stocksPortfolio *api.Portfolio
)

//WsCMD ...
type WsCMD struct {
	Command   string `json:"command"`
	EventType string `json:"eventype"`
	Figi      string `json:"figi"`
	Depth     string `json:"depth"`
}

func initEnv() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func handler(c echo.Context) error {

	c.Response().Header().Set(echo.HeaderAccessControlAllowOrigin, "*")
	c.Response().Header().Set(echo.HeaderContentType, "application/json")

	//loggerHTTP.Println(c.Request())

	if isRunAPI == false {
		loggerHTTP.Println("http.StatusNotAcceptable isRunAPI=", isRunAPI)
		return c.JSON(http.StatusNotAcceptable, nil)
	}

	switch c.QueryParam("type") {

	case "getportfolio":
		portfolio := stocksPortfolio.PrepareForSend()
		err := c.JSON(http.StatusOK, *portfolio)
		return err

	case "getdescription":
		ticker := c.QueryParam("ticker")
		loggerHTTP.Println("Try search", ticker)
		stock, err := api.GetDescription(ticker)
		if err != nil {
			loggerHTTP.Println(err)
			return c.JSON(http.StatusNotFound, nil)
		}
		return c.JSON(http.StatusOK, stock)

	case "registerorderbook":
		ticker := c.QueryParam("ticker")
		depth, err := strconv.Atoi(c.QueryParam("depth"))
		if err != nil {
			loggerHTTP.Println(err)
			return c.JSON(http.StatusBadRequest, err)
		}
		err = clientWebSoc.RegisterOrderBookByTicker(ticker, depth)
		if err != nil {
			return c.JSON(http.StatusNotModified, err)
		}
		return c.JSON(http.StatusOK, "succesfull")

	case "unregisterorderbook":
		ticker := c.QueryParam("ticker")
		depth, err := strconv.Atoi(c.QueryParam("depth"))
		if err != nil {
			loggerHTTP.Println(err)
			return c.JSON(http.StatusBadRequest, err)
		}
		err = clientWebSoc.UnregisterOrderBookByTicker(ticker, depth)
		if err != nil {
			loggerHTTP.Println(err)
			return c.JSON(http.StatusNotModified, err)
		}
		return c.JSON(http.StatusOK, "succesfull")

	case "getorders":
		orders, _ := api.GetOrders(stocksPortfolio.AccountID)
		err := c.JSON(http.StatusOK, orders)
		log.Println(orders)
		return err

	case "extorder":
		s, _ := ioutil.ReadAll(c.Request().Body)
		log.Printf("%s\n", s)
		var order api.TExtorder
		if err := json.Unmarshal(s, &order); err != nil {
			log.Println("Can't unmarshal register response err:", err)
			err = c.JSON(http.StatusBadRequest, err)
			return err
		}

		loggerHTTP.Println("Try search", order.Ticker)
		if _, err := api.GetDescription(order.Ticker); err != nil {
			loggerHTTP.Println(err)
			return c.JSON(http.StatusNotFound, err)
		}

		result, err := stocksPortfolio.ExecuteOrder(order)

		log.Println("1111", err)

		if err != nil {
			loggerHTTP.Println(err)
			return c.JSON(http.StatusNotFound, err)
		}
		c.JSON(http.StatusOK, result)
		return nil

	default:
		log.Println(c.Request())
		return c.JSON(http.StatusBadRequest, nil)
	}
}

func wsReadMsg(ws *websocket.Conn, ch chan string, connectID int, flExit *bool) {
	defer func() {
		fmt.Println("wsReadMsg Check recover ... ")
		if recoverMsg := recover(); recoverMsg != nil {
			fmt.Println("done", recoverMsg)
		} else {
			fmt.Println("normal exit.")
		}
	}()
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			fmt.Println("Exit from gorut wsReadMsg[", connectID, "]", err)
			*flExit = true
			return
		}
		//log.Printf("wsReadMsg[%+v] %s", connectID, string(msg))
		ch <- string(msg)
	}
}

func wsParseMsg(msg string) (cmd WsCMD) {
	err := json.Unmarshal([]byte(msg), &cmd)
	if err != nil {
		log.Println(err)
		return
	}
	//log.Println(cmd)
	return cmd
}

var (
	upgrader = websocket.Upgrader{}
)

func closeChan(ctx echo.Context, ch chan string, id int) {
	fmt.Println("Defer close respCh", id)
	close(ch)
}

func wsHandler(ctx echo.Context) error {
	counter := 0
	connectID := rand.Int()

	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	ws, err := upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		ctx.Logger().Error(err)
		return err
	}
	defer ws.Close()

	respCh := make(chan string, 10)
	defer closeChan(ctx, respCh, connectID)

	flExit := false
	go wsReadMsg(ws, respCh, connectID, &flExit)

	for {
		counter++
		msgPack, ok := <-respCh
		if !ok {
			ctx.Logger().Printf("respCh closed. Try close ws...")
			if err := ws.Close(); err != nil {
				ctx.Logger().Printf(" result %s %e ", err, err)
			}
			return nil
		}

		cmd := wsParseMsg(msgPack)

		if flExit {
			cmd.Command = "exit"
		}
		switch cmd.Command {
		case "exit":
			ctx.Logger().Printf("wsHandler exit")
			fmt.Println("wsHandler exit")
			if err := ws.Close(); err != nil {
				ctx.Logger().Printf("Try exit...%s", err)
			}
			return nil
		case "register":
			fmt.Println(cmd.EventType, cmd.Figi)
			register.AddListenerByFigi(cmd.EventType, cmd.Figi, ws)
			ctx.Logger().Printf("Register %s %s", cmd.EventType, cmd.Figi)

			switch cmd.EventType {
			case "candle":
				err := clientWebSoc.RegisterCandle(cmd.Figi, "5min")
				if err != nil {
					ctx.Logger().Printf("RegisterCandle %s %s %s", cmd.EventType, cmd.Figi, err)
					register.DeleteListener(cmd.EventType, cmd.Figi, ws)
					ctx.Logger().Printf("UnregisterCandle %s %s", cmd.EventType, cmd.Figi)
				}
			case "instrument_info":
				err := clientWebSoc.RegisterInfo(cmd.Figi)
				if err != nil {
					ctx.Logger().Printf("RegisterInfo %s %s %s", cmd.EventType, cmd.Figi, err)
					register.DeleteListener(cmd.EventType, cmd.Figi, ws)
					ctx.Logger().Printf("UnregisterInfo %s %s", cmd.EventType, cmd.Figi)
				}
			case "orderbook":
				depth, err := strconv.Atoi(cmd.Depth)
				if err != nil {
					depth = 5
				}
				err = clientWebSoc.RegisterOrderbook(cmd.Figi, depth)
				if err != nil {
					ctx.Logger().Printf("RegisterOrderbook %s %s %s", cmd.EventType, cmd.Figi, err)
					register.DeleteListener(cmd.EventType, cmd.Figi, ws)
					ctx.Logger().Printf("UnregisterOrderbook %s %s", cmd.EventType, cmd.Figi)
				}
			}

		case "unregister":
			register.DeleteListener(cmd.EventType, cmd.Figi, ws)
			switch cmd.EventType {
			case "candle":
				err := clientWebSoc.UnregisterCandle(cmd.Figi, "5min")
				if err != nil {
					ctx.Logger().Printf("UnregisterCandle %s %s %s", cmd.EventType, cmd.Figi, err)
				}
			case "instrument_info":
				err := clientWebSoc.UnregisterInfo(cmd.Figi)
				if err != nil {
					ctx.Logger().Printf("UnregisterInfo %s %s %s", cmd.EventType, cmd.Figi, err)
				}
			case "orderbook":
				depth, err := strconv.Atoi(cmd.Depth)
				if err != nil {
					depth = 5
				}
				err = clientWebSoc.UnregisterOrderbook(cmd.Figi, depth)
				if err != nil {
					ctx.Logger().Printf("UnregisterOrderBook %s %s %s", cmd.EventType, cmd.Figi, err)
				}
			}

			ctx.Logger().Printf("delEvent %s %s", cmd.EventType, cmd.Figi)
		}

	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	f, err := os.OpenFile("main.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		return
	}
	defer f.Close()

	initEnv()
	cfg := config.New()

	// assign it to the standard logger
	log.SetOutput(f)

	log.Println("//******************************************************//")
	log.Println("//                         Start                        //")
	log.Println("//******************************************************//")

	if err := biznes.Init(); err == nil {

		accounts := biznes.GetAccounts()
		accountID = accounts[0].BrokerAccountID

		register = controller.New("register")

		fmt.Println("PORTWS", cfg.Web.PortWS)
		ws := echo.New()
		ws.Logger.SetOutput(f)
		ws.Debug = true
		ws.Logger.SetPrefix("[pref_websoc_server]")
		ws.Logger.SetHeader("[websocSERVER]")
		ws.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}))

		ws.Use(middleware.Recover()) //CORS
		ws.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		}))
		ws.GET("/ws", wsHandler)

		go func() {
			fmt.Println("Start ws server")
			fmt.Println(cfg.Web.PortWS)
			ws.Logger.Fatal(ws.Start(cfg.Web.PortWS))
			fmt.Println("Stop ws server")
		}()

		loggerHTTP = log.New(f, "[restHTTP]", log.LstdFlags)       //os.Stdout
		loggerWebSoc = log.New(f, "[websocCLIENT]", log.LstdFlags) //os.Stdout
		clientWebSoc, err = api.NewWebSoc(loggerWebSoc, cfg.Connect.TokenReal, cfg.Connect.URLWebSoc)
		if err != nil {
			loggerWebSoc.Println(err)
			os.Exit(1)
		}
		defer clientWebSoc.Close()

		mydb.Init()

		stocksPortfolio = api.NewPortfolio("TinkofBroker", accountID, clientWebSoc, 1, api.CandleInterval1Day)
		go stocksPortfolio.Refresh()

		go clientWebSoc.ListenWebSoc(stocksPortfolio, register)
		time.Sleep(200 * time.Millisecond)

		isRunAPI = true

		//http server (rest)
		fmt.Println("PORT", cfg.Web.Port)
		httpServ := echo.New()

		httpServ.Logger.SetOutput(f)
		httpServ.Debug = true
		httpServ.Logger.SetPrefix("[pref_websoc_server]")
		httpServ.Logger.SetHeader("[websocSERVER]")

		httpServ.Use(middleware.Recover()) //CORS
		httpServ.Use(middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{echo.GET, echo.OPTIONS, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		}))

		httpServ.GET("/", handler)
		httpServ.POST("/", handler)
		go func() {
			fmt.Println("Start http server")
			httpServ.Logger.Fatal(httpServ.Start(cfg.Web.Port))
			fmt.Println("Stop http server")
		}()

		for isRunAPI {
			time.Sleep(1 * time.Second)
		}

		loggerWebSoc.Println("Stop server...")

		for _, value := range api.InfoMap {
			err = clientWebSoc.UnregisterListenerByFigi(value.FIGI, api.CandleInterval5Min)
			if err != nil {
				loggerWebSoc.Fatalln(err)
			}
		}

		for _, value := range api.OrderBookMap {
			err = clientWebSoc.UnregisterOrderBookByFigi(value.FIGI, api.OrderBookMap[value.FIGI].Depth)
			if err != nil {
				loggerWebSoc.Fatalln(err)
			}
		}
	}
	log.Println("//******************************************************//")
	log.Println("//                         Stop                         //")
	log.Println("//******************************************************//")
}

/*
	httpServ.Pre(middleware.HTTPSRedirect())
	httpServ.Pre(middleware.HTTPSWWWRedirect())
	httpServ.Pre(middleware.HTTPSNonWWWRedirect())
	httpServ.Pre(middleware.WWWRedirect())
	httpServ.Pre(middleware.NonWWWRedirect())
*/
/*
	httpServ.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
*/
