import type { ProviderResult, StockData } from './types';
import {
  SourceAuto,
  SourceMacrotrends,
  SourceYahoo,
  daysAgo,
  formatDate,
  isHKStock,
  reverseData,
  toNumber,
} from './helpers';
import { MacrotrendsFetcher } from './macrotrends';
import { fetchHistoricalData } from './yahoo';
import { D1Cache, coversRange, isFresh } from './cache';

// Fetches US stock data from macrotrends (prices + historical P/E).
export async function fetchUSStock(
  symbol: string,
  days: number,
): Promise<{ data: StockData[]; ttmEPS: number; companyName: string }> {
  const fetcher = new MacrotrendsFetcher();

  const peData = await fetcher.fetchPERatio(symbol);
  const latestEPS = peData.getLatestTTMEPS();
  const companyName = peData.companyName;

  const prices = await fetcher.fetchDailyPrices(symbol, days);

  const data: StockData[] = [];
  let prevClose = 0;
  let prevHigh = 0;

  for (const p of prices) {
    const close = toNumber(p.c);
    const open = toNumber(p.o);
    const high = toNumber(p.h);
    const low = toNumber(p.l);

    let change = '';
    if (prevClose > 0) {
      change = `${(((close - prevClose) / prevClose) * 100).toFixed(2)}%`;
    }

    let hchange = '';
    if (prevHigh > 0) {
      hchange = `${(((close - prevHigh) / prevHigh) * 100).toFixed(2)}%`;
    }

    let pe = '';
    const historicalEPS = peData.getEPSForDate(p.d);
    if (historicalEPS > 0) {
      pe = (close / historicalEPS).toFixed(2);
    }

    data.push({
      date: p.d,
      open: open.toFixed(2),
      high: high.toFixed(2),
      low: low.toFixed(2),
      close: close.toFixed(2),
      volume: p.v + 'M',
      change,
      hchange,
      pe,
    });

    prevClose = close;
    prevHigh = high;
  }

  return { data: reverseData(data), ttmEPS: latestEPS, companyName };
}

// Fetches HK stock data from Yahoo (no P/E).
export async function fetchHKStock(
  symbol: string,
  days: number,
): Promise<{ data: StockData[]; companyName: string }> {
  const endDate = new Date();
  const startDate = daysAgo(days, endDate);

  const yahoo = await fetchHistoricalData(symbol, startDate, endDate);
  return { data: reverseData(yahoo.data), companyName: yahoo.companyName };
}

// Fetches stock data directly from the upstream provider.
export async function fetchFromProvider(
  symbol: string,
  days: number,
  useYahoo: boolean,
  source: string,
): Promise<ProviderResult> {
  if (useYahoo) {
    const { data, companyName } = await fetchHKStock(symbol, days);
    return { data, ttmEPS: 0, companyName, source: SourceYahoo };
  }

  switch (source) {
    case SourceMacrotrends: {
      const { data, ttmEPS, companyName } = await fetchUSStock(symbol, days);
      return { data, ttmEPS, companyName, source: SourceMacrotrends };
    }
    case SourceYahoo: {
      const { data, companyName } = await fetchHKStock(symbol, days);
      return { data, ttmEPS: 0, companyName, source: SourceYahoo };
    }
    default:
      break;
  }

  // SourceAuto: macrotrends first, then Yahoo.
  try {
    const result = await fetchUSStock(symbol, days);
    return { ...result, source: SourceMacrotrends };
  } catch {
    // fall through to Yahoo
  }

  const { data, companyName } = await fetchHKStock(symbol, days);
  return { data, ttmEPS: 0, companyName, source: SourceYahoo };
}

// Fetches stock data, using the D1 cache when available. A specific source
// only reads cache entries produced by that same source; auto uses any cached
// source.
export async function fetchStockData(
  db: D1Database | null,
  symbol: string,
  days: number,
  useYahoo: boolean,
  source: string,
): Promise<ProviderResult> {
  const symbolUpper = symbol.toUpperCase();
  const startDate = formatDate(daysAgo(days));
  const today = formatDate(new Date());

  const cache = db ? new D1Cache(db) : null;

  if (cache) {
    let meta = null;
    try {
      meta = await cache.getFetchMeta(symbolUpper);
    } catch {
      // Cache read failures should not block the request.
    }
    const cacheMatches = !!meta && (source === SourceAuto || meta.source === source);

    // Cache hit: fresh today and covers the requested range.
    if (cacheMatches && meta && isFresh(meta) && coversRange(meta, startDate)) {
      try {
        const data = await cache.getDailyPrices(symbolUpper, startDate, today);
        if (data.length > 0) {
          return {
            data,
            ttmEPS: meta.ttmEPS,
            companyName: meta.companyName,
            source: meta.source,
          };
        }
      } catch {
        // Fall through to provider fetch on cache read errors.
      }
    }

    // Cache stale or doesn't cover range — fetch from provider. If we have
    // matching cached data, fetch only the delta.
    let fetchDays = days;
    if (cacheMatches && meta && coversRange(meta, startDate)) {
      const daysSinceLatest =
        Math.floor((Date.now() - meta.lastFetched.getTime()) / 86_400_000) + 5;
      if (daysSinceLatest < fetchDays) {
        fetchDays = daysSinceLatest;
      }
    }

    try {
      const fetched = await fetchFromProvider(symbol, fetchDays, useYahoo, source);

      if (fetched.data.length > 0) {
        try {
          await cache.storeDailyPrices(symbolUpper, fetched.data);

          let earliestDate = fetched.data[fetched.data.length - 1].date;
          const latestDate = fetched.data[0].date;
          if (cacheMatches && meta && meta.earliestDate < earliestDate) {
            earliestDate = meta.earliestDate;
          }

          await cache.updateFetchLog({
            symbol: symbolUpper,
            source: fetched.source,
            companyName: fetched.companyName,
            ttmEPS: fetched.ttmEPS,
            lastFetched: new Date(),
            latestDate,
            earliestDate,
          });
        } catch {
          // Cache write failures are non-fatal, matching the Go behavior.
        }
      }

      try {
        const cachedData = await cache.getDailyPrices(symbolUpper, startDate, today);
        if (cachedData.length > 0) {
          return { ...fetched, data: cachedData };
        }
      } catch {
        // Serve provider data directly if the cache read fails.
      }
      return fetched;
    } catch (err) {
      // Provider failed — try serving stale cache if it matches the source.
      if (cacheMatches && meta) {
        try {
          const staleData = await cache.getDailyPrices(symbolUpper, startDate, today);
          if (staleData.length > 0) {
            return {
              data: staleData,
              ttmEPS: meta.ttmEPS,
              companyName: meta.companyName,
              source: meta.source,
            };
          }
        } catch {
          // Fall through and rethrow the provider error.
        }
      }
      throw err;
    }
  }

  // No cache — fetch directly from provider.
  return fetchFromProvider(symbol, days, useYahoo, source);
}
