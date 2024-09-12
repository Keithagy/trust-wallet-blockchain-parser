package notif

// Txn describes requisite data for handling a transaction notification.
type Txn struct {
	From  string
	To    string
	Value string
	Hash  string
	Block int64
}
