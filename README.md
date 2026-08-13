# Stock Price Fetcher

A Go web server that fetches historical stock price data with P/E ratios.

## Features

- **US Stocks**: Daily prices with historical P/E ratio (via macrotrends.net, with SEC EDGAR + Yahoo fallback)
- **HK Stocks**: Daily prices via Yahoo Finance
- Period aggregation: weekly, monthly, quarterly, yearly
- Drop day analysis (2%–5%+ buckets, close-based and low-based)
- SQLite cache with delta fetching — first fetch ~10s, subsequent fetches ~20ms
- Excel export
- Responsive web UI with interactive charts (price, P/E) and EPS in tooltips
- Mobile-friendly — chart renders on all screen sizes
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

# Build from source
docker compose up -d --build
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
| `source` | auto | `auto`, `macrotrends`, `edgar`, `yahoo` |

`source` controls which provider is used. `auto` uses the fallback chain
(macrotrends → EDGAR → Yahoo). A specific source forces only that provider.
HK stocks (`.HK`) only support `auto` and `yahoo`.

### Examples

```bash
curl localhost:8080/api/stock/AAPL
curl localhost:8080/api/stock/AAPL?days=90\&period=daily
curl localhost:8080/api/stock/AAPL?source=edgar
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
| US Stocks (fallback) | SEC EDGAR + Yahoo Finance | ✅ Yes (TTM, historical) |
| HK Stocks (.HK) | Yahoo Finance | ❌ No |

US stocks are fetched in this order:

1. **macrotrends.net** — prices plus historical P/E (primary)
2. **SEC EDGAR + Yahoo Finance** — EDGAR provides historical quarterly diluted EPS
   (no API key required), Yahoo provides prices, and P/E is computed as
   price ÷ trailing-twelve-month EPS
3. **Yahoo Finance only** — prices without P/E (e.g., ETFs without EDGAR EPS data)

EDGAR requests use the SEC's fair-access rules (a declared User-Agent and a
rate limit well below 10 requests/second). Override the default contact with
`SEC_USER_AGENT`, e.g. `SEC_USER_AGENT="StockFetcher/1.0 (you@example.com)"`.

## Supported Indices

| Index | Stocks | Description |
|-------|--------|-------------|
| `sp500` | 502 | S&P 500 |
| `dow` | 30 | Dow Jones Industrial Average |
| `nasdaq100` | 102 | NASDAQ 100 |
| `hangseng` | 85 | Hang Seng Index |

## License

MIT License
