import staticData from './data/static-data.json';

export const CompanyNames: Record<string, string> = staticData.companies;

export interface IndexInfo {
  name: string;
  description: string;
  symbols: string[];
}

export const Indices: Record<string, IndexInfo> = staticData.indices;

export function getIndices(): Record<string, IndexInfo> {
  return Indices;
}

export function getCompanyNamesForSymbols(symbols: string[]): Record<string, string> {
  const out: Record<string, string> = {};
  for (const symbol of symbols) {
    if (CompanyNames[symbol]) out[symbol] = CompanyNames[symbol];
  }
  return out;
}
