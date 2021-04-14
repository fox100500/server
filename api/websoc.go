package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"server/controller"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

//
var CandleMap map[string]Candle

//
var InfoMap map[string]InstrumentInfo

//
var OrderBookMap map[string]OrderBook

var letterRunes = []rune("1234efghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//RequestID ... Генерируем уникальный ID для запроса
func RequestID() string {
	b := make([]rune, 12)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}

//WebSocClient ...
type WebSocClient struct {
	logger  *log.Logger
	hWebSoc *websocket.Conn
	token   string
	url     string
	isRun   bool
}

/*
//Logger ...
type Logger interface {
	Printf(format string, args ...interface{})
}
*/
//NewWebSoc ... Init WebSoC client
func NewWebSoc(logger *log.Logger, token string, url string) (*WebSocClient, error) {
	clientWebSoc := &WebSocClient{
		logger: logger,
		token:  token,
		url:    url,
		isRun:  true,
	}

	rand.Seed(time.Now().UnixNano()) // инициируем Seed рандома для функции requestID

	var err error
	clientWebSoc.hWebSoc, err = clientWebSoc.connect()
	if err != nil {
		clientWebSoc.isRun = false
		return nil, err
	}

	CandleMap = make(map[string]Candle, 10)
	InfoMap = make(map[string]InstrumentInfo, 10)
	OrderBookMap = make(map[string]OrderBook, 10)

	return clientWebSoc, nil
}

//connect ... WebSoc client connect
func (ctx WebSocClient) connect() (*websocket.Conn, error) {
	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 5 * time.Second,
	}

	hWebSoc, resp, err := dialer.Dial(ctx.url, http.Header{"Authorization": {"Bearer " + ctx.token}})
	if err != nil {
		if resp != nil {
			if resp.StatusCode == http.StatusForbidden {
				return nil, ErrForbidden
			}
			if resp.StatusCode == http.StatusUnauthorized {
				return nil, ErrUnauthorized
			}

			return nil, errors.Wrapf(err, "can't connect to %s %s", ctx.url, resp.Status)
		}
		return nil, errors.Wrapf(err, "can't connect to %s", ctx.url)
	}
	defer resp.Body.Close()

	hWebSoc.SetPingHandler(func(message string) error {
		err := hWebSoc.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(time.Second))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		return err
	})

	return hWebSoc, nil
}

//Close ... Close WebSoc client
func (ctx *WebSocClient) Close() error {
	ctx.isRun = false
	ctx.logger.Println("Close websoc client")
	return ctx.hWebSoc.Close()
}

//RegisterInfo ... Подписаться на получение событий по инструменту
func (ctx *WebSocClient) RegisterInfo(figi string) error {
	sub := `{"event": "instrument_info:subscribe", "request_id": "` + RequestID() + `", "figi": "` + figi + `"}`
	if err := ctx.hWebSoc.WriteMessage(websocket.TextMessage, []byte(sub)); err != nil {
		ctx.logger.Println("Регистрация на события не выполнена - can't subscribe to event:", figi)
		return errors.Wrap(err, "can't subscribe to event")
	}
	ctx.logger.Println("Регистрация на события выполнена:", figi)
	return nil
}

//UnregisterInfo ... Отписаться от получения событий по инструменту
func (ctx *WebSocClient) UnregisterInfo(figi string) error {
	sub := `{"event": "instrument_info:unsubscribe", "request_id": "` + RequestID() + `", "figi": "` + figi + `"}`
	if err := ctx.hWebSoc.WriteMessage(websocket.TextMessage, []byte(sub)); err != nil {
		ctx.logger.Println("Отмена регистрации на события не выполнена - can't unsubscribe from event", figi)
		return errors.Wrap(err, "can't unsubscribe from event")
	}
	ctx.logger.Println("Отмена регистрации на события:", figi)
	return nil
}

//RegisterCandle ... Подписаться на получение свечей по инструменту
func (ctx *WebSocClient) RegisterCandle(figi string, interval CandleInterval) error {
	sub := `{ 
		"event": "candle:subscribe",
		"request_id": "` + RequestID() +
		`", "figi": "` + figi +
		`", "interval": "` + string(interval) + `"}`
	if err := ctx.hWebSoc.WriteMessage(websocket.TextMessage, []byte(sub)); err != nil {
		ctx.logger.Println("Регистрация на свечи не выполнена - : can't subscribe to event", figi, interval)
		return errors.Wrap(err, "can't subscribe to event")
	}
	ctx.logger.Println("Регистрация на свечи выполнена:", figi, interval)
	return nil
}

//UnregisterCandle ... Отписаться от получения свечей по инструменту
func (ctx *WebSocClient) UnregisterCandle(figi string, interval CandleInterval) error {
	sub := `{ "event": "candle:unsubscribe", "request_id": "` + RequestID() + `", "figi": "` + figi + `", "interval": "` + string(interval) + `"}`
	if err := ctx.hWebSoc.WriteMessage(websocket.TextMessage, []byte(sub)); err != nil {
		ctx.logger.Println("Отписка от свечей не выполнена: can't unsubscribe from event:", figi, interval)
		return errors.Wrap(err, "can't unsubscribe from event")
	}
	ctx.logger.Println("Отписка от получения свечей выполнена:", figi, interval)
	return nil
}

//RegisterOrderbook ... Подписаться на получение стакана по инструменту
func (ctx *WebSocClient) RegisterOrderbook(figi string, depth int) error {
	if depth < 1 || depth > MaxOrderbookDepth {
		ctx.logger.Println("Подписка на стакана не выполнена - ", ErrDepth, ":", figi, depth)
		return ErrDepth
	}

	sub := `{ "event": "orderbook:subscribe", "request_id": "` + RequestID() + `", "figi": "` + figi + `", "depth": ` + strconv.Itoa(depth) + `}`
	if err := ctx.hWebSoc.WriteMessage(websocket.TextMessage, []byte(sub)); err != nil {
		ctx.logger.Println("Подписка на стакан не выполнена - can't subscribe to event:", figi, depth)
		return errors.Wrap(err, "can't subscribe to event")
	}
	ctx.logger.Println("Подписка на стакан выполнена:", figi, depth)
	return nil
}

//UnregisterOrderbook ... Отписаться от получения стакана по инструменту
func (ctx *WebSocClient) UnregisterOrderbook(figi string, depth int) error {
	if depth < 1 || depth > MaxOrderbookDepth {
		ctx.logger.Print("Отписка от стакана не выполнена - ", ErrDepth, ":", figi, depth)
		return ErrDepth
	}

	sub := `{ "event": "orderbook:unsubscribe", "request_id": "` + RequestID() + `", "figi": "` + figi + `", "depth": ` + strconv.Itoa(depth) + `}`
	if err := ctx.hWebSoc.WriteMessage(websocket.TextMessage, []byte(sub)); err != nil {
		ctx.logger.Print("Отписка от стакана не выполнена - can't unsubscribe from event:", figi, depth)
		return errors.Wrap(err, "can't unsubscribe from event")
	}
	ctx.logger.Print("Отписка от стакана выполнена:", figi, depth)
	return nil
}

func parseEvent(event interface{}, portfolio *Portfolio, register *controller.Register) error {

	return nil
}

//ListenWebSoc ... collect data from websoc
func (ctx *WebSocClient) ListenWebSoc(portfolio *Portfolio, register *controller.Register) { //register *controller.Register
	ctx.logger.Println("Listening websoc")
	err := ctx.parseMessage(portfolio, register)
	if err != nil {
		return
	}
	if ctx.isRun == false {
		return
	}
	ctx.logger.Println("Wait next websoc message ")
}

func (ctx *WebSocClient) wsReadMsg(chMsg chan []byte) error {
	for {
		_, msg, err := ctx.hWebSoc.ReadMessage()
		if err != nil {
			return errors.Wrap(err, "can't read message")
		}
		chMsg <- msg
	}
}

//
func (ctx *WebSocClient) parseMessage(portfolio *Portfolio, register *controller.Register) error {
	chMsg := make(chan []byte, 100)
	go ctx.wsReadMsg(chMsg)

	for {

		msg, ok := <-chMsg
		//ctx.logger.Println(string(msg))
		if !ok {
			ctx.logger.Println("chMsg closed. Try close ws...")
			if err := ctx.hWebSoc.Close(); err != nil {
				ctx.logger.Println(" result", err.Error(), err)
				return err
			}
			ctx.logger.Println("Ok")
			return nil
		}

		var event Event
		if err := json.Unmarshal(msg, &event); err != nil {
			ctx.logger.Println("Can't unmarshal event", string(msg))
			continue
		}

		switch event.Name {
		case "candle":
			var event CandleEvent
			if err := json.Unmarshal(msg, &event); err != nil {
				ctx.logger.Println("Can't unmarshal event candle", string(msg))
				continue
			}
			//ctx.logger.Println("CandleEvent:", event)
			portfolio.EventRefreshStock(nil, &event)
		case "instrument_info":
			var event InstrumentInfoEvent
			if err := json.Unmarshal(msg, &event); err != nil {
				ctx.logger.Println("Can't unmarshal event instrument_info", string(msg))
				continue
			}
			//ctx.logger.Println("InfoEvent:", event)
			portfolio.EventRefreshStock(&event, nil)
		case "orderbook":
			var event OrderBookEvent
			if err := json.Unmarshal(msg, &event); err != nil {
				ctx.logger.Println("Can't unmarshal event orderbook", string(msg))
				continue
			}
			//ctx.logger.Println("OrderBookEvent:", event)
			err := register.DispatchEvent(event.FullEvent.Name, event.OrderBook.FIGI, &event)
			if err != nil {
				go func() {
					err := ctx.UnregisterOrderBookByFigi(event.OrderBook.FIGI, event.OrderBook.Depth)
					if err != nil {
						log.Println("ListenWebSoc UnregisterOrderBook", err)
						fmt.Println("ListenWebSoc UnregisterOrderBook", err)
					}
					return
				}()
				continue
			}
		case "error":
			var event ErrorEvent
			if err := json.Unmarshal(msg, &event); err != nil {
				ctx.logger.Println("Can't unmarshal event error", string(msg))
				continue
			}
			ctx.logger.Println("ErrorEvent:", event)
		default:
			ctx.logger.Println("Get unknown event", string(msg))
		}
	}
}

//ErrDepth ... Недопустимое значение глубины стакана
var ErrDepth = errors.Errorf("invalid depth. Should be in interval 0 < x <= %d", MaxOrderbookDepth)

//ErrForbidden ... Ошибка доступа, неверный токен
var ErrForbidden = errors.New("invalid token")

//ErrUnauthorized ...
var ErrUnauthorized = errors.New("token not provided")

/*
func errorHandle(err error) error {
	if err == nil {
		return nil
	}

	if tradingErr, ok := err.(TradingError); ok {
		if tradingErr.InvalidTokenSpace() {
			tradingErr.Hint = "Do you use sandbox token in production environment or vise verse?"
			return tradingErr
		}
	}

	return err
}

func (t TradingError) Error() string {
	return fmt.Sprintf(
		"TrackingID: %s, Status: %s, Message: %s, Code: %s, Hint: %s",
		t.TrackingID, t.Status, t.Payload.Message, t.Payload.Code, t.Hint,
	)
}

func (t TradingError) NotEnoughBalance() bool {
	return t.Payload.Code == "NOT_ENOUGH_BALANCE"
}

func (t TradingError) InvalidTokenSpace() bool {
	return t.Payload.Message == "Invalid token scopes"
}
*/
//RegisterOrderBookByFigi ... Регистрация на инструмент
func (ctx *WebSocClient) RegisterOrderBookByFigi(figi string, depth int) error {
	err := ctx.RegisterOrderbook(figi, depth)
	if err != nil {
		ctx.logger.Println("RegOrderBook err", err)
		return err
	}

	return nil
}

//RegisterOrderBookByTicker ... Регистрация на стакан по тикеру
func (ctx *WebSocClient) RegisterOrderBookByTicker(ticker string, depth int) error {
	instrument, err := GetDescription(ticker)
	if err != nil {
		return err
	}

	return ctx.RegisterOrderBookByFigi(instrument.FIGI, depth)
}

//UnregisterOrderBookByFigi ... Отмена регистрации на стакан
func (ctx *WebSocClient) UnregisterOrderBookByFigi(figi string, depth int) error {
	err := ctx.UnregisterOrderbook(figi, depth)
	if err != nil {
		ctx.logger.Println("UnregOrderbook", err)
	}

	delete(OrderBookMap, figi)

	return nil
}

//UnregisterOrderBookByTicker ... Отмена регистрации на стакан по тикеру
func (ctx *WebSocClient) UnregisterOrderBookByTicker(ticker string, depth int) error {
	instrument, err := GetDescription(ticker)
	if err != nil {
		return err
	}

	return ctx.UnregisterOrderBookByFigi(instrument.FIGI, depth)
}

//RegisterListenerByFigi ... Комплексная регистрация на инструмент
func (ctx *WebSocClient) RegisterListenerByFigi(figi string, interval CandleInterval) error {
	err := ctx.RegisterInfo(figi)
	if err != nil {
		ctx.logger.Println("RegInfo err", err)
		return err
	}

	err = ctx.RegisterCandle(figi, interval)
	if err != nil {
		ctx.logger.Println("RegCandle err", err)
		return err
	}
	/*
		err = ctx.RegisterOrderbook(figi, 5)
		if err != nil {
			ctx.logger.Println("RegOrdeBook err", err)
			return err
		}
	*/
	return nil
}

//RegisterListenerByTicker ... Комплексная регистрация на инструмент
func (ctx *WebSocClient) RegisterListenerByTicker(ticker string, interval CandleInterval) error {
	instrument, err := GetDescription(ticker)
	if err != nil {
		return err
	}

	return ctx.RegisterListenerByFigi(instrument.FIGI, interval)
}

//UnregisterListenerByFigi ... Отмена комплексной регистрации на инструмент
func (ctx *WebSocClient) UnregisterListenerByFigi(figi string, interval CandleInterval) error {
	err := ctx.UnregisterInfo(figi)
	if err != nil {
		ctx.logger.Println("UnregInfo err", err)
	}

	err = ctx.UnregisterCandle(figi, interval)
	if err != nil {
		ctx.logger.Println("UnregCandle err", err)
	}

	delete(InfoMap, figi)
	delete(CandleMap, figi)

	return nil
}

//UnregisterListenerByTicker ... Отмена комплексной регистрации на инструмент
func (ctx *WebSocClient) UnregisterListenerByTicker(ticker string, interval CandleInterval) error {
	instrument, err := GetDescription(ticker)
	if err != nil {
		return err
	}

	return ctx.UnregisterListenerByFigi(instrument.FIGI, interval)
}

//RegisterPortfolio ... Регистрация инструментов из портфеля
func (ctx *WebSocClient) RegisterPortfolio(accountID string, interval CandleInterval) error {
	portfolioStocks, err := GetPortfolio(accountID)
	if err != nil {
		ctx.logger.Println("Register portfolio", err)
		return err
	}

	for i := range portfolioStocks {
		ctx.RegisterListenerByFigi(portfolioStocks[i].FIGI, interval)
	}

	return nil
}

/*
//GetInfo ...
func (ctx *WebSocClient) GetInfo(figi string) (*FullInfo, error) {

	info, ok := InfoMap[figi]
	if !ok {
		return nil, errors.New("figi_not_found")
	}

	finfo.mutex.Lock()
	finfo.instrument = info
	finfo.mutex.Unlock()

	candle, ok := CandleMap[figi]
	if ok {
		finfo.mutex.Lock()
		finfo.candle = candle
		finfo.mutex.Unlock()
	}

	orderBook, ok := OrderBookMap[figi]
	if ok {
		finfo.mutex.Lock()
		finfo.orderBook = orderBook
		finfo.mutex.Unlock()
	}

	return &finfo, nil
}

//RepackInfo ...
func RepackInfo(finfo *FullInfo) (info Info) {
	finfo.mutex.Lock()
	defer finfo.mutex.Unlock()

	info.Figi = finfo.instrument.FIGI
	info.TradeStatus = finfo.instrument.TradeStatus
	info.MinPriceIncrement = finfo.instrument.MinPriceIncrement
	info.Lot = finfo.instrument.Lot
	info.AccruedInterest = finfo.instrument.AccruedInterest
	info.LimitUp = finfo.instrument.LimitUp
	info.LimitDown = finfo.instrument.LimitDown

	info.Interval = finfo.candle.Interval
	info.OpenPrice = finfo.candle.OpenPrice
	info.ClosePrice = finfo.candle.ClosePrice
	info.HighPrice = finfo.candle.HighPrice
	info.LowPrice = finfo.candle.LowPrice
	info.Volume = finfo.candle.Volume
	info.Time = finfo.candle.Time

	info.Depth = finfo.orderBook.Depth

	var tmp RestPriceQuantity

	if len(finfo.orderBook.Bids) > 0 {
		for i := range finfo.orderBook.Bids {
			tmp.Price = finfo.orderBook.Bids[i][0]
			tmp.Quantity = finfo.orderBook.Bids[i][1]
			//fmt.Println("b", tmp)
			info.Bids = append(info.Bids, tmp)
			//fmt.Println("Bids", i, info.Bids[i].Price, info.Bids[i].Quantity)
		}

		for i := range finfo.orderBook.Asks {
			tmp.Price = finfo.orderBook.Asks[i][0]
			tmp.Quantity = finfo.orderBook.Asks[i][1]
			//fmt.Println("a", tmp)
			info.Asks = append(info.Asks, tmp)
			//fmt.Println("Asks", i, info.Asks[i].Price, info.Asks[i].Quantity)
		}
	}
	return info
}
*/
