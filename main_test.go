package main

import (
	"testing"
)

func TestIsHKStock(t *testing.T) {
	tests := []struct {
		symbol   string
		expected bool
	}{
		{"0700.HK", true},
		{"0700.hk", true},
		{"AAPL", false},
		{"MSFT", false},
		{"9988.HK", true},
		{"", false},
		{"HK", false},
		{".HK", true},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			result := isHKStock(tt.symbol)
			if result != tt.expected {
				t.Errorf("isHKStock(%q) = %v, want %v", tt.symbol, result, tt.expected)
			}
		})
	}
}

func TestReverseData(t *testing.T) {
	tests := []struct {
		name     string
		input    []StockData
		expected []StockData
	}{
		{
			name:     "empty slice",
			input:    []StockData{},
			expected: []StockData{},
		},
		{
			name:     "single element",
			input:    []StockData{{Date: "2024-01-01"}},
			expected: []StockData{{Date: "2024-01-01"}},
		},
		{
			name: "multiple elements",
			input: []StockData{
				{Date: "2024-01-01"},
				{Date: "2024-01-02"},
				{Date: "2024-01-03"},
			},
			expected: []StockData{
				{Date: "2024-01-03"},
				{Date: "2024-01-02"},
				{Date: "2024-01-01"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reverseData(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("reverseData() returned %d elements, want %d", len(result), len(tt.expected))
				return
			}
			for i := range result {
				if result[i].Date != tt.expected[i].Date {
					t.Errorf("reverseData()[%d].Date = %q, want %q", i, result[i].Date, tt.expected[i].Date)
				}
			}
		})
	}
}

func TestSourceHasPE(t *testing.T) {
	tests := []struct {
		source   string
		expected bool
	}{
		{SourceMacrotrends, true},
		{SourceEDGAR, true},
		{SourceYahoo, false},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := sourceHasPE(tt.source); got != tt.expected {
			t.Errorf("sourceHasPE(%q) = %v, want %v", tt.source, got, tt.expected)
		}
	}
}

func TestNormalizeSource(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"", SourceAuto, false},
		{"auto", SourceAuto, false},
		{"AUTO", SourceAuto, false},
		{"macrotrends", SourceMacrotrends, false},
		{"edgar", SourceEDGAR, false},
		{"yahoo", SourceYahoo, false},
		{"Yahoo", SourceYahoo, false},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		got, err := normalizeSource(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("normalizeSource(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("normalizeSource(%q) error = %v", tt.input, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("normalizeSource(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestValidateSourceForSymbol(t *testing.T) {
	tests := []struct {
		source  string
		symbol  string
		wantErr bool
	}{
		{SourceAuto, "AAPL", false},
		{SourceMacrotrends, "AAPL", false},
		{SourceEDGAR, "AAPL", false},
		{SourceYahoo, "AAPL", false},
		{SourceAuto, "0700.HK", false},
		{SourceYahoo, "0700.HK", false},
		{SourceMacrotrends, "0700.HK", true},
		{SourceEDGAR, "0700.HK", true},
	}

	for _, tt := range tests {
		err := validateSourceForSymbol(tt.source, tt.symbol)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateSourceForSymbol(%q, %q) error = %v, wantErr %v", tt.source, tt.symbol, err, tt.wantErr)
		}
	}
}
