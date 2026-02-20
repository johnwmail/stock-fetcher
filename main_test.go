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
