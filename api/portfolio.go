package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"server/mydb"
	"time"
)

//NewPortfolio ...
func NewPortfolio(idx string, account string, clientWS *WebSocClient, refreshTimeout int, candleInterval CandleInterval) (ctx *Portfolio) {
	ctx = &Portfolio{
		ID:              idx,
		AccountID:       account,
		clientWS:        clientWS,
		CountStocks:     0,
		CountCurrencies: 0,
		CandleInterval:  candleInterval,
		RefreshTimeout:  refreshTimeout,
	}

	ctx.Stocks = make(map[string]Instrument)
	ctx.Currencies = make(map[string]CurrencyPosition)
	ctx.Orders = make(map[string]Order)
	ctx.DefferedOrders = make(map[string]TDefferedOrder)

	items, err := GetPortfolio(ctx.AccountID)
	if err != nil {
		log.Println("GetPortfolio main items", err)
		return ctx
	}

	if len(items) > 0 {
		for _, value := range items {
			ctx.AddStock(ctx.CreateStock(&value))
			ctx.clientWS.RegisterListenerByFigi(value.FIGI, candleInterval)
		}
	}

	currencies, err := GetCurrencies(ctx.AccountID)
	if err != nil {
		log.Println("GetPortfolio main currencies", err)
		return ctx
	}

	if len(currencies) > 0 {
		for _, value := range currencies {
			ctx.AddCurrency(ctx.CreateCurrency(&value))
		}
	}

	ctx.CalculateBalance()

	return ctx
}

//CalculateBalance ...
func (ctx *Portfolio) CalculateBalance() {
	m := make(map[string]float64, 3)
	m["RUB"] = 1
	m["USD"] = ctx.Stocks["BBG0013HGFT4"].ClosePrice
	m["EUR"] = ctx.Stocks["BBG0013HJJ31"].ClosePrice

	//	log.Println(m)

	ctx.SumBalance = 0
	if len(ctx.Stocks) > 0 {
		for key := range ctx.Stocks {
			balance := ctx.Stocks[key].Balance*ctx.Stocks[key].AvgPriceValue + ctx.Stocks[key].EYvalue
			ctx.SumBalance += balance * m[string(ctx.Stocks[key].Currency)]
		}
	}

	if len(ctx.Currencies) > 0 {
		for key := range ctx.Currencies {
			switch ctx.Currencies[key].TCurrence {
			case "RUB":
				ctx.SumBalance += float64(ctx.Currencies[key].Balance) * m[string(ctx.Currencies[key].TCurrence)]
				ctx.RubBalance = float64(ctx.Currencies[key].Balance)
			case "USD":
				ctx.UsdBalance = float64(ctx.Currencies[key].Balance)
			case "EUR":
				ctx.EurBalance = float64(ctx.Currencies[key].Balance)
			default:
				log.Println("Not evalute in balance", ctx.Currencies[key].TCurrence, ctx.Currencies[key].Balance)
			}
		}
	}
	//log.Println(ctx.SumBalance, ctx.RubBalance, ctx.UsdBalance, ctx.EurBalance)
}

//Refresh ...
func (ctx *Portfolio) Refresh() error {
	for {
		time.Sleep(time.Duration(ctx.RefreshTimeout) * time.Second)
		items, err := GetPortfolio(ctx.AccountID)
		if err != nil {
			log.Println("portfolio not avaliable", ctx.ID, err)
			if err.Error() == "too many requests" {
				time.Sleep(10 * time.Second)
			} else {
				//return errors.New("portfolio not avaliable:" + err.Error())
			}
		} else {
			if len(items) > 0 {

				ctx.mutex.Lock()
				stocksOld := make(map[string]Instrument)
				for key, val := range ctx.Stocks {
					stocksOld[key] = val
				}
				ctx.mutex.Unlock()

				//Поиск позиций выбывших из портфеля и удаление их регистрации
				if len(stocksOld) > 0 {
					for _, stockOld := range stocksOld {
						isPresent := false
						for _, stockCurrent := range items {
							if stockOld.Figi == stockCurrent.FIGI {
								isPresent = true
								break
							}
						}
						if isPresent == false {
							ctx.DeleteStock(stockOld.Figi)
							ctx.clientWS.UnregisterListenerByFigi(stockOld.Figi, ctx.CandleInterval)
						}
					}
				}
				//Добавление новых позиций в портфель и их регистрация
				for _, stockCurrent := range items {
					if ctx.AddStock(ctx.CreateStock(&stockCurrent)) {
						ctx.clientWS.RegisterListenerByFigi(stockCurrent.FIGI, ctx.CandleInterval)
					} else {
						ctx.LoopRefreshStock(&stockCurrent)
					}
				}
			}
		}

		currencies, err := GetCurrencies(ctx.AccountID)
		if err != nil {
			log.Println("currencies not avaliable", ctx.ID, err)
			if err.Error() == "too many requests" {
				time.Sleep(10 * time.Second)
			} else {
				return errors.New("currencies not avaliable:" + err.Error())
			}
		} else {
			if len(currencies) > 0 {
				ctx.mutex.Lock()
				currenciesOld := make(map[string]CurrencyPosition)
				for key, val := range ctx.Currencies {
					currenciesOld[key] = val
				}
				ctx.mutex.Unlock()

				//Поиск позиций выбывших из портфеля и удаление их регистрации
				if len(currenciesOld) > 0 {
					for _, currencyOld := range currenciesOld {
						isPresent := false
						for _, currencyCurrent := range currencies {
							if currencyOld.TCurrence == currencyCurrent.TCurrence {
								isPresent = true
								break
							}
						}
						if isPresent == false {
							ctx.DeleteCurrency(currencyOld.TCurrence)
						}
					}
				}
				//Добавление новых позиций в портфель и их регистрация
				for _, currencyCurrent := range currencies {
					if !ctx.AddCurrency(ctx.CreateCurrency(&currencyCurrent)) {
						ctx.LoopRefreshCurrency(&currencyCurrent)
					}
				}
			}
		}

		ctx.mutex.Lock()
		ctx.CalculateBalance()
		ctx.mutex.Unlock()

		orderlist, err := GetOrders(ctx.AccountID)
		if err != nil {
			log.Println("orderslist not avaliable", ctx.ID, err)
			if err.Error() == "too many requests" {
				time.Sleep(10 * time.Second)
			} else {
				//return errors.New("portfolio not avaliable:" + err.Error())
			}
		} else {
			if len(orderlist) > 0 {
				log.Println("Orderslist")
				for key, value := range orderlist {
					log.Println(key, value)
				}
			} else {
				log.Println("Orderslist empty")
			}
		}

		if len(ctx.DefferedOrders) > 0 {
			log.Println("DefferedOrderslist")
			for key, item := range ctx.DefferedOrders {
				log.Println(key, item)
			}
		}

	}
}

//AddStock ... add item
func (ctx *Portfolio) AddStock(item *Instrument) bool {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if _, ok := ctx.Stocks[item.Figi]; ok {
		return false
	}
	ctx.Stocks[item.Figi] = *item
	ctx.CountStocks++
	return true
}

//AddCurrency ... add item
func (ctx *Portfolio) AddCurrency(item *CurrencyPosition) bool {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if _, ok := ctx.Currencies[string(item.TCurrence)]; ok {
		return false
	}
	ctx.Currencies[string(item.TCurrence)] = *item
	ctx.CountCurrencies++
	return true
}

//DeleteStock ... delete item
func (ctx *Portfolio) DeleteStock(figi string) error {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if ctx.CountStocks > 0 {
		if _, ok := ctx.Stocks[figi]; !ok {
			return errors.New(" DeleteStock figi not found " + figi)
		}
		ctx.CountStocks--
		delete(ctx.Stocks, figi)
		return nil
	}
	return errors.New("portfolio empty")
}

//DeleteCurrency ... delete item
func (ctx *Portfolio) DeleteCurrency(currency Currency) error {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if ctx.CountCurrencies > 0 {
		if _, ok := ctx.Currencies[string(currency)]; !ok {
			return errors.New("DeleteCurrency not found " + string(currency))
		}
		ctx.CountCurrencies--
		delete(ctx.Stocks, string(currency))
		return nil
	}
	return errors.New("currencies empty")
}

//CreateStock ...
func (ctx *Portfolio) CreateStock(itemData *PositionBalance) *Instrument {

	var item Instrument

	item.Ticker = itemData.Ticker
	item.Figi = itemData.FIGI
	item.Name = itemData.Name
	item.Type = itemData.InstrumentType
	item.InfoTime = time.Now()
	item.Lots = itemData.Lots
	item.Balance = itemData.Balance
	item.Blocked = itemData.Blocked
	item.Currency = string(itemData.AveragePositionPrice.Currency)
	item.EYvalue = itemData.ExpectedYield.Value
	item.AvgPriceValue = itemData.AveragePositionPrice.Value
	item.AvgPriceValueNoNkd = itemData.AveragePositionPriceNoNkd.Value

	return &item
}

//CreateCurrency ...
func (ctx *Portfolio) CreateCurrency(itemData *CurrencyPosition) *CurrencyPosition {

	var item CurrencyPosition

	item.Balance = itemData.Balance
	item.TCurrence = itemData.TCurrence

	return &item
}

//LoopRefreshStock ...
func (ctx *Portfolio) LoopRefreshStock(itemSrc *PositionBalance) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	itemDst := ctx.Stocks[itemSrc.FIGI]
	itemDst.InfoTime = time.Now()
	itemDst.Balance = itemSrc.Balance
	itemDst.Blocked = itemSrc.Blocked
	itemDst.EYvalue = itemSrc.ExpectedYield.Value
	itemDst.AvgPriceValue = itemSrc.AveragePositionPrice.Value
	itemDst.AvgPriceValueNoNkd = itemSrc.AveragePositionPriceNoNkd.Value
	itemDst.Currency = string(itemSrc.AveragePositionPrice.Currency)
	ctx.Stocks[itemSrc.FIGI] = itemDst
}

//LoopRefreshCurrency ...
func (ctx *Portfolio) LoopRefreshCurrency(itemSrc *CurrencyPosition) {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	itemDst := ctx.Currencies[string(itemSrc.TCurrence)]
	itemDst.Balance = itemSrc.Balance
	itemDst.TCurrence = itemSrc.TCurrence
	ctx.Currencies[string(itemSrc.TCurrence)] = itemDst
}

//EventRefreshStock ...
func (ctx *Portfolio) EventRefreshStock(infoEvent *InstrumentInfoEvent, candleEvent *CandleEvent) error {
	defer func() {
		recover()
	}()

	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if infoEvent != nil {
		figi := infoEvent.Info.FIGI

		item, ok := ctx.Stocks[figi]
		if !ok {
			log.Println("figi not found111")
			return errors.New("figi not found")
		}

		if infoEvent.Info.TradeStatus == "normal_trading" {
			item.TradeStatus = true
		} else {
			item.TradeStatus = false
		}

		item.InfoTime = infoEvent.Time
		item.MinPriceIncrement = infoEvent.Info.MinPriceIncrement
		item.Lot = infoEvent.Info.Lot
		item.LimitUp = infoEvent.Info.LimitUp
		item.LimitDown = infoEvent.Info.LimitDown
		item.AccruedInterest = infoEvent.Info.AccruedInterest
		ctx.Stocks[figi] = item
		return nil
	}

	if candleEvent != nil {
		figi := candleEvent.Candle.Figi

		item, ok := ctx.Stocks[figi]
		if !ok {
			log.Println("figi not found")
			return errors.New("figi not found")
		}

		item.CandleTime = candleEvent.Time
		item.Interval = candleEvent.Candle.Interval
		item.OpenPrice = candleEvent.Candle.OpenPrice
		item.ClosePrice = candleEvent.Candle.ClosePrice
		item.HighPrice = candleEvent.Candle.HighPrice
		item.LowPrice = candleEvent.Candle.LowPrice
		item.Volume = candleEvent.Candle.Volume

		ctx.Stocks[figi] = item

		ctx.checkDefferedOrders(item)

		return nil
	}
	return nil
}

//Print ...
func (ctx *Portfolio) Print() {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	log.Println("stocksPortfolio", ctx.CountStocks)
	for _, stock := range ctx.Stocks {

		data, err := json.Marshal(stock)
		if err != nil {
			fmt.Println("err")
		}
		log.Println(string(data))
	}
}

//PrepareForSend ...
func (ctx *Portfolio) PrepareForSend() *SmallPacket {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	defer func() {
		if recoverMsg := recover(); recoverMsg != nil {
			log.Println("done", recoverMsg)
			fmt.Println("done", recoverMsg)
		}
	}()

	var packet SmallPacket
	//	start := time.Now()
	packet.SumBalance = ctx.SumBalance
	packet.RubBalance = ctx.RubBalance
	packet.UsdBalance = ctx.UsdBalance
	packet.EurBalance = ctx.EurBalance

	for _, value := range ctx.Stocks {
		packet.Stocks = append(packet.Stocks, value)
	}

	//	fmt.Println(time.Since(start))

	return &packet
}

//ExecuteOrder ...
func (ctx *Portfolio) ExecuteOrder(order TExtorder) (*TPlacedOrder, error) {
	log.Println(order)
	var result *TPlacedOrder
	var err error

	if order.Operation == string(BUY) || order.Operation == string(SELL) {
		switch order.Typed {
		case "market":
			result, err = SetMarketOrder(ctx.AccountID, order.Figi, order.Lots, OperationType(order.Operation))
		case "limit":
			result, err = SetLimitOrder(ctx.AccountID, order.Figi, order.Lots, order.Price, OperationType(order.Operation))
			log.Println(result, err)
		case "iflimit":
			ctx.addDefferedOrder(order)
		default:
			return nil, errors.New("order untyped")
		}

		if order.Takeprofit.Enabled || order.Stoploss.Enabled || order.Trailingstop.Enabled {
			ctx.addDefferedOrder(order)
		}

		if err != nil {
			log.Println("try exit err")
			return nil, err
		}

		log.Println("try exit norm")
		return result, nil
	}

	if order.Operation == "closeall" {

	}

	return nil, errors.New("unknown operation")
}

func (ctx *Portfolio) addDefferedOrder(order TExtorder) bool {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if _, ok := ctx.DefferedOrders[order.Figi]; ok {
		return false
	}

	var defOrder TDefferedOrder

	defOrder.Ticker = order.Ticker
	defOrder.Figi = order.Figi
	defOrder.Typed = order.Typed
	defOrder.Operation = order.Operation
	defOrder.Price = order.Price
	defOrder.Takeprofit.Enabled = order.Takeprofit.Enabled
	defOrder.Takeprofit.Price = order.Takeprofit.Price
	defOrder.Takeprofit.Lots = order.Takeprofit.Lots
	defOrder.Stoploss.Enabled = order.Stoploss.Enabled
	defOrder.Stoploss.Price = order.Stoploss.Price
	defOrder.Stoploss.Lots = order.Stoploss.Lots
	defOrder.Trailingstop.Enabled = order.Trailingstop.Enabled
	defOrder.Trailingstop.Size = order.Trailingstop.Size
	defOrder.OrderID = RequestID()
	defOrder.Status = OrderStatusNew
	defOrder.RequestedLots = order.Lots
	defOrder.ExecutedLots = 0
	defOrder.Duratation = time.Second

	ctx.DefferedOrders[order.Figi] = defOrder

	dbOrder := mydb.TExtOrder{
		Ticker:                order.Ticker,
		Figi:                  order.Figi,
		Typed:                 order.Typed,
		Operation:             order.Operation,
		Price:                 order.Price,
		Lots:                  order.Lots,
		TakeprofitEnabled:     order.Takeprofit.Enabled,
		TakeprofitPrice:       order.Takeprofit.Price,
		TakeprofitLots:        order.Takeprofit.Lots,
		TakeprofitEndDataTime: time.Now(),
		StoplossEnabled:       order.Stoploss.Enabled,
		StoplossPrice:         order.Stoploss.Price,
		StoplossLots:          order.Stoploss.Lots,
		StopLossEndDataTime:   time.Now(),
		TrailingstopEnabled:   order.Trailingstop.Enabled,
		TrailingstopSize:      order.Trailingstop.Size,
		OrderStartDateTime:    time.Now(),
	}
	mydb.Insert(dbOrder)
	return true
}

func (ctx *Portfolio) checkDefferedOrders(item Instrument) {
	if order, ok := ctx.DefferedOrders[item.Figi]; ok {
		if item.ClosePrice >= order.Takeprofit.Price {
			log.Println("Takeprofit setLimitSellOrder")

			SetMarketOrder(ctx.AccountID, order.Figi, order.Takeprofit.Lots, SELL)
			ctx.DeleteDefferedOrder(item.Figi)
			return
		}

		if item.ClosePrice <= order.Stoploss.Price {
			log.Println("Stoploss setLimitSellOrder")

			SetMarketOrder(ctx.AccountID, order.Figi, order.Stoploss.Lots, SELL)
			ctx.DeleteDefferedOrder(item.Figi)
			return
		}

		if (item.ClosePrice - order.Stoploss.Price) > order.Trailingstop.Size {
			log.Println("Trailingstop change Stoploss")
			log.Println("item.ClosePrice", item.ClosePrice, "order.Trailingstop.Size", order.Trailingstop.Size)

			k := int((item.ClosePrice - order.Trailingstop.Size) / item.MinPriceIncrement)

			order.Stoploss.Price = order.Stoploss.Price + item.MinPriceIncrement*float64(k)
			ctx.DefferedOrders[item.Figi] = order
			log.Println("New order.Stoploss.Price", order.Stoploss.Price)
		}
	}
}

//DeleteDefferedOrder ... delete item
func (ctx *Portfolio) DeleteDefferedOrder(figi string) error {
	ctx.mutex.Lock()
	defer ctx.mutex.Unlock()

	if len(ctx.DefferedOrders) > 0 {
		if _, ok := ctx.DefferedOrders[figi]; !ok {
			return errors.New(" DeleteDefferedOrder figi not found " + figi)
		}
		delete(ctx.DefferedOrders, figi)
		mydb.Delete(figi)
		return nil
	}
	return errors.New("DefferedOrderList empty")
}
