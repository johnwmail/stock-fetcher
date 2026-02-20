# Stock Price Fetcher

A Go web server that fetches historical stock price data with P/E ratios.

## Features

- **US Stocks**: Daily prices with historical P/E ratio (via macrotrends.net)
- **HK Stocks**: Daily prices via Yahoo Finance
- Period aggregation: weekly, monthly, quarterly, yearly
- Drop day analysis (2%–5%+ buckets, close-based and low-based)
- SQLite cache — first fetch ~10s, subsequent fetches ~20ms
- Excel export
- Web UI with interactive charts (price, P/E, EPS)
- AWS Lambda support

## Running

```bash
go build
./stock-fetcher              # starts on :8080
PORT=3000 ./stock-fetcher    # starts on :3000
```

## Docker

```bash
# Using published image
docker compose up -d

# Local build
docker compose --profile local up stock-fetcher-local -d
```

Cache is persisted in a named volume (`cache-data` → `/data/cache.db`).

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check + version info |
| GET | `/api/stock/{symbol}` | Fetch stock data (JSON) |
| GET | `/api/stock-excel/{symbol}` | Download Excel file |
| GET | `/api/indices` | List available indices |
| GET | `/api/indices/{name}` | List symbols in an index |
| GET | `/` | Web UI |

### Query Parameters

| Param | Default | Values |
|-------|---------|--------|
| `days` | 1825 (5 years) | Number of days of historical data |
| `period` | monthly | `daily`, `weekly`, `monthly`, `quarterly`, `yearly` |

### Examples

```bash
curl localhost:8080/api/stock/AAPL
curl localhost:8080/api/stock/AAPL?days=90\&period=daily
curl localhost:8080/api/stock/0700.HK?days=365
curl localhost:8080/api/indices/dow
```

## Cache

Historical data is cached in SQLite. The DB path is auto-detected:

| Environment | Detection | DB path |
|---|---|---|
| Lambda | `AWS_LAMBDA_FUNCTION_NAME` env | `/tmp/cache.db` |
| Docker | `/data` directory exists | `/data/cache.db` |
| Local | fallback | `./cache.db` |

Override with `DB_PATH` env var. Set `DB_PATH=none` to disable caching.

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
