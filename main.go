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

// fetchFromProvider fetches stock data directly from the upstream provider
func fetchFromProvider(symbol string, days int, useYahoo bool) ([]StockData, float64, string, bool, error) {
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

// fetchStockData fetches stock data, using cache when available.
// The cache stores raw OHLCV+PE; Change/HChange are recomputed on read.
func fetchStockData(cache *Cache, symbol string, days int, useYahoo bool) ([]StockData, float64, string, bool, error) {
	symbolUpper := strings.ToUpper(symbol)
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	if cache != nil {
		meta, _ := cache.GetFetchMeta(symbolUpper)

		// Cache hit: fresh today and covers the requested range
		if meta != nil && meta.IsFresh() && meta.CoversRange(startDate) {
			data, err := cache.GetDailyPrices(symbolUpper, startDate, today)
			if err == nil && len(data) > 0 {
				includePE := meta.Source == "macrotrends"
				return data, meta.TTMEPS, meta.CompanyName, includePE, nil
			}
		}

		// Cache stale or doesn't cover range — fetch from provider
		// If we have some cached data, fetch only the delta
		fetchDays := days
		if meta != nil && meta.CoversRange(startDate) {
			// We have the range but it's stale — just fetch recent delta
			daysSinceLatest := int(time.Since(meta.LastFetched).Hours()/24) + 5
			if daysSinceLatest < fetchDays {
				fetchDays = daysSinceLatest
			}
		}

		data, ttmEPS, companyName, includePE, err := fetchFromProvider(symbol, fetchDays, useYahoo)
		if err != nil {
			// Provider failed — try serving stale cache if available
			if meta != nil {
				staleData, cacheErr := cache.GetDailyPrices(symbolUpper, startDate, today)
				if cacheErr == nil && len(staleData) > 0 {
					incPE := meta.Source == "macrotrends"
					return staleData, meta.TTMEPS, meta.CompanyName, incPE, nil
				}
			}
			return nil, 0, "", false, err
		}

		// Store new data in cache
		if len(data) > 0 {
			_ = cache.StoreDailyPrices(symbolUpper, data)

			source := "yahoo"
			if includePE {
				source = "macrotrends"
			}

			// Determine date range in cache
			earliestDate := data[len(data)-1].Date // data is newest-first
			latestDate := data[0].Date
			if meta != nil && meta.EarliestDate < earliestDate {
				earliestDate = meta.EarliestDate
			}

			_ = cache.UpdateFetchLog(FetchMeta{
				Symbol:       symbolUpper,
				Source:       source,
				CompanyName:  companyName,
				TTMEPS:       ttmEPS,
				LastFetched:  time.Now(),
				LatestDate:   latestDate,
				EarliestDate: earliestDate,
			})
		}

		// Serve full range from cache (includes old + new data)
		cachedData, cacheErr := cache.GetDailyPrices(symbolUpper, startDate, today)
		if cacheErr == nil && len(cachedData) > 0 {
			return cachedData, ttmEPS, companyName, includePE, nil
		}

		// Fallback: return provider data directly
		return data, ttmEPS, companyName, includePE, nil
	}

	// No cache — fetch directly from provider
	return fetchFromProvider(symbol, days, useYahoo)
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
