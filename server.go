package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

)

//go:embed web/*
var webFS embed.FS

// APIResponse is a standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// StockRequest represents a stock data request
type StockRequest struct {
	Symbol string `json:"symbol"`
	Days   int    `json:"days"`
	Period string `json:"period"` // daily, weekly, monthly, quarterly, yearly
}

// StockResponse represents the API response for stock data
type StockResponse struct {
	Symbol      string       `json:"symbol"`
	CompanyName string       `json:"company_name"`
	DataSource  string       `json:"data_source"`
	TTM_EPS     float64      `json:"ttm_eps,omitempty"`
	PeriodType  string       `json:"period_type"`
	RecordCount int          `json:"record_count"`
	DailyData   []StockData  `json:"daily_data,omitempty"`
	PeriodData  []PeriodData `json:"period_data,omitempty"`
}

// Server holds the HTTP server and its dependencies
type Server struct {
	port   string
	router *http.ServeMux
}

// NewServer creates a new HTTP server
func NewServer(port string) *Server {
	s := &Server{
		port:   port,
		router: http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// API routes
	s.router.HandleFunc("/api/health", s.handleHealth)
	s.router.HandleFunc("/api/stock/", s.handleStock)
	s.router.HandleFunc("/api/stock-excel/", s.handleStockExcel)
	s.router.HandleFunc("/api/indices", s.handleIndices)
	s.router.HandleFunc("/api/indices/", s.handleIndexSymbols)

	// Static files (frontend)
	webContent, _ := fs.Sub(webFS, "web")
	fileServer := http.FileServer(http.FS(webContent))
	s.router.Handle("/", fileServer)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	s.router.ServeHTTP(w, r)
}

// Start starts the HTTP server with graceful shutdown
func (s *Server) Start() error {
	server := &http.Server{
		Addr:         ":" + s.port,
		Handler:      s,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Server is shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
		close(done)
	}()

	log.Printf("Stock Fetcher %s (commit: %s, built: %s)", Version, CommitHash, BuildTime)
	log.Printf("Server starting on port %s", s.port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	<-done
	log.Println("Server stopped")
	return nil
}

// Handler returns the http.Handler for Lambda integration
func (s *Server) Handler() http.Handler {
	return s
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, APIResponse{
		Success: false,
		Error:   message,
	})
}

// writeSuccess writes a success response
func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeSuccess(w, map[string]string{
		"status":     "ok",
		"version":    Version,
		"commit":     CommitHash,
		"build_time": BuildTime,
	})
}

// handleStock handles stock data requests
// GET /api/stock/{symbol}?days=365&period=daily
func (s *Server) handleStock(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse symbol from path
	path := strings.TrimPrefix(r.URL.Path, "/api/stock/")
	symbol := strings.TrimSuffix(path, "/")
	if symbol == "" {
		writeError(w, http.StatusBadRequest, "Symbol is required")
		return
	}

	// Parse query parameters
	days := 365
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "daily"
	}

	// Validate period
	validPeriods := map[string]bool{
		"daily": true, "weekly": true, "monthly": true,
		"quarterly": true, "yearly": true,
	}
	if !validPeriods[period] {
		writeError(w, http.StatusBadRequest, "Invalid period. Use: daily, weekly, monthly, quarterly, yearly")
		return
	}

	// Fetch data
	useYahoo := isHKStock(symbol)
	data, ttmEPS, companyName, includePE, err := fetchStockData(symbol, days, useYahoo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch data: %v", err))
		return
	}

	if len(data) == 0 {
		writeError(w, http.StatusNotFound, "No data found for symbol")
		return
	}

	// Determine data source
	dataSource := "macrotrends"
	if useYahoo || !includePE {
		dataSource = "yahoo"
	}

	// Build response
	resp := StockResponse{
		Symbol:      strings.ToUpper(symbol),
		CompanyName: formatCompanyName(companyName),
		DataSource:  dataSource,
		PeriodType:  period,
	}

	if includePE {
		resp.TTM_EPS = ttmEPS
	}

	// Aggregate if period is not daily
	if period != "daily" {
		periodType, _ := ParsePeriodType(period)
		// Data is newest-first, AggregateToPeriods expects oldest-first
		reversedData := reverseData(data)
		periodData := AggregateToPeriods(reversedData, periodType)
		resp.PeriodData = periodData
		resp.RecordCount = len(periodData)
	} else {
		resp.DailyData = data
		resp.RecordCount = len(data)
	}

	writeSuccess(w, resp)
}

// handleIndices handles index list requests
// GET /api/indices
func (s *Server) handleIndices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	indices := GetIndices()
	result := make([]map[string]interface{}, 0, len(indices))

	for key, idx := range indices {
		result = append(result, map[string]interface{}{
			"key":         key,
			"name":        idx.Name,
			"description": idx.Description,
			"count":       len(idx.Symbols),
		})
	}

	writeSuccess(w, result)
}

// handleIndexSymbols handles index symbol list requests
// GET /api/indices/{name}
func (s *Server) handleIndexSymbols(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse index name from path
	path := strings.TrimPrefix(r.URL.Path, "/api/indices/")
	indexName := strings.TrimSuffix(path, "/")
	if indexName == "" {
		writeError(w, http.StatusBadRequest, "Index name is required")
		return
	}

	indices := GetIndices()
	idx, exists := indices[strings.ToLower(indexName)]
	if !exists {
		writeError(w, http.StatusNotFound, "Index not found")
		return
	}

	writeSuccess(w, map[string]interface{}{
		"key":         indexName,
		"name":        idx.Name,
		"description": idx.Description,
		"symbols":     idx.Symbols,
		"companies":   GetCompanyNamesForSymbols(idx.Symbols),
		"count":       len(idx.Symbols),
	})
}

// handleStockExcel handles Excel export requests
// GET /api/stock-excel/{symbol}?days=365&period=daily
func (s *Server) handleStockExcel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse symbol from path
	path := strings.TrimPrefix(r.URL.Path, "/api/stock-excel/")
	symbol := strings.TrimSuffix(path, "/")
	if symbol == "" {
		writeError(w, http.StatusBadRequest, "Symbol is required")
		return
	}
	symbol = strings.ToUpper(symbol)

	// Parse query parameters
	query := r.URL.Query()
	days := 365
	if d := query.Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}
	period := query.Get("period")
	if period == "" {
		period = "daily"
	}

	// Determine data source
	useYahoo := isHKStock(symbol)

	// Fetch stock data
	data, ttmEPS, companyName, includePE, err := fetchStockData(symbol, days, useYahoo)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Prepare Excel params
	params := ExcelParams{
		Symbol:      symbol,
		CompanyName: companyName,
		Period:      period,
		TTMEPS:      ttmEPS,
		IncludePE:   includePE,
	}

	if period == "daily" {
		params.Data = data
	} else {
		periodType, _ := ParsePeriodType(period)
		reversedData := reverseData(data)
		params.PeriodData = AggregateToPeriods(reversedData, periodType)
	}

	// Generate Excel file
	f, err := GenerateExcel(params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate Excel")
		return
	}
	defer func() { _ = f.Close() }()

	// Set response headers
	filename := fmt.Sprintf("%s_%s.xlsx", symbol, period)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// Write to response
	if err := f.Write(w); err != nil {
		log.Printf("Error writing Excel file: %v", err)
	}
}

// runServer starts the web server (called from main)
func runServer(port string) error {
	server := NewServer(port)
	return server.Start()
}
