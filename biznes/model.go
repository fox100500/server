package biznes

import (
	"errors"
	"fmt"
	"log"
	"server/api"
)

//Balance ...
var Balance float64

//Orders ...
var Orders []api.Order

//ID ...
var ID string

//TBalance ...
type TBalance struct {
	SumBalance float64 `json:"summa"`
	RubBalance float64 `json:"rub"`
	UsdBalance float64 `json:"usd"`
	EurBalance float64 `json:"eur"`
}

//Init ...
func Init() error { // open a file
	return api.Init()
}

//GetBalance ...
func GetBalance(accountID string) (*TBalance, error) {

	var Balance = &TBalance{
		SumBalance: 0,
		RubBalance: 0,
		UsdBalance: 0,
		EurBalance: 0,
	}

	portfolioStocks, err := api.GetPortfolio(accountID)
	if err != nil {
		fmt.Println("model portfolioStocks", err)
		return nil, err
	}

	portfolioCurrencies, err := api.GetCurrencies(accountID)
	if err != nil {
		fmt.Println("model portfolioCurrencies", err)
		return nil, err
	}

	m := make(map[string]float64, 3)
	m["RUB"] = 1
	m["USD"] = api.GetCursWebSoc("USD000UTSTOM")
	m["EUR"] = api.GetCursWebSoc("EUR_RUB__TOM")

	if len(portfolioStocks) > 0 {
		for i := range portfolioStocks {
			balance := portfolioStocks[i].Balance*portfolioStocks[i].AveragePositionPrice.Value + portfolioStocks[i].ExpectedYield.Value
			Balance.SumBalance += balance * m[string(portfolioStocks[i].AveragePositionPrice.Currency)]
		}
	}

	if len(portfolioCurrencies) > 0 {
		for i := range portfolioCurrencies {
			switch portfolioCurrencies[i].TCurrence {
			case "RUB":
				//				log.Println(i, portfolioCurrencies[i].TCurrence, "kurs", m[string(portfolioCurrencies[i].TCurrence)], portfolioCurrencies[i].Balance)
				Balance.SumBalance += float64(portfolioCurrencies[i].Balance) * m[string(portfolioCurrencies[i].TCurrence)]
				Balance.RubBalance = float64(portfolioCurrencies[i].Balance)
			case "USD":
				Balance.UsdBalance = float64(portfolioCurrencies[i].Balance)
			case "EUR":
				Balance.EurBalance = float64(portfolioCurrencies[i].Balance)
			default:
				log.Println("Not evalute in balance", portfolioCurrencies[i].TCurrence, portfolioCurrencies[i].Balance)
			}
		}
	}
	//	fmt.Println("Balance", *Balance)
	return Balance, nil
}

//GetAccounts ...
func GetAccounts() []api.UserAccount {
	return api.GetUserAccounts()
}

//GetOrderBookOld ...
func GetOrderBookOld(ticker string, depth int) (*api.OrderBookOld, error) {
	instrument, err := api.GetDescription(ticker)
	if err != nil {
		return nil, err
	}
	orderbook, err := api.GetOrderBookOld(instrument.FIGI, depth)
	if err != nil {
		return nil, err
	}
	return orderbook, nil
}

//SearchByTicker ...
func searchByTicker(ticker string, instruments []api.MarketInstrument) (api.MarketInstrument, error) {

	if len(instruments) > 0 {
		for key := range instruments {
			if instruments[key].Ticker == ticker {
				return instruments[key], nil
			}
		}
	}
	return instruments[0], errors.New("ticker not found")
}

//
func RegisterOrderBook(ticker string, depth int) (*api.Candle, error) {
	instrument, err := api.GetDescription(ticker)
	if err != nil {
		return nil, err
	}
	fmt.Println(instrument.FIGI, depth)
	//api.RegisterOrderBook(instrument.FIGI, depth)

	return nil, nil
}

/*
func RefreshOrderBookData(ticker) (*api.Candle, error) {
	return api.RefreshOrderBookData(ticker)
}

func WebGetOrderBook(ticker) (*api.Candle, error) {
	return api.WebGetOrderBook(ticker)
}
*/
