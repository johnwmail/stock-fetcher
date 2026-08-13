// Shared domain types for the Stock Fetcher Workers backend.
// These mirror the JSON shapes produced by the Go implementation so the
// existing web frontend keeps working unchanged.

export interface StockData {
  date: string;
  open: string;
  high: string;
  low: string;
  close: string;
  volume: string;
  change: string;
  hchange: string;
  pe?: string;
}

export interface PERatioData {
  date: string;
  stockPrice: number;
  eps: number;
  peRatio: number;
}

export class FundamentalData {
  constructor(
    public symbol: string,
    public companyName: string,
    public historicalData: PERatioData[] = [],
    public currentPE = 0,
    public currentEPS = 0,
    public currentPrice = 0,
  ) {}

  // Returns the latest trailing twelve months EPS.
  // Macrotrends EPS values are already TTM. EDGAR values are converted to
  // TTM by buildTTMEPSSeries() before being stored here.
  getLatestTTMEPS(): number {
    for (let i = this.historicalData.length - 1; i >= 0; i--) {
      if (this.historicalData[i].eps > 0) {
        return this.historicalData[i].eps;
      }
    }
    return 0;
  }

  // Returns the TTM EPS that was valid on or before the given YYYY-MM-DD date.
  getEPSForDate(date: string): number {
    let eps = 0;
    for (const d of this.historicalData) {
      if (d.date <= date && d.eps > 0) {
        eps = d.eps;
      }
      if (d.date > date) {
        break;
      }
    }
    return eps;
  }
}

export interface DropCount {
  close: number;
  low: number;
}

export interface PeriodData {
  period: string;
  start_date: string;
  end_date: string;
  open: string;
  high: string;
  low: string;
  close: string;
  volume: string;
  change: string;
  hchange: string;
  pe?: string;
  days: number;
  drop_2pct: DropCount;
  drop_3pct: DropCount;
  drop_4pct: DropCount;
  drop_5pct: DropCount;
}

export type PeriodType = 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'yearly';

export interface FetchMeta {
  symbol: string;
  source: string;
  companyName: string;
  ttmEPS: number;
  lastFetched: Date;
  latestDate: string;
  earliestDate: string;
}

export interface StockResponse {
  symbol: string;
  company_name: string;
  data_source: string;
  provider_url: string;
  currency: string;
  ttm_eps?: number;
  period_type: string;
  record_count: number;
  daily_data?: StockData[];
  period_data?: PeriodData[];
}

export interface ProviderResult {
  data: StockData[];
  ttmEPS: number;
  companyName: string;
  source: string;
}
