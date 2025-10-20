package transaction

// Source represents the source of transaction
// @Description Transaction source for balance changes
type Source string

const (
	DAILY_LOGIN Source = "DAILY_LOGIN"
	VIDEO_AD    Source = "VIDEO_AD"
	STORE       Source = "STORE_PURCHASE"
	SPIN        Source = "SPIN"
)
