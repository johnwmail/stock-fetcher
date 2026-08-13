import { describe, expect, it } from 'vitest';
import {
  buildTTMEPSSeries,
  daysBetween,
  deriveMissingQ4,
  selectAnnualEPS,
  selectQuarterlyEPS,
} from './edgar';
import type { EdgarEPSFact } from './edgar';

const q = (
  start: string,
  end: string,
  val: number,
  form = '10-Q',
  filed = '2024-05-01',
): EdgarEPSFact => ({ start, end, val, form, fp: 'Q1', filed });

describe('selectQuarterlyEPS', () => {
  it('keeps only single-quarter facts and prefers latest filed', () => {
    const facts = [
      q('2024-01-01', '2024-06-30', 5.0), // YTD, excluded
      q('2024-01-01', '2024-03-31', 2.0, '10-Q', '2024-05-01'),
      q('2024-01-01', '2024-03-31', 2.1, '10-Q', '2024-05-15'),
      q('2024-04-01', '2024-06-30', 3.0, '10-Q', '2024-08-01'),
    ];
    const result = selectQuarterlyEPS(facts);
    expect(result.map((f) => f.val)).toEqual([2.1, 3.0]);
  });
});

describe('selectAnnualEPS', () => {
  it('keeps annual 10-K facts', () => {
    const facts = [
      q('2024-01-01', '2024-12-31', 10.0, '10-K', '2025-02-01'),
      q('2024-01-01', '2024-03-31', 2.0, '10-K', '2024-05-01'),
    ];
    expect(selectAnnualEPS(facts).map((f) => f.val)).toEqual([10.0]);
  });
});

describe('deriveMissingQ4', () => {
  it('derives Q4 from annual minus first three quarters', () => {
    const quarters = [
      q('2022-01-01', '2022-03-31', 2.0),
      q('2022-04-01', '2022-06-30', 2.5),
      q('2022-07-01', '2022-09-30', 3.0),
    ];
    const annuals = [q('2022-01-01', '2022-12-31', 10.0, '10-K', '2023-02-15')];
    const result = deriveMissingQ4(quarters, annuals);
    const q4 = result.find((f) => f.fp === 'Q4');
    expect(q4?.val).toBe(2.5);
  });
});

describe('buildTTMEPSSeries', () => {
  it('builds trailing twelve month EPS', () => {
    const quarters = [
      q('2022-01-01', '2022-03-31', 2.0),
      q('2022-04-01', '2022-06-30', 2.5),
      q('2022-07-01', '2022-09-30', 3.0),
      q('2022-10-01', '2022-12-31', 2.5),
    ];
    const series = buildTTMEPSSeries(quarters);
    expect(series[0].eps).toBe(0);
    expect(series[3].eps).toBe(10.0);
  });
});

describe('daysBetween', () => {
  it('calculates day differences', () => {
    expect(daysBetween('2024-01-01', '2024-03-31')).toBe(90);
    expect(daysBetween('bad', '2024-01-01')).toBe(0);
  });
});
