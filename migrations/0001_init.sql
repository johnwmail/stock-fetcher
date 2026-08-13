CREATE TABLE IF NOT EXISTS daily_prices (
    symbol TEXT NOT NULL,
    date   TEXT NOT NULL,
    open   TEXT,
    high   TEXT,
    low    TEXT,
    close  TEXT,
    volume TEXT,
    pe     TEXT,
    PRIMARY KEY (symbol, date)
);

CREATE TABLE IF NOT EXISTS fetch_log (
    symbol        TEXT PRIMARY KEY,
    source        TEXT,
    company_name  TEXT,
    ttm_eps       REAL,
    last_fetched  TEXT,
    latest_date   TEXT,
    earliest_date TEXT
);
