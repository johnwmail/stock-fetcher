package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PeriodData represents aggregated data for a period (week, month, quarter, year)
type PeriodData struct {
	Period    string `json:"period"`     // Period label (e.g., "2024-W01", "2024-01", "2024-Q1", "2024")
	StartDate string `json:"start_date"` // First trading day in period
	EndDate   string `json:"end_date"`   // Last trading day in period
	Open      string `json:"open"`       // Open price of first day
	High      string `json:"high"`       // Highest price in period
	Low       string `json:"low"`        // Lowest price in period
	Close     string `json:"close"`      // Close price of last day
	Volume    string `json:"volume"`     // Total volume in period
	Change    string `json:"change"`     // Period change percentage
	PE        string `json:"pe,omitempty"`
	Days      int    `json:"days"`       // Number of trading days
	Drop2Pct  int    `json:"drop_2pct"`  // Days with 2-3% drop
	Drop3Pct  int    `json:"drop_3pct"`  // Days with 3-4% drop
	Drop4Pct  int    `json:"drop_4pct"`  // Days with 4-5% drop
	Drop5Pct  int    `json:"drop_5pct"`  // Days with 5%+ drop
}

// PeriodType represents the type of period aggregation
type PeriodType string

const (
	PeriodWeekly    PeriodType = "weekly"
	PeriodMonthly   PeriodType = "monthly"
	PeriodQuarterly PeriodType = "quarterly"
	PeriodYearly    PeriodType = "yearly"
)

// ParsePeriodType parses a string into a PeriodType
func ParsePeriodType(s string) (PeriodType, error) {
	switch strings.ToLower(s) {
	case "weekly", "week", "w":
		return PeriodWeekly, nil
	case "monthly", "month", "m":
		return PeriodMonthly, nil
	case "quarterly", "quarter", "q":
		return PeriodQuarterly, nil
	case "yearly", "year", "y":
		return PeriodYearly, nil
	default:
		return "", fmt.Errorf("invalid period type: %s (use weekly, monthly, quarterly, or yearly)", s)
	}
}

// getPeriodKey returns a unique key for grouping dates into periods
func getPeriodKey(date time.Time, periodType PeriodType) string {
	switch periodType {
	case PeriodWeekly:
		year, week := date.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case PeriodMonthly:
		return date.Format("2006-01")
	case PeriodQuarterly:
		quarter := (date.Month()-1)/3 + 1
		return fmt.Sprintf("%d-Q%d", date.Year(), quarter)
	case PeriodYearly:
		return fmt.Sprintf("%d", date.Year())
	default:
		return date.Format("2006-01-02")
	}
}

// classifyDrop returns which drop bucket a change percentage falls into
// Returns 0 if no significant drop, or 2, 3, 4, 5 for the drop bucket
func classifyDrop(changeStr string) int {
	// Parse change string like "-2.50%" or "1.5%"
	changeStr = strings.TrimSuffix(changeStr, "%")
	change, err := strconv.ParseFloat(changeStr, 64)
	if err != nil {
		return 0
	}

	// Only count negative changes (drops)
	if change >= 0 {
		return 0
	}

	// Use absolute value for comparison
	absChange := -change

	// Classify into exclusive buckets (largest drop wins)
	if absChange >= 5.0 {
		return 5
	} else if absChange >= 4.0 {
		return 4
	} else if absChange >= 3.0 {
		return 3
	} else if absChange >= 2.0 {
		return 2
	}

	return 0
}

// AggregateToPeriods converts daily stock data into period aggregates
// Input data should be sorted with oldest first
func AggregateToPeriods(data []StockData, periodType PeriodType) []PeriodData {
	if len(data) == 0 {
		return nil
	}

	// Group data by period
	periodGroups := make(map[string][]StockData)
	periodOrder := make([]string, 0)

	for _, d := range data {
		date, err := time.Parse("2006-01-02", d.Date)
		if err != nil {
			continue
		}

		key := getPeriodKey(date, periodType)
		if _, exists := periodGroups[key]; !exists {
			periodOrder = append(periodOrder, key)
		}
		periodGroups[key] = append(periodGroups[key], d)
	}

	// Sort period keys chronologically
	sort.Strings(periodOrder)

	// Aggregate each period
	var result []PeriodData
	var prevPeriodClose float64

	for _, key := range periodOrder {
		days := periodGroups[key]
		if len(days) == 0 {
			continue
		}

		// Sort days by date (oldest first)
		sort.Slice(days, func(i, j int) bool {
			return days[i].Date < days[j].Date
		})

		// Calculate aggregates
		firstDay := days[0]
		lastDay := days[len(days)-1]

		var highVal, lowVal float64
		var totalVolume float64
		var drop2, drop3, drop4, drop5 int

		for i, d := range days {
			high := parseFloat(d.High)
			low := parseFloat(d.Low)
			vol := parseVolume(d.Volume)

			if i == 0 || high > highVal {
				highVal = high
			}
			if i == 0 || low < lowVal {
				lowVal = low
			}
			totalVolume += vol

			// Classify drop
			switch classifyDrop(d.Change) {
			case 2:
				drop2++
			case 3:
				drop3++
			case 4:
				drop4++
			case 5:
				drop5++
			}
		}

		// Calculate period change
		closeVal := parseFloat(lastDay.Close)
		change := ""
		if prevPeriodClose > 0 {
			pctChange := ((closeVal - prevPeriodClose) / prevPeriodClose) * 100
			change = fmt.Sprintf("%.2f%%", pctChange)
		}

		period := PeriodData{
			Period:    key,
			StartDate: firstDay.Date,
			EndDate:   lastDay.Date,
			Open:      firstDay.Open,
			High:      fmt.Sprintf("%.2f", highVal),
			Low:       fmt.Sprintf("%.2f", lowVal),
			Close:     lastDay.Close,
			Volume:    formatVolumeFloat(totalVolume),
			Change:    change,
			PE:        lastDay.PE,
			Days:      len(days),
			Drop2Pct:  drop2,
			Drop3Pct:  drop3,
			Drop4Pct:  drop4,
			Drop5Pct:  drop5,
		}

		result = append(result, period)
		prevPeriodClose = closeVal
	}

	// Reverse so newest is first (consistent with daily output)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// parseFloat parses a string to float64, returns 0 on error
func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// parseVolume parses volume string like "1.5M" or "500K" to float64
func parseVolume(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	multiplier := 1.0
	if strings.HasSuffix(s, "B") {
		multiplier = 1e9
		s = strings.TrimSuffix(s, "B")
	} else if strings.HasSuffix(s, "M") {
		multiplier = 1e6
		s = strings.TrimSuffix(s, "M")
	} else if strings.HasSuffix(s, "K") {
		multiplier = 1e3
		s = strings.TrimSuffix(s, "K")
	}

	v, _ := strconv.ParseFloat(s, 64)
	return v * multiplier
}

// formatVolumeFloat formats a volume float to a human-readable string
func formatVolumeFloat(v float64) string {
	if v >= 1e9 {
		return fmt.Sprintf("%.2fB", v/1e9)
	}
	if v >= 1e6 {
		return fmt.Sprintf("%.2fM", v/1e6)
	}
	if v >= 1e3 {
		return fmt.Sprintf("%.2fK", v/1e3)
	}
	return fmt.Sprintf("%.0f", v)
}

// WritePeriodCSV writes period data to a CSV file
func WritePeriodCSV(data []PeriodData, filename string, includePE bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if includePE {
		if err := writer.Write([]string{"Period", "Start", "End", "Open", "High", "Low", "Close", "Volume", "Change", "PE", "Days", "Drop2%", "Drop3%", "Drop4%", "Drop5%"}); err != nil {
			return err
		}
		for _, d := range data {
			if err := writer.Write([]string{
				d.Period, d.StartDate, d.EndDate, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change, d.PE,
				strconv.Itoa(d.Days), strconv.Itoa(d.Drop2Pct), strconv.Itoa(d.Drop3Pct), strconv.Itoa(d.Drop4Pct), strconv.Itoa(d.Drop5Pct),
			}); err != nil {
				return err
			}
		}
	} else {
		if err := writer.Write([]string{"Period", "Start", "End", "Open", "High", "Low", "Close", "Volume", "Change", "Days", "Drop2%", "Drop3%", "Drop4%", "Drop5%"}); err != nil {
			return err
		}
		for _, d := range data {
			if err := writer.Write([]string{
				d.Period, d.StartDate, d.EndDate, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change,
				strconv.Itoa(d.Days), strconv.Itoa(d.Drop2Pct), strconv.Itoa(d.Drop3Pct), strconv.Itoa(d.Drop4Pct), strconv.Itoa(d.Drop5Pct),
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// WritePeriodJSON writes period data to a JSON file
func WritePeriodJSON(data []PeriodData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// WritePeriodTable writes period data in a formatted table
func WritePeriodTable(data []PeriodData, filename string, includePE bool) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if includePE {
		fmt.Fprintf(file, "%-10s %-12s %-12s %10s %10s %10s %10s %10s %8s %8s %5s %5s %5s %5s %5s\n",
			"Period", "Start", "End", "Open", "High", "Low", "Close", "Volume", "Change", "PE", "Days", "D2%", "D3%", "D4%", "D5%")
		fmt.Fprintln(file, strings.Repeat("-", 140))
		for _, d := range data {
			fmt.Fprintf(file, "%-10s %-12s %-12s %10s %10s %10s %10s %10s %8s %8s %5d %5d %5d %5d %5d\n",
				d.Period, d.StartDate, d.EndDate, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change, d.PE,
				d.Days, d.Drop2Pct, d.Drop3Pct, d.Drop4Pct, d.Drop5Pct)
		}
	} else {
		fmt.Fprintf(file, "%-10s %-12s %-12s %10s %10s %10s %10s %10s %8s %5s %5s %5s %5s %5s\n",
			"Period", "Start", "End", "Open", "High", "Low", "Close", "Volume", "Change", "Days", "D2%", "D3%", "D4%", "D5%")
		fmt.Fprintln(file, strings.Repeat("-", 130))
		for _, d := range data {
			fmt.Fprintf(file, "%-10s %-12s %-12s %10s %10s %10s %10s %10s %8s %5d %5d %5d %5d %5d\n",
				d.Period, d.StartDate, d.EndDate, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change,
				d.Days, d.Drop2Pct, d.Drop3Pct, d.Drop4Pct, d.Drop5Pct)
		}
	}

	return nil
}

// PrintPeriodPreview prints a preview of period data to stdout
func PrintPeriodPreview(data []PeriodData, count int, includePE bool) {
	if includePE {
		fmt.Printf("%-10s %-12s %-12s %10s %10s %10s %10s %10s %8s %8s %5s %5s %5s %5s %5s\n",
			"Period", "Start", "End", "Open", "High", "Low", "Close", "Volume", "Change", "PE", "Days", "D2%", "D3%", "D4%", "D5%")
		fmt.Println(strings.Repeat("-", 140))
		for i, d := range data {
			if i >= count {
				break
			}
			fmt.Printf("%-10s %-12s %-12s %10s %10s %10s %10s %10s %8s %8s %5d %5d %5d %5d %5d\n",
				d.Period, d.StartDate, d.EndDate, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change, d.PE,
				d.Days, d.Drop2Pct, d.Drop3Pct, d.Drop4Pct, d.Drop5Pct)
		}
	} else {
		fmt.Printf("%-10s %-12s %-12s %10s %10s %10s %10s %10s %8s %5s %5s %5s %5s %5s\n",
			"Period", "Start", "End", "Open", "High", "Low", "Close", "Volume", "Change", "Days", "D2%", "D3%", "D4%", "D5%")
		fmt.Println(strings.Repeat("-", 130))
		for i, d := range data {
			if i >= count {
				break
			}
			fmt.Printf("%-10s %-12s %-12s %10s %10s %10s %10s %10s %8s %5d %5d %5d %5d %5d\n",
				d.Period, d.StartDate, d.EndDate, d.Open, d.High, d.Low, d.Close, d.Volume, d.Change,
				d.Days, d.Drop2Pct, d.Drop3Pct, d.Drop4Pct, d.Drop5Pct)
		}
	}
}
