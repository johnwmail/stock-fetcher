import { describe, expect, it } from 'vitest';
import { aggregateToPeriods, getPeriodKey, parsePeriodType } from './period';
import type { StockData } from './types';

function record(date: string, close: number, low: number): StockData {
  return {
    date,
    open: String(close),
    high: String(close),
    low: String(low),
    close: String(close),
    volume: '1M',
    change: '',
    hchange: '',
  };
}

describe('parsePeriodType', () => {
  it('parses aliases', () => {
    expect(parsePeriodType('m')).toBe('monthly');
    expect(parsePeriodType('year')).toBe('yearly');
  });
});

describe('aggregateToPeriods', () => {
  it('aggregates daily data into monthly buckets newest-first', () => {
    const data = [
      record('2024-01-01', 100, 95),
      record('2024-01-02', 102, 98),
      record('2024-02-01', 101, 97),
    ];
    const result = aggregateToPeriods(data, 'monthly');
    expect(result.map((p) => p.period)).toEqual(['2024-02', '2024-01']);
    expect(result[1].close).toBe('102');
    expect(result[1].days).toBe(2);
  });

  it('classifies drops', () => {
    const data = [
      record('2024-01-01', 100, 100),
      record('2024-01-02', 95, 94), // -5% close, -6% low
    ];
    const result = aggregateToPeriods(data, 'monthly');
    expect(result[0].drop_5pct).toEqual({ close: 1, low: 1 });
  });
});

describe('getPeriodKey', () => {
  it('formats weekly keys with zero padding', () => {
    const date = new Date('2024-01-01T00:00:00Z');
    expect(getPeriodKey(date, 'weekly')).toMatch(/^2024-W\d{2}$/);
  });
});
