import type { StockData } from './types';

export const SourceAuto = 'auto';
export const SourceMacrotrends = 'macrotrends';
export const SourceYahoo = 'yahoo';

export type DataSource =
  | typeof SourceAuto
  | typeof SourceMacrotrends
  | typeof SourceYahoo;

// Checks whether a symbol is a Hong Kong stock.
export function isHKStock(symbol: string): boolean {
  return symbol.toUpperCase().endsWith('.HK');
}

// Reverses the slice so newest data is first (the app's output convention).
export function reverseData(data: StockData[]): StockData[] {
  const result: StockData[] = [];
  for (let i = data.length - 1; i >= 0; i--) {
    result.push(data[i]);
  }
  return result;
}

// Reports whether a data source includes historical P/E data.
export function sourceHasPE(source: string): boolean {
  return source === SourceMacrotrends;
}

// Converts a user-supplied source value into a known source identifier.
export function normalizeSource(raw: string | null | undefined): DataSource {
  const value = (raw ?? '').trim().toLowerCase();
  switch (value) {
    case '':
    case SourceAuto:
      return SourceAuto;
    case SourceMacrotrends:
    case SourceYahoo:
      return value as DataSource;
    default:
      throw new Error(
        `invalid source: "${raw}" (use auto, macrotrends, or yahoo)`,
      );
  }
}

// Rejects sources that cannot serve a symbol. HK stocks only support Yahoo.
export function validateSourceForSymbol(source: DataSource, symbol: string): void {
  if (isHKStock(symbol) && source !== SourceAuto && source !== SourceYahoo) {
    throw new Error(
      `source "${source}" is not supported for HK stocks; use auto or yahoo`,
    );
  }
}

// Formats the macrotrends company slug for display.
export function formatCompanyName(slug: string): string {
  if (!slug) return '';
  const name = slug.replace(/-/g, ' ');
  return name
    .split(/\s+/)
    .filter((word) => word.length > 0)
    .map((word) => word[0].toUpperCase() + word.slice(1).toLowerCase())
    .join(' ');
}

// Parses a float and returns 0 on error.
export function toNumber(value: string | number | undefined | null): number {
  if (value === undefined || value === null || value === '') return 0;
  const n = typeof value === 'number' ? value : Number(value);
  return Number.isFinite(n) ? n : 0;
}

// Formats a number with two fixed decimals.
export function formatFloat(value: number): string {
  return value.toFixed(2);
}

// Formats an integer volume as K/M/B with two decimals.
export function formatVolume(value: number): string {
  if (value >= 1_000_000_000) return `${(value / 1_000_000_000).toFixed(2)}B`;
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(2)}M`;
  if (value >= 1_000) return `${(value / 1_000).toFixed(2)}K`;
  return String(value);
}

// Parses a human-readable volume string like "1.5M" or "500K".
export function parseVolume(value: string): number {
  const s = (value ?? '').trim();
  if (s === '') return 0;

  let multiplier = 1;
  let n = s;
  if (s.endsWith('B')) {
    multiplier = 1e9;
    n = s.slice(0, -1);
  } else if (s.endsWith('M')) {
    multiplier = 1e6;
    n = s.slice(0, -1);
  } else if (s.endsWith('K')) {
    multiplier = 1e3;
    n = s.slice(0, -1);
  }

  const parsed = Number(n);
  return (Number.isFinite(parsed) ? parsed : 0) * multiplier;
}

// Formats a float volume as K/M/B.
export function formatVolumeFloat(value: number): string {
  if (value >= 1e9) return `${(value / 1e9).toFixed(2)}B`;
  if (value >= 1e6) return `${(value / 1e6).toFixed(2)}M`;
  if (value >= 1e3) return `${(value / 1e3).toFixed(2)}K`;
  return value.toFixed(0);
}

// Formats a Date as YYYY-MM-DD using UTC (matches the Docker TZ=UTC default).
export function formatDate(date: Date): string {
  return date.toISOString().slice(0, 10);
}

// Returns a Date N days in the past.
export function daysAgo(days: number, from = new Date()): Date {
  return new Date(from.getTime() - days * 86_400_000);
}
