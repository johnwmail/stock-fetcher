import type { FetchMeta, StockData } from './types';
import { formatFloat } from './helpers';

interface PriceRow {
  date: string;
  open: string | null;
  high: string | null;
  low: string | null;
  close: string | null;
  volume: string | null;
  pe: string | null;
}

interface FetchMetaRow {
  symbol: string;
  source: string | null;
  company_name: string | null;
  ttm_eps: number | null;
  last_fetched: string;
  latest_date: string | null;
  earliest_date: string | null;
}

const D1_BATCH_LIMIT = 100;

// SQLite-backed cache, equivalent to the Go cache but running on Cloudflare D1.
export class D1Cache {
  constructor(private readonly db: D1Database) {}

  async getFetchMeta(symbol: string): Promise<FetchMeta | null> {
    const row = await this.db
      .prepare(
        `SELECT symbol, source, company_name, ttm_eps, last_fetched, latest_date, earliest_date
         FROM fetch_log WHERE symbol = ?`,
      )
      .bind(symbol)
      .first<FetchMetaRow>();

    if (!row) return null;

    const lastFetched = new Date(row.last_fetched);
    return {
      symbol: row.symbol,
      source: row.source ?? '',
      companyName: row.company_name ?? '',
      ttmEPS: row.ttm_eps ?? 0,
      lastFetched: Number.isNaN(lastFetched.getTime()) ? new Date(0) : lastFetched,
      latestDate: row.latest_date ?? '',
      earliestDate: row.earliest_date ?? '',
    };
  }

  // Returns cached daily prices newest-first, recomputing Change/HChange.
  async getDailyPrices(
    symbol: string,
    startDate: string,
    endDate: string,
  ): Promise<StockData[]> {
    const result = await this.db
      .prepare(
        `SELECT date, open, high, low, close, volume, pe
         FROM daily_prices
         WHERE symbol = ? AND date >= ? AND date <= ?
         ORDER BY date ASC`,
      )
      .bind(symbol, startDate, endDate)
      .all<PriceRow>();

    const data: StockData[] = [];
    let prevClose = 0;
    let prevHigh = 0;

    for (const row of result.results ?? []) {
      const close = Number(row.close ?? 0) || 0;
      const high = Number(row.high ?? 0) || 0;

      const record: StockData = {
        date: row.date,
        open: row.open ?? '',
        high: row.high ?? '',
        low: row.low ?? '',
        close: row.close ?? '',
        volume: row.volume ?? '',
        change: '',
        hchange: '',
      };
      if (row.pe) record.pe = row.pe;

      if (prevClose > 0) {
        record.change = formatFloat(((close - prevClose) / prevClose) * 100) + '%';
      }
      if (prevHigh > 0) {
        record.hchange = formatFloat(((close - prevHigh) / prevHigh) * 100) + '%';
      }

      data.push(record);
      prevClose = close;
      prevHigh = high;
    }

    data.reverse();
    return data;
  }

  // Stores daily price records, using INSERT OR REPLACE.
  async storeDailyPrices(symbol: string, data: StockData[]): Promise<void> {
    const statements: D1PreparedStatement[] = [];
    for (const d of data) {
      statements.push(
        this.db
          .prepare(
            `INSERT OR REPLACE INTO daily_prices (symbol, date, open, high, low, close, volume, pe)
             VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
          )
          .bind(symbol, d.date, d.open, d.high, d.low, d.close, d.volume, d.pe ?? null),
      );

      if (statements.length >= D1_BATCH_LIMIT) {
        await this.db.batch(statements);
        statements.length = 0;
      }
    }

    if (statements.length > 0) {
      await this.db.batch(statements);
    }
  }

  // Upserts fetch metadata for a symbol.
  async updateFetchLog(meta: FetchMeta): Promise<void> {
    await this.db
      .prepare(
        `INSERT OR REPLACE INTO fetch_log
           (symbol, source, company_name, ttm_eps, last_fetched, latest_date, earliest_date)
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
      )
      .bind(
        meta.symbol,
        meta.source,
        meta.companyName,
        meta.ttmEPS,
        meta.lastFetched.toISOString(),
        meta.latestDate,
        meta.earliestDate,
      )
      .run();
  }
}

// True when the symbol was fetched today (UTC).
export function isFresh(meta: FetchMeta): boolean {
  const now = new Date();
  return (
    meta.lastFetched.getUTCFullYear() === now.getUTCFullYear() &&
    meta.lastFetched.getUTCMonth() === now.getUTCMonth() &&
    meta.lastFetched.getUTCDate() === now.getUTCDate()
  );
}

// True when cached data covers the requested start date.
export function coversRange(meta: FetchMeta, startDate: string): boolean {
  return meta.earliestDate <= startDate;
}
