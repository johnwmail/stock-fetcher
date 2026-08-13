import type { StockData } from './types';
import { formatFloat, formatVolume } from './helpers';

const YAHOO_USER_AGENT =
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36';

interface YahooQuote {
  open: (number | null)[];
  high: (number | null)[];
  low: (number | null)[];
  close: (number | null)[];
  volume: (number | null)[];
}

export interface YahooChartResponse {
  chart: {
    result?: Array<{
      meta?: {
        longName?: string;
        shortName?: string;
        symbol?: string;
      };
      timestamp?: number[];
      indicators?: {
        quote?: YahooQuote[];
        adjclose?: Array<{ adjclose?: (number | null)[] }>;
      };
    }>;
    error?: { code?: string; description?: string } | null;
  };
}

// Fetches a company name from Yahoo Finance.
export async function fetchCompanyName(symbol: string): Promise<string> {
  const url = `https://query1.finance.yahoo.com/v8/finance/chart/${encodeURIComponent(symbol.toUpperCase())}?interval=1d&range=1d`;
  const resp = await fetch(url, {
    headers: {
      'User-Agent': YAHOO_USER_AGENT,
      Accept: 'application/json',
    },
  });

  const chartResp = (await resp.json()) as YahooChartResponse;
  const result = chartResp.chart.result?.[0];
  if (!result) return '';

  return result.meta?.longName || result.meta?.shortName || '';
}

// Fetches historical OHLCV data from the Yahoo chart API.
// Returns data in chronological order (oldest first).
export async function fetchHistoricalData(
  symbol: string,
  startDate: Date,
  endDate: Date,
): Promise<{ data: StockData[]; companyName: string }> {
  const period1 = Math.floor(startDate.getTime() / 1000);
  const period2 = Math.floor(endDate.getTime() / 1000);

  const url =
    `https://query1.finance.yahoo.com/v8/finance/chart/${encodeURIComponent(symbol.toUpperCase())}` +
    `?period1=${period1}&period2=${period2}&interval=1d&includePrePost=false`;

  const resp = await fetch(url, {
    headers: {
      'User-Agent': YAHOO_USER_AGENT,
      Accept: 'application/json',
    },
  });

  if (!resp.ok) {
    const body = (await resp.text()).slice(0, 500);
    throw new Error(`API returned status ${resp.status}: ${body}`);
  }

  const chartResp = (await resp.json()) as YahooChartResponse;

  if (chartResp.chart.error) {
    throw new Error(
      `API error: ${chartResp.chart.error.code} - ${chartResp.chart.error.description}`,
    );
  }

  const result = chartResp.chart.result?.[0];
  if (!result) {
    throw new Error(`no data returned for symbol ${symbol}`);
  }

  const companyName = result.meta?.longName || result.meta?.shortName || '';
  const data = parseYahooChartData(chartResp);
  return { data, companyName };
}

// Converts a Yahoo chart response into StockData records.
export function parseYahooChartData(resp: YahooChartResponse): StockData[] {
  const result = resp.chart.result?.[0];
  if (!result) return [];

  const timestamps = result.timestamp ?? [];
  const quote = result.indicators?.quote?.[0];
  if (!quote) throw new Error('no quote data in response');

  const data: StockData[] = [];
  let prevClose = 0;
  let prevHigh = 0;

  for (let i = 0; i < timestamps.length; i++) {
    if (i >= (quote.close?.length ?? 0)) break;

    const close = quote.close[i];
    if (close === null || close === undefined || close === 0) continue;

    const date = new Date(timestamps[i] * 1000).toISOString().slice(0, 10);

    const openVal = quote.open?.[i] ?? 0;
    const highVal = quote.high?.[i] ?? 0;
    const lowVal = quote.low?.[i] ?? 0;
    const volume = quote.volume?.[i] ?? 0;

    let change = '';
    if (prevClose > 0) {
      change = `${(((close - prevClose) / prevClose) * 100).toFixed(2)}%`;
    }

    let hchange = '';
    if (prevHigh > 0) {
      hchange = `${(((close - prevHigh) / prevHigh) * 100).toFixed(2)}%`;
    }

    data.push({
      date,
      open: formatFloat(openVal ?? 0),
      high: formatFloat(highVal ?? 0),
      low: formatFloat(lowVal ?? 0),
      close: formatFloat(close),
      volume: formatVolume(volume ?? 0),
      change,
      hchange,
    });

    prevClose = close;
    prevHigh = highVal ?? 0;
  }

  return data;
}
