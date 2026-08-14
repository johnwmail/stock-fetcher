# Stock Fetcher

A stock data service that fetches historical price data with P/E ratios,
running on Cloudflare Workers (`workerd`) with a TypeScript backend and D1
(SQLite) cache.

## Features

- **US Stocks**: Daily prices with historical P/E ratio (via macrotrends.net)
- **HK Stocks**: Daily prices via Yahoo Finance
- Period aggregation: weekly, monthly, quarterly, yearly
- Drop day analysis (2%–5%+ buckets, close-based and low-based)
- D1-backed SQLite cache with delta fetching
- Excel export
- Responsive web UI with interactive charts (price, P/E) and EPS in tooltips
- Mobile-friendly — chart renders on all screen sizes

## Quick start

Requires **Node.js 22 or newer** (24.x recommended).

```bash
npm ci
npm run typecheck
npm test

# Create the D1 database and apply migrations
npx wrangler login
npx wrangler d1 create stock-fetcher
# paste the returned database_id into wrangler.toml
npx wrangler d1 migrations apply stock-fetcher --local

# Run locally
npm run dev
```

## Deployment

```bash
npx wrangler d1 migrations apply stock-fetcher --remote
npx wrangler deploy
```

The deployed worker is available at:

```
https://stock.<your-subdomain>.workers.dev
```

See [WORKERS.md](WORKERS.md) for full architecture, D1, and GitHub Actions
deployment documentation.

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check + version info |
| GET | `/api/stock/{symbol}` | Fetch stock data (JSON) |
| GET | `/api/stock-excel/{symbol}` | Download Excel file |
| GET | `/api/indices` | List available indices |
| GET | `/api/indices/{name}` | List symbols in an index |
| GET | `/` | Web UI |

### Query parameters

| Param | Default | Values |
|-------|---------|--------|
| `days` | `1825` (5 years) | Number of days of historical data |
| `period` | `monthly` | `daily`, `weekly`, `monthly`, `quarterly`, `yearly` |
| `source` | `auto` | `auto`, `macrotrends`, `yahoo` |

`source` controls which provider is used:

- `auto` tries macrotrends first, then Yahoo Finance.
- A specific source forces only that provider.
- HK stocks (`.HK`) only support `auto` and `yahoo`.
- The `data_source` field in the response reports which provider was used.

## Data sources

| Stock Type | Source | P/E Ratio |
|------------|--------|----------|
| US Stocks | macrotrends.net | ✅ Yes (TTM, historical) |
| US Stocks (fallback) | Yahoo Finance | ❌ No |
| HK Stocks (.HK) | Yahoo Finance | ❌ No |

US stocks are fetched in this order:

1. **macrotrends.net** — prices plus historical P/E
2. **Yahoo Finance** — prices without P/E (fallback)

## License

MIT License
