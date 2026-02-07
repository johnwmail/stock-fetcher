package main

// CompanyNames maps stock symbols to company names
var CompanyNames = map[string]string{
	// Dow Jones
	"AAPL": "Apple", "AMGN": "Amgen", "AXP": "American Express", "BA": "Boeing",
	"CAT": "Caterpillar", "CRM": "Salesforce", "CSCO": "Cisco", "CVX": "Chevron",
	"DIS": "Disney", "DOW": "Dow Inc", "GS": "Goldman Sachs", "HD": "Home Depot",
	"HON": "Honeywell", "IBM": "IBM", "JNJ": "Johnson & Johnson", "JPM": "JPMorgan Chase",
	"KO": "Coca-Cola", "MCD": "McDonald's", "MMM": "3M", "MRK": "Merck",
	"MSFT": "Microsoft", "NKE": "Nike", "NVDA": "NVIDIA", "PG": "Procter & Gamble",
	"SHW": "Sherwin-Williams", "TRV": "Travelers", "UNH": "UnitedHealth",
	"V": "Visa", "VZ": "Verizon", "WMT": "Walmart",

	// NASDAQ 100
	"ABNB": "Airbnb", "ADBE": "Adobe", "ADI": "Analog Devices",
	"ADP": "ADP", "ADSK": "Autodesk", "AEP": "American Electric Power",
	"AMAT": "Applied Materials", "AMD": "AMD", "AMZN": "Amazon",
	"ANSS": "ANSYS", "APP": "AppLovin", "ARM": "ARM Holdings",
	"ASML": "ASML", "AVGO": "Broadcom", "AZN": "AstraZeneca",
	"BIIB": "Biogen", "BKNG": "Booking Holdings", "BKR": "Baker Hughes",
	"CCEP": "Coca-Cola Europacific", "CDNS": "Cadence Design", "CDW": "CDW",
	"CEG": "Constellation Energy", "CHTR": "Charter Communications", "CMCSA": "Comcast",
	"COST": "Costco", "CPRT": "Copart", "CRWD": "CrowdStrike",
	"CSGP": "CoStar Group", "CSX": "CSX", "CTAS": "Cintas",
	"CTSH": "Cognizant", "DASH": "DoorDash", "DDOG": "Datadog",
	"DLTR": "Dollar Tree", "DXCM": "DexCom", "EA": "Electronic Arts",
	"EXC": "Exelon", "FANG": "Diamondback Energy", "FAST": "Fastenal",
	"FTNT": "Fortinet", "GEHC": "GE HealthCare", "GFS": "GlobalFoundries",
	"GILD": "Gilead Sciences", "GOOG": "Alphabet (C)", "GOOGL": "Alphabet (A)",
	"IDXX": "IDEXX Laboratories", "ILMN": "Illumina", "INTC": "Intel",
	"INTU": "Intuit", "ISRG": "Intuitive Surgical", "KDP": "Keurig Dr Pepper",
	"KHC": "Kraft Heinz", "KLAC": "KLA Corporation", "LIN": "Linde",
	"LRCX": "Lam Research", "LULU": "Lululemon", "MAR": "Marriott",
	"MCHP": "Microchip Technology", "MDB": "MongoDB", "MDLZ": "Mondelez",
	"MELI": "MercadoLibre", "META": "Meta Platforms", "MNST": "Monster Beverage",
	"MRNA": "Moderna", "MRVL": "Marvell Technology", "MU": "Micron",
	"NFLX": "Netflix", "NXPI": "NXP Semiconductors", "ODFL": "Old Dominion Freight",
	"ON": "ON Semiconductor", "ORLY": "O'Reilly Automotive", "PANW": "Palo Alto Networks",
	"PAYX": "Paychex", "PCAR": "PACCAR", "PDD": "PDD Holdings",
	"PEP": "PepsiCo", "PYPL": "PayPal", "QCOM": "Qualcomm",
	"REGN": "Regeneron", "ROP": "Roper Technologies", "ROST": "Ross Stores",
	"SBUX": "Starbucks", "SMCI": "Super Micro Computer", "SNPS": "Synopsys",
	"TEAM": "Atlassian", "TMUS": "T-Mobile", "TSLA": "Tesla",
	"TTD": "The Trade Desk", "TTWO": "Take-Two Interactive", "TXN": "Texas Instruments",
	"VRSK": "Verisk Analytics", "VRTX": "Vertex Pharmaceuticals", "WBD": "Warner Bros Discovery",
	"WDAY": "Workday", "XEL": "Xcel Energy", "ZS": "Zscaler",

	// S&P 500 (major additions beyond Dow and NASDAQ)
	"ABBV": "AbbVie", "ABT": "Abbott", "ACN": "Accenture", "AFL": "Aflac",
	"AIG": "AIG", "ALL": "Allstate", "AMT": "American Tower", "AON": "Aon",
	"APD": "Air Products", "APH": "Amphenol",
	"BAC": "Bank of America", "BDX": "Becton Dickinson", "BK": "BNY Mellon",
	"BLK": "BlackRock", "BMY": "Bristol-Myers Squibb", "BRK.A": "Berkshire (A)",
	"BRK.B": "Berkshire (B)", "C": "Citigroup", "CB": "Chubb",
	"CCI": "Crown Castle", "CI": "Cigna", "CL": "Colgate-Palmolive",
	"COP": "ConocoPhillips", "DE": "Deere & Co", "DHR": "Danaher",
	"DUK": "Duke Energy", "ECL": "Ecolab", "EL": "Estee Lauder",
	"ELV": "Elevance Health", "EMR": "Emerson Electric", "EOG": "EOG Resources",
	"ETN": "Eaton", "F": "Ford", "FCX": "Freeport-McMoRan",
	"FDX": "FedEx", "GD": "General Dynamics", "GE": "GE Aerospace",
	"GM": "General Motors", "HCA": "HCA Healthcare", "HUM": "Humana",
	"ICE": "Intercontinental Exchange", "ITW": "Illinois Tool Works", "JCI": "Johnson Controls",
	"LLY": "Eli Lilly", "LMT": "Lockheed Martin", "LOW": "Lowe's",
	"MA": "Mastercard", "MDT": "Medtronic", "MET": "MetLife",
	"MO": "Altria", "MPC": "Marathon Petroleum", "MS": "Morgan Stanley",
	"NEE": "NextEra Energy", "NOC": "Northrop Grumman", "NSC": "Norfolk Southern",
	"ORCL": "Oracle", "OXY": "Occidental Petroleum", "PFE": "Pfizer",
	"PGR": "Progressive", "PLD": "Prologis", "PM": "Philip Morris",
	"PNC": "PNC Financial", "PSA": "Public Storage", "PSX": "Phillips 66",
	"RTX": "RTX", "SCHW": "Charles Schwab", "SLB": "Schlumberger",
	"SO": "Southern Company", "SPG": "Simon Property Group", "SPGI": "S&P Global",
	"SYK": "Stryker", "T": "AT&T", "TFC": "Truist",
	"TGT": "Target", "TMO": "Thermo Fisher", "UNP": "Union Pacific",
	"UPS": "UPS", "USB": "US Bancorp", "VLO": "Valero Energy",
	"WELL": "Welltower", "WFC": "Wells Fargo", "WM": "Waste Management",
	"XOM": "Exxon Mobil", "ZTS": "Zoetis",

	// Hang Seng (Hong Kong)
	"0001.HK": "CK Hutchison", "0002.HK": "CLP Holdings", "0003.HK": "HK & China Gas",
	"0005.HK": "HSBC Holdings", "0006.HK": "Power Assets", "0011.HK": "Hang Seng Bank",
	"0012.HK": "Henderson Land", "0016.HK": "SHK Properties", "0017.HK": "New World Dev",
	"0019.HK": "Swire Pacific A", "0027.HK": "Galaxy Entertainment", "0066.HK": "MTR Corporation",
	"0083.HK": "Sino Land", "0101.HK": "Hang Lung Properties", "0175.HK": "Geely Automobile",
	"0241.HK": "Alibaba Health", "0267.HK": "CITIC", "0288.HK": "WH Group",
	"0291.HK": "China Resources Beer", "0316.HK": "Orient Overseas", "0386.HK": "China Petroleum",
	"0388.HK": "HK Exchanges", "0669.HK": "Techtronic", "0688.HK": "China Overseas",
	"0700.HK": "Tencent", "0762.HK": "China Unicom", "0823.HK": "Link REIT",
	"0857.HK": "PetroChina", "0868.HK": "Xinyi Glass", "0881.HK": "Zhongsheng Group",
	"0883.HK": "CNOOC", "0939.HK": "CCB", "0941.HK": "China Mobile",
	"0960.HK": "Longfor Group", "0968.HK": "Xinyi Solar", "0981.HK": "SMIC",
	"0992.HK": "Lenovo Group", "1038.HK": "CK Infrastructure", "1044.HK": "Hengan International",
	"1093.HK": "CSPC Pharmaceutical", "1109.HK": "China Resources Land", "1113.HK": "CK Asset",
	"1177.HK": "Sino Biopharm", "1209.HK": "China Resources Mix", "1211.HK": "BYD",
	"1299.HK": "AIA Group", "1378.HK": "China Hongqiao", "1398.HK": "ICBC",
	"1810.HK": "Xiaomi", "1876.HK": "Budweiser APAC", "1928.HK": "Sands China",
	"1997.HK": "Wharf REIC", "2007.HK": "Country Garden", "2018.HK": "AAC Technologies",
	"2020.HK": "ANTA Sports", "2269.HK": "WuXi Biologics", "2313.HK": "Shenzhou International",
	"2318.HK": "Ping An Insurance", "2319.HK": "Mengniu Dairy", "2331.HK": "Li Ning",
	"2382.HK": "Sunny Optical", "2388.HK": "BOC Hong Kong", "2628.HK": "China Life",
	"2688.HK": "ENN Energy", "3328.HK": "Bank of Communications", "3690.HK": "Meituan",
	"3968.HK": "China Merchants Bank", "3988.HK": "Bank of China", "6098.HK": "Country Garden Services",
	"6862.HK": "Haidilao", "9618.HK": "JD.com", "9633.HK": "Nongfu Spring",
	"9888.HK": "Baidu", "9961.HK": "Trip.com", "9988.HK": "Alibaba",
	"9999.HK": "NetEase",
}

// GetCompanyName returns the company name for a symbol
func GetCompanyName(symbol string) string {
	if name, ok := CompanyNames[symbol]; ok {
		return name
	}
	return ""
}

// GetCompanyNamesForSymbols returns a map of company names for a list of symbols
func GetCompanyNamesForSymbols(symbols []string) map[string]string {
	result := make(map[string]string)
	for _, sym := range symbols {
		if name := GetCompanyName(sym); name != "" {
			result[sym] = name
		}
	}
	return result
}
