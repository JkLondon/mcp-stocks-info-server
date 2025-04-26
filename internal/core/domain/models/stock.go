package models

import (
	"time"
)

// Stock представляет собой информацию об акции
type Stock struct {
	Ticker     string    `json:"ticker" bson:"ticker"`
	Name       string    `json:"name" bson:"name"`
	Price      float64   `json:"price" bson:"price"`
	Change     float64   `json:"change" bson:"change"`
	ChangePerc float64   `json:"change_perc" bson:"change_perc"`
	Volume     int64     `json:"volume" bson:"volume"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

// StockQuote представляет котировки акции
type StockQuote struct {
	Ticker         string    `json:"ticker" bson:"ticker"`
	Open           float64   `json:"open" bson:"open"`
	High           float64   `json:"high" bson:"high"`
	Low            float64   `json:"low" bson:"low"`
	Close          float64   `json:"close" bson:"close"`
	Volume         int64     `json:"volume" bson:"volume"`
	Date           time.Time `json:"date" bson:"date"`
	MarketCapBln   float64   `json:"market_cap_bln" bson:"market_cap_bln"`
	PE             float64   `json:"pe" bson:"pe"`
	DividendYield  float64   `json:"dividend_yield" bson:"dividend_yield"`
	Sector         string    `json:"sector" bson:"sector"`
	TradingSession string    `json:"trading_session" bson:"trading_session"`
}
