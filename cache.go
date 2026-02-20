package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

// Cache provides SQLite-backed caching for stock data
type Cache struct {
	db *sql.DB
}

// FetchMeta holds metadata about a cached symbol
type FetchMeta struct {
	Symbol       string
	Source       string
	CompanyName  string
	TTMEPS       float64
	LastFetched  time.Time
	LatestDate   string
	EarliestDate string
}

// NewCache creates a new SQLite cache
func NewCache(dbPath string) (*Cache, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open cache db: %w", err)
	}

	// Enable WAL mode for better concurrent read performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	c := &Cache{db: db}
	if err := c.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate cache db: %w", err)
	}

	return c, nil
}

// Close closes the cache database
func (c *Cache) Close() error {
	return c.db.Close()
}

func (c *Cache) migrate() error {
	_, err := c.db.Exec(`
		CREATE TABLE IF NOT EXISTS daily_prices (
			symbol TEXT NOT NULL,
			date   TEXT NOT NULL,
			open   TEXT,
			high   TEXT,
			low    TEXT,
			close  TEXT,
			volume TEXT,
			pe     TEXT,
			PRIMARY KEY (symbol, date)
		);

		CREATE TABLE IF NOT EXISTS fetch_log (
			symbol        TEXT PRIMARY KEY,
			source        TEXT,
			company_name  TEXT,
			ttm_eps       REAL,
			last_fetched  TEXT,
			latest_date   TEXT,
			earliest_date TEXT
		);
	`)
	return err
}

// GetFetchMeta returns fetch metadata for a symbol, or nil if not cached
func (c *Cache) GetFetchMeta(symbol string) (*FetchMeta, error) {
	row := c.db.QueryRow(
		`SELECT symbol, source, company_name, ttm_eps, last_fetched, latest_date, earliest_date
		 FROM fetch_log WHERE symbol = ?`, symbol)

	var m FetchMeta
	var lastFetched string
	err := row.Scan(&m.Symbol, &m.Source, &m.CompanyName, &m.TTMEPS,
		&lastFetched, &m.LatestDate, &m.EarliestDate)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	m.LastFetched, _ = time.Parse(time.RFC3339, lastFetched)
	return &m, nil
}

// GetDailyPrices returns cached daily prices for a symbol in a date range.
// Returns data sorted newest-first (consistent with the app convention).
// Change and HChange are recomputed from the raw OHLC data.
func (c *Cache) GetDailyPrices(symbol, startDate, endDate string) ([]StockData, error) {
	rows, err := c.db.Query(
		`SELECT date, open, high, low, close, volume, pe
		 FROM daily_prices
		 WHERE symbol = ? AND date >= ? AND date <= ?
		 ORDER BY date ASC`, symbol, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []StockData
	var prevClose, prevHigh float64

	for rows.Next() {
		var d StockData
		if err := rows.Scan(&d.Date, &d.Open, &d.High, &d.Low, &d.Close, &d.Volume, &d.PE); err != nil {
			return nil, err
		}

		// Recompute Change and HChange from raw data
		close := parseFloat(d.Close)
		high := parseFloat(d.High)

		if prevClose > 0 {
			d.Change = fmt.Sprintf("%.2f%%", ((close-prevClose)/prevClose)*100)
		}
		if prevHigh > 0 {
			d.HChange = fmt.Sprintf("%.2f%%", ((close-prevHigh)/prevHigh)*100)
		}

		data = append(data, d)
		prevClose = close
		prevHigh = high
	}

	// Reverse to newest-first
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}

	return data, rows.Err()
}

// StoreDailyPrices stores daily price records in the cache.
// Uses INSERT OR REPLACE so newer data overwrites older cached values.
func (c *Cache) StoreDailyPrices(symbol string, data []StockData) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT OR REPLACE INTO daily_prices (symbol, date, open, high, low, close, volume, pe)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, d := range data {
		if _, err := stmt.Exec(symbol, d.Date, d.Open, d.High, d.Low, d.Close, d.Volume, d.PE); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateFetchLog updates the fetch metadata for a symbol
func (c *Cache) UpdateFetchLog(m FetchMeta) error {
	_, err := c.db.Exec(
		`INSERT OR REPLACE INTO fetch_log (symbol, source, company_name, ttm_eps, last_fetched, latest_date, earliest_date)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.Symbol, m.Source, m.CompanyName, m.TTMEPS,
		m.LastFetched.Format(time.RFC3339), m.LatestDate, m.EarliestDate)
	return err
}

// IsFresh returns true if the symbol was fetched today
func (m *FetchMeta) IsFresh() bool {
	now := time.Now()
	return m.LastFetched.Year() == now.Year() &&
		m.LastFetched.YearDay() == now.YearDay()
}

// CoversRange returns true if cached data covers the requested date range
func (m *FetchMeta) CoversRange(startDate string) bool {
	return m.EarliestDate <= startDate
}

// InitCache initializes the global cache from DB_PATH env var.
// Returns nil (no cache) if DB_PATH is explicitly set to empty or "none".
// detectDBPath picks a DB path based on the runtime environment.
//   - DB_PATH env set       → use that ("none" disables cache)
//   - AWS Lambda detected   → /tmp/cache.db
//   - /data dir exists (Docker volume) → /data/cache.db
//   - otherwise             → ./cache.db
func detectDBPath() string {
	// Explicit override always wins
	if p, set := os.LookupEnv("DB_PATH"); set {
		return p
	}

	// Lambda: AWS_LAMBDA_FUNCTION_NAME is always set in Lambda
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		return "/tmp/cache.db"
	}

	// Docker/container: /data volume mount
	if info, err := os.Stat("/data"); err == nil && info.IsDir() {
		return "/data/cache.db"
	}

	return "cache.db"
}

func InitCache() *Cache {
	dbPath := detectDBPath()
	if dbPath == "none" || dbPath == "" {
		log.Println("Cache disabled")
		return nil
	}

	cache, err := NewCache(dbPath)
	if err != nil {
		log.Printf("Warning: failed to init cache at %s: %v (running without cache)", dbPath, err)
		return nil
	}

	log.Printf("Cache initialized at %s", dbPath)
	return cache
}
