package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// MacrotrendsFetcher fetches fundamental data from macrotrends.net
type MacrotrendsFetcher struct {
	client *http.Client
}

// NewMacrotrendsFetcher creates a new Macrotrends fetcher
func NewMacrotrendsFetcher() *MacrotrendsFetcher {
	return &MacrotrendsFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// PERatioData represents P/E ratio data for a single period
type PERatioData struct {
	Date       string  `json:"date"`
	StockPrice float64 `json:"v1"`
	EPS        float64 `json:"v2"`
	PERatio    float64 `json:"v3"`
}

// DailyPriceData represents daily stock price from macrotrends
type DailyPriceData struct {
	Date   string `json:"d"`
	Open   string `json:"o"`
	High   string `json:"h"`
	Low    string `json:"l"`
	Close  string `json:"c"`
	Volume string `json:"v"`
}

// FundamentalData represents fundamental metrics for a stock
type FundamentalData struct {
	Symbol         string
	CompanyName    string
	CurrentPE      float64
	CurrentEPS     float64
	CurrentPrice   float64
	HistoricalData []PERatioData
}

// getCompanySlug tries to find the macrotrends URL slug for a symbol
func (f *MacrotrendsFetcher) getCompanySlug(symbol string) (string, error) {
	// Search for the company
	searchURL := fmt.Sprintf("https://www.macrotrends.net/production/stocks/desktop/ticker_search_list.php?q=%s", symbol)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Parse search results - format: [{"n":"Apple Inc.","s":"AAPL/apple"},...]
	var results []struct {
		Name   string `json:"n"`
		Symbol string `json:"s"` // Format: "AAPL/apple"
	}

	if err := json.Unmarshal(body, &results); err != nil {
		return "", fmt.Errorf("failed to parse search results: %w", err)
	}

	if len(results) == 0 {
		return "", fmt.Errorf("no results found for symbol %s", symbol)
	}

	// Find exact match only - don't fall back to first result
	for _, r := range results {
		parts := strings.Split(r.Symbol, "/")
		if len(parts) == 2 && strings.EqualFold(parts[0], symbol) {
			return r.Symbol, nil
		}
	}

	// No exact match found
	return "", fmt.Errorf("symbol %s not found on macrotrends (may be an ETF or unsupported stock)", symbol)
}

// FetchPERatio fetches P/E ratio data for a symbol
func (f *MacrotrendsFetcher) FetchPERatio(symbol string) (*FundamentalData, error) {
	// Get company slug
	slug, err := f.getCompanySlug(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to find company: %w", err)
	}

	parts := strings.Split(slug, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid slug format: %s", slug)
	}
	ticker := parts[0]
	companySlug := parts[1]

	// Fetch the iframe with chart data
	iframeURL := fmt.Sprintf("https://www.macrotrends.net/production/stocks/desktop/fundamental_iframe.php?t=%s&type=pe-ratio&statement=price-ratios&freq=Q&sub=", ticker)

	req, err := http.NewRequest("GET", iframeURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", fmt.Sprintf("https://www.macrotrends.net/stocks/charts/%s/%s/pe-ratio", ticker, companySlug))

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("iframe returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract chartData JSON from the HTML
	bodyStr := string(body)
	startMarker := "var chartData = "
	startIdx := strings.Index(bodyStr, startMarker)
	if startIdx == -1 {
		return nil, fmt.Errorf("could not find chart data in response")
	}
	startIdx += len(startMarker)

	// Find the end of the JSON array - look for ]\n or ]; or just ]
	subStr := bodyStr[startIdx:]
	bracketCount := 0
	endIdx := -1
	for i, c := range subStr {
		if c == '[' {
			bracketCount++
		} else if c == ']' {
			bracketCount--
			if bracketCount == 0 {
				endIdx = i + 1
				break
			}
		}
	}
	if endIdx == -1 {
		return nil, fmt.Errorf("could not find end of chart data")
	}

	jsonData := subStr[:endIdx]

	var peData []PERatioData
	if err := json.Unmarshal([]byte(jsonData), &peData); err != nil {
		return nil, fmt.Errorf("failed to parse P/E data: %w", err)
	}

	if len(peData) == 0 {
		return nil, fmt.Errorf("no P/E data found")
	}

	// Get latest data point
	latest := peData[len(peData)-1]

	return &FundamentalData{
		Symbol:         strings.ToUpper(ticker),
		CompanyName:    companySlug,
		CurrentPE:      latest.PERatio,
		CurrentEPS:     latest.EPS,
		CurrentPrice:   latest.StockPrice,
		HistoricalData: peData,
	}, nil
}

// FetchDailyPrices fetches daily stock prices from macrotrends
func (f *MacrotrendsFetcher) FetchDailyPrices(symbol string, days int) ([]DailyPriceData, error) {
	// Get company slug
	slug, err := f.getCompanySlug(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to find company: %w", err)
	}

	parts := strings.Split(slug, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid slug format: %s", slug)
	}
	ticker := parts[0]
	companySlug := parts[1]

	// Fetch the stock price history iframe
	iframeURL := fmt.Sprintf("https://www.macrotrends.net/production/stocks/desktop/stock_price_history.php?t=%s", ticker)

	req, err := http.NewRequest("GET", iframeURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", fmt.Sprintf("https://www.macrotrends.net/stocks/charts/%s/%s/stock-price-history", ticker, companySlug))

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("price history returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract dataDaily JSON from the HTML
	bodyStr := string(body)
	startMarker := "var dataDaily = "
	startIdx := strings.Index(bodyStr, startMarker)
	if startIdx == -1 {
		return nil, fmt.Errorf("could not find daily price data in response")
	}
	startIdx += len(startMarker)

	// Find the end of the JSON array
	subStr := bodyStr[startIdx:]
	bracketCount := 0
	endIdx := -1
	for i, c := range subStr {
		if c == '[' {
			bracketCount++
		} else if c == ']' {
			bracketCount--
			if bracketCount == 0 {
				endIdx = i + 1
				break
			}
		}
	}
	if endIdx == -1 {
		return nil, fmt.Errorf("could not find end of daily price data")
	}

	jsonData := subStr[:endIdx]

	var allData []DailyPriceData
	if err := json.Unmarshal([]byte(jsonData), &allData); err != nil {
		return nil, fmt.Errorf("failed to parse daily price data: %w", err)
	}

	if len(allData) == 0 {
		return nil, fmt.Errorf("no daily price data found")
	}

	// Return only the last N days
	if days > 0 && days < len(allData) {
		return allData[len(allData)-days:], nil
	}

	return allData, nil
}

// GetLatestTTM_EPS returns the latest trailing twelve months EPS
// Note: The EPS values from macrotrends are already TTM (not quarterly)
func (data *FundamentalData) GetLatestTTM_EPS() float64 {
	if len(data.HistoricalData) == 0 {
		return 0
	}
	// Find the most recent quarter with valid EPS data
	for i := len(data.HistoricalData) - 1; i >= 0; i-- {
		if data.HistoricalData[i].EPS > 0 {
			return data.HistoricalData[i].EPS
		}
	}
	return 0
}

// GetEPSForDate returns the TTM EPS that was valid on a given date
// It finds the most recent EPS data point on or before the given date
func (data *FundamentalData) GetEPSForDate(date string) float64 {
	if len(data.HistoricalData) == 0 {
		return 0
	}

	// Historical data is sorted oldest to newest
	// Find the last entry with date <= target date
	var eps float64
	for _, d := range data.HistoricalData {
		if d.Date <= date && d.EPS > 0 {
			eps = d.EPS
		}
		if d.Date > date {
			break
		}
	}
	return eps
}


