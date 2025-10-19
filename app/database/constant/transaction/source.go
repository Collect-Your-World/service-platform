package transaction

type Source string

const (
	DAILY_LOGIN Source = "DAILY_LOGIN"
	VIDEO_AD    Source = "VIDEO_AD"
	STORE       Source = "STORE_PURCHASE"
	SPIN        Source = "SPIN"
)
