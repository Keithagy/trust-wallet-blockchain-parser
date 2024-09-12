package parser

import (
	"log/slog"
	"slices"

	"github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth"
	customIter "github.com/keithagy/trust-wallet-blockchain-parser/pkg/iter"
	"github.com/keithagy/trust-wallet-blockchain-parser/pkg/notif"
)

// GetCurrentBlock returns the current block of the [Parser].
func (p *Parser) GetCurrentBlock() int64 {
	return p.getCurrentBlock()
}

// Subscribe adds an address if it has not already been added, returning `true`, and returns `false` otherwise.
func (p *Parser) Subscribe(address string) bool {
	return p.subscribe(address)
}

// Unsubscribe removes an address if it has been added, returning `true`, and returns `false` otherwise.
func (p *Parser) Unsubscribe(address string) bool {
	return p.unsubscribe(address)
}

// unsubscribe removes an address from the subscriptions.
func (p *Parser) unsubscribe(address string) bool {
	return p.subscriptions.Remove(address)
}

// GetTransactions returns the list of transactions already seen for an address.
func (p *Parser) GetTransactions(address string) []notif.Txn {
	txns, err := p.getTransactions(address)
	if err != nil {
		slog.Error("Get transactions by address failed", "address", address, "error", err)
		return nil
	}
	return slices.Collect(customIter.Map(txns, intoNotifTxn))
}

func intoNotifTxn(t eth.Txn) notif.Txn {
	return notif.Txn{
		From:  t.From,
		To:    t.To,
		Value: t.Value.String(),
		Hash:  t.Hash,
		Block: t.BlockNumber.Int64(),
	}
}

var _ notif.Parser = (*Parser)(nil)
