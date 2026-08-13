import { describe, expect, it } from 'vitest';
import {
  SourceAuto,
  SourceEDGAR,
  SourceMacrotrends,
  SourceYahoo,
  isHKStock,
  normalizeSource,
  reverseData,
  sourceHasPE,
  validateSourceForSymbol,
} from './helpers';

describe('isHKStock', () => {
  it('detects HK suffix case-insensitively', () => {
    expect(isHKStock('0700.HK')).toBe(true);
    expect(isHKStock('0700.hk')).toBe(true);
    expect(isHKStock('AAPL')).toBe(false);
  });
});

describe('reverseData', () => {
  it('reverses records', () => {
    const data = [{ date: '2024-01-01' }, { date: '2024-01-02' }].map((d) => ({
      date: d.date,
      open: '',
      high: '',
      low: '',
      close: '',
      volume: '',
      change: '',
      hchange: '',
    }));
    expect(reverseData(data).map((d) => d.date)).toEqual(['2024-01-02', '2024-01-01']);
  });
});

describe('normalizeSource', () => {
  it('normalizes valid sources', () => {
    expect(normalizeSource('')).toBe(SourceAuto);
    expect(normalizeSource('auto')).toBe(SourceAuto);
    expect(normalizeSource('AUTO')).toBe(SourceAuto);
    expect(normalizeSource('edgar')).toBe(SourceEDGAR);
    expect(normalizeSource('yahoo')).toBe(SourceYahoo);
    expect(normalizeSource('macrotrends')).toBe(SourceMacrotrends);
  });

  it('rejects unknown sources', () => {
    expect(() => normalizeSource('foo')).toThrow(/invalid source/);
  });
});

describe('sourceHasPE', () => {
  it('returns true for PE providers', () => {
    expect(sourceHasPE(SourceMacrotrends)).toBe(true);
    expect(sourceHasPE(SourceEDGAR)).toBe(true);
    expect(sourceHasPE(SourceYahoo)).toBe(false);
  });
});

describe('validateSourceForSymbol', () => {
  it('rejects non-Yahoo sources for HK stocks', () => {
    expect(() => validateSourceForSymbol(SourceEDGAR, '0700.HK')).toThrow(/not supported/);
    expect(() => validateSourceForSymbol(SourceYahoo, '0700.HK')).not.toThrow();
    expect(() => validateSourceForSymbol(SourceMacrotrends, 'AAPL')).not.toThrow();
  });
});
