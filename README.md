# Stock Price Fetcher

A Go web server that fetches historical stock price data with P/E ratios.

## Features

- **US Stocks**: Daily prices with historical P/E ratio (via macrotrends.net)
- **HK Stocks**: Daily prices via Yahoo Finance
- Period aggregation: weekly, monthly, quarterly, yearly
- Drop day analysis (2%–5%+ buckets, close-based and low-based)
- Excel export
- Web UI at `/`
- AWS Lambda support

## Running

```bash
go build
./stock-fetcher              # starts on :8080
PORT=3000 ./stock-fetcher    # starts on :3000
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check + version info |
| GET | `/api/stock/{symbol}?days=365&period=daily` | Fetch stock data (JSON) |
| GET | `/api/stock-excel/{symbol}?days=365&period=daily` | Download Excel file |
| GET | `/api/indices` | List available indices |
| GET | `/api/indices/{name}` | List symbols in an index |
| GET | `/` | Web UI |

### Query Parameters

| Param | Default | Values |
|-------|---------|--------|
| `days` | 365 | Number of days of historical data |
| `period` | daily | `daily`, `weekly`, `monthly`, `quarterly`, `yearly` |

### Example

```bash
curl localhost:8080/api/stock/AAPL?days=90
curl localhost:8080/api/stock/AAPL?days=365\&period=monthly
curl localhost:8080/api/indices
curl localhost:8080/api/indices/dow
```

## Data Sources

| Stock Type | Source | P/E Ratio |
|------------|--------|----------|
| US Stocks | macrotrends.net | ✅ Yes (TTM, historical) |
| HK Stocks (.HK) | Yahoo Finance | ❌ No |

US stocks automatically fall back to Yahoo Finance if macrotrends fails (e.g., ETFs).

## Supported Indices

| Index | Stocks | Description |
|-------|--------|-------------|
| `sp500` | 502 | S&P 500 |
| `dow` | 30 | Dow Jones Industrial Average |
| `nasdaq100` | 102 | NASDAQ 100 |
| `hangseng` | 85 | Hang Seng Index |

## License

MIT License
