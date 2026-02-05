package main

import (
	"testing"
)

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{100.0, "100.00"},
		{100.123, "100.12"},
		{100.126, "100.13"},
		{0.0, "0.00"},
		{-50.5, "-50.50"},
		{1234567.89, "1234567.89"},
	}

	for _, tt := range tests {
		result := formatFloat(tt.input)
		if result != tt.expected {
			t.Errorf("formatFloat(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatVolume(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{500, "500"},
		{1000, "1.00K"},
		{1500, "1.50K"},
		{1000000, "1.00M"},
		{1500000, "1.50M"},
		{1000000000, "1.00B"},
		{1500000000, "1.50B"},
		{0, "0"},
	}

	for _, tt := range tests {
		result := formatVolume(tt.input)
		if result != tt.expected {
			t.Errorf("formatVolume(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNewYahooFetcher(t *testing.T) {
	fetcher := NewYahooFetcher()
	if fetcher == nil {
		t.Error("NewYahooFetcher() returned nil")
	}
	if fetcher.client == nil {
		t.Error("NewYahooFetcher().client is nil")
	}
}

func TestParseYahooChartData(t *testing.T) {
	// Test with valid data
	resp := YahooChartResponse{}
	resp.Chart.Result = []struct {
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
	}{
		{
			Timestamp: []int64{1704067200, 1704153600}, // 2024-01-01, 2024-01-02
			Indicators: struct {
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
			}{
				Quote: []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				}{
					{
						Open:   []float64{100.0, 102.0},
						High:   []float64{105.0, 108.0},
						Low:    []float64{99.0, 101.0},
						Close:  []float64{104.0, 107.0},
						Volume: []int64{1000000, 2000000},
					},
				},
			},
		},
	}

	data, err := parseYahooChartData(resp)
	if err != nil {
		t.Fatalf("parseYahooChartData() error = %v", err)
	}

	if len(data) != 2 {
		t.Errorf("Expected 2 records, got %d", len(data))
	}

	// First record should have no change (no previous close)
	if data[0].Change != "" {
		t.Errorf("First record change should be empty, got %q", data[0].Change)
	}

	// Second record should have change calculated
	if data[1].Change == "" {
		t.Error("Second record change should not be empty")
	}
}

func TestParseYahooChartData_EmptyQuote(t *testing.T) {
	resp := YahooChartResponse{}
	resp.Chart.Result = []struct {
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
	}{
		{
			Timestamp: []int64{1704067200},
			// Empty indicators
		},
	}

	_, err := parseYahooChartData(resp)
	if err == nil {
		t.Error("Expected error for empty quote data")
	}
}
