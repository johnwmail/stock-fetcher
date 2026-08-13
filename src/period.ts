import type { DropCount, PeriodData, PeriodType, StockData } from './types';
import { formatVolumeFloat, toNumber, parseVolume } from './helpers';

// Valid period types accepted by the API.
const VALID_PERIODS = new Set<PeriodType>([
  'daily',
  'weekly',
  'monthly',
  'quarterly',
  'yearly',
]);

export function isValidPeriod(value: string): value is PeriodType {
  return VALID_PERIODS.has(value as PeriodType);
}

// Parses a string into a PeriodType. "daily" is allowed for validation but
// aggregation is only meaningful for weekly and longer periods.
export function parsePeriodType(value: string): PeriodType {
  switch (value.toLowerCase()) {
    case 'weekly':
    case 'week':
    case 'w':
      return 'weekly';
    case 'monthly':
    case 'month':
    case 'm':
      return 'monthly';
    case 'quarterly':
    case 'quarter':
    case 'q':
      return 'quarterly';
    case 'yearly':
    case 'year':
    case 'y':
      return 'yearly';
    case 'daily':
    case 'day':
    case 'd':
      return 'daily';
    default:
      throw new Error(
        `invalid period type: ${value} (use weekly, monthly, quarterly, or yearly)`,
      );
  }
}

// Returns the ISO-8601 week number for a date.
function isoWeek(date: Date): { year: number; week: number } {
  const d = new Date(Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), date.getUTCDate()));
  const dayNum = d.getUTCDay() || 7;
  d.setUTCDate(d.getUTCDate() + 4 - dayNum);
  const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
  const week = Math.ceil(((d.getTime() - yearStart.getTime()) / 86_400_000 + 1) / 7);
  return { year: d.getUTCFullYear(), week };
}

// Returns a unique, lexicographically sortable key for a period.
export function getPeriodKey(date: Date, periodType: PeriodType): string {
  switch (periodType) {
    case 'weekly': {
      const { year, week } = isoWeek(date);
      return `${year}-W${String(week).padStart(2, '0')}`;
    }
    case 'monthly':
      return formatDateKey(date.getUTCFullYear(), date.getUTCMonth() + 1);
    case 'quarterly': {
      const quarter = Math.floor(date.getUTCMonth() / 3) + 1;
      return `${date.getUTCFullYear()}-Q${quarter}`;
    }
    case 'yearly':
      return String(date.getUTCFullYear());
    case 'daily':
    default:
      return formatDateUTC(date);
  }
}

function formatDateUTC(date: Date): string {
  return date.toISOString().slice(0, 10);
}

function formatDateKey(year: number, month: number): string {
  return `${year}-${String(month).padStart(2, '0')}`;
}

// Classifies a percentage change into a drop bucket (2, 3, 4, 5) or 0.
function classifyDropPct(pctChange: number): number {
  if (pctChange >= 0) return 0;
  const absChange = -pctChange;
  if (absChange >= 5.0) return 5;
  if (absChange >= 4.0) return 4;
  if (absChange >= 3.0) return 3;
  if (absChange >= 2.0) return 2;
  return 0;
}

// Calculates both close-based and low-based drop buckets.
function calculateDrops(close: number, low: number, prevClose: number): [number, number] {
  if (prevClose <= 0) return [0, 0];
  const closePct = ((close - prevClose) / prevClose) * 100;
  const lowPct = ((low - prevClose) / prevClose) * 100;
  return [classifyDropPct(closePct), classifyDropPct(lowPct)];
}

// Aggregates daily data into weekly/monthly/quarterly/yearly periods.
// Input data must be sorted oldest-first.
export function aggregateToPeriods(data: StockData[], periodType: PeriodType): PeriodData[] {
  if (data.length === 0) return [];

  const groups = new Map<string, StockData[]>();
  const order: string[] = [];

  for (const d of data) {
    const date = new Date(d.date + 'T00:00:00Z');
    if (Number.isNaN(date.getTime())) continue;

    const key = getPeriodKey(date, periodType);
    if (!groups.has(key)) {
      groups.set(key, []);
      order.push(key);
    }
    groups.get(key)!.push(d);
  }

  order.sort();

  const result: PeriodData[] = [];
  let prevPeriodClose = 0;
  let prevPeriodHigh = 0;

  for (const key of order) {
    const days = groups.get(key)!;
    if (days.length === 0) continue;

    days.sort((a, b) => (a.date < b.date ? -1 : a.date > b.date ? 1 : 0));

    const firstDay = days[0];
    const lastDay = days[days.length - 1];

    let highVal = 0;
    let lowVal = 0;
    let totalVolume = 0;
    let drop2C = 0, drop3C = 0, drop4C = 0, drop5C = 0;
    let drop2L = 0, drop3L = 0, drop4L = 0, drop5L = 0;
    let dayPrevClose = 0;

    for (let i = 0; i < days.length; i++) {
      const d = days[i];
      const high = toNumber(d.high);
      const low = toNumber(d.low);
      const close = toNumber(d.close);
      const vol = parseVolume(d.volume);

      if (i === 0 || high > highVal) highVal = high;
      if (i === 0 || low < lowVal) lowVal = low;
      totalVolume += vol;

      if (dayPrevClose > 0) {
        const [closeDrop, lowDrop] = calculateDrops(close, low, dayPrevClose);
        if (closeDrop === 2) drop2C++;
        else if (closeDrop === 3) drop3C++;
        else if (closeDrop === 4) drop4C++;
        else if (closeDrop === 5) drop5C++;

        if (lowDrop === 2) drop2L++;
        else if (lowDrop === 3) drop3L++;
        else if (lowDrop === 4) drop4L++;
        else if (lowDrop === 5) drop5L++;
      }

      dayPrevClose = close;
    }

    const closeVal = toNumber(lastDay.close);
    let change = '';
    if (prevPeriodClose > 0) {
      change = `${(((closeVal - prevPeriodClose) / prevPeriodClose) * 100).toFixed(2)}%`;
    }

    let hchange = '';
    if (prevPeriodHigh > 0) {
      hchange = `${(((closeVal - prevPeriodHigh) / prevPeriodHigh) * 100).toFixed(2)}%`;
    }

    result.push({
      period: key,
      start_date: firstDay.date,
      end_date: lastDay.date,
      open: firstDay.open,
      high: highVal.toFixed(2),
      low: lowVal.toFixed(2),
      close: lastDay.close,
      volume: formatVolumeFloat(totalVolume),
      change,
      hchange,
      pe: lastDay.pe,
      days: days.length,
      drop_2pct: { close: drop2C, low: drop2L },
      drop_3pct: { close: drop3C, low: drop3L },
      drop_4pct: { close: drop4C, low: drop4L },
      drop_5pct: { close: drop5C, low: drop5L },
    });

    prevPeriodClose = closeVal;
    prevPeriodHigh = highVal;
  }

  // Reverse so newest is first, matching the daily output convention.
  result.reverse();
  return result;
}
