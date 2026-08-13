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

// Data source identifiers stored in the cache and returned to the API layer.
const (
	SourceAuto        = "auto"
	SourceMacrotrends = "macrotrends"
	SourceEDGAR       = "edgar"
	SourceYahoo       = "yahoo"
)

// sourceHasPE reports whether a data source includes historical P/E data.
func sourceHasPE(source string) bool {
	return source == SourceMacrotrends || source == SourceEDGAR
}

// normalizeSource converts a user-supplied source value into a known source.
// Empty and "auto" both mean the automatic fallback chain.
func normalizeSource(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", SourceAuto:
		return SourceAuto, nil
	case SourceMacrotrends:
		return SourceMacrotrends, nil
	case SourceEDGAR:
		return SourceEDGAR, nil
	case SourceYahoo:
		return SourceYahoo, nil
	default:
		return "", fmt.Errorf("invalid source: %q (use auto, macrotrends, edgar, or yahoo)", raw)
	}
}

// validateSourceForSymbol rejects sources that cannot serve a symbol.
// HK stocks are only available via Yahoo Finance.
func validateSourceForSymbol(source, symbol string) error {
	if isHKStock(symbol) && source != SourceAuto && source != SourceYahoo {
		return fmt.Errorf("source %q is not supported for HK stocks; use auto or yahoo", source)
	}
	return nil
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

// fetchUSStockWithEDGAR fetches US stock prices from Yahoo and computes
// historical P/E ratios using trailing-twelve-month EPS from SEC EDGAR.
// EDGAR has no market prices, so Yahoo supplies the OHLCV data.
func fetchUSStockWithEDGAR(symbol string, days int) ([]StockData, float64, string, error) {
	peData, err := sharedEDGARFetcher.FetchFundamental(symbol)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to fetch EDGAR EPS data: %w", err)
	}

	yahoo := NewYahooFetcher()
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	yahooData, companyName, err := yahoo.FetchHistoricalData(symbol, startDate, endDate)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to fetch Yahoo price data: %w", err)
	}
	if companyName == "" {
		companyName = peData.CompanyName
	}

	for i := range yahooData {
		eps := peData.GetEPSForDate(yahooData[i].Date)
		if eps <= 0 {
			continue
		}

		closePrice := parseFloat(yahooData[i].Close)
		if closePrice > 0 {
			yahooData[i].PE = fmt.Sprintf("%.2f", closePrice/eps)
		}
	}

	return reverseData(yahooData), peData.GetLatestTTM_EPS(), companyName, nil
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

// fetchFromProvider fetches stock data directly from the upstream provider.
// It returns the data source identifier so callers can decide whether the
// data includes historical P/E.
//
// source controls provider selection:
//   - SourceAuto: try macrotrends, then EDGAR+Yahoo, then Yahoo.
//   - SourceMacrotrends / SourceEDGAR / SourceYahoo: use only that source.
func fetchFromProvider(symbol string, days int, useYahoo bool, source string) ([]StockData, float64, string, string, error) {
	if useYahoo {
		data, companyName, err := fetchHKStock(symbol, days)
		return data, 0, companyName, SourceYahoo, err
	}

	switch source {
	case SourceMacrotrends:
		data, ttmEPS, companyName, err := fetchUSStock(symbol, days)
		if err != nil {
			return nil, 0, "", SourceMacrotrends, err
		}
		return data, ttmEPS, companyName, SourceMacrotrends, nil

	case SourceEDGAR:
		data, ttmEPS, companyName, err := fetchUSStockWithEDGAR(symbol, days)
		if err != nil {
			return nil, 0, "", SourceEDGAR, err
		}
		return data, ttmEPS, companyName, SourceEDGAR, nil

	case SourceYahoo:
		data, companyName, err := fetchHKStock(symbol, days)
		if err != nil {
			return nil, 0, "", SourceYahoo, err
		}
		return data, 0, companyName, SourceYahoo, nil
	}

	// SourceAuto: primary is macrotrends (prices + historical P/E).
	data, ttmEPS, companyName, err := fetchUSStock(symbol, days)
	if err == nil {
		return data, ttmEPS, companyName, SourceMacrotrends, nil
	}

	// Fallback: EDGAR provides historical EPS while Yahoo provides prices.
	data, ttmEPS, companyName, err = fetchUSStockWithEDGAR(symbol, days)
	if err == nil {
		return data, ttmEPS, companyName, SourceEDGAR, nil
	}

	// Last resort: Yahoo prices only (no P/E) for ETFs or unsupported stocks.
	data, companyName, err = fetchHKStock(symbol, days)
	if err != nil {
		return nil, 0, "", "", err
	}
	return data, 0, companyName, SourceYahoo, nil
}

// fetchStockData fetches stock data, using cache when available.
// The cache stores raw OHLCV+PE; Change/HChange are recomputed on read.
// The returned source string is one of SourceMacrotrends, SourceEDGAR, or
// SourceYahoo.
//
// A specific source only reads cache entries produced by that same source;
// SourceAuto uses the cached source whatever it is.
func fetchStockData(cache *Cache, symbol string, days int, useYahoo bool, source string) ([]StockData, float64, string, string, error) {
	symbolUpper := strings.ToUpper(symbol)
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	if cache != nil {
		meta, _ := cache.GetFetchMeta(symbolUpper)
		cacheMatches := meta != nil && (source == SourceAuto || meta.Source == source)

		// Cache hit: fresh today and covers the requested range
		if cacheMatches && meta != nil && meta.IsFresh() && meta.CoversRange(startDate) {
			data, err := cache.GetDailyPrices(symbolUpper, startDate, today)
			if err == nil && len(data) > 0 {
				return data, meta.TTMEPS, meta.CompanyName, meta.Source, nil
			}
		}

		// Cache stale or doesn't cover range — fetch from provider.
		// If we have some matching cached data, fetch only the delta.
		fetchDays := days
		if cacheMatches && meta != nil && meta.CoversRange(startDate) {
			// We have the range but it's stale — just fetch recent delta
			daysSinceLatest := int(time.Since(meta.LastFetched).Hours()/24) + 5
			if daysSinceLatest < fetchDays {
				fetchDays = daysSinceLatest
			}
		}

		data, ttmEPS, companyName, fetchedSource, err := fetchFromProvider(symbol, fetchDays, useYahoo, source)
		if err != nil {
			// Provider failed — try serving stale cache if it matches the
			// requested source.
			if cacheMatches && meta != nil {
				staleData, cacheErr := cache.GetDailyPrices(symbolUpper, startDate, today)
				if cacheErr == nil && len(staleData) > 0 {
					return staleData, meta.TTMEPS, meta.CompanyName, meta.Source, nil
				}
			}
			return nil, 0, "", "", err
		}

		// Store new data in cache
		if len(data) > 0 {
			_ = cache.StoreDailyPrices(symbolUpper, data)

			// Determine date range in cache
			earliestDate := data[len(data)-1].Date // data is newest-first
			latestDate := data[0].Date
			if cacheMatches && meta != nil && meta.EarliestDate < earliestDate {
				earliestDate = meta.EarliestDate
			}

			_ = cache.UpdateFetchLog(FetchMeta{
				Symbol:       symbolUpper,
				Source:       fetchedSource,
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
			return cachedData, ttmEPS, companyName, fetchedSource, nil
		}

		// Fallback: return provider data directly
		return data, ttmEPS, companyName, fetchedSource, nil
	}

	// No cache — fetch directly from provider
	return fetchFromProvider(symbol, days, useYahoo, source)
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
