package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// StockData represents a single day's stock data
type StockData struct {
	Date    string `json:"date"`
	Open    string `json:"open"`
	High    string `json:"high"`
	Low     string `json:"low"`
	Close   string `json:"close"`
	Volume  string `json:"volume"`
	Change  string `json:"change"`
	HChange string `json:"hchange"`
	PE      string `json:"pe,omitempty"`
}

// isHKStock checks if the symbol is a Hong Kong stock
func isHKStock(symbol string) bool {
	return strings.HasSuffix(strings.ToUpper(symbol), ".HK")
}

// reverseData reverses the slice so newest data is first
func reverseData(data []StockData) []StockData {
	result := make([]StockData, len(data))
	for i, d := range data {
		result[len(data)-1-i] = d
	}
	return result
}

// fetchUSStock fetches US stock data from macrotrends (with P/E)
func fetchUSStock(symbol string, days int) ([]StockData, float64, string, error) {
	fetcher := NewMacrotrendsFetcher()

	peData, err := fetcher.FetchPERatio(symbol)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to fetch P/E data: %w", err)
	}
	latestEPS := peData.GetLatestTTM_EPS()
	companyName := peData.CompanyName

	prices, err := fetcher.FetchDailyPrices(symbol, days)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to fetch price data: %w", err)
	}

	var data []StockData
	var prevClose, prevHigh float64

	for _, p := range prices {
		close, _ := strconv.ParseFloat(p.Close, 64)
		open, _ := strconv.ParseFloat(p.Open, 64)
		high, _ := strconv.ParseFloat(p.High, 64)
		low, _ := strconv.ParseFloat(p.Low, 64)

		change := ""
		if prevClose > 0 {
			pctChange := ((close - prevClose) / prevClose) * 100
			change = fmt.Sprintf("%.2f%%", pctChange)
		}

		hchange := ""
		if prevHigh > 0 {
			pctHChange := ((close - prevHigh) / prevHigh) * 100
			hchange = fmt.Sprintf("%.2f%%", pctHChange)
		}

		pe := ""
		historicalEPS := peData.GetEPSForDate(p.Date)
		if historicalEPS > 0 {
			pe = fmt.Sprintf("%.2f", close/historicalEPS)
		}

		data = append(data, StockData{
			Date:    p.Date,
			Open:    fmt.Sprintf("%.2f", open),
			High:    fmt.Sprintf("%.2f", high),
			Low:     fmt.Sprintf("%.2f", low),
			Close:   fmt.Sprintf("%.2f", close),
			Volume:  p.Volume + "M",
			Change:  change,
			HChange: hchange,
			PE:      pe,
		})

		prevClose = close
		prevHigh = high
	}

	return reverseData(data), latestEPS, companyName, nil
}

// fetchHKStock fetches HK stock data from Yahoo (no P/E)
func fetchHKStock(symbol string, days int) ([]StockData, string, error) {
	fetcher := NewYahooFetcher()
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	yahooData, companyName, err := fetcher.FetchHistoricalData(symbol, startDate, endDate)
	if err != nil {
		return nil, "", err
	}

	return reverseData(yahooData), companyName, nil
}

// formatCompanyName formats the company slug for display
func formatCompanyName(slug string) string {
	if slug == "" {
		return ""
	}
	name := strings.ReplaceAll(slug, "-", " ")
	words := strings.Fields(name)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

// fetchStockData fetches stock data from appropriate source
func fetchStockData(symbol string, days int, useYahoo bool) ([]StockData, float64, string, bool, error) {
	var data []StockData
	var companyName string
	var ttmEPS float64
	var err error
	includePE := false

	if useYahoo {
		data, companyName, err = fetchHKStock(symbol, days)
	} else {
		data, ttmEPS, companyName, err = fetchUSStock(symbol, days)
		if err != nil {
			// Fallback to Yahoo Finance for ETFs or unsupported stocks
			data, companyName, err = fetchHKStock(symbol, days)
		} else {
			includePE = true
		}
	}

	return data, ttmEPS, companyName, includePE, err
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := runServer(port); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
