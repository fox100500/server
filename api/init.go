package api

import (
	"net/http"
	"server/config"
)

//
func Init() (err error) {

	cfg := config.New()

	const reqName = "Register"

	hClient = http.Client{}

	if cfg.Connect.IsSendBox {
		apiURL = cfg.Connect.URLSandBox
		token = cfg.Connect.TokenSandBox
		Register(AccountTinkoff)
		Register(AccountTinkoffIIS)
	} else {
		apiURL = cfg.Connect.URLReal
		token = cfg.Connect.TokenReal
	}
	websocURL = cfg.Connect.URLWebSoc

	Stocks, err = GetMarketStocks()
	if err != nil {
		return err
	}

	Bonds, err = GetMarketBonds()
	if err != nil {
		return err
	}
	Etfs, err = GetMarketEtfs()
	if err != nil {
		return err
	}
	CurrenciesAvailable, err = GetAvalableCurrencies()
	if err != nil {
		return err
	}
	return nil
	//Загрузка данных однократно
	/*

		SetBalance(accountID, 101, "RUB")
		SetBalance(accountID, 102, "USD")
		SetBalance(accountID, 103, "EUR")
		SetSandboxPositions(ID, "BBG00B3T3HD3", 20)

		GetOperations(
			ID,
			"BBG00B3T3HD3",
			"2020-12-19T18:38:33.131642+03:00",
			"2020-12-24T18:38:33.131642+03:00",
		)

		GetOrderBook("BBG00B3T3HD3", 20)

		GetCandles(
			"BBG00B3T3HD3",
			"2020-12-17T18:38:33.131642+03:00",
			"2020-12-24T18:38:33.131642+03:00",
			CandleInterval1Day,
		)

		result := SetLimitOrder(
			accounts.PayLoad.Accounts[0].BrokerAccountID,
			"BBG00B3T3HD3",
			5,
			20,
			BUY,
		)

		SetMarketOrder(
			accounts.PayLoad.Accounts[0].BrokerAccountID,
			"BBG00B3T3HD3", //Alcoa
			1,
			SELL,
		)

		SearchByFigi("BBG00B3T3HD3")

		SearchByTicker("AA")

		OrderCancel(ID, result.Payload.OrderID)

		SandboxClear(ID)

		SandboxRemove(ID)
	*/

}
