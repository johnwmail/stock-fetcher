import { FundamentalData, type PERatioData } from './types';

const SEC_COMPANY_TICKERS_URL = 'https://www.sec.gov/files/company_tickers.json';
const SEC_DATA_BASE_URL = 'https://data.sec.gov';

// SEC requires a declared User-Agent with a contact. A URL in the User-Agent
// is rejected, so keep it to name + contact. Override via SEC_USER_AGENT.
const DEFAULT_SEC_USER_AGENT = 'stock-fetcher/1.0 (admin@example.com)';

// SEC fair-access limit is 10 requests/second. Serializing at 5/second keeps
// comfortably below the threshold.
const SEC_MIN_REQUEST_INTERVAL_MS = 200;

export interface EdgarEPSFact {
  start: string;
  end: string;
  val: number;
  form: string;
  fp: string;
  filed: string;
  frame?: string;
}

interface EdgarCompanyFacts {
  entityName: string;
  facts: {
    'us-gaap': {
      EarningsPerShareDiluted: {
        units: Record<string, EdgarEPSFact[]>;
      };
    };
  };
}

// Per-isolate cache for the ~800KB SEC ticker map.
let tickersCache: Record<string, string> | null = null;
let tickersInFlight: Promise<Record<string, string>> | null = null;

let lastRequestTime = 0;
let requestQueue: Promise<unknown> = Promise.resolve();

let configuredUserAgent = DEFAULT_SEC_USER_AGENT;

// Allows the request handler to override the SEC User-Agent from env.
export function setSECUserAgent(value: string | undefined): void {
  configuredUserAgent = value || DEFAULT_SEC_USER_AGENT;
}

function userAgent(): string {
  return configuredUserAgent;
}

// Performs a serialized, rate-limited SEC request.
async function secFetch(url: string): Promise<Response> {
  const run = async (): Promise<Response> => {
    const now = Date.now();
    const wait = Math.max(0, SEC_MIN_REQUEST_INTERVAL_MS - (now - lastRequestTime));
    if (wait > 0) {
      await new Promise((resolve) => setTimeout(resolve, wait));
    }
    lastRequestTime = Date.now();
    return fetch(url, {
      headers: {
        'User-Agent': userAgent(),
        Accept: 'application/json',
      },
    });
  };

  const result = requestQueue.then(run, run);
  requestQueue = result.then(
    () => undefined,
    () => undefined,
  );
  return result;
}

async function readJSON<T>(resp: Response, context: string): Promise<T> {
  if (!resp.ok) {
    const body = (await resp.text()).slice(0, 512);
    throw new Error(`SEC returned status ${resp.status}: ${body}`);
  }
  try {
    return (await resp.json()) as T;
  } catch (err) {
    throw new Error(`failed to parse SEC response (${context}): ${err}`);
  }
}

// Resolves a ticker to its zero-padded 10-digit CIK.
export async function getCIK(symbol: string): Promise<string> {
  const upper = symbol.toUpperCase();

  if (tickersCache?.[upper]) return tickersCache[upper];

  if (!tickersInFlight) {
    tickersInFlight = (async () => {
      const resp = await secFetch(SEC_COMPANY_TICKERS_URL);
      const data = await readJSON<Record<string, { cik_str: number; ticker: string }>>(
        resp,
        'ticker map',
      );

      const map: Record<string, string> = {};
      for (const value of Object.values(data)) {
        if (value.ticker) {
          map[value.ticker.toUpperCase()] = String(value.cik_str).padStart(10, '0');
        }
      }
      tickersCache = map;
      return map;
    })().finally(() => {
      tickersInFlight = null;
    });
  }

  const map = await tickersInFlight;
  const cik = map[upper];
  if (!cik) {
    throw new Error(`symbol ${symbol} not found in SEC ticker map`);
  }
  return cik;
}

// Fetches historical TTM EPS from EDGAR in the same FundamentalData shape
// used by the macrotrends fetcher.
export async function fetchEDGARFundamental(symbol: string): Promise<FundamentalData> {
  const cik = await getCIK(symbol);
  const factsURL = `${SEC_DATA_BASE_URL}/api/xbrl/companyfacts/CIK${cik}.json`;

  const resp = await secFetch(factsURL);
  const facts = await readJSON<EdgarCompanyFacts>(resp, 'company facts');

  const rawFacts = facts.facts['us-gaap'].EarningsPerShareDiluted.units['USD/shares'] ?? [];
  const quarters = selectQuarterlyEPS(rawFacts);
  if (quarters.length === 0) {
    throw new Error(`no quarterly diluted EPS data found for ${symbol}`);
  }

  const withQ4 = deriveMissingQ4(quarters, selectAnnualEPS(rawFacts));
  const history = buildTTMEPSSeries(withQ4);

  return new FundamentalData(
    symbol.toUpperCase(),
    facts.entityName,
    history,
    0,
    latestPositiveEPS(history),
    0,
  );
}

// Selects one diluted EPS value per fiscal quarter from raw XBRL facts.
export function selectQuarterlyEPS(facts: EdgarEPSFact[]): EdgarEPSFact[] {
  const best = new Map<string, EdgarEPSFact>();

  for (const fact of facts) {
    if (!fact.start || !fact.end) continue;
    if (fact.form !== '10-Q' && fact.form !== '10-K' && fact.form !== '10-K/A') continue;

    const duration = daysBetween(fact.start, fact.end);
    if (duration <= 0 || duration >= 120) continue;

    const current = best.get(fact.end);
    if (
      !current ||
      fact.filed > current.filed ||
      (fact.filed === current.filed && duration < daysBetween(current.start, current.end))
    ) {
      best.set(fact.end, fact);
    }
  }

  return [...best.values()].sort((a, b) => (a.end < b.end ? -1 : a.end > b.end ? 1 : 0));
}

// Selects one diluted EPS value per fiscal year from 10-K/10-K/A facts.
export function selectAnnualEPS(facts: EdgarEPSFact[]): EdgarEPSFact[] {
  const best = new Map<string, EdgarEPSFact>();

  for (const fact of facts) {
    if (!fact.start || !fact.end) continue;
    if (fact.form !== '10-K' && fact.form !== '10-K/A') continue;

    const duration = daysBetween(fact.start, fact.end);
    if (duration < 300) continue;

    const current = best.get(fact.end);
    if (!current || fact.filed > current.filed) {
      best.set(fact.end, fact);
    }
  }

  return [...best.values()].sort((a, b) => (a.end < b.end ? -1 : a.end > b.end ? 1 : 0));
}

// Fills in the fourth fiscal quarter for years where only annual EPS was
// tagged in the 10-K: Q4 = annual EPS - the three available quarters.
export function deriveMissingQ4(quarters: EdgarEPSFact[], annuals: EdgarEPSFact[]): EdgarEPSFact[] {
  const covered = new Set(quarters.map((q) => q.end));
  const result = [...quarters];

  for (const annual of annuals) {
    if (covered.has(annual.end)) continue;

    const candidates = quarters.filter((q) => {
      const daysBefore = daysBetween(q.end, annual.end);
      return daysBefore > 0 && daysBefore < 320;
    });

    if (candidates.length < 3) continue;

    const last3 = candidates.slice(-3);
    const q4Value = annual.val - last3[0].val - last3[1].val - last3[2].val;

    result.push({
      start: last3[2].end,
      end: annual.end,
      val: q4Value,
      form: annual.form,
      fp: 'Q4',
      filed: annual.filed,
    });
  }

  return result.sort((a, b) => (a.end < b.end ? -1 : a.end > b.end ? 1 : 0));
}

// Builds trailing-twelve-month EPS data points from quarterly EPS facts.
export function buildTTMEPSSeries(quarters: EdgarEPSFact[]): PERatioData[] {
  const series: PERatioData[] = [];

  for (let i = 0; i < quarters.length; i++) {
    const quarter = quarters[i];
    let eps = 0;
    if (i >= 3) {
      eps = quarters[i - 3].val + quarters[i - 2].val + quarters[i - 1].val + quarter.val;
    }

    series.push({
      date: quarter.end,
      stockPrice: 0,
      eps,
      peRatio: 0,
    });
  }

  return series;
}

// Returns the number of days between two YYYY-MM-DD dates, or 0 on error.
export function daysBetween(start: string, end: string): number {
  const startMs = Date.parse(start + 'T00:00:00Z');
  const endMs = Date.parse(end + 'T00:00:00Z');
  if (Number.isNaN(startMs) || Number.isNaN(endMs)) return 0;
  return Math.round((endMs - startMs) / 86_400_000);
}

// Returns the most recent positive EPS value, or 0 if none exists.
export function latestPositiveEPS(history: PERatioData[]): number {
  for (let i = history.length - 1; i >= 0; i--) {
    if (history[i].eps > 0) return history[i].eps;
  }
  return 0;
}
