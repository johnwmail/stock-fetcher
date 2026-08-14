# Cloudflare Workers (TypeScript) backend

This is the TypeScript backend for the Cloudflare Workers (`workerd`)
platform. It is stateless, serves a JSON API plus the static frontend, and
uses D1 (SQLite) for caching.

## Layout

| Path | Purpose |
|------|---------|
| `src/index.ts` | Worker entry point, HTTP router, API handlers |
| `src/helpers.ts` | Source/symbol/formatting helpers |
| `src/providers.ts` | Provider fallback chain and cache orchestration |
| `src/yahoo.ts` | Yahoo Finance chart API fetcher |
| `src/macrotrends.ts` | macrotrends.net P/E + price fetcher |
| `src/cache.ts` | D1-backed SQLite cache |
| `src/period.ts` | Period aggregation + drop analysis |
| `src/excel.ts` | Excel export via exceljs |
| `src/data/static-data.json` | Company names and index constituents |
| `web/` | Static frontend served by Workers Static Assets |
| `migrations/` | D1 schema migrations |
| `wrangler.toml` | Workers config |

## Local development

Requires **Node.js 22 or newer** (24.x is recommended and used in CI).

```bash
npm ci
npm run typecheck
npm test

# Create a local D1 database and update wrangler.toml database_id first:
npx wrangler d1 create stock-fetcher
npx wrangler d1 migrations apply stock-fetcher --local

npm run dev
```

The worker serves the API on `/api/*` and the frontend on `/`.

## Data source selection

`GET /api/stock/{symbol}?source=auto|macrotrends|yahoo`

- `auto` (default) tries macrotrends, then Yahoo.
- HK stocks (`.HK`) only support `auto` and `yahoo`.
- The `data_source` field in the response reports which provider was used.

## Deployment

### 1. Create the D1 database

```bash
npx wrangler login
npx wrangler d1 create stock-fetcher
```

Paste the returned `database_id` into `wrangler.toml`, or provide it via the
`D1_DATABASE_ID` GitHub secret (the deploy workflow replaces the placeholder).

### Look up the D1 database id again

If you already created the database and need its id again:

```bash
npx wrangler d1 list
```

The id is shown next to the database name. You can also find it in the
Cloudflare Dashboard under **Workers & Pages → D1 SQL Database →
stock-fetcher**.

### 2. Configure GitHub secrets

| Secret | Description |
|--------|-------------|
| `CLOUDFLARE_API_TOKEN` | Cloudflare API token with Workers + D1 edit permissions |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare account ID |
| `D1_DATABASE_ID` | D1 database ID from `wrangler d1 create` |

### 3. Push to `main`

The `.github/workflows/deploy-workers.yml` workflow runs typecheck + tests,
applies D1 migrations, and deploys with `wrangler deploy`. It can also be
triggered manually from the Actions tab.

## Notes and caveats

- **macrotrends** is Cloudflare-protected and frequently returns `403 Just a
  moment` to non-browser IPs, including Workers egress. Verify it works from
  your account before relying on it.
- **Yahoo Finance** is an undocumented endpoint and may rate-limit datacenter
  IPs.
- Excel generation uses `nodejs_compat` for Node `Buffer` support.
