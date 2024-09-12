package eth

import (
	"encoding/json"
	"math/big"
	"strings"
)

// Txn represents an Ethereum transaction
type Txn struct {
	From             string   `json:"from"`
	To               string   `json:"to"`
	Hash             string   `json:"hash"`
	Value            *big.Int `json:"value"`
	BlockNumber      *big.Int `json:"blockNumber"`
	TransactionIndex *big.Int `json:"transactionIndex"`
}

// UnmarshalJSON implements a custom unmarshaler for Txn to handle hex number parsing.
func (t *Txn) UnmarshalJSON(data []byte) error {
	type TxnAlias Txn
	temp := &struct {
		Value            string `json:"value"`
		BlockNumber      string `json:"blockNumber"`
		TransactionIndex string `json:"transactionIndex"`
		*TxnAlias
	}{
		TxnAlias: (*TxnAlias)(t),
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Parse Value
	t.Value = parseHexBigInt(temp.Value)

	// Parse BlockNumber
	t.BlockNumber = parseHexBigInt(temp.BlockNumber)

	// Parse TransactionIndex
	t.TransactionIndex = parseHexBigInt(temp.TransactionIndex)

	return nil
}

// parseHexBigInt parses a hexadecimal string into a big.Int
func parseHexBigInt(hex string) *big.Int {
	// Remove "0x" prefix if present
	hex = strings.TrimPrefix(hex, "0x")

	if hex == "" {
		return big.NewInt(0)
	}

	// Parse the hexadecimal string
	n := new(big.Int)
	n.SetString(hex, 16)
	return n
}
