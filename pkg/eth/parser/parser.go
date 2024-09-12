package parser

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth"
)

const LOOKBACK_FROM_LATEST = 20

// Parser implements the ability to subscribe to Etherium transaction updates, and schedules blockchain queries.
type Parser struct {
	currentBlock  int64
	subscriptions subscriptions
	client        client
	store         store
}

type subscriptions interface {
	Add(address string) bool
	Check(address string) bool
	Remove(address string) bool
}

// client wraps the methods required by [Parser] to query the blockchain
type client interface {
	GetLatestBlock() (int64, error)
	GetBlockTransactions(blockNumber int64) ([]eth.Txn, error)
}

// store wraps the methods required by [Parser] to persist results of blockchain block / transaction queries
type store interface {
	SaveBlock(blockNumber int64) error
	GetLastBlock() (int64, error)
	SaveTransaction(tx eth.Txn) error
	GetTransactions(address string) ([]eth.Txn, error)
}

// New constructs a new [Parser].
func New(subs subscriptions, client client, store store) *Parser {
	return &Parser{
		subscriptions: subs,
		client:        client,
		store:         store,
	}
}

// Start begins the [Parser], and is the entry point for the transaction-polling dataflow.
func (p *Parser) Start(ctx context.Context) {
	latestBlock, err := p.client.GetLatestBlock()
	if err != nil {
		log.Fatal("Failed to get latest block.", "error", err)
	}
	startBlock := latestBlock - LOOKBACK_FROM_LATEST
	if startBlock < 0 {
		startBlock = 0
	}
	p.indexSubbedTransactionsForBlockRange(startBlock, latestBlock)
	p.parse(ctx)
}

func (p *Parser) parse(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled. Eth parser shutting down.")
			return
		case <-ticker.C:
			slog.Info("Checking for new transactions")
			latestBlock, err := p.client.GetLatestBlock()
			if err != nil {
				slog.Error("Failed to get latest block.", "error", err)
				continue
			}
			startBlock := p.currentBlock
			p.indexSubbedTransactionsForBlockRange(startBlock, latestBlock)
		}
	}
}

func (p *Parser) indexSubbedTransactionsForBlockRange(startBlock int64, endBlock int64) {

	for blockNumber := startBlock; blockNumber <= endBlock; blockNumber++ {
		slog.Info("Updating block", "blockNumber", blockNumber)
		transactions, err := p.client.GetBlockTransactions(blockNumber)
		if err != nil {
			slog.Error("Failed to get transactions for block.", "blockNumber", blockNumber, "error", err)
			continue
		}

		slog.Info("Writing transactions for block", "blockNumber", blockNumber, "transactionCount", len(transactions))
		for _, tx := range transactions {
			p.processTransaction(tx)
		}
		slog.Info("Written transactions for block", "blockNumber", blockNumber, "transactionCount", len(transactions))

		p.updateCurrentBlock(blockNumber)
	}
}

func (p *Parser) getCurrentBlock() int64 {
	return p.currentBlock
}

func (p *Parser) subscribe(address string) bool {
	if p.subscriptions.Check(address) {
		return false
	}
	p.subscriptions.Add(address)
	go p.backfill()
	return true
}
func (p *Parser) backfill() {
	startBlock := p.currentBlock - LOOKBACK_FROM_LATEST
	if startBlock < 0 {
		startBlock = 0
	}
	p.indexSubbedTransactionsForBlockRange(startBlock, p.currentBlock)
}

func (p *Parser) getTransactions(address string) ([]eth.Txn, error) {
	if !p.subscriptions.Check(address) {
		return nil, fmt.Errorf("parser.getTransactions():: address not subscribed: %s", address)
	}
	txns, err := p.store.GetTransactions(address)
	if err != nil {
		return nil, fmt.Errorf("parser.getTransactions():: error retrieving transactions: %w", err)
	}
	return txns, nil
}

func (p *Parser) processTransaction(tx eth.Txn) {
	// print out addresses
	if p.subscriptions.Check(tx.From) || p.subscriptions.Check(tx.To) {
		err := p.store.SaveTransaction(tx)
		if err != nil {
			slog.Error("Failed to save transaction for sender", "address", tx.From, "error", err)
		}
	}
}

func (p *Parser) updateCurrentBlock(blockNumber int64) {
	p.currentBlock = blockNumber
	p.store.SaveBlock(blockNumber)
}
