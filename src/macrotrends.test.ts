import { describe, expect, it } from 'vitest';
import { extractJSONArray } from './macrotrends';

describe('extractJSONArray', () => {
  it('extracts the JSON array after a marker', () => {
    const body = `var chartData = [{"date":"2024-01-01","v1":1,"v2":2,"v3":3}];`;
    expect(extractJSONArray(body, 'var chartData = ')).toEqual([
      { date: '2024-01-01', v1: 1, v2: 2, v3: 3 },
    ]);
  });

  it('returns null when marker is missing', () => {
    expect(extractJSONArray('nothing here', 'var chartData = ')).toBeNull();
  });
});
