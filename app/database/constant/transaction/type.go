package transaction

// Type represents the type of transaction
// @Description Transaction type for balance changes
type Type string

const (
	DAILY_REWARD Type = "DAILY_REWARD"
	AD_WATCH     Type = "AD_WATCH"
	PURCHASE     Type = "PURCHASE"
	SPIN_USE     Type = "SPIN_USE"
)
