package main

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCacheRoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	cache, err := NewCache(dbPath)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer cache.Close()

	symbol := "AAPL"

	// Store some daily prices (newest-first, as returned by fetchers)
	data := []StockData{
		{Date: "2024-01-05", Open: "150.00", High: "155.00", Low: "149.00", Close: "154.00", Volume: "10M", PE: "30.00"},
		{Date: "2024-01-04", Open: "148.00", High: "152.00", Low: "147.00", Close: "150.00", Volume: "8M", PE: "29.00"},
		{Date: "2024-01-03", Open: "145.00", High: "149.00", Low: "144.00", Close: "148.00", Volume: "9M", PE: "28.50"},
	}

	if err := cache.StoreDailyPrices(symbol, data); err != nil {
		t.Fatalf("StoreDailyPrices: %v", err)
	}

	// Update fetch log
	meta := FetchMeta{
		Symbol:       symbol,
		Source:       "macrotrends",
		CompanyName:  "apple",
		TTMEPS:       7.50,
		LastFetched:  time.Now(),
		LatestDate:   "2024-01-05",
		EarliestDate: "2024-01-03",
	}
	if err := cache.UpdateFetchLog(meta); err != nil {
		t.Fatalf("UpdateFetchLog: %v", err)
	}

	// Read back fetch meta
	gotMeta, err := cache.GetFetchMeta(symbol)
	if err != nil {
		t.Fatalf("GetFetchMeta: %v", err)
	}
	if gotMeta == nil {
		t.Fatal("GetFetchMeta returned nil")
	}
	if gotMeta.Source != "macrotrends" {
		t.Errorf("Source = %q, want %q", gotMeta.Source, "macrotrends")
	}
	if gotMeta.TTMEPS != 7.50 {
		t.Errorf("TTMEPS = %v, want %v", gotMeta.TTMEPS, 7.50)
	}
	if !gotMeta.IsFresh() {
		t.Error("Expected meta to be fresh (fetched just now)")
	}

	// Read back prices
	gotData, err := cache.GetDailyPrices(symbol, "2024-01-03", "2024-01-05")
	if err != nil {
		t.Fatalf("GetDailyPrices: %v", err)
	}
	if len(gotData) != 3 {
		t.Fatalf("GetDailyPrices returned %d records, want 3", len(gotData))
	}

	// Should be newest-first
	if gotData[0].Date != "2024-01-05" {
		t.Errorf("First record date = %q, want %q", gotData[0].Date, "2024-01-05")
	}
	if gotData[2].Date != "2024-01-03" {
		t.Errorf("Last record date = %q, want %q", gotData[2].Date, "2024-01-03")
	}

	// Change should be recomputed (first chronological day has no change)
	if gotData[2].Change != "" {
		t.Errorf("Oldest record should have no change, got %q", gotData[2].Change)
	}
	if gotData[1].Change == "" {
		t.Error("Middle record should have computed change")
	}

	// PE should be preserved
	if gotData[0].PE != "30.00" {
		t.Errorf("PE = %q, want %q", gotData[0].PE, "30.00")
	}
}

func TestCachePartialRange(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	cache, err := NewCache(dbPath)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer cache.Close()

	// Store 5 days
	data := []StockData{
		{Date: "2024-01-05", Close: "154.00", High: "155.00"},
		{Date: "2024-01-04", Close: "150.00", High: "152.00"},
		{Date: "2024-01-03", Close: "148.00", High: "149.00"},
		{Date: "2024-01-02", Close: "145.00", High: "146.00"},
		{Date: "2024-01-01", Close: "142.00", High: "143.00"},
	}
	_ = cache.StoreDailyPrices("TEST", data)

	// Query partial range
	result, _ := cache.GetDailyPrices("TEST", "2024-01-03", "2024-01-05")
	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}
}

func TestCacheNonExistent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	cache, err := NewCache(dbPath)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer cache.Close()

	meta, err := cache.GetFetchMeta("NONEXISTENT")
	if err != nil {
		t.Fatalf("GetFetchMeta: %v", err)
	}
	if meta != nil {
		t.Error("Expected nil for non-existent symbol")
	}

	data, err := cache.GetDailyPrices("NONEXISTENT", "2024-01-01", "2024-12-31")
	if err != nil {
		t.Fatalf("GetDailyPrices: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("Expected 0 records, got %d", len(data))
	}
}

func TestCacheUpsert(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	cache, err := NewCache(dbPath)
	if err != nil {
		t.Fatalf("NewCache: %v", err)
	}
	defer cache.Close()

	// Store initial data
	_ = cache.StoreDailyPrices("AAPL", []StockData{
		{Date: "2024-01-03", Close: "148.00"},
		{Date: "2024-01-02", Close: "145.00"},
	})

	// Store overlapping + new data (should upsert)
	_ = cache.StoreDailyPrices("AAPL", []StockData{
		{Date: "2024-01-04", Close: "150.00"},
		{Date: "2024-01-03", Close: "149.00"}, // updated value
	})

	result, _ := cache.GetDailyPrices("AAPL", "2024-01-02", "2024-01-04")
	if len(result) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(result))
	}

	// Newest first, so [0] = Jan 4, [1] = Jan 3 (updated), [2] = Jan 2
	if result[1].Close != "149.00" {
		t.Errorf("Jan 3 close should be updated to 149.00, got %s", result[1].Close)
	}
}

func TestFetchMetaFreshness(t *testing.T) {
	// Fresh: fetched today
	fresh := FetchMeta{LastFetched: time.Now()}
	if !fresh.IsFresh() {
		t.Error("Expected fresh")
	}

	// Stale: fetched yesterday
	stale := FetchMeta{LastFetched: time.Now().AddDate(0, 0, -1)}
	if stale.IsFresh() {
		t.Error("Expected stale")
	}
}

func TestFetchMetaCoversRange(t *testing.T) {
	m := FetchMeta{EarliestDate: "2023-01-01"}

	if !m.CoversRange("2023-06-01") {
		t.Error("Should cover 2023-06-01")
	}
	if !m.CoversRange("2023-01-01") {
		t.Error("Should cover exact start")
	}
	if m.CoversRange("2022-12-31") {
		t.Error("Should not cover date before earliest")
	}
}
