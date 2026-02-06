package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// StockData represents a single day's stock data
type StockData struct {
	Date   string `json:"date"`
	Open   string `json:"open"`
	High   string `json:"high"`
	Low    string `json:"low"`
	Close  string `json:"close"`
	Volume string `json:"volume"`
	Change string `json:"change"`
	PE     string `json:"pe,omitempty"`
}

// isHKStock checks if the symbol is a Hong Kong stock
func isHKStock(symbol string) bool {
	return strings.HasSuffix(strings.ToUpper(symbol), ".HK")
}

// WriteCSV writes stock data to a CSV file
func WriteCSV(data []StockData, filename string, includePE bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if includePE {
		if err := writer.Write([]string{"Date", "Open", "High", "Low", "Close", "Volume", "Change", "PE"}); err != nil {
			return err
		}
		for _, d := range data {
			if err := writer.Write([]string{d.Date, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change, d.PE}); err != nil {
				return err
			}
		}
	} else {
		if err := writer.Write([]string{"Date", "Open", "High", "Low", "Close", "Volume", "Change"}); err != nil {
			return err
		}
		for _, d := range data {
			if err := writer.Write([]string{d.Date, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change}); err != nil {
				return err
			}
		}
	}

	return nil
}

// WriteJSON writes stock data to a JSON file
func WriteJSON(data []StockData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// WriteTable writes stock data in a formatted table
func WriteTable(data []StockData, filename string, includePE bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if includePE {
		fmt.Fprintf(file, "%-12s %12s %12s %12s %12s %12s %10s %10s\n",
			"Date", "Open", "High", "Low", "Close", "Volume", "Change", "PE")
		fmt.Fprintln(file, strings.Repeat("-", 95))
		for _, d := range data {
			fmt.Fprintf(file, "%-12s %12s %12s %12s %12s %12s %10s %10s\n",
				d.Date, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change, d.PE)
		}
	} else {
		fmt.Fprintf(file, "%-12s %12s %12s %12s %12s %12s %10s\n",
			"Date", "Open", "High", "Low", "Close", "Volume", "Change")
		fmt.Fprintln(file, strings.Repeat("-", 85))
		for _, d := range data {
			fmt.Fprintf(file, "%-12s %12s %12s %12s %12s %12s %10s\n",
				d.Date, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change)
		}
	}

	return nil
}

// expandListAlias expands list aliases to full names
func expandListAlias(name string) string {
	aliases := map[string]string{
		"sp":     "sp500",
		"hk":     "hangseng",
		"nasdaq": "nasdaq100",
	}
	if expanded, ok := aliases[strings.ToLower(name)]; ok {
		return expanded
	}
	return name
}

// listSymbols prints symbols for the specified index
func listSymbols(indexName string) {
	indices := GetIndices()

	if indexName == "all" || indexName == "help" {
		fmt.Println("Available indices:")
		fmt.Println()
		for key, idx := range indices {
			fmt.Printf("  %-12s - %s (%d stocks)\n", key, idx.Name, len(idx.Symbols))
		}
		fmt.Println()
		fmt.Println("Aliases: sp=sp500, hk=hangseng, nasdaq=nasdaq100")
		fmt.Println()
		fmt.Println("Usage: ./stock-fetcher -l <index>")
		fmt.Println("Example: ./stock-fetcher -l sp")
		return
	}

	// Expand alias
	indexName = expandListAlias(indexName)

	idx, ok := indices[strings.ToLower(indexName)]
	if !ok {
		fmt.Printf("Unknown index: %s\n", indexName)
		fmt.Println("\nAvailable: sp500 (sp), dow, nasdaq100 (nasdaq), hangseng (hk)")
		fmt.Println("Use '-l all' to see details.")
		return
	}

	fmt.Printf("%s\n", idx.Name)
	fmt.Printf("%s\n", idx.Description)
	fmt.Printf("Total: %d stocks\n", len(idx.Symbols))
	fmt.Println(strings.Repeat("-", 60))

	cols := 8
	for i, sym := range idx.Symbols {
		fmt.Printf("%-10s", sym)
		if (i+1)%cols == 0 {
			fmt.Println()
		}
	}
	if len(idx.Symbols)%cols != 0 {
		fmt.Println()
	}
}

// reverseData reverses the slice so newest data is first
func reverseData(data []StockData) []StockData {
	result := make([]StockData, len(data))
	for i, d := range data {
		result[len(data)-1-i] = d
	}
	return result
}

// fetchUSStock fetches US stock data from macrotrends (with P/E)
func fetchUSStock(symbol string, days int) ([]StockData, float64, error) {
	fetcher := NewMacrotrendsFetcher()

	// Get TTM EPS for P/E calculation
	peData, err := fetcher.FetchPERatio(symbol)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch P/E data: %w", err)
	}
	ttmEPS := peData.GetLatestTTM_EPS()

	// Get daily prices
	prices, err := fetcher.FetchDailyPrices(symbol, days)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch price data: %w", err)
	}

	// Convert to StockData with P/E (calculate change from old to new first)
	var data []StockData
	var prevClose float64

	for _, p := range prices {
		close, _ := strconv.ParseFloat(p.Close, 64)
		open, _ := strconv.ParseFloat(p.Open, 64)
		high, _ := strconv.ParseFloat(p.High, 64)
		low, _ := strconv.ParseFloat(p.Low, 64)

		// Calculate change %
		change := ""
		if prevClose > 0 {
			pctChange := ((close - prevClose) / prevClose) * 100
			change = fmt.Sprintf("%.2f%%", pctChange)
		}

		// Calculate P/E
		pe := ""
		if ttmEPS > 0 {
			pe = fmt.Sprintf("%.2f", close/ttmEPS)
		}

		data = append(data, StockData{
			Date:   p.Date,
			Open:   fmt.Sprintf("%.2f", open),
			High:   fmt.Sprintf("%.2f", high),
			Low:    fmt.Sprintf("%.2f", low),
			Close:  fmt.Sprintf("%.2f", close),
			Volume: p.Volume + "M",
			Change: change,
			PE:     pe,
		})

		prevClose = close
	}

	// Reverse so newest is first
	return reverseData(data), ttmEPS, nil
}

// fetchHKStock fetches HK stock data from Yahoo (no P/E)
func fetchHKStock(symbol string, days int) ([]StockData, error) {
	fetcher := NewYahooFetcher()
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	yahooData, err := fetcher.FetchHistoricalData(symbol, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Reverse so newest is first
	return reverseData(yahooData), nil
}

func printUsage() {
	fmt.Println("Stock Price Fetcher - Fetch historical stock data with P/E ratio")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  stock-fetcher -s <SYMBOL> [options]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  stock-fetcher -s AAPL                # US stock, 3 years, with P/E")
	fmt.Println("  stock-fetcher -s AAPL -d 30          # US stock, 30 days")
	fmt.Println("  stock-fetcher -s AAPL -y             # US stock, use Yahoo (no P/E)")
	fmt.Println("  stock-fetcher -s 0700.HK             # HK stock (Yahoo, no P/E)")
	fmt.Println("  stock-fetcher -s AAPL -p weekly      # Weekly aggregated report")
	fmt.Println("  stock-fetcher -s AAPL -p monthly     # Monthly aggregated report")
	fmt.Println("  stock-fetcher -l sp                  # List S&P 500 symbols")
	fmt.Println("  stock-fetcher -l hk                  # List Hang Seng symbols")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -s, -sym, -symbol     Stock symbol (e.g., MSFT, AAPL, 0700.HK)")
	fmt.Println("  -d, -days             Number of days (default: 1095 = 3 years)")
	fmt.Println("  -p, -period           Period aggregation: weekly, monthly, quarterly, yearly")
	fmt.Println("  -source, -src         Data source: macrotrends or yahoo")
	fmt.Println("  -y                    Use Yahoo Finance (alias for -source yahoo)")
	fmt.Println("  -m                    Use macrotrends.net (alias for -source macrotrends)")
	fmt.Println("  -l, -list             List index: sp500/sp, dow, nasdaq100/nasdaq, hangseng/hk, all")
	fmt.Println("  -format               Output format: csv, json, table (default: csv)")
	fmt.Println("  -output               Output filename (default: <SYMBOL>_historical.csv)")
	fmt.Println()
	fmt.Println("Period Reports:")
	fmt.Println("  Period reports aggregate daily data and include drop day counts:")
	fmt.Println("  - Drop2%: Days with 2-3% price drop")
	fmt.Println("  - Drop3%: Days with 3-4% price drop")
	fmt.Println("  - Drop4%: Days with 4-5% price drop")
	fmt.Println("  - Drop5%: Days with 5%+ price drop")
	fmt.Println()
	fmt.Println("Data Sources:")
	fmt.Println("  macrotrends  - Default for US stocks (includes P/E ratio)")
	fmt.Println("  yahoo        - Default for HK stocks (no P/E)")
}

func main() {
	// Main flags
	symbol := flag.String("symbol", "", "Stock symbol (e.g., MSFT, AAPL, 0700.HK)")
	days := flag.Int("days", 1095, "Number of days of historical data (default 3 years)")
	output := flag.String("output", "", "Output filename (default: <symbol>_historical.csv)")
	format := flag.String("format", "csv", "Output format: csv, json, or table")
	source := flag.String("source", "", "Data source: macrotrends (with P/E) or yahoo (no P/E)")
	listIndex := flag.String("list", "", "List symbols: sp500, dow, nasdaq100, hangseng, or 'all'")
	period := flag.String("period", "", "Period aggregation: weekly, monthly, quarterly, yearly")

	// Short aliases
	flag.StringVar(symbol, "s", "", "Alias for -symbol")
	flag.StringVar(symbol, "sym", "", "Alias for -symbol")
	flag.IntVar(days, "d", 1095, "Alias for -days")
	flag.StringVar(listIndex, "l", "", "Alias for -list")
	flag.StringVar(period, "p", "", "Alias for -period")
	yahooSource := flag.Bool("y", false, "Use Yahoo Finance as data source (alias for -source yahoo)")
	macroSource := flag.Bool("m", false, "Use macrotrends as data source (alias for -source macrotrends)")
	flag.StringVar(source, "src", "", "Alias for -source")

	flag.Parse()

	// Handle source aliases
	if *yahooSource {
		*source = "yahoo"
	} else if *macroSource {
		*source = "macrotrends"
	}

	// Show usage if no arguments
	if *symbol == "" && *listIndex == "" {
		printUsage()
		return
	}

	if *listIndex != "" {
		listSymbols(*listIndex)
		return
	}

	// Parse period type if specified
	var periodType PeriodType
	if *period != "" {
		var err error
		periodType, err = ParsePeriodType(*period)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Set default output filename
	if *output == "" {
		ext := "csv"
		switch *format {
		case "json":
			ext = "json"
		case "table":
			ext = "txt"
		}
		if *period != "" {
			*output = fmt.Sprintf("%s_%s.%s", strings.ToUpper(*symbol), *period, ext)
		} else {
			*output = fmt.Sprintf("%s_historical.%s", strings.ToUpper(*symbol), ext)
		}
	}

	var data []StockData
	var err error
	var ttmEPS float64
	includePE := false

	// Determine data source
	useYahoo := isHKStock(*symbol) || *source == "yahoo"

	if useYahoo {
		// Use Yahoo Finance (no P/E)
		fmt.Printf("Fetching %d days of data for %s from Yahoo Finance...\n", *days, strings.ToUpper(*symbol))
		data, err = fetchHKStock(*symbol, *days)
	} else {
		// Use macrotrends (with P/E)
		fmt.Printf("Fetching %d days of data for %s from macrotrends.net...\n", *days, strings.ToUpper(*symbol))
		data, ttmEPS, err = fetchUSStock(*symbol, *days)
		includePE = true
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching data: %v\n", err)
		os.Exit(1)
	}

	if len(data) == 0 {
		fmt.Println("No data received.")
		os.Exit(1)
	}

	fmt.Printf("Received %d daily records\n", len(data))
	if includePE && ttmEPS > 0 {
		fmt.Printf("TTM EPS: $%.2f\n", ttmEPS)
	}

	// Handle period aggregation
	if *period != "" {
		// Data is newest-first, but AggregateToPeriods expects oldest-first
		reversedData := reverseData(data)
		periodData := AggregateToPeriods(reversedData, periodType)

		if len(periodData) == 0 {
			fmt.Println("No period data generated.")
			os.Exit(1)
		}

		fmt.Printf("Aggregated into %d %s periods\n", len(periodData), *period)

		// Write period output
		switch *format {
		case "json":
			if !strings.HasSuffix(*output, ".json") {
				*output = strings.TrimSuffix(*output, ".csv") + ".json"
			}
			err = WritePeriodJSON(periodData, *output)
		case "table":
			if !strings.HasSuffix(*output, ".txt") {
				*output = strings.TrimSuffix(*output, ".csv") + ".txt"
			}
			err = WritePeriodTable(periodData, *output, includePE)
		default:
			err = WritePeriodCSV(periodData, *output, includePE)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Data saved to %s\n", *output)

		// Show preview
		fmt.Println("\nPreview (first 5 periods):")
		PrintPeriodPreview(periodData, 5, includePE)
		return
	}

	// Write daily output
	switch *format {
	case "json":
		if !strings.HasSuffix(*output, ".json") {
			*output = strings.TrimSuffix(*output, ".csv") + ".json"
		}
		err = WriteJSON(data, *output)
	case "table":
		if !strings.HasSuffix(*output, ".txt") {
			*output = strings.TrimSuffix(*output, ".csv") + ".txt"
		}
		err = WriteTable(data, *output, includePE)
	default:
		err = WriteCSV(data, *output, includePE)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Data saved to %s\n", *output)

	// Show preview
	fmt.Println("\nPreview (first 5 records):")
	if includePE {
		fmt.Printf("%-12s %12s %12s %12s %12s %12s %10s %10s\n",
			"Date", "Open", "High", "Low", "Close", "Volume", "Change", "PE")
		fmt.Println(strings.Repeat("-", 95))
		for i, d := range data {
			if i >= 5 {
				break
			}
			fmt.Printf("%-12s %12s %12s %12s %12s %12s %10s %10s\n",
				d.Date, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change, d.PE)
		}
	} else {
		fmt.Printf("%-12s %12s %12s %12s %12s %12s %10s\n",
			"Date", "Open", "High", "Low", "Close", "Volume", "Change")
		fmt.Println(strings.Repeat("-", 85))
		for i, d := range data {
			if i >= 5 {
				break
			}
			fmt.Printf("%-12s %12s %12s %12s %12s %12s %10s\n",
				d.Date, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change)
		}
	}
}
