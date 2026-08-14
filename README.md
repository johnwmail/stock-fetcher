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

See [WORKERS.md](WORKERS.md) for full architecture, API, data source,
D1, and GitHub Actions deployment documentation.

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
