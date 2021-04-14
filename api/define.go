package api

import (
	"net/http"
	"sync"
	"time"
)

//Stocks - массив описаний элементов
var Stocks []MarketInstrument

//Bonds - массив описаний элементов
var Bonds []MarketInstrument

//Etfs - массив описаний элементов
var Etfs []MarketInstrument

//CurrenciesAvailable ...
var CurrenciesAvailable []MarketInstrument

//Balance current
var Balance float64

//**************************************************
var apiURL string
var hClient http.Client
var token string
var websocURL string

const MaxOrderbookDepth = 20

//
type AccountType string

//
const (
	AccountTinkoff    AccountType = "Tinkoff"
	AccountTinkoffIIS AccountType = "TinkoffIis"
)

//
type Currency string

//
const (
	RUB Currency = "RUB"
	USD Currency = "USD"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
	HKD Currency = "HKD"
	CHF Currency = "CHF"
	JPY Currency = "JPY"
	CNY Currency = "CNY"
	TRY Currency = "TRY"
)

//
type InstrumentType string

//
const (
	Stock     InstrumentType = "Stock"
	Currency_ InstrumentType = "Currency"
	Bond      InstrumentType = "Bond"
	Etf       InstrumentType = "Etf"
)

//
type UserAccount struct {
	BrokerAccountType AccountType `json:"brokerAccountType"`
	BrokerAccountID   string      `json:"brokerAccountId"`
}

//
type UserAccounts struct {
	Accounts []UserAccount
}

//
type UserAccountsResponse struct {
	TrackingID string       `json:"trackingId"`
	Status     string       `json:"status"`
	PayLoad    UserAccounts `json:"payload"`
}

//****************************************
//
type CurrencyPosition struct {
	TCurrence Currency `json:"currency"`
	Balance   float32  `json:"balance"`
}

//
type Currencies struct {
	Currencies []CurrencyPosition `json:"currencies"`
}

//
type PortfolioCurrenciesResponse struct {
	TrackingID string `json:"trackingId"`
	Status     string `json:"status"`
	PayLoad    Currencies
}

//
type MoneyAmount struct {
	Currency Currency `json:"currency"`
	Value    float64  `json:"value"`
}

//
type PositionBalance struct {
	FIGI                      string         `json:"figi"`
	Ticker                    string         `json:"ticker"`
	ISIN                      string         `json:"isin"`
	InstrumentType            InstrumentType `json:"instrumentType"`
	Balance                   float64        `json:"balance"`
	Blocked                   float64        `json:"blocked"`
	Lots                      int            `json:"lots"`
	ExpectedYield             MoneyAmount    `json:"expectedYield"`
	AveragePositionPrice      MoneyAmount    `json:"averagePositionPrice"`
	AveragePositionPriceNoNkd MoneyAmount    `json:"averagePositionPriceNoNkd"`
	Name                      string         `json:"name"`
}

//
type PortfolioResponse struct {
	TrackingID string `json:"trackingId"`
	Status     string `json:"status"`
	Payload    struct {
		Positions []PositionBalance `json:"positions"`
	} `json:"payload"`
}

//
type MarketInstrument struct {
	FIGI                      string         `json:"figi"`
	Ticker                    string         `json:"ticker"`
	ISIN                      string         `json:"isin"`
	InstrumentType            InstrumentType `json:"instrumentType"`
	Balance                   float64        `json:"balance"`
	Blocked                   float64        `json:"blocked"`
	Lots                      int            `json:"lots"`
	ExpectedYield             MoneyAmount    `json:"expectedYield"`
	AveragePositionPrice      MoneyAmount    `json:"averagePositionPrice"`
	AveragePositionPriceNoNkd MoneyAmount    `json:"averagePositionPriceNoNkd"`
	Name                      string         `json:"name"`
}

//
type StocksInfo struct {
	Total       int32              `json:"total"`
	Instruments []MarketInstrument `json:"instruments"`
}

//
type MarketInstrumentListResponse struct {
	TrackingID string     `json:"trackingId"`
	Status     string     `json:"status"`
	Payload    StocksInfo `json:"payload"`
}

//***************************************
//
type OperationType string

//
const (
	BUY                             OperationType = "Buy"
	SELL                            OperationType = "Sell"
	OperationTypeBrokerCommission   OperationType = "BrokerCommission"
	OperationTypeExchangeCommission OperationType = "ExchangeCommission"
	OperationTypeServiceCommission  OperationType = "ServiceCommission"
	OperationTypeMarginCommission   OperationType = "MarginCommission"
	OperationTypeOtherCommission    OperationType = "OtherCommission"
	OperationTypePayIn              OperationType = "PayIn"
	OperationTypePayOut             OperationType = "PayOut"
	OperationTypeTax                OperationType = "Tax"
	OperationTypeTaxLucre           OperationType = "TaxLucre"
	OperationTypeTaxDividend        OperationType = "TaxDividend"
	OperationTypeTaxCoupon          OperationType = "TaxCoupon"
	OperationTypeTaxBack            OperationType = "TaxBack"
	OperationTypeRepayment          OperationType = "Repayment"
	OperationTypePartRepayment      OperationType = "PartRepayment"
	OperationTypeCoupon             OperationType = "Coupon"
	OperationTypeDividend           OperationType = "Dividend"
	OperationTypeSecurityIn         OperationType = "SecurityIn"
	OperationTypeSecurityOut        OperationType = "SecurityOut"
	OperationTypeBuyCard            OperationType = "BuyCard"
)

//
type OrderStatus string

//
const (
	OrderStatusNew            OrderStatus = "New"
	OrderStatusPartiallyFill  OrderStatus = "PartiallyFill"
	OrderStatusFill           OrderStatus = "Fill"
	OrderStatusCancelled      OrderStatus = "Cancelled"
	OrderStatusReplaced       OrderStatus = "Replaced"
	OrderStatusPendingCancel  OrderStatus = "PendingCancel"
	OrderStatusRejected       OrderStatus = "Rejected"
	OrderStatusPendingReplace OrderStatus = "PendingReplace"
	OrderStatusPendingNew     OrderStatus = "PendingNew"
)

//
type OrderType string

//
const (
	OrderTypeLimit  OrderType = "Limit"
	OrderTypeMarket OrderType = "Market"
)

//
type Order struct {
	OrderID       string        `json:"orderId"`
	FIGI          string        `json:"figi"`
	Operation     OperationType `json:"operation"`
	Status        OrderStatus   `json:"status"`
	RequestedLots int           `json:"requestedLots"`
	ExecutedLots  int           `json:"executedLots"`
	Type          OrderType     `json:"type"`
	Price         float64       `json:"price"`
}

//
type OrdersResponse struct {
	TrackingID string  `json:"trackingId"`
	Status     string  `json:"status"`
	Payload    []Order `json:"payload"`
}

//**************************
//
type OperationStatus string

//
const (
	OperationStatusDone     OperationStatus = "Done"
	OperationStatusDecline  OperationStatus = "Decline"
	OperationStatusProgress OperationStatus = "Progress"
)

//
type Trade struct {
	ID       string    `json:"tradeId"`
	DateTime time.Time `json:"date"`
	Price    float64   `json:"price"`
	Quantity int       `json:"quantity"`
}

//
type Operation struct {
	ID               string          `json:"id"`
	Status           OperationStatus `json:"status"`
	Trades           []Trade         `json:"trades"`
	Commission       MoneyAmount     `json:"commission"`
	Currency         Currency        `json:"currency"`
	Payment          float64         `json:"payment"`
	Price            float64         `json:"price"`
	Quantity         int             `json:"quantity"`
	QuantityExecuted int             `json:"quantityExecuted"`
	FIGI             string          `json:"figi"`
	InstrumentType   InstrumentType  `json:"instrumentType"`
	IsMarginCall     bool            `json:"isMarginCall"`
	DateTime         time.Time       `json:"date"`
	OperationType    OperationType   `json:"operationType"`
}

//
type OperationResponse struct {
	TrackingID string `json:"trackingId"`
	Status     string `json:"status"`
	Payload    struct {
		Operations []Operation `json:"operations"`
	} `json:"payload"`
}

//******************************************************
//
type TradingStatus string

//
const (
	NormalTrading          TradingStatus = "NormalTrading"
	NotAvailableForTrading TradingStatus = "NotAvailableForTrading"
)

//
type RestPriceQuantity struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

//OrderBookOld ...
type OrderBookOld struct {
	FIGI              string              `json:"figi"`
	Depth             int                 `json:"depth"`
	Bids              []RestPriceQuantity `json:"bids"`
	Asks              []RestPriceQuantity `json:"asks"`
	TradeStatus       TradingStatus       `json:"tradeStatus"`
	MinPriceIncrement float64             `json:"minPriceIncrement"`
	LastPrice         float64             `json:"lastPrice,omitempty"`
	ClosePrice        float64             `json:"closePrice,omitempty"`
	LimitUp           float64             `json:"limitUp,omitempty"`
	LimitDown         float64             `json:"limitDown,omitempty"`
	FaceValue         float64             `json:"faceValue,omitempty"`
}

//
type OrderBookResponse struct {
	TrackingID string       `json:"trackingId"`
	Status     string       `json:"status"`
	Payload    OrderBookOld `json:"payload"`
}

//**************************************************
//TPlacedOrder ...
type TPlacedOrder struct {
	OrderID       string        `json:"orderId"`
	Operation     OperationType `json:"operation"`
	Status        OrderStatus   `json:"status"`
	RejectReason  string        `json:"rejectReason"`
	RequestedLots int           `json:"requestedLots"`
	ExecutedLots  int           `json:"executedLots"`
	Commission    MoneyAmount   `json:"commission"`
	Message       string        `json:"message,omitempty"`
}

//
type OrderResponse struct {
	TrackingID string       `json:"trackingId"`
	Status     string       `json:"status"`
	Payload    TPlacedOrder `json:"payload"`
}

//*******************************************
//
type EmptyResponce struct {
	TrackingID string `json:"trackingId"`
	Status     string `json:"status"`
	Payload    struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"payload"`
}

//****************************************************
type CandleInterval string

//
const (
	CandleInterval1Min   CandleInterval = "1min"
	CandleInterval2Min   CandleInterval = "2min"
	CandleInterval3Min   CandleInterval = "3min"
	CandleInterval5Min   CandleInterval = "5min"
	CandleInterval10Min  CandleInterval = "10min"
	CandleInterval15Min  CandleInterval = "15min"
	CandleInterval30Min  CandleInterval = "30min"
	CandleInterval1Hour  CandleInterval = "hour"
	CandleInterval2Hour  CandleInterval = "2hour"
	CandleInterval4Hour  CandleInterval = "4hour"
	CandleInterval1Day   CandleInterval = "day"
	CandleInterval1Week  CandleInterval = "week"
	CandleInterval1Month CandleInterval = "month"
)

var CandleIntervals = []interface{}{
	"1min",
	"2min",
	"3min",
	"5min",
	"10min",
	"15min",
	"30min",
	"hour",
	"2hour",
	"4hour",
	"day",
	"week",
	"month",
}

//*********************************************
//
type Candle struct {
	Figi       string         `json:"figi"`
	Interval   CandleInterval `json:"interval"`
	OpenPrice  float64        `json:"o"`
	ClosePrice float64        `json:"c"`
	HighPrice  float64        `json:"h"`
	LowPrice   float64        `json:"l"`
	Volume     float64        `json:"v"`
	Time       string         `json:"time"`
}

//
type CandlesResponse struct {
	TrackingID string `json:"trackingId"`
	Status     string `json:"status"`
	Payload    struct {
		Figi     string   `json:"figi"`
		Interval string   `json:"interval"`
		Candles  []Candle `json:"candles"`
	} `json:"payload"`
}

//**********************************************
//
type SearchMarketInstrument struct {
	Figi              string         `json:"figi"`
	Ticker            string         `json:"ticker"`
	Isin              string         `json:"isin"`
	MinPriceIncrement float64        `json:"minPriceIncrement"`
	Lot               int            `json:"lot"`
	Currency          Currency       `json:"currency"`
	Name              string         `json:"name"`
	Type              InstrumentType `json:"type"`
}

//
type SearchMarketInstrumentResponse struct {
	TrackingID string                 `json:"trackingId"`
	Status     string                 `json:"status"`
	Payload    SearchMarketInstrument `json:"payload"`
}

//**********************************************
//
type SearchMarketInstrumentByTicker struct {
	Figi              string         `json:"figi"`
	Ticker            string         `json:"ticker"`
	Isin              string         `json:"isin"`
	MinPriceIncrement float64        `json:"minPriceIncrement"`
	Lot               int            `json:"lot"`
	MinQuantity       int            `json:"minQuantity"`
	Currency          Currency       `json:"currency"`
	Name              string         `json:"name"`
	Type              InstrumentType `json:"type"`
}

//
type SearchMarketInstrumentList struct {
	Total       int                              `json:"total"`
	Instruments []SearchMarketInstrumentByTicker `json:"instruments"`
}

//
type SearchMarketInstrumentListResponse struct {
	TrackingID string                     `json:"trackingId"`
	Status     string                     `json:"status"`
	Payload    SearchMarketInstrumentList `json:"payload"`
}

//*************************
type Event struct {
	Name string `json:"event"`
}

type FullEvent struct {
	Name string    `json:"event"`
	Time time.Time `json:"time"`
}

type CandleEvent struct {
	FullEvent
	Candle Candle `json:"payload"`
}

type OrderBookEvent struct {
	FullEvent
	OrderBook OrderBook `json:"payload"`
}

//
type OrderBook struct {
	FIGI  string          `json:"figi"`
	Depth int             `json:"depth"`
	Bids  []PriceQuantity `json:"bids"`
	Asks  []PriceQuantity `json:"asks"`
}

//
type PriceQuantity [2]float64

type InstrumentInfoEvent struct {
	FullEvent
	Info InstrumentInfo `json:"payload"`
}

type InstrumentInfo struct {
	FIGI              string        `json:"figi"`
	TradeStatus       TradingStatus `json:"trade_status"`
	MinPriceIncrement float64       `json:"min_price_increment"`
	Lot               float64       `json:"lot"`
	AccruedInterest   float64       `json:"accrued_interest,omitempty"`
	LimitUp           float64       `json:"limit_up,omitempty"`
	LimitDown         float64       `json:"limit_down,omitempty"`
}

type ErrorEvent struct {
	FullEvent
	Error Error `json:"payload"`
}

type Error struct {
	RequestID string `json:"request_id,omitempty"`
	Error     string `json:"error"`
}

type TradingError struct {
	TrackingID string `json:"trackingId"`
	Status     string `json:"status"`
	Hint       string
	Payload    struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"payload"`
}

//Instrument ... Stock description
type Instrument struct {
	//instrument_info
	Ticker      string         `json:"ticker"`
	Figi        string         `json:"figi"`
	TradeStatus bool           `json:"tradingstatus"`
	Name        string         `json:"name"`
	Type        InstrumentType `json:"instrumentType"`

	InfoTime           time.Time `json:"infotime"`
	MinPriceIncrement  float64   `json:"min_price_increment"`
	Lot                float64   `json:"lot"`
	Lots               int       `json:"lots"`
	AccruedInterest    float64   `json:"accrued_interest,omitempty"`
	LimitUp            float64   `json:"limit_up,omitempty"`
	LimitDown          float64   `json:"limit_down,omitempty"`
	Balance            float64   `json:"balance"`
	Blocked            float64   `json:"blocked"`
	Currency           string    `json:"currency"`
	EYvalue            float64   `json:"eyvalue"`
	AvgPriceValue      float64   `json:"avgpricevalue"`
	AvgPriceValueNoNkd float64   `json:"avgpricevaluenonkd"`

	//candle
	CandleTime time.Time      `json:"candletime"`
	Interval   CandleInterval `json:"interval"`
	OpenPrice  float64        `json:"open"`
	ClosePrice float64        `json:"close"`
	HighPrice  float64        `json:"high"`
	LowPrice   float64        `json:"low"`
	Volume     float64        `json:"volume"`
}

//TExtorder ...
type TExtorder struct {
	Ticker     string  `json:"ticker"`
	Figi       string  `json:"figi"`
	Typed      string  `json:"typed"`
	Operation  string  `json:"operation"`
	Price      float64 `json:"price"`
	Lots       int     `json:"lots"`
	Takeprofit struct {
		Enabled bool    `json:"enabled"`
		Price   float64 `json:"price"`
		Lots    int     `json:"lots"`
	} `json:"takeprofit"`
	Stoploss struct {
		Enabled bool    `json:"enabled"`
		Price   float64 `json:"price"`
		Lots    int     `json:"lots"`
	} `json:"stoploss"`
	Trailingstop struct {
		Enabled bool    `json:"enabled"`
		Size    float64 `json:"size"`
	} `json:"trailingstop"`
}

//TDefferedOrder ...
type TDefferedOrder struct {
	Ticker     string  `json:"ticker"`
	Figi       string  `json:"figi"`
	Typed      string  `json:"typed"`
	Operation  string  `json:"operation"`
	Price      float64 `json:"price"`
	Takeprofit struct {
		Enabled bool    `json:"enabled"`
		Price   float64 `json:"price"`
		Lots    int     `json:"lots"`
	} `json:"takeprofit"`
	Stoploss struct {
		Enabled bool    `json:"enabled"`
		Price   float64 `json:"price"`
		Lots    int     `json:"lots"`
	} `json:"stoploss"`
	Trailingstop struct {
		Enabled bool    `json:"enabled"`
		Size    float64 `json:"size"`
	} `json:"trailingstop"`

	OrderID       string        `json:"orderId"`
	Status        OrderStatus   `json:"status"`
	RequestedLots int           `json:"requestedLots"`
	ExecutedLots  int           `json:"executedLots"`
	Duratation    time.Duration `json:"duratation"`
}

//SmallPacket ...
type SmallPacket struct {
	SumBalance float64      `json:"summa"`
	RubBalance float64      `json:"rub"`
	UsdBalance float64      `json:"usd"`
	EurBalance float64      `json:"eur"`
	Stocks     []Instrument `json:"stocks"`
}

//Portfolio ...
type Portfolio struct {
	ID              string
	AccountID       string
	RefreshTimeout  int
	clientWS        *WebSocClient
	CountStocks     int            `json:"countstocks"`
	CountCurrencies int            `json:"countcurrencies"`
	CandleInterval  CandleInterval `json:"candleinterval"`
	SumBalance      float64        `json:"summa"`
	RubBalance      float64        `json:"rub"`
	UsdBalance      float64        `json:"usd"`
	EurBalance      float64        `json:"eur"`
	mutex           sync.RWMutex
	Stocks          map[string]Instrument       `json:"stocks"`
	Currencies      map[string]CurrencyPosition `json:"currencies"`
	Orders          map[string]Order            `json:"orders"`
	DefferedOrders  map[string]TDefferedOrder   `json:"defferedorders"`
}
