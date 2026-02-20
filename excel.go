package main

import (
	"fmt"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// ExcelParams contains parameters for Excel generation
type ExcelParams struct {
	Symbol      string
	CompanyName string
	Period      string
	TTMEPS      float64
	IncludePE   bool
	Data        []StockData
	PeriodData  []PeriodData
}

// GenerateExcel creates an Excel file from stock data
func GenerateExcel(params ExcelParams) (*excelize.File, error) {
	f := excelize.NewFile()

	sheetName := "Stock Data"
	_ = f.SetSheetName("Sheet1", sheetName)

	// Style for header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"4472C4"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	// Style for numbers
	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4, // #,##0.00
	})

	// Add metadata
	setCell(f, sheetName, 1, 1, "Symbol:")
	setCell(f, sheetName, 2, 1, params.Symbol)
	setCell(f, sheetName, 1, 2, "Company:")
	setCell(f, sheetName, 2, 2, params.CompanyName)
	setCell(f, sheetName, 1, 3, "Period:")
	setCell(f, sheetName, 2, 3, params.Period)
	if params.IncludePE {
		setCell(f, sheetName, 1, 4, "TTM EPS:")
		setCell(f, sheetName, 2, 4, params.TTMEPS)
	}

	row := 6

	if params.PeriodData != nil {
		row = writePeriodData(f, sheetName, row, params.PeriodData, params.IncludePE, headerStyle)
	} else {
		row = writeDailyData(f, sheetName, row, params.Data, params.IncludePE, headerStyle, numberStyle)
	}

	// Auto-fit columns
	_ = row // silence unused warning
	for col := 1; col <= 16; col++ {
		colName, _ := excelize.ColumnNumberToName(col)
		_ = f.SetColWidth(sheetName, colName, colName, 12)
	}

	return f, nil
}

// writeDailyData writes daily stock data to Excel
func writeDailyData(f *excelize.File, sheet string, startRow int, data []StockData, includePE bool, headerStyle, numberStyle int) int {
	headers := []string{"Date", "Open", "High", "Low", "Close", "Volume", "Change", "HChange"}
	if includePE {
		headers = append(headers, "PE")
	}

	// Write headers
	for col, h := range headers {
		setCellWithStyle(f, sheet, col+1, startRow, h, headerStyle)
	}
	startRow++

	// Write data rows
	for _, d := range data {
		setCell(f, sheet, 1, startRow, d.Date)
		setCellNum(f, sheet, 2, startRow, d.Open, numberStyle)
		setCellNum(f, sheet, 3, startRow, d.High, numberStyle)
		setCellNum(f, sheet, 4, startRow, d.Low, numberStyle)
		setCellNum(f, sheet, 5, startRow, d.Close, numberStyle)
		setCell(f, sheet, 6, startRow, d.Volume)
		setCell(f, sheet, 7, startRow, d.Change)
		setCell(f, sheet, 8, startRow, d.HChange)
		if includePE {
			setCell(f, sheet, 9, startRow, d.PE)
		}
		startRow++
	}
	return startRow
}

// writePeriodData writes period aggregated data to Excel
func writePeriodData(f *excelize.File, sheet string, startRow int, data []PeriodData, includePE bool, headerStyle int) int {
	headers := []string{"Period", "Start", "End", "Open", "High", "Low", "Close", "Volume", "Change", "HChange"}
	if includePE {
		headers = append(headers, "PE")
	}
	headers = append(headers, "Days", "C/L-2%", "C/L-3%", "C/L-4%", "C/L-5%")

	// Write headers
	for col, h := range headers {
		setCellWithStyle(f, sheet, col+1, startRow, h, headerStyle)
	}
	startRow++

	// Write data rows
	for _, p := range data {
		col := 1
		setCell(f, sheet, col, startRow, p.Period)
		col++
		setCell(f, sheet, col, startRow, p.StartDate)
		col++
		setCell(f, sheet, col, startRow, p.EndDate)
		col++
		setCell(f, sheet, col, startRow, parseFloatStr(p.Open))
		col++
		setCell(f, sheet, col, startRow, parseFloatStr(p.High))
		col++
		setCell(f, sheet, col, startRow, parseFloatStr(p.Low))
		col++
		setCell(f, sheet, col, startRow, parseFloatStr(p.Close))
		col++
		setCell(f, sheet, col, startRow, p.Volume)
		col++
		setCell(f, sheet, col, startRow, p.Change)
		col++
		setCell(f, sheet, col, startRow, p.HChange)
		col++
		if includePE {
			setCell(f, sheet, col, startRow, p.PE)
			col++
		}
		setCell(f, sheet, col, startRow, p.Days)
		col++
		setCell(f, sheet, col, startRow, fmt.Sprintf("%d/%d", p.Drop2Pct.Close, p.Drop2Pct.Low))
		col++
		setCell(f, sheet, col, startRow, fmt.Sprintf("%d/%d", p.Drop3Pct.Close, p.Drop3Pct.Low))
		col++
		setCell(f, sheet, col, startRow, fmt.Sprintf("%d/%d", p.Drop4Pct.Close, p.Drop4Pct.Low))
		col++
		setCell(f, sheet, col, startRow, fmt.Sprintf("%d/%d", p.Drop5Pct.Close, p.Drop5Pct.Low))
		startRow++
	}
	return startRow
}

// Helper functions
func setCell(f *excelize.File, sheet string, col, row int, value interface{}) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	_ = f.SetCellValue(sheet, cell, value)
}

func setCellWithStyle(f *excelize.File, sheet string, col, row int, value interface{}, style int) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	_ = f.SetCellValue(sheet, cell, value)
	_ = f.SetCellStyle(sheet, cell, cell, style)
}

func setCellNum(f *excelize.File, sheet string, col, row int, value string, style int) {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	v, _ := strconv.ParseFloat(value, 64)
	_ = f.SetCellValue(sheet, cell, v)
	_ = f.SetCellStyle(sheet, cell, cell, style)
}

func parseFloatStr(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
