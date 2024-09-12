package notif

// Parser describes the API that blockchain parsers need to fulfil in order to connect to notification service dataflows
type Parser interface {
	// GetCurrentBlock returns the last parsed block
	GetCurrentBlock() int64

	// Subscribe registers a particular address and returns the success status of registering that particular address
	Subscribe(address string) bool

	// GetTransactions provides a list of inbound or outbound transactions for an address
	// NOTE: This does not, but should, support more granular query parameters (e.g. time bounds, txn types, txn counterparties)
	GetTransactions(address string) []Txn
}
