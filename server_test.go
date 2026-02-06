package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	server := NewServer("0")

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestIndicesEndpoint(t *testing.T) {
	server := NewServer("0")

	req := httptest.NewRequest("GET", "/api/indices", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	// Check that we have indices data
	data, ok := resp.Data.([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}

	if len(data) == 0 {
		t.Error("Expected at least one index")
	}
}

func TestIndexSymbolsEndpoint(t *testing.T) {
	server := NewServer("0")

	req := httptest.NewRequest("GET", "/api/indices/dow", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestIndexSymbolsNotFound(t *testing.T) {
	server := NewServer("0")

	req := httptest.NewRequest("GET", "/api/indices/nonexistent", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestStockEndpointMissingSymbol(t *testing.T) {
	server := NewServer("0")

	req := httptest.NewRequest("GET", "/api/stock/", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestStockEndpointInvalidPeriod(t *testing.T) {
	server := NewServer("0")

	req := httptest.NewRequest("GET", "/api/stock/AAPL?period=invalid", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCORSHeaders(t *testing.T) {
	server := NewServer("0")

	req := httptest.NewRequest("OPTIONS", "/api/health", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if cors := w.Header().Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Errorf("Expected CORS header *, got %q", cors)
	}
}

func TestStaticFiles(t *testing.T) {
	server := NewServer("0")

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "Stock Fetcher") {
		t.Error("Expected HTML to contain 'Stock Fetcher'")
	}
}
