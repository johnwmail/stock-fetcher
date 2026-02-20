// Global state
let currentData = null;
let priceChart = null;

// API base URL (empty for same origin)
const API_BASE = '';

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    loadIndices();
    
    // Enter key to fetch
    document.getElementById('symbol').addEventListener('keypress', (e) => {
        if (e.key === 'Enter') fetchStock();
    });
});

// Fetch stock data
async function fetchStock() {
    const symbol = document.getElementById('symbol').value.trim().toUpperCase();
    const days = document.getElementById('days').value || 1825;
    const period = document.getElementById('period').value;

    if (!symbol) {
        showError('Please enter a stock symbol');
        return;
    }

    showLoading(true);
    hideError();
    hideResults();

    try {
        const response = await fetch(`${API_BASE}/api/stock/${symbol}?days=${days}&period=${period}`);
        const result = await response.json();

        if (!result.success) {
            throw new Error(result.error || 'Failed to fetch data');
        }

        currentData = result.data;
        displayResults(result.data);
    } catch (error) {
        showError(error.message);
    } finally {
        showLoading(false);
    }
}

// Quick fetch helper
function quickFetch(symbol) {
    document.getElementById('symbol').value = symbol;
    fetchStock();
}

// Display results
function displayResults(data) {
    // Update info
    document.getElementById('companyName').textContent = data.company_name || data.symbol;
    document.getElementById('stockSymbol').textContent = data.symbol;
    document.getElementById('dataSource').textContent = data.data_source;
    document.getElementById('recordCount').textContent = data.record_count;
    
    // TTM EPS
    const epsContainer = document.getElementById('epsContainer');
    if (data.ttm_eps > 0) {
        document.getElementById('ttmEps').textContent = `$${data.ttm_eps.toFixed(2)}`;
        epsContainer.classList.remove('hidden');
    } else {
        epsContainer.classList.add('hidden');
    }

    // Determine if daily or period data
    const isDaily = data.period_type === 'daily';
    const records = isDaily ? data.daily_data : data.period_data;

    // Build table
    buildTable(records, isDaily, data.ttm_eps > 0);

    // Build chart
    buildChart(records, isDaily, data.symbol);

    // Show results
    document.getElementById('results').classList.remove('hidden');
}

// Build data table
function buildTable(records, isDaily, hasPE) {
    const headerRow = document.getElementById('tableHeader');
    const tableBody = document.getElementById('tableBody');

    // Clear existing
    headerRow.innerHTML = '';
    tableBody.innerHTML = '';

    if (!records || records.length === 0) return;

    // Define columns based on data type
    let columns;
    if (isDaily) {
        columns = [
            { key: 'date', label: 'Date' },
            { key: 'open', label: 'Open', align: 'right' },
            { key: 'high', label: 'High', align: 'right' },
            { key: 'low', label: 'Low', align: 'right' },
            { key: 'close', label: 'Close', align: 'right' },
            { key: 'volume', label: 'Volume', align: 'right' },
            { key: 'change', label: 'Change', align: 'right' },
            { key: 'hchange', label: 'HChange', align: 'right' },
        ];
        if (hasPE) columns.push({ key: 'pe', label: 'P/E', align: 'right' });
    } else {
        columns = [
            { key: 'period', label: 'Period' },
            { key: 'start_date', label: 'Start' },
            { key: 'end_date', label: 'End' },
            { key: 'open', label: 'Open', align: 'right' },
            { key: 'high', label: 'High', align: 'right' },
            { key: 'low', label: 'Low', align: 'right' },
            { key: 'close', label: 'Close', align: 'right' },
            { key: 'volume', label: 'Volume', align: 'right' },
            { key: 'change', label: 'Change', align: 'right' },
            { key: 'hchange', label: 'HChange', align: 'right' },
            { key: 'days', label: 'Days', align: 'right' },
        ];
        if (hasPE) columns.push({ key: 'pe', label: 'P/E', align: 'right' });
        // Add drop columns
        columns.push(
            { key: 'drop_2pct', label: 'C/L-2%', align: 'right', format: 'drop' },
            { key: 'drop_3pct', label: 'C/L-3%', align: 'right', format: 'drop' },
            { key: 'drop_4pct', label: 'C/L-4%', align: 'right', format: 'drop' },
            { key: 'drop_5pct', label: 'C/L-5%', align: 'right', format: 'drop' }
        );
    }

    // Build header
    columns.forEach(col => {
        const th = document.createElement('th');
        th.className = `px-4 py-3 ${col.align === 'right' ? 'text-right' : 'text-left'}`;
        th.textContent = col.label;
        headerRow.appendChild(th);
    });

    // Build rows
    records.forEach(record => {
        const tr = document.createElement('tr');
        tr.className = 'hover:bg-gray-700/50';

        columns.forEach(col => {
            const td = document.createElement('td');
            td.className = `px-4 py-2 ${col.align === 'right' ? 'text-right' : 'text-left'}`;
            
            let value = record[col.key];
            
            // Format drop counts (C/L format)
            if (col.format === 'drop' && value) {
                value = `${value.close}/${value.low}`;
            }
            
            // Color code change values
            if ((col.key === 'change' || col.key === 'hchange') && value) {
                const numVal = parseFloat(value);
                if (numVal > 0) td.className += ' text-green-400';
                else if (numVal < 0) td.className += ' text-red-400';
            }
            
            td.textContent = value || '-';
            tr.appendChild(td);
        });

        tableBody.appendChild(tr);
    });
}

// Build price chart
function buildChart(records, isDaily, symbol) {
    const ctx = document.getElementById('priceChart').getContext('2d');

    // Destroy existing chart
    if (priceChart) {
        priceChart.destroy();
    }

    // Prepare data (reverse to chronological order)
    const chartData = [...records].reverse();
    const labels = chartData.map(r => isDaily ? r.date : r.period);
    const closes = chartData.map(r => parseFloat(r.close));
    const highs = chartData.map(r => parseFloat(r.high));
    const lows = chartData.map(r => parseFloat(r.low));

    // Check if P/E data is available
    const hasPE = chartData.some(r => r.pe && r.pe !== '' && parseFloat(r.pe) > 0);
    const peValues = hasPE ? chartData.map(r => {
        const v = parseFloat(r.pe);
        return v > 0 ? v : null;
    }) : [];

    // Compute EPS from close/pe where available
    const epsValues = hasPE ? chartData.map(r => {
        const close = parseFloat(r.close);
        const pe = parseFloat(r.pe);
        if (pe > 0 && close > 0) return parseFloat((close / pe).toFixed(2));
        return null;
    }) : [];

    const datasets = [
        {
            label: 'Close',
            data: closes,
            borderColor: 'rgb(59, 130, 246)',
            backgroundColor: 'rgba(59, 130, 246, 0.1)',
            fill: true,
            tension: 0.1,
            pointRadius: 0,
            borderWidth: 2,
            yAxisID: 'yPrice',
        },
        {
            label: 'High',
            data: highs,
            borderColor: 'rgba(34, 197, 94, 0.5)',
            borderWidth: 1,
            pointRadius: 0,
            borderDash: [5, 5],
            yAxisID: 'yPrice',
        },
        {
            label: 'Low',
            data: lows,
            borderColor: 'rgba(239, 68, 68, 0.5)',
            borderWidth: 1,
            pointRadius: 0,
            borderDash: [5, 5],
            yAxisID: 'yPrice',
        }
    ];

    if (hasPE) {
        datasets.push({
            label: 'P/E',
            data: peValues,
            borderColor: 'rgb(251, 191, 36)',
            backgroundColor: 'rgba(251, 191, 36, 0.1)',
            borderWidth: 2,
            pointRadius: 0,
            tension: 0.1,
            yAxisID: 'yPE',
        });
    }

    const scales = {
        x: {
            ticks: { color: '#9ca3af', maxTicksLimit: 12 },
            grid: { color: 'rgba(75, 85, 99, 0.3)' }
        },
        yPrice: {
            type: 'linear',
            position: 'left',
            ticks: { color: '#9ca3af' },
            grid: { color: 'rgba(75, 85, 99, 0.3)' },
            title: { display: true, text: 'Price', color: '#9ca3af' },
        },
    };

    if (hasPE) {
        scales.yPE = {
            type: 'linear',
            position: 'right',
            ticks: { color: 'rgb(251, 191, 36)' },
            grid: { drawOnChartArea: false },
            title: { display: true, text: 'P/E', color: 'rgb(251, 191, 36)' },
        };
    }

    priceChart = new Chart(ctx, {
        type: 'line',
        data: { labels, datasets },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                intersect: false,
                mode: 'index',
            },
            plugins: {
                legend: {
                    labels: { color: '#9ca3af' }
                },
                title: {
                    display: true,
                    text: `${symbol} Price History`,
                    color: '#e5e7eb',
                    font: { size: 16 }
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const label = context.dataset.label;
                            const val = context.parsed.y;
                            if (val == null) return null;
                            if (label === 'P/E') return `P/E: ${val.toFixed(2)}`;
                            return `${label}: $${val.toFixed(2)}`;
                        },
                        afterBody: function(contexts) {
                            if (!hasPE) return '';
                            const idx = contexts[0].dataIndex;
                            const eps = epsValues[idx];
                            if (eps != null) return `EPS: $${eps.toFixed(2)}`;
                            return '';
                        }
                    }
                }
            },
            scales,
        }
    });
}

// Load available indices
async function loadIndices() {
    try {
        const response = await fetch(`${API_BASE}/api/indices`);
        const result = await response.json();

        if (!result.success) return;

        const container = document.getElementById('indices');
        container.innerHTML = '';

        result.data.forEach(idx => {
            const card = document.createElement('div');
            card.className = 'bg-gray-800 rounded-lg p-4 hover:bg-gray-700 cursor-pointer transition-colors';
            card.onclick = () => showIndexSymbols(idx.key);
            card.innerHTML = `
                <h4 class="font-bold text-blue-400">${idx.name}</h4>
                <p class="text-sm text-gray-400">${idx.description}</p>
                <p class="text-xs text-gray-500 mt-2">${idx.count} symbols</p>
            `;
            container.appendChild(card);
        });
    } catch (error) {
        console.error('Failed to load indices:', error);
    }
}

// Show index symbols in table
async function showIndexSymbols(indexKey) {
    try {
        const response = await fetch(`${API_BASE}/api/indices/${indexKey}`);
        const result = await response.json();

        if (!result.success) return;

        const section = document.getElementById('indexSymbolsSection');
        const title = document.getElementById('indexSymbolsTitle');
        const desc = document.getElementById('indexSymbolsDesc');
        const tbody = document.getElementById('indexSymbolsBody');

        title.textContent = `${result.data.name} (${result.data.count} symbols)`;
        desc.textContent = result.data.description || '';
        
        tbody.innerHTML = '';
        
        // Check if we have company names in the response
        const hasCompanyNames = result.data.companies && Object.keys(result.data.companies).length > 0;
        
        result.data.symbols.forEach((symbol, idx) => {
            const tr = document.createElement('tr');
            tr.className = 'border-b border-gray-700 hover:bg-gray-700';
            const companyName = hasCompanyNames ? (result.data.companies[symbol] || '-') : '-';
            tr.innerHTML = `
                <td class="py-2 px-2 text-gray-500">${idx + 1}</td>
                <td class="py-2 px-2">
                    <span onclick="fetchSymbol('${symbol}')" class="font-mono text-blue-400 hover:text-blue-300 cursor-pointer hover:underline">${symbol}</span>
                </td>
                <td class="py-2 px-2 text-gray-400">${companyName}</td>
            `;
            tbody.appendChild(tr);
        });

        section.classList.remove('hidden');
        section.scrollIntoView({ behavior: 'smooth', block: 'start' });
    } catch (error) {
        console.error('Failed to load index symbols:', error);
    }
}

// Hide index symbols table
function hideIndexSymbols() {
    document.getElementById('indexSymbolsSection').classList.add('hidden');
}

// Fetch a symbol from the index table
function fetchSymbol(symbol) {
    document.getElementById('symbol').value = symbol;
    document.getElementById('fetchBtn').click();
    hideIndexSymbols();
    window.scrollTo({ top: 0, behavior: 'smooth' });
}

// Export to CSV
function exportCSV() {
    if (!currentData) return;

    const isDaily = currentData.period_type === 'daily';
    const records = isDaily ? currentData.daily_data : currentData.period_data;
    
    if (!records || records.length === 0) return;

    // Get headers from first record
    const headers = Object.keys(records[0]);
    
    // Build CSV
    let csv = headers.join(',') + '\n';
    records.forEach(record => {
        const row = headers.map(h => {
            let val = record[h];
            if (typeof val === 'object') val = JSON.stringify(val);
            if (typeof val === 'string' && val.includes(',')) val = `"${val}"`;
            return val ?? '';
        });
        csv += row.join(',') + '\n';
    });

    downloadFile(csv, `${currentData.symbol}_${currentData.period_type}.csv`, 'text/csv');
}

// Export to JSON
function exportJSON() {
    if (!currentData) return;
    const json = JSON.stringify(currentData, null, 2);
    downloadFile(json, `${currentData.symbol}_${currentData.period_type}.json`, 'application/json');
}

// Export table format (fixed-width columns like CLI output)
function exportExcel() {
    if (!currentData) return;
    const symbol = currentData.symbol;
    const period = currentData.period_type;
    const days = document.getElementById('days').value || 1825;
    
    // Trigger download via server endpoint
    const url = `${API_BASE}/api/stock-excel/${symbol}?days=${days}&period=${period}`;
    window.location.href = url;
}

// Format drop count object as C/L string
function formatDropCount(drop) {
    if (!drop) return '0/0';
    return `${drop.close || 0}/${drop.low || 0}`;
}

// Download helper
function downloadFile(content, filename, mimeType) {
    const blob = new Blob([content], { type: mimeType });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
}

// UI helpers
function showLoading(show) {
    document.getElementById('loading').classList.toggle('hidden', !show);
    document.getElementById('fetchBtn').disabled = show;
}

function showError(message) {
    document.getElementById('errorMsg').textContent = message;
    document.getElementById('error').classList.remove('hidden');
}

function hideError() {
    document.getElementById('error').classList.add('hidden');
}

function hideResults() {
    document.getElementById('results').classList.add('hidden');
}
