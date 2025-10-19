package transaction

type Type string

const (
	DAILY_REWARD Type = "DAILY_REWARD"
	AD_WATCH     Type = "AD_WATCH"
	PURCHASE     Type = "PURCHASE"
	SPIN_USE     Type = "SPIN_USE"
)
