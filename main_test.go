package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestExpandListAlias(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sp", "sp500"},
		{"SP", "sp500"},
		{"hk", "hangseng"},
		{"HK", "hangseng"},
		{"nasdaq", "nasdaq100"},
		{"NASDAQ", "nasdaq100"},
		{"dow", "dow"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandListAlias(tt.input)
			if result != tt.expected {
				t.Errorf("expandListAlias(%q) = %q, want %q", tt.input, result, tt.expected)
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

func TestWriteCSV(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		data      []StockData
		includePE bool
		wantCols  int
	}{
		{
			name: "without PE",
			data: []StockData{
				{Date: "2024-01-01", Open: "100.00", High: "105.00", Low: "99.00", Close: "104.00", Volume: "1M", Change: "1.5%", HChange: "-0.5%"},
			},
			includePE: false,
			wantCols:  8, // Added HChange column
		},
		{
			name: "with PE",
			data: []StockData{
				{Date: "2024-01-01", Open: "100.00", High: "105.00", Low: "99.00", Close: "104.00", Volume: "1M", Change: "1.5%", HChange: "-0.5%", PE: "25.5"},
			},
			includePE: true,
			wantCols:  9, // Added HChange column
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(tmpDir, tt.name+".csv")
			err := WriteCSV(tt.data, filename, tt.includePE)
			if err != nil {
				t.Fatalf("WriteCSV() error = %v", err)
			}

			// Read and verify
			file, err := os.Open(filename)
			if err != nil {
				t.Fatalf("Failed to open output file: %v", err)
			}
			defer file.Close()

			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil {
				t.Fatalf("Failed to read CSV: %v", err)
			}

			// Check header + 1 data row
			if len(records) != 2 {
				t.Errorf("Expected 2 rows, got %d", len(records))
			}

			// Check column count
			if len(records[0]) != tt.wantCols {
				t.Errorf("Expected %d columns, got %d", tt.wantCols, len(records[0]))
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	data := []StockData{
		{Date: "2024-01-01", Open: "100.00", High: "105.00", Low: "99.00", Close: "104.00", Volume: "1M", Change: "1.5%", PE: "25.5"},
		{Date: "2024-01-02", Open: "104.00", High: "110.00", Low: "103.00", Close: "108.00", Volume: "2M", Change: "3.8%", PE: "26.0"},
	}

	err := WriteJSON(data, filename)
	if err != nil {
		t.Fatalf("WriteJSON() error = %v", err)
	}

	// Read and verify
	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	var result []StockData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(result) != len(data) {
		t.Errorf("Expected %d records, got %d", len(data), len(result))
	}

	if result[0].Date != data[0].Date {
		t.Errorf("Expected date %q, got %q", data[0].Date, result[0].Date)
	}
}

func TestWriteTable(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		includePE bool
		wantPE    bool
	}{
		{"without PE", false, false},
		{"with PE", true, true},
	}

	data := []StockData{
		{Date: "2024-01-01", Open: "100.00", High: "105.00", Low: "99.00", Close: "104.00", Volume: "1M", Change: "1.5%", PE: "25.5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(tmpDir, tt.name+".txt")
			err := WriteTable(data, filename, tt.includePE)
			if err != nil {
				t.Fatalf("WriteTable() error = %v", err)
			}

			content, err := os.ReadFile(filename)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			contentStr := string(content)
			hasPE := strings.Contains(contentStr, "PE")
			if hasPE != tt.wantPE {
				t.Errorf("Table contains PE header = %v, want %v", hasPE, tt.wantPE)
			}
		})
	}
}
