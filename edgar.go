package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	secCompanyTickersURL = "https://www.sec.gov/files/company_tickers.json"
	secDataBaseURL       = "https://data.sec.gov"

	// SEC requires a declared, non-browser User-Agent (see sec.gov/os/accessing-edgar-data).
	// A URL in the User-Agent is rejected, so keep it to name + contact.
	secUserAgent = "stock-fetcher/1.0 (admin@example.com)"

	// SEC fair-access limit is 10 requests/second. Serializing at 5/second
	// keeps comfortably below the threshold.
	secMinRequestInterval = 200 * time.Millisecond
)

// EDGARFetcher fetches historical EPS data from the SEC EDGAR XBRL API.
// EDGAR is free, official, and requires no API key.
//
// It does not provide market prices, so callers combine EDGAR fundamentals
// with prices from another source (e.g. Yahoo Finance).
type EDGARFetcher struct {
	client *http.Client

	// Test seam for swapping the upstream URLs.
	tickersURL   string
	factsBaseURL string
	userAgent    string

	// tickers caches the ticker -> CIK map after the first lookup.
	tickerMu sync.Mutex
	tickers  map[string]string

	// rateMu serializes SEC requests and enforces the rate interval.
	rateMu  sync.Mutex
	lastReq time.Time
}

// NewEDGARFetcher creates a new SEC EDGAR fetcher.
// Set SEC_USER_AGENT to override the User-Agent sent to SEC (SEC requires a
// declared contact, e.g. "StockFetcher/1.0 (you@example.com)").
func NewEDGARFetcher() *EDGARFetcher {
	userAgent := os.Getenv("SEC_USER_AGENT")
	if userAgent == "" {
		userAgent = secUserAgent
	}

	return &EDGARFetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		tickersURL:   secCompanyTickersURL,
		factsBaseURL: secDataBaseURL,
		userAgent:    userAgent,
		tickers:      make(map[string]string),
	}
}

// EdgarEPSFact is a single XBRL fact for us-gaap:EarningsPerShareDiluted.
type EdgarEPSFact struct {
	Start string  `json:"start"`
	End   string  `json:"end"`
	Val   float64 `json:"val"`
	Form  string  `json:"form"`
	FP    string  `json:"fp"`
	Filed string  `json:"filed"`
	Frame string  `json:"frame"`
}

// sharedEDGARFetcher is used by the fallback path so the ~800KB SEC ticker
// map is only downloaded once per process instead of once per request.
var sharedEDGARFetcher = NewEDGARFetcher()

// EdgarCompanyFacts is the subset of the EDGAR companyfacts JSON we need.
type EdgarCompanyFacts struct {
	EntityName string `json:"entityName"`
	Facts      struct {
		UsGAAP struct {
			EarningsPerShareDiluted struct {
				Units map[string][]EdgarEPSFact `json:"units"`
			} `json:"EarningsPerShareDiluted"`
		} `json:"us-gaap"`
	} `json:"facts"`
}

// fetchJSON performs a rate-limited GET request and decodes the JSON body.
func (f *EDGARFetcher) fetchJSON(url string, out interface{}) error {
	f.rateMu.Lock()
	defer f.rateMu.Unlock()

	if wait := secMinRequestInterval - time.Since(f.lastReq); wait > 0 {
		time.Sleep(wait)
	}
	f.lastReq = time.Now()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", f.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("SEC returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to parse SEC response: %w", err)
	}
	return nil
}

// getCIK resolves a ticker symbol to its zero-padded 10-digit CIK.
func (f *EDGARFetcher) getCIK(symbol string) (string, error) {
	upper := strings.ToUpper(symbol)

	f.tickerMu.Lock()
	defer f.tickerMu.Unlock()

	if cik, ok := f.tickers[upper]; ok {
		return cik, nil
	}

	var data map[string]struct {
		CIK    int    `json:"cik_str"`
		Ticker string `json:"ticker"`
	}
	if err := f.fetchJSON(f.tickersURL, &data); err != nil {
		return "", fmt.Errorf("failed to load SEC ticker map: %w", err)
	}

	for _, v := range data {
		if v.Ticker != "" {
			f.tickers[strings.ToUpper(v.Ticker)] = fmt.Sprintf("%010d", v.CIK)
		}
	}

	cik, ok := f.tickers[upper]
	if !ok {
		return "", fmt.Errorf("symbol %s not found in SEC ticker map", symbol)
	}
	return cik, nil
}

// FetchFundamental fetches historical TTM EPS data from EDGAR and returns it
// in the same FundamentalData shape used by the macrotrends fetcher.
// StockPrice and PERatio are left zero because EDGAR has no market prices.
func (f *EDGARFetcher) FetchFundamental(symbol string) (*FundamentalData, error) {
	cik, err := f.getCIK(symbol)
	if err != nil {
		return nil, err
	}

	factsURL := fmt.Sprintf("%s/api/xbrl/companyfacts/CIK%s.json", f.factsBaseURL, cik)
	var facts EdgarCompanyFacts
	if err := f.fetchJSON(factsURL, &facts); err != nil {
		return nil, fmt.Errorf("failed to fetch SEC company facts: %w", err)
	}

	rawFacts := facts.Facts.UsGAAP.EarningsPerShareDiluted.Units["USD/shares"]
	quarters := selectQuarterlyEPS(rawFacts)
	if len(quarters) == 0 {
		return nil, fmt.Errorf("no quarterly diluted EPS data found for %s", symbol)
	}

	quarters = deriveMissingQ4(quarters, selectAnnualEPS(rawFacts))
	history := buildTTMEPSSeries(quarters)

	return &FundamentalData{
		Symbol:         strings.ToUpper(symbol),
		CompanyName:    facts.EntityName,
		CurrentEPS:     latestPositiveEPS(history),
		HistoricalData: history,
	}, nil
}

// selectQuarterlyEPS reduces raw EDGAR facts to one diluted EPS value per
// fiscal quarter. It keeps facts whose start->end duration looks like a
// single quarter (under 120 days), discarding year-to-date and annual facts.
// When the same quarter appears in multiple filings (original + amended),
// the most recently filed value wins.
func selectQuarterlyEPS(facts []EdgarEPSFact) []EdgarEPSFact {
	best := make(map[string]EdgarEPSFact)

	for _, fact := range facts {
		if fact.Start == "" || fact.End == "" {
			continue
		}
		if fact.Form != "10-Q" && fact.Form != "10-K" && fact.Form != "10-K/A" {
			continue
		}

		duration := daysBetween(fact.Start, fact.End)
		if duration <= 0 || duration >= 120 {
			continue
		}

		current, exists := best[fact.End]
		if !exists || fact.Filed > current.Filed ||
			(fact.Filed == current.Filed && duration < daysBetween(current.Start, current.End)) {
			best[fact.End] = fact
		}
	}

	result := make([]EdgarEPSFact, 0, len(best))
	for _, fact := range best {
		result = append(result, fact)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].End < result[j].End
	})

	return result
}

// selectAnnualEPS reduces raw EDGAR facts to one diluted EPS value per fiscal
// year. It keeps 10-K/10-K/A facts whose duration looks like a fiscal year.
func selectAnnualEPS(facts []EdgarEPSFact) []EdgarEPSFact {
	best := make(map[string]EdgarEPSFact)

	for _, fact := range facts {
		if fact.Start == "" || fact.End == "" {
			continue
		}
		if fact.Form != "10-K" && fact.Form != "10-K/A" {
			continue
		}

		duration := daysBetween(fact.Start, fact.End)
		if duration < 300 {
			continue
		}

		current, exists := best[fact.End]
		if !exists || fact.Filed > current.Filed {
			best[fact.End] = fact
		}
	}

	result := make([]EdgarEPSFact, 0, len(best))
	for _, fact := range best {
		result = append(result, fact)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].End < result[j].End
	})

	return result
}

// deriveMissingQ4 fills in the fourth fiscal quarter for years where the
// filer only tagged annual EPS in the 10-K. The Q4 value is annual EPS minus
// the three quarterly values already available for that fiscal year.
func deriveMissingQ4(quarters, annuals []EdgarEPSFact) []EdgarEPSFact {
	covered := make(map[string]bool, len(quarters))
	for _, q := range quarters {
		covered[q.End] = true
	}

	result := append([]EdgarEPSFact(nil), quarters...)

	for _, annual := range annuals {
		if covered[annual.End] {
			continue
		}

		// Find the three most recent quarter ends that fall within the fiscal
		// year (within 320 days before the fiscal year end).
		var candidates []EdgarEPSFact
		for _, q := range quarters {
			daysBefore := daysBetween(q.End, annual.End)
			if daysBefore > 0 && daysBefore < 320 {
				candidates = append(candidates, q)
			}
		}

		if len(candidates) < 3 {
			continue
		}

		last3 := candidates[len(candidates)-3:]
		q4Value := annual.Val - last3[0].Val - last3[1].Val - last3[2].Val

		result = append(result, EdgarEPSFact{
			Start: last3[2].End,
			End:   annual.End,
			Val:   q4Value,
			Form:  annual.Form,
			FP:    "Q4",
			Filed: annual.Filed,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].End < result[j].End
	})

	return result
}

// buildTTMEPSSeries converts quarterly EPS facts into trailing-twelve-month
// EPS data points. The first three quarters have EPS 0 because a full TTM
// window is not yet available.
func buildTTMEPSSeries(quarters []EdgarEPSFact) []PERatioData {
	series := make([]PERatioData, 0, len(quarters))

	for i, quarter := range quarters {
		eps := 0.0
		if i >= 3 {
			eps = quarters[i-3].Val + quarters[i-2].Val + quarters[i-1].Val + quarter.Val
		}

		series = append(series, PERatioData{
			Date: quarter.End,
			EPS:  eps,
		})
	}

	return series
}

// daysBetween returns the number of days between two YYYY-MM-DD dates.
// It returns 0 if either date cannot be parsed.
func daysBetween(start, end string) int {
	startDate, startErr := time.Parse("2006-01-02", start)
	endDate, endErr := time.Parse("2006-01-02", end)
	if startErr != nil || endErr != nil {
		return 0
	}
	return int(endDate.Sub(startDate).Hours() / 24)
}

// latestPositiveEPS returns the most recent positive EPS value in the series,
// or 0 if none exists.
func latestPositiveEPS(history []PERatioData) float64 {
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].EPS > 0 {
			return history[i].EPS
		}
	}
	return 0
}
