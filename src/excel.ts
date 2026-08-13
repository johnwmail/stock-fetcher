import ExcelJS from 'exceljs';
import type { PeriodData, StockData } from './types';

export interface ExcelParams {
  symbol: string;
  companyName: string;
  period: string;
  ttmEPS: number;
  includePE: boolean;
  data?: StockData[];
  periodData?: PeriodData[];
}

const HEADER_STYLE: Partial<ExcelJS.Style> = {
  font: { bold: true, color: { argb: 'FFFFFFFF' } },
  fill: { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FF4472C4' } },
  alignment: { horizontal: 'right' },
};

const RIGHT_ALIGN_STYLE: Partial<ExcelJS.Style> = {
  alignment: { horizontal: 'right' },
};

const NUMBER_STYLE: Partial<ExcelJS.Style> = {
  numFmt: '#,##0.00',
  alignment: { horizontal: 'right' },
};

const NEG_STYLE: Partial<ExcelJS.Style> = {
  fill: { type: 'pattern', pattern: 'solid', fgColor: { argb: 'FFFFC7CE' } },
  alignment: { horizontal: 'right' },
};

// Generates the same Excel workbook as the Go excelize implementation.
export async function generateExcel(params: ExcelParams): Promise<Uint8Array> {
  const workbook = new ExcelJS.Workbook();
  const sheet = workbook.addWorksheet('Stock Data');

  setCell(sheet, 1, 1, 'Symbol:', RIGHT_ALIGN_STYLE);
  setCell(sheet, 2, 1, params.symbol, RIGHT_ALIGN_STYLE);
  setCell(sheet, 1, 2, 'Company:', RIGHT_ALIGN_STYLE);
  setCell(sheet, 2, 2, params.companyName, RIGHT_ALIGN_STYLE);
  setCell(sheet, 1, 3, 'Period:', RIGHT_ALIGN_STYLE);
  setCell(sheet, 2, 3, params.period, RIGHT_ALIGN_STYLE);
  if (params.includePE) {
    setCell(sheet, 1, 4, 'TTM EPS:', RIGHT_ALIGN_STYLE);
    setCell(sheet, 2, 4, params.ttmEPS, RIGHT_ALIGN_STYLE);
  }

  let row = 6;

  if (params.periodData) {
    row = writePeriodData(sheet, row, params.periodData, params.includePE);
  } else {
    row = writeDailyData(sheet, row, params.data ?? [], params.includePE);
  }

  // Auto-fit columns (same fixed width as the Go implementation).
  void row;
  for (let col = 1; col <= 16; col++) {
    sheet.getColumn(col).width = 12;
  }

  const buffer = await workbook.xlsx.writeBuffer();
  return new Uint8Array(buffer);
}

function writeDailyData(
  sheet: ExcelJS.Worksheet,
  startRow: number,
  data: StockData[],
  includePE: boolean,
): number {
  const headers = ['Date', 'Open', 'High', 'Low', 'Close', 'Volume', 'Change', 'HChange'];
  if (includePE) headers.push('PE');

  headers.forEach((h, i) => setCell(sheet, i + 1, startRow, h, HEADER_STYLE));
  let row = startRow + 1;

  for (const d of data) {
    setCell(sheet, 1, row, d.date, RIGHT_ALIGN_STYLE);
    setNumberCell(sheet, 2, row, d.open);
    setNumberCell(sheet, 3, row, d.high);
    setNumberCell(sheet, 4, row, d.low);
    setNumberCell(sheet, 5, row, d.close);
    setCell(sheet, 6, row, d.volume, RIGHT_ALIGN_STYLE);
    setChangeCell(sheet, 7, row, d.change);
    setChangeCell(sheet, 8, row, d.hchange);
    if (includePE) setCell(sheet, 9, row, d.pe ?? '', RIGHT_ALIGN_STYLE);
    row++;
  }

  return row;
}

function writePeriodData(
  sheet: ExcelJS.Worksheet,
  startRow: number,
  data: PeriodData[],
  includePE: boolean,
): number {
  const headers = [
    'Period', 'Start', 'End', 'Open', 'High', 'Low', 'Close',
    'Volume', 'Change', 'HChange',
  ];
  if (includePE) headers.push('PE');
  headers.push('Days', 'C/L-2%', 'C/L-3%', 'C/L-4%', 'C/L-5%');

  headers.forEach((h, i) => setCell(sheet, i + 1, startRow, h, HEADER_STYLE));
  let row = startRow + 1;

  for (const p of data) {
    let col = 1;
    setCell(sheet, col++, row, p.period, RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, p.start_date, RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, p.end_date, RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, toNumberOrZero(p.open), RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, toNumberOrZero(p.high), RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, toNumberOrZero(p.low), RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, toNumberOrZero(p.close), RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, p.volume, RIGHT_ALIGN_STYLE);
    setChangeCell(sheet, col++, row, p.change);
    setChangeCell(sheet, col++, row, p.hchange);
    if (includePE) setCell(sheet, col++, row, p.pe ?? '', RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, p.days, RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, `${p.drop_2pct.close}/${p.drop_2pct.low}`, RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, `${p.drop_3pct.close}/${p.drop_3pct.low}`, RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, `${p.drop_4pct.close}/${p.drop_4pct.low}`, RIGHT_ALIGN_STYLE);
    setCell(sheet, col++, row, `${p.drop_5pct.close}/${p.drop_5pct.low}`, RIGHT_ALIGN_STYLE);
    row++;
  }

  return row;
}

function setCell(
  sheet: ExcelJS.Worksheet,
  col: number,
  row: number,
  value: ExcelJS.CellValue,
  style: Partial<ExcelJS.Style>,
): void {
  const cell = sheet.getCell(row, col);
  cell.value = value;
  cell.style = style as ExcelJS.Style;
}

function setNumberCell(sheet: ExcelJS.Worksheet, col: number, row: number, value: string): void {
  setCell(sheet, col, row, toNumberOrZero(value), NUMBER_STYLE);
}

function setChangeCell(sheet: ExcelJS.Worksheet, col: number, row: number, value: string): void {
  setCell(sheet, col, row, value, isNegativePct(value) ? NEG_STYLE : RIGHT_ALIGN_STYLE);
}

function isNegativePct(value: string): boolean {
  const n = Number((value ?? '').trim().replace(/%$/, ''));
  return Number.isFinite(n) && n < 0;
}

function toNumberOrZero(value: string): number {
  const n = Number(value);
  return Number.isFinite(n) ? n : 0;
}
