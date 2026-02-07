package main

// Index represents a stock market index with its constituent symbols
type Index struct {
	Name        string
	Description string
	Symbols     []string
}

// GetIndices returns all supported indices
func GetIndices() map[string]Index {
	return map[string]Index{
		"sp500":     SP500Index,
		"dow":       DowIndex,
		"nasdaq100": Nasdaq100Index,
		"hangseng":  HangSengIndex,
	}
}

// DowIndex - Dow Jones Industrial Average (30 stocks)
var DowIndex = Index{
	Name:        "Dow Jones Industrial Average",
	Description: "30 large-cap US stocks",
	Symbols: []string{
		"AAPL", "AMGN", "AXP", "BA", "CAT", "CRM", "CSCO", "CVX", "DIS", "DOW",
		"GS", "HD", "HON", "IBM", "JNJ", "JPM", "KO", "MCD", "MMM", "MRK",
		"MSFT", "NKE", "NVDA", "PG", "SHW", "TRV", "UNH", "V", "VZ", "WMT",
	},
}

// Nasdaq100Index - NASDAQ 100 stocks (as of 2025)
var Nasdaq100Index = Index{
	Name:        "NASDAQ 100",
	Description: "100 largest non-financial companies on NASDAQ",
	Symbols: []string{
		"AAPL",  // Apple
		"ABNB",  // Airbnb
		"ADBE",  // Adobe
		"ADI",   // Analog Devices
		"ADP",   // Automatic Data Processing
		"ADSK",  // Autodesk
		"AEP",   // American Electric Power
		"AMAT",  // Applied Materials
		"AMD",   // AMD
		"AMGN",  // Amgen
		"AMZN",  // Amazon
		"ANSS",  // ANSYS
		"APP",   // AppLovin
		"ARM",   // ARM Holdings
		"ASML",  // ASML
		"AVGO",  // Broadcom
		"AZN",   // AstraZeneca
		"BIIB",  // Biogen
		"BKNG",  // Booking Holdings
		"BKR",   // Baker Hughes
		"CCEP",  // Coca-Cola Europacific
		"CDNS",  // Cadence Design
		"CDW",   // CDW
		"CEG",   // Constellation Energy
		"CHTR",  // Charter Communications
		"CMCSA", // Comcast
		"COST",  // Costco
		"CPRT",  // Copart
		"CRWD",  // CrowdStrike
		"CSCO",  // Cisco
		"CSGP",  // CoStar Group
		"CSX",   // CSX
		"CTAS",  // Cintas
		"CTSH",  // Cognizant
		"DASH",  // DoorDash
		"DDOG",  // Datadog
		"DLTR",  // Dollar Tree
		"DXCM",  // DexCom
		"EA",    // Electronic Arts
		"EXC",   // Exelon
		"FANG",  // Diamondback Energy
		"FAST",  // Fastenal
		"FTNT",  // Fortinet
		"GEHC",  // GE HealthCare
		"GFS",   // GlobalFoundries
		"GILD",  // Gilead Sciences
		"GOOG",  // Alphabet Class C
		"GOOGL", // Alphabet Class A
		"HON",   // Honeywell
		"IDXX",  // IDEXX Laboratories
		"ILMN",  // Illumina
		"INTC",  // Intel
		"INTU",  // Intuit
		"ISRG",  // Intuitive Surgical
		"KDP",   // Keurig Dr Pepper
		"KHC",   // Kraft Heinz
		"KLAC",  // KLA Corporation
		"LIN",   // Linde
		"LRCX",  // Lam Research
		"LULU",  // Lululemon
		"MAR",   // Marriott
		"MCHP",  // Microchip Technology
		"MDB",   // MongoDB
		"MDLZ",  // Mondelez
		"MELI",  // MercadoLibre
		"META",  // Meta Platforms
		"MNST",  // Monster Beverage
		"MRNA",  // Moderna
		"MRVL",  // Marvell Technology
		"MSFT",  // Microsoft
		"MU",    // Micron
		"NFLX",  // Netflix
		"NVDA",  // NVIDIA
		"NXPI",  // NXP Semiconductors
		"ODFL",  // Old Dominion Freight
		"ON",    // ON Semiconductor
		"ORLY",  // O'Reilly Automotive
		"PANW",  // Palo Alto Networks
		"PAYX",  // Paychex
		"PCAR",  // PACCAR
		"PDD",   // PDD Holdings
		"PEP",   // PepsiCo
		"PYPL",  // PayPal
		"QCOM",  // Qualcomm
		"REGN",  // Regeneron
		"ROP",   // Roper Technologies
		"ROST",  // Ross Stores
		"SBUX",  // Starbucks
		"SMCI",  // Super Micro Computer
		"SNPS",  // Synopsys
		"TEAM",  // Atlassian
		"TMUS",  // T-Mobile
		"TSLA",  // Tesla
		"TTD",   // The Trade Desk
		"TTWO",  // Take-Two Interactive
		"TXN",   // Texas Instruments
		"VRSK",  // Verisk Analytics
		"VRTX",  // Vertex Pharmaceuticals
		"WBD",   // Warner Bros Discovery
		"WDAY",  // Workday
		"XEL",   // Xcel Energy
		"ZS",    // Zscaler
	},
}

// SP500Index - S&P 500 stocks (as of 2025)
var SP500Index = Index{
	Name:        "S&P 500",
	Description: "500 largest US companies by market cap",
	Symbols: []string{
		// A
		"A", "AAPL", "ABBV", "ABNB", "ABT", "ACGL", "ACN", "ADBE", "ADI", "ADM",
		"ADP", "ADSK", "AEE", "AEP", "AES", "AFL", "AIG", "AIZ", "AJG", "AKAM",
		"ALB", "ALGN", "ALL", "ALLE", "AMAT", "AMCR", "AMD", "AME", "AMGN", "AMP",
		"AMT", "AMZN", "ANET", "ANSS", "AON", "AOS", "APA", "APD", "APH", "APTV",
		"ARE", "ATO", "AVB", "AVGO", "AVY", "AWK", "AXON", "AXP", "AZO",
		// B
		"BA", "BAC", "BALL", "BAX", "BBWI", "BBY", "BDX", "BEN", "BF.B", "BG",
		"BIIB", "BIO", "BK", "BKNG", "BKR", "BLDR", "BLK", "BMY", "BR", "BRK.B",
		"BRO", "BSX", "BWA", "BX", "BXP",
		// C
		"C", "CAG", "CAH", "CARR", "CAT", "CB", "CBOE", "CBRE", "CCI", "CCL",
		"CDNS", "CDW", "CE", "CEG", "CF", "CFG", "CHD", "CHRW", "CHTR", "CI",
		"CINF", "CL", "CLX", "CMA", "CMCSA", "CME", "CMG", "CMI", "CMS", "CNC",
		"CNP", "COF", "COO", "COP", "COR", "COST", "CPAY", "CPB", "CPRT", "CPT",
		"CRL", "CRM", "CRWD", "CSCO", "CSGP", "CSX", "CTAS", "CTLT", "CTRA", "CTSH",
		"CTVA", "CVS", "CVX",
		// D
		"D", "DAL", "DAY", "DD", "DE", "DECK", "DELL", "DFS", "DG", "DGX",
		"DHI", "DHR", "DIS", "DLR", "DLTR", "DOC", "DOV", "DOW", "DPZ", "DRI",
		"DTE", "DUK", "DVA", "DVN", "DXCM",
		// E
		"EA", "EBAY", "ECL", "ED", "EFX", "EG", "EIX", "EL", "ELV", "EMN",
		"EMR", "ENPH", "EOG", "EPAM", "EQIX", "EQR", "EQT", "ERIE", "ES", "ESS",
		"ETN", "ETR", "EVRG", "EW", "EXC", "EXPD", "EXPE", "EXR",
		// F
		"F", "FANG", "FAST", "FCX", "FDS", "FDX", "FE", "FFIV", "FI", "FICO",
		"FIS", "FITB", "FMC", "FOX", "FOXA", "FRT", "FSLR", "FTNT", "FTV",
		// G
		"GD", "GDDY", "GE", "GEHC", "GEN", "GEV", "GILD", "GIS", "GL", "GLW",
		"GM", "GNRC", "GOOG", "GOOGL", "GPC", "GPN", "GRMN", "GS", "GWW",
		// H
		"HAL", "HAS", "HBAN", "HCA", "HD", "HES", "HIG", "HII", "HLT", "HOLX",
		"HON", "HPE", "HPQ", "HRL", "HSIC", "HST", "HSY", "HUBB", "HUM", "HWM",
		// I
		"IBM", "ICE", "IDXX", "IEX", "IFF", "ILMN", "INCY", "INTC", "INTU", "INVH",
		"IP", "IPG", "IQV", "IR", "IRM", "ISRG", "IT", "ITW", "IVZ",
		// J
		"J", "JBHT", "JBL", "JCI", "JKHY", "JNJ", "JNPR", "JPM",
		// K
		"K", "KDP", "KEY", "KEYS", "KHC", "KIM", "KKR", "KLAC", "KMB", "KMI",
		"KMX", "KO", "KR",
		// L
		"L", "LDOS", "LEN", "LH", "LHX", "LIN", "LKQ", "LLY", "LMT", "LNT",
		"LOW", "LRCX", "LULU", "LUV", "LVS", "LW", "LYB", "LYV",
		// M
		"MA", "MAA", "MAR", "MAS", "MCD", "MCHP", "MCK", "MCO", "MDLZ", "MDT",
		"MET", "META", "MGM", "MHK", "MKC", "MKTX", "MLM", "MMC", "MMM", "MNST",
		"MO", "MOH", "MOS", "MPC", "MPWR", "MRK", "MRNA", "MRO", "MS", "MSCI",
		"MSFT", "MSI", "MTB", "MTCH", "MTD", "MU",
		// N
		"NCLH", "NDAQ", "NDSN", "NEE", "NEM", "NFLX", "NI", "NKE", "NOC", "NOW",
		"NRG", "NSC", "NTAP", "NTRS", "NUE", "NVDA", "NVR", "NWS", "NWSA", "NXPI",
		// O
		"O", "ODFL", "OKE", "OMC", "ON", "ORCL", "ORLY", "OTIS", "OXY",
		// P
		"PANW", "PARA", "PAYC", "PAYX", "PCAR", "PCG", "PEG", "PEP", "PFE", "PFG",
		"PG", "PGR", "PH", "PHM", "PKG", "PLD", "PLTR", "PM", "PNC", "PNR",
		"PNW", "PODD", "POOL", "PPG", "PPL", "PRU", "PSA", "PSX", "PTC", "PWR",
		"PYPL",
		// Q
		"QCOM", "QRVO",
		// R
		"RCL", "REG", "REGN", "RF", "RJF", "RL", "RMD", "ROK", "ROL", "ROP",
		"ROST", "RSG", "RTX",
		// S
		"SBAC", "SBUX", "SCHW", "SHW", "SJM", "SLB", "SMCI", "SNA", "SNPS", "SO",
		"SOLV", "SPG", "SPGI", "SRE", "STE", "STLD", "STT", "STX", "STZ", "SWK",
		"SWKS", "SYF", "SYK", "SYY",
		// T
		"T", "TAP", "TDG", "TDY", "TECH", "TEL", "TER", "TFC", "TFX", "TGT",
		"TJX", "TMO", "TMUS", "TPR", "TRGP", "TRMB", "TROW", "TRV", "TSCO", "TSLA",
		"TSN", "TT", "TTWO", "TXN", "TXT", "TYL",
		// U
		"UAL", "UBER", "UDR", "UHS", "ULTA", "UNH", "UNP", "UPS", "URI", "USB",
		// V
		"V", "VICI", "VLO", "VLTO", "VMC", "VRSK", "VRSN", "VRTX", "VST", "VTR",
		"VTRS", "VZ",
		// W
		"WAB", "WAT", "WBA", "WBD", "WDC", "WEC", "WELL", "WFC", "WM", "WMB",
		"WMT", "WRB", "WST", "WTW", "WY", "WYNN",
		// X
		"XEL", "XOM", "XYL",
		// Y
		"YUM",
		// Z
		"ZBH", "ZBRA", "ZTS",
	},
}

// HangSengIndex - Hong Kong Hang Seng Index constituents (as of 2025)
// Note: These use .HK suffix for Yahoo Finance
var HangSengIndex = Index{
	Name:        "Hang Seng Index",
	Description: "Major Hong Kong stocks (use .HK suffix)",
	Symbols: []string{
		// Financials
		"0005.HK", // HSBC Holdings
		"0011.HK", // Hang Seng Bank
		"0388.HK", // Hong Kong Exchanges
		"0939.HK", // China Construction Bank
		"1299.HK", // AIA Group
		"1398.HK", // ICBC
		"2318.HK", // Ping An Insurance
		"2388.HK", // BOC Hong Kong
		"2628.HK", // China Life Insurance
		"3328.HK", // Bank of Communications
		"3988.HK", // Bank of China
		"0066.HK", // MTR Corporation
		"1038.HK", // CK Infrastructure
		"1113.HK", // CK Asset Holdings
		"2007.HK", // Country Garden
		// Technology
		"0700.HK", // Tencent Holdings
		"0981.HK", // SMIC
		"1810.HK", // Xiaomi
		"2382.HK", // Sunny Optical
		"3690.HK", // Meituan
		"9618.HK", // JD.com
		"9888.HK", // Baidu
		"9988.HK", // Alibaba
		"9999.HK", // NetEase
		"0241.HK", // Alibaba Health
		"0268.HK", // Kingdee International
		"0285.HK", // BYD Electronic
		"0772.HK", // China Literature
		"1024.HK", // Kuaishou Technology
		"1347.HK", // Hua Hong Semi
		"1833.HK", // Ping An Healthcare
		"2018.HK", // AAC Technologies
		"6060.HK", // ZhongAn Online
		"6618.HK", // JD Health
		"9626.HK", // Bilibili
		"9698.HK", // GDS Holdings
		// Consumer & Healthcare
		"0027.HK", // Galaxy Entertainment
		"0175.HK", // Geely Automobile
		"0267.HK", // CITIC
		"0291.HK", // China Resources Beer
		"1044.HK", // Hengan International
		"1093.HK", // CSPC Pharmaceutical
		"1177.HK", // Sino Biopharmaceutical
		"1211.HK", // BYD Company
		"1876.HK", // Budweiser APAC
		"1928.HK", // Sands China
		"2020.HK", // ANTA Sports
		"2269.HK", // WuXi Biologics
		"2313.HK", // Shenzhou International
		"2319.HK", // Mengniu Dairy
		"2331.HK", // Li Ning
		"3692.HK", // Hansoh Pharmaceutical
		"6098.HK", // Country Garden Services
		"6160.HK", // BeiGene
		"6862.HK", // Haidilao
		"9633.HK", // Nongfu Spring
		"9901.HK", // New Oriental Education
		// Electric Vehicles
		"2015.HK", // Li Auto
		"9866.HK", // NIO
		"9868.HK", // XPeng
		// Telecom & Utilities
		"0762.HK", // China Unicom
		"0883.HK", // CNOOC
		"0941.HK", // China Mobile
		"1088.HK", // China Shenhua Energy
		"2688.HK", // ENN Energy
		// Properties & Infrastructure
		"0001.HK", // CK Hutchison
		"0002.HK", // CLP Holdings
		"0003.HK", // HK & China Gas
		"0006.HK", // Power Assets
		"0012.HK", // Henderson Land
		"0016.HK", // Sun Hung Kai Properties
		"0017.HK", // New World Development
		"0019.HK", // Swire Pacific
		"0083.HK", // Sino Land
		"0101.HK", // Hang Lung Properties
		"0288.HK", // WH Group
		"0386.HK", // Sinopec
		"0688.HK", // China Overseas Land
		"0823.HK", // Link REIT
		"0857.HK", // PetroChina
		"0960.HK", // Longfor Group
		"1109.HK", // China Resources Land
		"1997.HK", // Wharf Real Estate
		"1972.HK", // Swire Properties
		"2057.HK", // ZTO Express
	},
}
