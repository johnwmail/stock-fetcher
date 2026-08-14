import { FundamentalData, type PERatioData } from './types';

const MACROTRENDS_USER_AGENT =
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36';

interface MacrotrendsSearchResult {
  n: string;
  s: string; // Format: "AAPL/apple"
}

interface DailyPriceData {
  d: string;
  o: string;
  h: string;
  l: string;
  c: string;
  v: string;
}

export class MacrotrendsFetcher {
  // Finds the macrotrends URL slug ("TICKER/company-slug") for a symbol.
  async getCompanySlug(symbol: string): Promise<string> {
    const searchURL = `https://www.macrotrends.net/production/stocks/desktop/ticker_search_list.php?q=${encodeURIComponent(symbol)}`;

    const resp = await fetch(searchURL, {
      headers: {
        'User-Agent': MACROTRENDS_USER_AGENT,
        Accept: 'application/json',
      },
    });

    if (!resp.ok) {
      throw new Error(`search returned status ${resp.status}`);
    }

    const results = (await resp.json()) as MacrotrendsSearchResult[];
    if (results.length === 0) {
      throw new Error(`no results found for symbol ${symbol}`);
    }

    for (const result of results) {
      const parts = result.s.split('/');
      if (parts.length === 2 && parts[0].toUpperCase() === symbol.toUpperCase()) {
        return result.s;
      }
    }

    throw new Error(
      `symbol ${symbol} not found on macrotrends (may be an ETF or unsupported stock)`,
    );
  }

  // Fetches quarterly P/E data for a symbol.
  async fetchPERatio(symbol: string): Promise<FundamentalData> {
    const slug = await this.getCompanySlug(symbol);
    const parts = slug.split('/');
    if (parts.length !== 2) {
      throw new Error(`invalid slug format: ${slug}`);
    }

    const ticker = parts[0];
    const companySlug = parts[1];
    const iframeURL =
      `https://www.macrotrends.net/production/stocks/desktop/fundamental_iframe.php` +
      `?t=${ticker}&type=pe-ratio&statement=price-ratios&freq=Q&sub=`;

    const resp = await fetch(iframeURL, {
      headers: {
        'User-Agent': MACROTRENDS_USER_AGENT,
        Referer: `https://www.macrotrends.net/stocks/charts/${ticker}/${companySlug}/pe-ratio`,
      },
    });

    if (!resp.ok) {
      throw new Error(`iframe returned status ${resp.status}`);
    }

    const body = await resp.text();
    const raw = extractJSONArray(body, 'var chartData = ');
    if (raw === null) {
      throw new Error('could not find chart data in response');
    }

    const peData = raw as Array<[string, number, number, number]>;
    if (peData.length === 0) {
      throw new Error('no P/E data found');
    }

    const historicalData: PERatioData[] = peData.map((row) => ({
      date: String(row[0]),
      stockPrice: Number(row[1]) || 0,
      eps: Number(row[2]) || 0,
      peRatio: Number(row[3]) || 0,
    }));

    const latest = historicalData[historicalData.length - 1];

    return new FundamentalData(
      ticker.toUpperCase(),
      companySlug,
      historicalData,
      latest.peRatio,
      latest.eps,
      latest.stockPrice,
    );
  }

  // Fetches daily OHLCV prices from macrotrends.
  async fetchDailyPrices(symbol: string, days: number): Promise<DailyPriceData[]> {
    const slug = await this.getCompanySlug(symbol);
    const parts = slug.split('/');
    if (parts.length !== 2) {
      throw new Error(`invalid slug format: ${slug}`);
    }

    const ticker = parts[0];
    const companySlug = parts[1];
    const iframeURL =
      `https://www.macrotrends.net/production/stocks/desktop/stock_price_history.php?t=${ticker}`;

    const resp = await fetch(iframeURL, {
      headers: {
        'User-Agent': MACROTRENDS_USER_AGENT,
        Referer: `https://www.macrotrends.net/stocks/charts/${ticker}/${companySlug}/stock-price-history`,
      },
    });

    if (!resp.ok) {
      throw new Error(`price history returned status ${resp.status}`);
    }

    const body = await resp.text();
    const raw = extractJSONArray(body, 'var dataDaily = ');
    if (raw === null) {
      throw new Error('could not find daily price data in response');
    }

    const allData = raw as DailyPriceData[];
    if (allData.length === 0) {
      throw new Error('no daily price data found');
    }

    if (days > 0 && days < allData.length) {
      return allData.slice(allData.length - days);
    }
    return allData;
  }
}

// Extracts the JSON array immediately following a marker like
// "var chartData = " or "var dataDaily = " from an HTML body.
export function extractJSONArray(body: string, marker: string): unknown | null {
  const startIdx = body.indexOf(marker);
  if (startIdx === -1) return null;

  const sub = body.slice(startIdx + marker.length);
  let bracketCount = 0;
  let endIdx = -1;

  for (let i = 0; i < sub.length; i++) {
    const c = sub[i];
    if (c === '[') {
      bracketCount++;
    } else if (c === ']') {
      bracketCount--;
      if (bracketCount === 0) {
        endIdx = i + 1;
        break;
      }
    }
  }

  if (endIdx === -1) return null;

  try {
    return JSON.parse(sub.slice(0, endIdx));
  } catch {
    return null;
  }
}
