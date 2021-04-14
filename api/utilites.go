package api

import (
	"errors"
	"log"
	"time"
)

func GetDescription(ticker string) (*MarketInstrument, error) {
	where := CurrenciesAvailable
	for i := range where {
		if where[i].Ticker == ticker {
			//log.Println("figi["+ticker+"]=", where[i].FIGI)
			return &where[i], nil
		}
	}

	where = Stocks
	for i := range where {
		if where[i].Ticker == ticker {
			//log.Println("figi["+ticker+"]=", where[i].FIGI)
			return &where[i], nil
		}
	}

	where = Etfs
	for i := range where {
		if where[i].Ticker == ticker {
			//log.Println("figi["+ticker+"]=", where[i].FIGI)
			return &where[i], nil
		}
	}

	where = Bonds
	for i := range where {
		if where[i].Ticker == ticker {
			//log.Println("figi["+ticker+"]=", where[i].FIGI)
			return &where[i], nil
		}
	}

	log.Println("figi[" + ticker + "]= not found")

	return nil, errors.New("Ticker " + ticker + " not found")
}

//GetLastCandle ...
func GetLastCandle(ticker string, interval CandleInterval) (*Candle, error) {
	chInterval := interval
	_, _, err := find(interval, CandleIntervals)
	if err != nil {
		chInterval = CandleInterval5Min
		log.Println("Bad interval", interval, "default", chInterval)
	}
	log.Println("chInterval", chInterval)
	offset := time.Minute
	switch chInterval {
	case CandleInterval1Min:
		offset = 1 * time.Minute
	case CandleInterval2Min:
		offset = 2 * time.Minute
	case CandleInterval3Min:
		offset = 3 * time.Minute
	case CandleInterval5Min:
		offset = 5 * time.Minute
	case CandleInterval10Min:
		offset = 10 * time.Minute
	case CandleInterval15Min:
		offset = 15 * time.Minute
	case CandleInterval30Min:
		offset = 30 * time.Minute
	case CandleInterval1Hour:
		offset = 1 * time.Hour
	case CandleInterval2Hour:
		offset = 2 * time.Hour
	case CandleInterval4Hour:
		offset = 4 * time.Hour
	case CandleInterval1Day:
		offset = 24 * time.Hour
	case CandleInterval1Week:
		offset = 7 * 24 * time.Hour
	case CandleInterval1Month:
		offset = 31 * 24 * time.Hour
	}

	//fmt.Println(offset)

	stopTime := time.Now()
	startTime := stopTime.Add(-offset)
	startRFC3339 := startTime.Format(time.RFC3339)
	stopRFC3339 := stopTime.Format(time.RFC3339)

	log.Println("startRFC3339", startRFC3339)
	log.Println("stoptRFC3339", stopRFC3339)

	instrument, err := GetDescription(ticker)
	if err != nil {
		return nil, err
	}
	candles, err := GetCandles(instrument.FIGI, startRFC3339, stopRFC3339, chInterval)
	if err != nil {
		return nil, err
	}

	return &candles[0], nil
}

//GetCurs ...
func GetCurs(ticker string) float64 {
	candle, err := GetLastCandle(ticker, "day")
	if err != nil {
		return 1
	}
	return *&candle.ClosePrice
}

//
func GetCursWebSoc(ticker string) float64 {
	instrument, err := GetDescription(ticker)
	if err != nil {
		return 1
	}

	candle, ok := CandleMap[instrument.FIGI]
	if ok {
		return candle.ClosePrice
	}
	return 1
}

func find(what interface{}, where []interface{}) (idx int, value interface{}, err error) {
	for i, value := range where {
		if value == what {
			return i, value, nil
		}
	}
	return -1, nil, errors.New("value not found")
}

//fmt.Println(find(what, where))
