package transaction

type TransactionType string

const (
	DAILY_REWARD TransactionType = "DAILY_REWARD"
	AD_WATCH     TransactionType = "AD_WATCH"
	PURCHASE     TransactionType = "PURCHASE"
	SPIN_USE     TransactionType = "SPIN_USE"
)
