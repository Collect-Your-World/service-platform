package currency

import "fmt"

type Currency struct {
	Code   string `json:"code"`
	Symbol string `json:"symbol"`
}

func (c Currency) String() string {
	return fmt.Sprintf("%s (%s)", c.Code, c.Symbol)
}

type Code string

// Currency codes as constants
const (
	USD Code = "USD" // United States Dollar
	EUR Code = "EUR" // Euro
	GBP Code = "GBP" // British Pound Sterling
	JPY Code = "JPY" // Japanese Yen
	CHF Code = "CHF" // Swiss Franc
	CAD Code = "CAD" // Canadian Dollar
	AUD Code = "AUD" // Australian Dollar
	NZD Code = "NZD" // New Zealand Dollar
	SEK Code = "SEK" // Swedish Krona
	NOK Code = "NOK" // Norwegian Krone
	DKK Code = "DKK" // Danish Krone
	PLN Code = "PLN" // Polish Złoty
	CZK Code = "CZK" // Czech Koruna
	HUF Code = "HUF" // Hungarian Forint
	BGN Code = "BGN" // Bulgarian Lev
	RON Code = "RON" // Romanian Leu
	HRK Code = "HRK" // Croatian Kuna
	RSD Code = "RSD" // Serbian Dinar
	UAH Code = "UAH" // Ukrainian Hryvnia
	RUB Code = "RUB" // Russian Ruble
	CNY Code = "CNY" // Chinese Yuan
	KRW Code = "KRW" // South Korean Won
	INR Code = "INR" // Indian Rupee
	SGD Code = "SGD" // Singapore Dollar
	HKD Code = "HKD" // Hong Kong Dollar
	TWD Code = "TWD" // Taiwan Dollar
	THB Code = "THB" // Thai Baht
	MYR Code = "MYR" // Malaysian Ringgit
	IDR Code = "IDR" // Indonesian Rupiah
	PHP Code = "PHP" // Philippine Peso
	VND Code = "VND" // Vietnamese Dong
	BRL Code = "BRL" // Brazilian Real
	ARS Code = "ARS" // Argentine Peso
	CLP Code = "CLP" // Chilean Peso
	COP Code = "COP" // Colombian Peso
	PEN Code = "PEN" // Peruvian Sol
	UYU Code = "UYU" // Uruguayan Peso
	MXN Code = "MXN" // Mexican Peso
	ZAR Code = "ZAR" // South African Rand
	EGP Code = "EGP" // Egyptian Pound
	NGN Code = "NGN" // Nigerian Naira
	KES Code = "KES" // Kenyan Shilling
	MAD Code = "MAD" // Moroccan Dirham
	TND Code = "TND" // Tunisian Dinar
	SAR Code = "SAR" // Saudi Riyal
	AED Code = "AED" // UAE Dirham
	QAR Code = "QAR" // Qatari Riyal
	KWD Code = "KWD" // Kuwaiti Dinar
	BHD Code = "BHD" // Bahraini Dinar
	OMR Code = "OMR" // Omani Rial
	JOD Code = "JOD" // Jordanian Dinar
	LBP Code = "LBP" // Lebanese Pound
	ILS Code = "ILS" // Israeli Shekel
	TRY Code = "TRY" // Turkish Lira
	IRR Code = "IRR" // Iranian Rial
	PKR Code = "PKR" // Pakistani Rupee
	BDT Code = "BDT" // Bangladeshi Taka
	LKR Code = "LKR" // Sri Lankan Rupee
	NPR Code = "NPR" // Nepalese Rupee
	AFN Code = "AFN" // Afghan Afghani
	MMK Code = "MMK" // Myanmar Kyat
	LAK Code = "LAK" // Lao Kip
	KHR Code = "KHR" // Cambodian Riel
	BND Code = "BND" // Brunei Dollar
	MNT Code = "MNT" // Mongolian Tugrik
	KZT Code = "KZT" // Kazakhstani Tenge
	UZS Code = "UZS" // Uzbekistani Som
	AMD Code = "AMD" // Armenian Dram
	AZN Code = "AZN" // Azerbaijani Manat
	GEL Code = "GEL" // Georgian Lari
	MDL Code = "MDL" // Moldovan Leu
	BYN Code = "BYN" // Belarusian Ruble
	ISK Code = "ISK" // Icelandic Króna
)

var currencyMap = map[Code]Currency{
	USD: {Code: "USD", Symbol: "$"},
	EUR: {Code: "EUR", Symbol: "€"},
	GBP: {Code: "GBP", Symbol: "£"},
	JPY: {Code: "JPY", Symbol: "¥"},
	CHF: {Code: "CHF", Symbol: "₣"},
	CAD: {Code: "CAD", Symbol: "C$"},
	AUD: {Code: "AUD", Symbol: "A$"},
	NZD: {Code: "NZD", Symbol: "NZ$"},
	SEK: {Code: "SEK", Symbol: "kr"},
	NOK: {Code: "NOK", Symbol: "kr"},
	DKK: {Code: "DKK", Symbol: "kr"},
	PLN: {Code: "PLN", Symbol: "zł"},
	CZK: {Code: "CZK", Symbol: "Kč"},
	HUF: {Code: "HUF", Symbol: "Ft"},
	BGN: {Code: "BGN", Symbol: "лв"},
	RON: {Code: "RON", Symbol: "lei"},
	HRK: {Code: "HRK", Symbol: "kn"},
	RSD: {Code: "RSD", Symbol: "дин"},
	UAH: {Code: "UAH", Symbol: "₴"},
	RUB: {Code: "RUB", Symbol: "₽"},
	CNY: {Code: "CNY", Symbol: "¥"},
	KRW: {Code: "KRW", Symbol: "₩"},
	INR: {Code: "INR", Symbol: "₹"},
	SGD: {Code: "SGD", Symbol: "S$"},
	HKD: {Code: "HKD", Symbol: "HK$"},
	TWD: {Code: "TWD", Symbol: "NT$"},
	THB: {Code: "THB", Symbol: "฿"},
	MYR: {Code: "MYR", Symbol: "RM"},
	IDR: {Code: "IDR", Symbol: "Rp"},
	PHP: {Code: "PHP", Symbol: "₱"},
	VND: {Code: "VND", Symbol: "₫"},
	BRL: {Code: "BRL", Symbol: "R$"},
	ARS: {Code: "ARS", Symbol: "$"},
	CLP: {Code: "CLP", Symbol: "$"},
	COP: {Code: "COP", Symbol: "$"},
	PEN: {Code: "PEN", Symbol: "S/"},
	UYU: {Code: "UYU", Symbol: "$U"},
	MXN: {Code: "MXN", Symbol: "$"},
	ZAR: {Code: "ZAR", Symbol: "R"},
	EGP: {Code: "EGP", Symbol: "£"},
	NGN: {Code: "NGN", Symbol: "₦"},
	KES: {Code: "KES", Symbol: "Sh"},
	MAD: {Code: "MAD", Symbol: "د.م."},
	TND: {Code: "TND", Symbol: "د.ت"},
	SAR: {Code: "SAR", Symbol: "﷼"},
	AED: {Code: "AED", Symbol: "د.إ"},
	QAR: {Code: "QAR", Symbol: "ر.ق"},
	KWD: {Code: "KWD", Symbol: "د.ك"},
	BHD: {Code: "BHD", Symbol: ".د.ب"},
	OMR: {Code: "OMR", Symbol: "ر.ع."},
	JOD: {Code: "JOD", Symbol: "د.ا"},
	LBP: {Code: "LBP", Symbol: "ل.ل"},
	ILS: {Code: "ILS", Symbol: "₪"},
	TRY: {Code: "TRY", Symbol: "₺"},
	IRR: {Code: "IRR", Symbol: "﷼"},
	PKR: {Code: "PKR", Symbol: "₨"},
	BDT: {Code: "BDT", Symbol: "৳"},
	LKR: {Code: "LKR", Symbol: "Rs"},
	NPR: {Code: "NPR", Symbol: "₨"},
	AFN: {Code: "AFN", Symbol: "؋"},
	MMK: {Code: "MMK", Symbol: "K"},
	LAK: {Code: "LAK", Symbol: "₭"},
	KHR: {Code: "KHR", Symbol: "៛"},
	BND: {Code: "BND", Symbol: "B$"},
	MNT: {Code: "MNT", Symbol: "₮"},
	KZT: {Code: "KZT", Symbol: "₸"},
	UZS: {Code: "UZS", Symbol: "лв"},
	AMD: {Code: "AMD", Symbol: "֏"},
	AZN: {Code: "AZN", Symbol: "₼"},
	GEL: {Code: "GEL", Symbol: "₾"},
	MDL: {Code: "MDL", Symbol: "lei"},
	BYN: {Code: "BYN", Symbol: "Br"},
	ISK: {Code: "ISK", Symbol: "kr"},
}

// GetCurrency returns the Currency struct for a given currency code
func GetCurrency(code Code) *Currency {
	currency := currencyMap[code]
	return &currency
}

// GetAllCurrencies returns all available currencies
func GetAllCurrencies() map[Code]Currency {
	return currencyMap
}

// GetCurrencyByString returns the Currency struct for a given currency code string
func GetCurrencyByString(codeStr string) (*Currency, bool) {
	currency := GetCurrency(Code(codeStr))
	return currency, currency != nil
}

// IsValidCurrency checks if a currency code is valid
func IsValidCurrency(code Code) bool {
	_, exists := currencyMap[code]
	return exists
}
