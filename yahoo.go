package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// YahooFetcher fetches data from Yahoo Finance
type YahooFetcher struct {
	client *http.Client
}

// NewYahooFetcher creates a new Yahoo Finance fetcher
func NewYahooFetcher() *YahooFetcher {
	return &YahooFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// YahooChartResponse represents the Yahoo Finance chart API response
type YahooChartResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
				AdjClose []struct {
					AdjClose []float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
		Error *struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	} `json:"chart"`
}

// FetchHistoricalData fetches historical data from Yahoo Finance using the chart API
func (f *YahooFetcher) FetchHistoricalData(symbol string, startDate, endDate time.Time) ([]StockData, error) {
	period1 := startDate.Unix()
	period2 := endDate.Unix()

	// Use the chart API which doesn't require authentication
	url := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d&includePrePost=false",
		strings.ToUpper(symbol),
		period1,
		period2,
	)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body[:min(500, len(body))]))
	}

	var chartResp YahooChartResponse
	if err := json.Unmarshal(body, &chartResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if chartResp.Chart.Error != nil {
		return nil, fmt.Errorf("API error: %s - %s", chartResp.Chart.Error.Code, chartResp.Chart.Error.Description)
	}

	if len(chartResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned for symbol %s", symbol)
	}

	return parseYahooChartData(chartResp)
}

// parseYahooChartData converts Yahoo chart response to StockData
func parseYahooChartData(resp YahooChartResponse) ([]StockData, error) {
	result := resp.Chart.Result[0]
	timestamps := result.Timestamp

	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data in response")
	}

	quote := result.Indicators.Quote[0]

	var data []StockData
	var prevClose, prevHigh float64

	for i, ts := range timestamps {
		if i >= len(quote.Close) {
			break
		}

		// Skip if data is missing (null values become 0)
		if quote.Close[i] == 0 {
			continue
		}

		t := time.Unix(ts, 0)
		date := t.Format("2006-01-02")

		var openVal, highVal, lowVal float64
		var volume int64

		if i < len(quote.Open) {
			openVal = quote.Open[i]
		}
		if i < len(quote.High) {
			highVal = quote.High[i]
		}
		if i < len(quote.Low) {
			lowVal = quote.Low[i]
		}
		if i < len(quote.Volume) {
			volume = quote.Volume[i]
		}

		// Calculate change % (close to close)
		change := ""
		if prevClose > 0 {
			pctChange := ((quote.Close[i] - prevClose) / prevClose) * 100
			change = fmt.Sprintf("%.2f%%", pctChange)
		}

		// Calculate HChange % (close relative to previous high)
		hchange := ""
		if prevHigh > 0 {
			pctHChange := ((quote.Close[i] - prevHigh) / prevHigh) * 100
			hchange = fmt.Sprintf("%.2f%%", pctHChange)
		}

		sd := StockData{
			Date:    date,
			Open:    formatFloat(openVal),
			High:    formatFloat(highVal),
			Low:     formatFloat(lowVal),
			Close:   formatFloat(quote.Close[i]),
			Volume:  formatVolume(volume),
			Change:  change,
			HChange: hchange,
		}

		data = append(data, sd)
		prevClose = quote.Close[i]
		prevHigh = highVal
	}

	return data, nil
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func formatVolume(v int64) string {
	if v >= 1000000000 {
		return fmt.Sprintf("%.2fB", float64(v)/1000000000)
	}
	if v >= 1000000 {
		return fmt.Sprintf("%.2fM", float64(v)/1000000)
	}
	if v >= 1000 {
		return fmt.Sprintf("%.2fK", float64(v)/1000)
	}
	return strconv.FormatInt(v, 10)
}
