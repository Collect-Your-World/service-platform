package eodhd

import (
	"context"
)

type FundamentalData struct {
	General struct {
		Code        string `json:"Code"`
		Name        string `json:"Name"`
		Exchange    string `json:"Exchange"`
		Description string `json:"Description"`
		WebURL      string `json:"WebURL"`
	} `json:"General"`
	SharesStats struct {
		SharesOutstanding float64 `json:"SharesOutstanding"`
	} `json:"SharesStats"`
	Highlights struct {
		MarketCapitalization float64 `json:"MarketCapitalization"`
	} `json:"Highlights"`
}

type RealtimeData struct {
	Code          string  `json:"code"`
	Timestamp     int64   `json:"timestamp"`
	GMTOffset     int     `json:"gmtoffset"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        int64   `json:"volume"`
	PreviousClose float64 `json:"previousClose"`
	Change        float64 `json:"change"`
	ChangePercent float64 `json:"change_p"`
}

type NewsItem struct {
	Date     string `json:"date"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Link     string `json:"link"`
	LinkHash string
}

type NewsData []NewsItem

type EODHDClient interface {
	GetFundamental(ctx context.Context, ticker string) (FundamentalData, error)
	GetRealtime(ctx context.Context, ticker string) (RealtimeData, error)
	GetNews(ctx context.Context, ticker string) (NewsData, error)
}
