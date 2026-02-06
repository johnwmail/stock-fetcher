package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParsePeriodType(t *testing.T) {
	tests := []struct {
		input    string
		expected PeriodType
		wantErr  bool
	}{
		{"weekly", PeriodWeekly, false},
		{"week", PeriodWeekly, false},
		{"w", PeriodWeekly, false},
		{"WEEKLY", PeriodWeekly, false},
		{"monthly", PeriodMonthly, false},
		{"month", PeriodMonthly, false},
		{"m", PeriodMonthly, false},
		{"quarterly", PeriodQuarterly, false},
		{"quarter", PeriodQuarterly, false},
		{"q", PeriodQuarterly, false},
		{"yearly", PeriodYearly, false},
		{"year", PeriodYearly, false},
		{"y", PeriodYearly, false},
		{"invalid", "", true},
		{"", "", true},
		{"daily", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParsePeriodType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePeriodType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParsePeriodType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetPeriodKey(t *testing.T) {
	tests := []struct {
		date       string
		periodType PeriodType
		expected   string
	}{
		{"2024-01-15", PeriodWeekly, "2024-W03"},
		{"2024-01-01", PeriodWeekly, "2024-W01"},
		{"2024-01-15", PeriodMonthly, "2024-01"},
		{"2024-12-31", PeriodMonthly, "2024-12"},
		{"2024-01-15", PeriodQuarterly, "2024-Q1"},
		{"2024-04-15", PeriodQuarterly, "2024-Q2"},
		{"2024-07-15", PeriodQuarterly, "2024-Q3"},
		{"2024-10-15", PeriodQuarterly, "2024-Q4"},
		{"2024-06-15", PeriodYearly, "2024"},
		{"2023-01-01", PeriodYearly, "2023"},
	}

	for _, tt := range tests {
		t.Run(tt.date+"_"+string(tt.periodType), func(t *testing.T) {
			date, _ := parseDate(tt.date)
			result := getPeriodKey(date, tt.periodType)
			if result != tt.expected {
				t.Errorf("getPeriodKey(%s, %s) = %q, want %q", tt.date, tt.periodType, result, tt.expected)
			}
		})
	}
}

func TestClassifyDrop(t *testing.T) {
	tests := []struct {
		change   string
		expected int
	}{
		// Positive changes - no drop
		{"1.5%", 0},
		{"0.0%", 0},
		{"5.5%", 0},
		// Small drops - no bucket
		{"-0.5%", 0},
		{"-1.99%", 0},
		// 2% bucket (2-3%)
		{"-2.0%", 2},
		{"-2.5%", 2},
		{"-2.99%", 2},
		// 3% bucket (3-4%)
		{"-3.0%", 3},
		{"-3.5%", 3},
		{"-3.99%", 3},
		// 4% bucket (4-5%)
		{"-4.0%", 4},
		{"-4.5%", 4},
		{"-4.99%", 4},
		// 5% bucket (5%+)
		{"-5.0%", 5},
		{"-5.5%", 5},
		{"-10.0%", 5},
		{"-50.0%", 5},
		// Edge cases
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.change, func(t *testing.T) {
			result := classifyDrop(tt.change)
			if result != tt.expected {
				t.Errorf("classifyDrop(%q) = %d, want %d", tt.change, result, tt.expected)
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"100.50", 100.50},
		{"0", 0},
		{"-50.25", -50.25},
		{"invalid", 0},
		{"", 0},
	}

	for _, tt := range tests {
		result := parseFloat(tt.input)
		if result != tt.expected {
			t.Errorf("parseFloat(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestParseVolume(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1.5M", 1500000},
		{"2.0B", 2000000000},
		{"500K", 500000},
		{"1000", 1000},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		result := parseVolume(tt.input)
		if result != tt.expected {
			t.Errorf("parseVolume(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestFormatVolumeFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{1500000000, "1.50B"},
		{1500000, "1.50M"},
		{1500, "1.50K"},
		{500, "500"},
		{0, "0"},
	}

	for _, tt := range tests {
		result := formatVolumeFloat(tt.input)
		if result != tt.expected {
			t.Errorf("formatVolumeFloat(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestAggregateToPeriods(t *testing.T) {
	// Create test data for one week (oldest first)
	data := []StockData{
		{Date: "2024-01-08", Open: "100.00", High: "105.00", Low: "99.00", Close: "104.00", Volume: "1M", Change: ""},
		{Date: "2024-01-09", Open: "104.00", High: "106.00", Low: "102.00", Close: "103.00", Volume: "1.5M", Change: "-0.96%"},
		{Date: "2024-01-10", Open: "103.00", High: "104.00", Low: "98.00", Close: "99.00", Volume: "2M", Change: "-3.88%"},  // 3% drop
		{Date: "2024-01-11", Open: "99.00", High: "101.00", Low: "97.00", Close: "100.00", Volume: "1.2M", Change: "1.01%"},
		{Date: "2024-01-12", Open: "100.00", High: "102.00", Low: "95.00", Close: "96.00", Volume: "1.8M", Change: "-4.00%"}, // 4% drop
	}

	result := AggregateToPeriods(data, PeriodWeekly)

	if len(result) != 1 {
		t.Fatalf("Expected 1 period, got %d", len(result))
	}

	period := result[0]

	// Check period key
	if period.Period != "2024-W02" {
		t.Errorf("Period = %q, want %q", period.Period, "2024-W02")
	}

	// Check dates
	if period.StartDate != "2024-01-08" {
		t.Errorf("StartDate = %q, want %q", period.StartDate, "2024-01-08")
	}
	if period.EndDate != "2024-01-12" {
		t.Errorf("EndDate = %q, want %q", period.EndDate, "2024-01-12")
	}

	// Check OHLC
	if period.Open != "100.00" {
		t.Errorf("Open = %q, want %q", period.Open, "100.00")
	}
	if period.Close != "96.00" {
		t.Errorf("Close = %q, want %q", period.Close, "96.00")
	}
	if period.High != "106.00" {
		t.Errorf("High = %q, want %q", period.High, "106.00")
	}
	if period.Low != "95.00" {
		t.Errorf("Low = %q, want %q", period.Low, "95.00")
	}

	// Check days
	if period.Days != 5 {
		t.Errorf("Days = %d, want %d", period.Days, 5)
	}

	// Check drop counts
	if period.Drop2Pct != 0 {
		t.Errorf("Drop2Pct = %d, want %d", period.Drop2Pct, 0)
	}
	if period.Drop3Pct != 1 {
		t.Errorf("Drop3Pct = %d, want %d", period.Drop3Pct, 1)
	}
	if period.Drop4Pct != 1 {
		t.Errorf("Drop4Pct = %d, want %d", period.Drop4Pct, 1)
	}
	if period.Drop5Pct != 0 {
		t.Errorf("Drop5Pct = %d, want %d", period.Drop5Pct, 0)
	}
}

func TestAggregateToPeriods_Empty(t *testing.T) {
	result := AggregateToPeriods([]StockData{}, PeriodWeekly)
	if result != nil {
		t.Errorf("Expected nil for empty input, got %v", result)
	}
}

func TestAggregateToPeriods_MultiplePeriods(t *testing.T) {
	// Create data spanning two months
	data := []StockData{
		{Date: "2024-01-15", Open: "100.00", High: "105.00", Low: "99.00", Close: "104.00", Volume: "1M", Change: ""},
		{Date: "2024-01-16", Open: "104.00", High: "106.00", Low: "102.00", Close: "105.00", Volume: "1M", Change: "0.96%"},
		{Date: "2024-02-01", Open: "105.00", High: "110.00", Low: "104.00", Close: "108.00", Volume: "1M", Change: "2.86%"},
		{Date: "2024-02-02", Open: "108.00", High: "112.00", Low: "107.00", Close: "110.00", Volume: "1M", Change: "1.85%"},
	}

	result := AggregateToPeriods(data, PeriodMonthly)

	if len(result) != 2 {
		t.Fatalf("Expected 2 periods, got %d", len(result))
	}

	// Result should be newest first
	if result[0].Period != "2024-02" {
		t.Errorf("First period = %q, want %q", result[0].Period, "2024-02")
	}
	if result[1].Period != "2024-01" {
		t.Errorf("Second period = %q, want %q", result[1].Period, "2024-01")
	}
}

func TestWritePeriodCSV(t *testing.T) {
	tmpDir := t.TempDir()

	data := []PeriodData{
		{
			Period: "2024-01", StartDate: "2024-01-02", EndDate: "2024-01-31",
			Open: "100.00", High: "110.00", Low: "95.00", Close: "105.00",
			Volume: "50M", Change: "5.00%", PE: "25.5",
			Days: 21, Drop2Pct: 2, Drop3Pct: 1, Drop4Pct: 0, Drop5Pct: 0,
		},
	}

	tests := []struct {
		name      string
		includePE bool
		wantCols  int
	}{
		{"with_pe", true, 15},
		{"without_pe", false, 14},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(tmpDir, tt.name+".csv")
			err := WritePeriodCSV(data, filename, tt.includePE)
			if err != nil {
				t.Fatalf("WritePeriodCSV() error = %v", err)
			}

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

			if len(records) != 2 {
				t.Errorf("Expected 2 rows, got %d", len(records))
			}

			if len(records[0]) != tt.wantCols {
				t.Errorf("Expected %d columns, got %d", tt.wantCols, len(records[0]))
			}

			// Check that drop columns exist
			header := strings.Join(records[0], ",")
			if !strings.Contains(header, "Drop2%") {
				t.Error("Header missing Drop2%")
			}
		})
	}
}

func TestWritePeriodJSON(t *testing.T) {
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.json")

	data := []PeriodData{
		{
			Period: "2024-01", StartDate: "2024-01-02", EndDate: "2024-01-31",
			Open: "100.00", High: "110.00", Low: "95.00", Close: "105.00",
			Volume: "50M", Change: "5.00%",
			Days: 21, Drop2Pct: 2, Drop3Pct: 1, Drop4Pct: 0, Drop5Pct: 0,
		},
	}

	err := WritePeriodJSON(data, filename)
	if err != nil {
		t.Fatalf("WritePeriodJSON() error = %v", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open output file: %v", err)
	}
	defer file.Close()

	var result []PeriodData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 record, got %d", len(result))
	}

	if result[0].Drop3Pct != 1 {
		t.Errorf("Drop3Pct = %d, want 1", result[0].Drop3Pct)
	}
}

func TestWritePeriodTable(t *testing.T) {
	tmpDir := t.TempDir()

	data := []PeriodData{
		{
			Period: "2024-01", StartDate: "2024-01-02", EndDate: "2024-01-31",
			Open: "100.00", High: "110.00", Low: "95.00", Close: "105.00",
			Volume: "50M", Change: "5.00%", PE: "25.5",
			Days: 21, Drop2Pct: 2, Drop3Pct: 1, Drop4Pct: 0, Drop5Pct: 0,
		},
	}

	filename := filepath.Join(tmpDir, "test.txt")
	err := WritePeriodTable(data, filename, true)
	if err != nil {
		t.Fatalf("WritePeriodTable() error = %v", err)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	// Check for expected content
	if !strings.Contains(contentStr, "D2%") {
		t.Error("Table missing D2% header")
	}
	if !strings.Contains(contentStr, "D5%") {
		t.Error("Table missing D5% header")
	}
	if !strings.Contains(contentStr, "2024-01") {
		t.Error("Table missing period data")
	}
}

// Helper function for tests
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
