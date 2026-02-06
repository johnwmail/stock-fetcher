package main

import (
	"testing"
)

func TestNewMacrotrendsFetcher(t *testing.T) {
	fetcher := NewMacrotrendsFetcher()
	if fetcher == nil {
		t.Error("NewMacrotrendsFetcher() returned nil")
	}
	if fetcher.client == nil {
		t.Error("NewMacrotrendsFetcher().client is nil")
	}
}

func TestGetLatestTTM_EPS(t *testing.T) {
	tests := []struct {
		name     string
		data     *FundamentalData
		expected float64
	}{
		{
			name:     "empty historical data",
			data:     &FundamentalData{HistoricalData: []PERatioData{}},
			expected: 0,
		},
		{
			name: "single positive EPS",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2024-01-01", EPS: 5.5},
				},
			},
			expected: 5.5,
		},
		{
			name: "multiple entries - returns latest positive",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2023-01-01", EPS: 4.0},
					{Date: "2024-01-01", EPS: 5.5},
				},
			},
			expected: 5.5,
		},
		{
			name: "latest is zero - returns previous positive",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2023-01-01", EPS: 4.0},
					{Date: "2024-01-01", EPS: 0},
				},
			},
			expected: 4.0,
		},
		{
			name: "latest is negative - returns previous positive",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2023-01-01", EPS: 4.0},
					{Date: "2024-01-01", EPS: -2.0},
				},
			},
			expected: 4.0,
		},
		{
			name: "all negative or zero",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2023-01-01", EPS: -1.0},
					{Date: "2024-01-01", EPS: 0},
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetLatestTTM_EPS()
			if result != tt.expected {
				t.Errorf("GetLatestTTM_EPS() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPERatioDataStruct(t *testing.T) {
	// Test that the struct can be initialized correctly
	data := PERatioData{
		Date:       "2024-01-01",
		StockPrice: 150.0,
		EPS:        5.0,
		PERatio:    30.0,
	}

	if data.Date != "2024-01-01" {
		t.Errorf("Date = %q, want %q", data.Date, "2024-01-01")
	}
	if data.StockPrice != 150.0 {
		t.Errorf("StockPrice = %v, want %v", data.StockPrice, 150.0)
	}
	if data.EPS != 5.0 {
		t.Errorf("EPS = %v, want %v", data.EPS, 5.0)
	}
	if data.PERatio != 30.0 {
		t.Errorf("PERatio = %v, want %v", data.PERatio, 30.0)
	}
}

func TestDailyPriceDataStruct(t *testing.T) {
	data := DailyPriceData{
		Date:   "2024-01-01",
		Open:   "150.00",
		High:   "155.00",
		Low:    "148.00",
		Close:  "154.00",
		Volume: "10.5",
	}

	if data.Date != "2024-01-01" {
		t.Errorf("Date = %q, want %q", data.Date, "2024-01-01")
	}
	if data.Open != "150.00" {
		t.Errorf("Open = %q, want %q", data.Open, "150.00")
	}
	if data.High != "155.00" {
		t.Errorf("High = %q, want %q", data.High, "155.00")
	}
	if data.Low != "148.00" {
		t.Errorf("Low = %q, want %q", data.Low, "148.00")
	}
	if data.Close != "154.00" {
		t.Errorf("Close = %q, want %q", data.Close, "154.00")
	}
	if data.Volume != "10.5" {
		t.Errorf("Volume = %q, want %q", data.Volume, "10.5")
	}
}

func TestFundamentalDataStruct(t *testing.T) {
	data := FundamentalData{
		Symbol:       "AAPL",
		CompanyName:  "apple",
		CurrentPE:    30.0,
		CurrentEPS:   5.0,
		CurrentPrice: 150.0,
		HistoricalData: []PERatioData{
			{Date: "2024-01-01", EPS: 5.0},
		},
	}

	if data.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", data.Symbol, "AAPL")
	}
	if data.CompanyName != "apple" {
		t.Errorf("CompanyName = %q, want %q", data.CompanyName, "apple")
	}
	if len(data.HistoricalData) != 1 {
		t.Errorf("HistoricalData length = %d, want 1", len(data.HistoricalData))
	}
}

func TestGetEPSForDate(t *testing.T) {
	tests := []struct {
		name     string
		data     *FundamentalData
		date     string
		expected float64
	}{
		{
			name:     "empty historical data",
			data:     &FundamentalData{HistoricalData: []PERatioData{}},
			date:     "2024-01-15",
			expected: 0,
		},
		{
			name: "date before all data",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2024-01-01", EPS: 5.0},
					{Date: "2024-04-01", EPS: 5.5},
				},
			},
			date:     "2023-06-15",
			expected: 0,
		},
		{
			name: "date in first quarter",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2024-01-01", EPS: 5.0},
					{Date: "2024-04-01", EPS: 5.5},
					{Date: "2024-07-01", EPS: 6.0},
				},
			},
			date:     "2024-02-15",
			expected: 5.0,
		},
		{
			name: "date in second quarter",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2024-01-01", EPS: 5.0},
					{Date: "2024-04-01", EPS: 5.5},
					{Date: "2024-07-01", EPS: 6.0},
				},
			},
			date:     "2024-05-15",
			expected: 5.5,
		},
		{
			name: "date after all data - use latest",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2024-01-01", EPS: 5.0},
					{Date: "2024-04-01", EPS: 5.5},
				},
			},
			date:     "2024-12-15",
			expected: 5.5,
		},
		{
			name: "exact date match",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2024-01-01", EPS: 5.0},
					{Date: "2024-04-01", EPS: 5.5},
				},
			},
			date:     "2024-04-01",
			expected: 5.5,
		},
		{
			name: "skip zero EPS entries",
			data: &FundamentalData{
				HistoricalData: []PERatioData{
					{Date: "2024-01-01", EPS: 5.0},
					{Date: "2024-04-01", EPS: 0},
					{Date: "2024-07-01", EPS: 6.0},
				},
			},
			date:     "2024-05-15",
			expected: 5.0, // Uses Q1 since Q2 has zero EPS
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.data.GetEPSForDate(tt.date)
			if result != tt.expected {
				t.Errorf("GetEPSForDate(%q) = %v, want %v", tt.date, result, tt.expected)
			}
		})
	}
}
