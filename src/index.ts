import {
  SourceEDGAR,
  SourceMacrotrends,
  SourceYahoo,
  formatCompanyName,
  isHKStock,
  normalizeSource,
  reverseData,
  sourceHasPE,
  validateSourceForSymbol,
} from './helpers';
import type { DataSource } from './helpers';
import { isValidPeriod, parsePeriodType, aggregateToPeriods } from './period';
import { fetchStockData } from './providers';
import { setSECUserAgent } from './edgar';
import { generateExcel } from './excel';
import { getCompanyNamesForSymbols, getIndices } from './data';
import type { StockResponse } from './types';

export interface Env {
  DB?: D1Database;
  ASSETS?: Fetcher;
  SEC_USER_AGENT?: string;
  VERSION?: string;
}

const API_PREFIX = '/api/';

function json(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

function success(data: unknown): Response {
  return json({ success: true, data });
}

function error(message: string, status: number): Response {
  return json({ success: false, error: message }, status);
}

function withCORS(resp: Response): Response {
  resp.headers.set('Access-Control-Allow-Origin', '*');
  resp.headers.set('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
  resp.headers.set('Access-Control-Allow-Headers', 'Content-Type');
  return resp;
}

async function serveAssets(request: Request, env: Env): Promise<Response> {
  if (!env.ASSETS) {
    return error('Static assets binding is not configured', 404);
  }
  const assetResp = await env.ASSETS.fetch(request);
  const resp = new Response(assetResp.body, assetResp);
  return withCORS(resp);
}

function providerURL(source: string, symbol: string, companyName: string): string {
  const upper = symbol.toUpperCase();
  switch (source) {
    case SourceMacrotrends: {
      const slug = companyName || symbol.toLowerCase();
      return `https://www.macrotrends.net/stocks/charts/${upper}/${slug}/stock-price-history`;
    }
    case SourceEDGAR:
      return `https://www.sec.gov/cgi-bin/browse-edgar?action=getcompany&company=${upper}&type=10-K`;
    default:
      return `https://finance.yahoo.com/quote/${upper}`;
  }
}

async function handleHealth(env: Env): Promise<Response> {
  return success({
    status: 'ok',
    version: env.VERSION ?? 'vDev',
    commit: 'sha-unknown',
    build_time: 'timeless',
  });
}

async function handleStock(request: Request, env: Env): Promise<Response> {
  const url = new URL(request.url);
  const symbol = decodeURIComponent(url.pathname.slice('/api/stock/'.length)).replace(/\/+$/, '');
  if (!symbol) return error('Symbol is required', 400);

  const daysParam = url.searchParams.get('days');
  let days = 1825;
  if (daysParam) {
    const parsed = Number(daysParam);
    if (Number.isInteger(parsed) && parsed > 0) days = parsed;
  }

  const period = url.searchParams.get('period') || 'monthly';
  if (!isValidPeriod(period)) {
    return error('Invalid period. Use: daily, weekly, monthly, quarterly, yearly', 400);
  }

  let source: DataSource;
  try {
    source = normalizeSource(url.searchParams.get('source'));
    validateSourceForSymbol(source, symbol);
  } catch (err) {
    return error(err instanceof Error ? err.message : String(err), 400);
  }

  const useYahoo = isHKStock(symbol);
  setSECUserAgent(env.SEC_USER_AGENT);

  const result = await fetchStockData(env.DB ?? null, symbol, days, useYahoo, source);
  if (result.data.length === 0) {
    return error('No data found for symbol', 404);
  }

  const includePE = sourceHasPE(result.source);
  const resp: StockResponse = {
    symbol: symbol.toUpperCase(),
    company_name: formatCompanyName(result.companyName),
    data_source: result.source,
    provider_url: providerURL(result.source, symbol, result.companyName),
    currency: useYahoo ? 'HKD' : 'USD',
    period_type: period,
    record_count: 0,
  };

  if (includePE && result.ttmEPS > 0) {
    resp.ttm_eps = result.ttmEPS;
  }

  if (period !== 'daily') {
    const periodType = parsePeriodType(period);
    const oldestFirst = reverseData(result.data);
    const periodData = aggregateToPeriods(oldestFirst, periodType);
    resp.period_data = periodData;
    resp.record_count = periodData.length;
  } else {
    resp.daily_data = result.data;
    resp.record_count = result.data.length;
  }

  return success(resp);
}

async function handleStockExcel(request: Request, env: Env): Promise<Response> {
  const url = new URL(request.url);
  let symbol = decodeURIComponent(url.pathname.slice('/api/stock-excel/'.length)).replace(/\/+$/, '');
  if (!symbol) return error('Symbol is required', 400);
  symbol = symbol.toUpperCase();

  const daysParam = url.searchParams.get('days');
  let days = 1825;
  if (daysParam) {
    const parsed = Number(daysParam);
    if (Number.isInteger(parsed) && parsed > 0) days = parsed;
  }

  const period = url.searchParams.get('period') || 'monthly';
  if (!isValidPeriod(period)) {
    return error('Invalid period. Use: daily, weekly, monthly, quarterly, yearly', 400);
  }

  let source: DataSource;
  try {
    source = normalizeSource(url.searchParams.get('source'));
    validateSourceForSymbol(source, symbol);
  } catch (err) {
    return error(err instanceof Error ? err.message : String(err), 400);
  }

  const useYahoo = isHKStock(symbol);
  setSECUserAgent(env.SEC_USER_AGENT);

  const result = await fetchStockData(env.DB ?? null, symbol, days, useYahoo, source);
  const includePE = sourceHasPE(result.source);

  const workbook = await generateExcel({
    symbol,
    companyName: result.companyName,
    period,
    ttmEPS: result.ttmEPS,
    includePE,
    data: period === 'daily' ? result.data : undefined,
    periodData:
      period === 'daily'
        ? undefined
        : aggregateToPeriods(reverseData(result.data), parsePeriodType(period)),
  });

  return withCORS(
    new Response(workbook, {
      status: 200,
      headers: {
        'Content-Type': 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
        'Content-Disposition': `attachment; filename=${symbol}_${period}.xlsx`,
      },
    }),
  );
}

async function handleIndices(): Promise<Response> {
  const indices = getIndices();
  const result = Object.entries(indices).map(([key, idx]) => ({
    key,
    name: idx.name,
    description: idx.description,
    count: idx.symbols.length,
  }));
  return success(result);
}

async function handleIndexSymbols(request: Request): Promise<Response> {
  const url = new URL(request.url);
  const indexName = decodeURIComponent(url.pathname.slice('/api/indices/'.length))
    .replace(/\/+$/, '')
    .toLowerCase();
  if (!indexName) return error('Index name is required', 400);

  const indices = getIndices();
  const idx = indices[indexName];
  if (!idx) return error('Index not found', 404);

  return success({
    key: indexName,
    name: idx.name,
    description: idx.description,
    symbols: idx.symbols,
    companies: getCompanyNamesForSymbols(idx.symbols),
    count: idx.symbols.length,
  });
}

async function handleRequest(request: Request, env: Env): Promise<Response> {
  if (request.method === 'OPTIONS') {
    return withCORS(new Response(null, { status: 204 }));
  }

  const url = new URL(request.url);
  const path = url.pathname;

  let resp: Response;
  try {
    if (path === '/api/health') {
      resp = await handleHealth(env);
    } else if (path === '/api/indices') {
      resp = await handleIndices();
    } else if (path.startsWith('/api/indices/')) {
      resp = await handleIndexSymbols(request);
    } else if (path.startsWith('/api/stock-excel/')) {
      resp = await handleStockExcel(request, env);
    } else if (path.startsWith('/api/stock/')) {
      resp = await handleStock(request, env);
    } else if (path.startsWith(API_PREFIX)) {
      resp = error('Not found', 404);
    } else {
      return serveAssets(request, env);
    }
  } catch (err) {
    resp = error(
      err instanceof Error ? `Failed to fetch data: ${err.message}` : String(err),
      500,
    );
  }

  return withCORS(resp);
}

export default {
  async fetch(request: Request, env: Env, _ctx: ExecutionContext): Promise<Response> {
    return handleRequest(request, env);
  },
};
