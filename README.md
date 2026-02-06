# Stock Price Fetcher

A simple Go application to fetch historical stock price data with P/E ratio.

## Features

- **US Stocks**: Fetches data from macrotrends.net with daily P/E ratio
- **HK Stocks**: Fetches data from Yahoo Finance (no P/E)
- Support for CSV, JSON, and table output formats
- Configurable date range (by number of days)
- Built-in symbol lists for major indices (S&P 500, DOW, NASDAQ 100, Hang Seng)

## Installation

```bash
cd stock-fetcher
go build
```

## Usage

```bash
# Fetch US stock (includes P/E ratio)
./stock-fetcher -symbol AAPL -days 365

# Fetch Hong Kong stock (no P/E)
./stock-fetcher -symbol 0700.HK -days 365

# Export to JSON format
./stock-fetcher -symbol GOOGL -format json

# Export to table format
./stock-fetcher -symbol MSFT -format table

# Specify custom output file
./stock-fetcher -symbol TSLA -days 365 -output tesla_2025.csv
```

## Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-symbol` | MSFT | Stock ticker symbol |
| `-days` | 365 | Number of days of historical data |
| `-output` | `<SYMBOL>_historical.csv` | Output filename |
| `-format` | csv | Output format: `csv`, `json`, or `table` |
| `-list` | | List symbols: `sp500`, `dow`, `nasdaq100`, `hangseng`, or `all` |

## Listing Index Symbols

```bash
./stock-fetcher -list all      # List available indices
./stock-fetcher -list sp500    # List S&P 500 stocks
./stock-fetcher -list dow      # List Dow Jones stocks
./stock-fetcher -list nasdaq100 # List NASDAQ 100 stocks
./stock-fetcher -list hangseng # List Hang Seng stocks
```

## Output Formats

### CSV - US Stock (with P/E)
```csv
Date,Open,High,Low,Close,Volume,Change,PE
2025-12-22,272.86,273.88,270.51,270.97,36.572M,,34.30
2025-12-23,270.84,272.50,269.56,272.36,29.642M,0.51%,34.48
```

### CSV - HK Stock (no P/E)
```csv
Date,Open,High,Low,Close,Volume,Change
2026-01-06,627.00,638.50,626.00,632.50,24.17M,
2026-01-07,627.50,629.50,615.00,624.50,21.38M,-1.26%
```

### JSON
```json
[
  {
    "date": "2025-12-22",
    "open": "272.86",
    "high": "273.88",
    "low": "270.51",
    "close": "270.97",
    "volume": "36.572M",
    "change": "",
    "pe": "34.30"
  }
]
```

## Examples

### Fetch Apple stock for 1 year
```bash
./stock-fetcher -symbol AAPL -days 365
```

Output:
```
Fetching 365 days of data for AAPL from macrotrends.net...
Received 252 records
TTM EPS: $7.90
Data saved to AAPL_historical.csv

Preview (first 5 records):
Date                 Open         High          Low        Close       Volume     Change         PE
-----------------------------------------------------------------------------------------------
2025-12-22         272.86       273.88       270.51       270.97      36.572M                 34.30
2025-12-23         270.84       272.50       269.56       272.36      29.642M      0.51%      34.48
...
```

### Fetch Tencent (Hong Kong)
```bash
./stock-fetcher -symbol 0700.HK -days 365
```

### Fetch multiple stocks
```bash
for sym in AAPL MSFT GOOGL AMZN; do
  ./stock-fetcher -symbol $sym -days 365
done
```

## Data Sources

| Stock Type | Source | P/E Ratio |
|------------|--------|----------|
| US Stocks | macrotrends.net | ✅ Yes (TTM, historical) |
| HK Stocks (.HK) | Yahoo Finance | ❌ No |

### P/E Ratio Data Availability

P/E ratios are calculated using historical TTM EPS data from macrotrends.net. The EPS data availability varies by company:

- Most companies have EPS data starting from 2010-2012
- For dates before the first available EPS, P/E will be empty
- Example: MSFT EPS data starts from 2011-12-31

This is a data source limitation, not a calculation error.

## Supported Indices

| Index | Stocks | Description |
|-------|--------|-------------|
| `sp500` | 502 | S&P 500 - Largest US companies |
| `dow` | 30 | Dow Jones Industrial Average |
| `nasdaq100` | 102 | NASDAQ 100 |
| `hangseng` | 85 | Hong Kong Hang Seng Index |

## License

MIT License
