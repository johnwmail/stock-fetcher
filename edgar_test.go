package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDaysBetween(t *testing.T) {
	tests := []struct {
		start, end string
		expected   int
	}{
		{"2024-01-01", "2024-03-31", 90},
		{"2024-01-01", "2024-01-01", 0},
		{"bad-date", "2024-01-01", 0},
		{"2024-01-01", "bad-date", 0},
	}

	for _, tt := range tests {
		if got := daysBetween(tt.start, tt.end); got != tt.expected {
			t.Errorf("daysBetween(%q, %q) = %d, want %d", tt.start, tt.end, got, tt.expected)
		}
	}
}

func TestSelectQuarterlyEPS(t *testing.T) {
	facts := []EdgarEPSFact{
		// Year-to-date (duration >= 120) — should be excluded.
		{Start: "2024-01-01", End: "2024-06-30", Val: 5.0, Form: "10-Q", Filed: "2024-08-01"},
		// Annual (duration >= 120) — should be excluded.
		{Start: "2024-01-01", End: "2024-12-31", Val: 10.0, Form: "10-K", Filed: "2025-02-01"},
		// 8-K — should be excluded.
		{Start: "2024-01-01", End: "2024-03-31", Val: 2.0, Form: "8-K", Filed: "2024-05-01"},
		// Original + amended duplicate — latest filed wins.
		{Start: "2024-01-01", End: "2024-03-31", Val: 2.0, Form: "10-Q", Filed: "2024-05-01"},
		{Start: "2024-01-01", End: "2024-03-31", Val: 2.1, Form: "10-Q", Filed: "2024-05-15"},
		// A valid later quarter.
		{Start: "2024-04-01", End: "2024-06-30", Val: 3.0, Form: "10-Q", Filed: "2024-08-01"},
	}

	got := selectQuarterlyEPS(facts)
	if len(got) != 2 {
		t.Fatalf("selectQuarterlyEPS() returned %d facts, want 2", len(got))
	}
	if got[0].End != "2024-03-31" || got[0].Val != 2.1 {
		t.Errorf("first fact = %+v, want amended Q1", got[0])
	}
	if got[1].End != "2024-06-30" || got[1].Val != 3.0 {
		t.Errorf("second fact = %+v, want Q2", got[1])
	}
}

func TestSelectAnnualEPS(t *testing.T) {
	facts := []EdgarEPSFact{
		// Original + amended duplicate — latest filed wins.
		{Start: "2024-01-01", End: "2024-12-31", Val: 10.0, Form: "10-K", Filed: "2025-02-01"},
		{Start: "2024-01-01", End: "2024-12-31", Val: 10.5, Form: "10-K", Filed: "2025-03-01"},
		// Quarterly fact — should be excluded.
		{Start: "2024-01-01", End: "2024-03-31", Val: 2.0, Form: "10-K", Filed: "2024-05-01"},
	}

	got := selectAnnualEPS(facts)
	if len(got) != 1 {
		t.Fatalf("selectAnnualEPS() returned %d facts, want 1", len(got))
	}
	if got[0].Val != 10.5 {
		t.Errorf("annual EPS = %v, want 10.5", got[0].Val)
	}
}

func TestDeriveMissingQ4(t *testing.T) {
	quarters := []EdgarEPSFact{
		{Start: "2022-01-01", End: "2022-03-31", Val: 2.0, Form: "10-Q", Filed: "2022-04-20"},
		{Start: "2022-04-01", End: "2022-06-30", Val: 2.5, Form: "10-Q", Filed: "2022-07-20"},
		{Start: "2022-07-01", End: "2022-09-30", Val: 3.0, Form: "10-Q", Filed: "2022-10-20"},
		{Start: "2023-01-01", End: "2023-03-31", Val: 3.0, Form: "10-Q", Filed: "2023-04-20"},
		{Start: "2023-04-01", End: "2023-06-30", Val: 3.5, Form: "10-Q", Filed: "2023-07-20"},
		{Start: "2023-07-01", End: "2023-09-30", Val: 4.0, Form: "10-Q", Filed: "2023-10-20"},
	}
	annuals := []EdgarEPSFact{
		{Start: "2022-01-01", End: "2022-12-31", Val: 10.0, Form: "10-K", Filed: "2023-02-15"},
		{Start: "2023-01-01", End: "2023-12-31", Val: 14.0, Form: "10-K", Filed: "2024-02-15"},
	}

	got := deriveMissingQ4(quarters, annuals)
	if len(got) != 8 {
		t.Fatalf("deriveMissingQ4() returned %d facts, want 8", len(got))
	}

	// The two derived Q4 facts should end on the fiscal year end dates.
	q4s := make(map[string]float64)
	for _, f := range got {
		if f.FP == "Q4" {
			q4s[f.End] = f.Val
		}
	}
	if len(q4s) != 2 {
		t.Fatalf("derived Q4 count = %d, want 2", len(q4s))
	}
	if q4s["2022-12-31"] != 2.5 {
		t.Errorf("FY2022 derived Q4 = %v, want 2.5", q4s["2022-12-31"])
	}
	if q4s["2023-12-31"] != 3.5 {
		t.Errorf("FY2023 derived Q4 = %v, want 3.5", q4s["2023-12-31"])
	}
}

func TestBuildTTMEPSSeries(t *testing.T) {
	quarters := []EdgarEPSFact{
		{Start: "2022-01-01", End: "2022-03-31", Val: 2.0},
		{Start: "2022-04-01", End: "2022-06-30", Val: 2.5},
		{Start: "2022-07-01", End: "2022-09-30", Val: 3.0},
		{Start: "2022-10-01", End: "2022-12-31", Val: 2.5},
		{Start: "2023-01-01", End: "2023-03-31", Val: 3.0},
	}

	got := buildTTMEPSSeries(quarters)
	if len(got) != 5 {
		t.Fatalf("buildTTMEPSSeries() returned %d points, want 5", len(got))
	}

	// First three quarters do not yet have a full TTM window.
	for i := 0; i < 3; i++ {
		if got[i].EPS != 0 {
			t.Errorf("EPS[%d] = %v, want 0", i, got[i].EPS)
		}
	}

	if got[3].Date != "2022-12-31" || got[3].EPS != 10.0 {
		t.Errorf("TTM[3] = %+v, want 2022-12-31 EPS 10.0", got[3])
	}
	if got[4].Date != "2023-03-31" || got[4].EPS != 11.0 {
		t.Errorf("TTM[4] = %+v, want 2023-03-31 EPS 11.0", got[4])
	}
}

func TestLatestPositiveEPS(t *testing.T) {
	history := []PERatioData{
		{Date: "2024-01-01", EPS: 1.0},
		{Date: "2024-04-01", EPS: 0},
		{Date: "2024-07-01", EPS: -1.0},
	}

	if got := latestPositiveEPS(history); got != 1.0 {
		t.Errorf("latestPositiveEPS() = %v, want 1.0", got)
	}
	if got := latestPositiveEPS(nil); got != 0 {
		t.Errorf("latestPositiveEPS(nil) = %v, want 0", got)
	}
}

func TestEDGARFetcherFetchFundamental(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/files/company_tickers.json", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"0":{"cik_str":320193,"ticker":"AAPL","title":"Apple Inc."}}`))
	})
	mux.HandleFunc("/api/xbrl/companyfacts/CIK0000320193.json", func(w http.ResponseWriter, r *http.Request) {
		facts := `{
			"entityName": "Apple Inc.",
			"facts": {
				"us-gaap": {
					"EarningsPerShareDiluted": {
						"units": {
							"USD/shares": [
								{"start":"2022-01-01","end":"2022-03-31","val":2.0,"form":"10-Q","fp":"Q1","filed":"2022-04-20"},
								{"start":"2022-04-01","end":"2022-06-30","val":2.5,"form":"10-Q","fp":"Q2","filed":"2022-07-20"},
								{"start":"2022-07-01","end":"2022-09-30","val":3.0,"form":"10-Q","fp":"Q3","filed":"2022-10-20"},
								{"start":"2022-01-01","end":"2022-12-31","val":10.0,"form":"10-K","fp":"FY","filed":"2023-02-15"},
								{"start":"2023-01-01","end":"2023-03-31","val":3.0,"form":"10-Q","fp":"Q1","filed":"2023-04-20"},
								{"start":"2023-04-01","end":"2023-06-30","val":3.5,"form":"10-Q","fp":"Q2","filed":"2023-07-20"},
								{"start":"2023-07-01","end":"2023-09-30","val":4.0,"form":"10-Q","fp":"Q3","filed":"2023-10-20"},
								{"start":"2023-01-01","end":"2023-12-31","val":14.0,"form":"10-K","fp":"FY","filed":"2024-02-15"}
							]
						}
					}
				}
			}
		}`
		_, _ = fmt.Fprint(w, facts)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	fetcher := NewEDGARFetcher()
	fetcher.tickersURL = server.URL + "/files/company_tickers.json"
	fetcher.factsBaseURL = server.URL

	data, err := fetcher.FetchFundamental("aapl")
	if err != nil {
		t.Fatalf("FetchFundamental() error = %v", err)
	}

	if data.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", data.Symbol)
	}
	if data.CompanyName != "Apple Inc." {
		t.Errorf("CompanyName = %q, want Apple Inc.", data.CompanyName)
	}
	if data.CurrentEPS != 14.0 {
		t.Errorf("CurrentEPS = %v, want 14.0", data.CurrentEPS)
	}
	if len(data.HistoricalData) != 8 {
		t.Fatalf("HistoricalData length = %d, want 8", len(data.HistoricalData))
	}
	if got := data.GetLatestTTM_EPS(); got != 14.0 {
		t.Errorf("GetLatestTTM_EPS() = %v, want 14.0", got)
	}
	if got := data.GetEPSForDate("2023-06-30"); got != 12.0 {
		t.Errorf("GetEPSForDate(2023-06-30) = %v, want 12.0", got)
	}
}
