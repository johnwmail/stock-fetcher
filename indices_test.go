package main

import (
	"testing"
)

func TestGetIndices(t *testing.T) {
	indices := GetIndices()

	// Check that all expected indices exist
	expectedKeys := []string{"sp500", "dow", "nasdaq100", "hangseng"}
	for _, key := range expectedKeys {
		if _, ok := indices[key]; !ok {
			t.Errorf("GetIndices() missing key %q", key)
		}
	}

	// Check that we have exactly 4 indices
	if len(indices) != 4 {
		t.Errorf("GetIndices() returned %d indices, want 4", len(indices))
	}
}

func TestDowIndex(t *testing.T) {
	if DowIndex.Name == "" {
		t.Error("DowIndex.Name is empty")
	}
	if DowIndex.Description == "" {
		t.Error("DowIndex.Description is empty")
	}
	if len(DowIndex.Symbols) != 30 {
		t.Errorf("DowIndex has %d symbols, want 30", len(DowIndex.Symbols))
	}

	// Check for some known Dow stocks
	knownStocks := []string{"AAPL", "MSFT", "JPM", "V", "WMT"}
	for _, stock := range knownStocks {
		found := false
		for _, sym := range DowIndex.Symbols {
			if sym == stock {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DowIndex missing expected stock %q", stock)
		}
	}
}

func TestNasdaq100Index(t *testing.T) {
	if Nasdaq100Index.Name == "" {
		t.Error("Nasdaq100Index.Name is empty")
	}
	if Nasdaq100Index.Description == "" {
		t.Error("Nasdaq100Index.Description is empty")
	}
	// NASDAQ 100 should have around 100 symbols (can vary slightly)
	if len(Nasdaq100Index.Symbols) < 90 || len(Nasdaq100Index.Symbols) > 110 {
		t.Errorf("Nasdaq100Index has %d symbols, expected ~100", len(Nasdaq100Index.Symbols))
	}

	// Check for some known NASDAQ stocks
	knownStocks := []string{"AAPL", "MSFT", "GOOGL", "AMZN", "NVDA"}
	for _, stock := range knownStocks {
		found := false
		for _, sym := range Nasdaq100Index.Symbols {
			if sym == stock {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Nasdaq100Index missing expected stock %q", stock)
		}
	}
}

func TestSP500Index(t *testing.T) {
	if SP500Index.Name == "" {
		t.Error("SP500Index.Name is empty")
	}
	if SP500Index.Description == "" {
		t.Error("SP500Index.Description is empty")
	}
	// S&P 500 should have around 500 symbols (can vary slightly due to multi-class shares)
	if len(SP500Index.Symbols) < 450 || len(SP500Index.Symbols) > 550 {
		t.Errorf("SP500Index has %d symbols, expected ~500", len(SP500Index.Symbols))
	}
}

func TestHangSengIndex(t *testing.T) {
	if HangSengIndex.Name == "" {
		t.Error("HangSengIndex.Name is empty")
	}
	if HangSengIndex.Description == "" {
		t.Error("HangSengIndex.Description is empty")
	}
	if len(HangSengIndex.Symbols) == 0 {
		t.Error("HangSengIndex has no symbols")
	}

	// Check that HK stocks have .HK suffix
	for _, sym := range HangSengIndex.Symbols {
		if len(sym) < 4 || sym[len(sym)-3:] != ".HK" {
			t.Errorf("HangSengIndex symbol %q doesn't have .HK suffix", sym)
		}
	}

	// Check for some known HK stocks
	knownStocks := []string{"0700.HK", "9988.HK", "0005.HK"}
	for _, stock := range knownStocks {
		found := false
		for _, sym := range HangSengIndex.Symbols {
			if sym == stock {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("HangSengIndex missing expected stock %q", stock)
		}
	}
}

func TestIndexStruct(t *testing.T) {
	idx := Index{
		Name:        "Test Index",
		Description: "Test description",
		Symbols:     []string{"TEST1", "TEST2"},
	}

	if idx.Name != "Test Index" {
		t.Errorf("Index.Name = %q, want %q", idx.Name, "Test Index")
	}
	if idx.Description != "Test description" {
		t.Errorf("Index.Description = %q, want %q", idx.Description, "Test description")
	}
	if len(idx.Symbols) != 2 {
		t.Errorf("Index.Symbols length = %d, want 2", len(idx.Symbols))
	}
}
