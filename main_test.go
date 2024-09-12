package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth"
	ethparser "github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth/parser"
	ethstore "github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth/store"
	ethsubs "github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth/subscriptions"
	"github.com/keithagy/trust-wallet-blockchain-parser/pkg/notif"
)

type mockClient struct {
	latestBlock int64
	blocks      map[int64][]eth.Txn
}

func (m *mockClient) GetLatestBlock() (int64, error) {
	return m.latestBlock, nil
}

func (m *mockClient) GetBlockTransactions(blockNumber int64) ([]eth.Txn, error) {
	return m.blocks[blockNumber], nil
}

func TestParserEndToEnd(t *testing.T) {
	// Step 1: Initialize parser with mock client, store, and subscriptions
	client := &mockClient{
		latestBlock: 2,
		blocks: map[int64][]eth.Txn{
			1: {
				{
					From:        "0x1234",
					To:          "0x5678",
					Hash:        "0xabcd",
					Value:       big.NewInt(100),
					BlockNumber: big.NewInt(1),
				},
			},
			2: {
				{
					From:        "0x5678",
					To:          "0x1234",
					Hash:        "0xefgh",
					Value:       big.NewInt(200),
					BlockNumber: big.NewInt(2),
				},
			},
		},
	}
	store := ethstore.NewInMemStore()
	subscriptions := ethsubs.New()
	parser := ethparser.New(subscriptions, client, store)

	// Start the parser in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		parser.Start(ctx)
	}()

	// Step 3: Before subscription, parser should return empty list of transactions
	time.Sleep(100 * time.Millisecond) // Give parser time to process initial blocks
	txns := parser.GetTransactions("0x1234")
	if len(txns) != 0 {
		t.Fatalf("Expected 0 transactions before subscription, got %d", len(txns))
	}

	// Step 4: After subscription, parser should return list of transactions
	parser.Subscribe("0x1234")
	time.Sleep(100 * time.Millisecond) // Give parser time to process subscription
	txns = parser.GetTransactions("0x1234")
	expectedTxns := []notif.Txn{
		{
			From:  "0x1234",
			To:    "0x5678",
			Value: "100",
			Hash:  "0xabcd",
			Block: 1,
		},
		{
			From:  "0x5678",
			To:    "0x1234",
			Value: "200",
			Hash:  "0xefgh",
			Block: 2,
		},
	}
	if !reflect.DeepEqual(txns, expectedTxns) {
		t.Fatalf("Transactions mismatch after subscription.\nGot: %+v\nWant: %+v", txns, expectedTxns)
	}

	// Step 5: After unsubscription, parser should return empty list of transactions
	parser.Unsubscribe("0x1234")
	time.Sleep(100 * time.Millisecond) // Give parser time to process unsubscription
	txns = parser.GetTransactions("0x1234")
	if len(txns) != 0 {
		t.Fatalf("Expected 0 transactions after unsubscription, got %d", len(txns))
	}

	// Clean up
	cancel()
	wg.Wait()
}

func (m *mockClient) call(method ethJsonRpcMethod, params []any) (json.RawMessage, error) {
	switch method {
	case EthBlockNumberMethod:
		result := fmt.Sprintf("0x%x", m.latestBlock)
		return json.Marshal(result)
	case EthGetBlockByNumberMethod:
		blockNumber := params[0].(string)
		blockNum, _ := hexToInt64(blockNumber)
		block := eth.Block{Transactions: m.blocks[blockNum]}
		return json.Marshal(block)
	default:
		return nil, fmt.Errorf("unsupported method: %s", method)
	}
}

func hexToInt64(hex string) (int64, error) {
	i, err := strconv.ParseInt(hex, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("hexToInt64: string parse error %w", err)
	}
	return i, nil
}

type ethJsonRpcMethod string

const (
	EthBlockNumberMethod      ethJsonRpcMethod = "eth_blockNumber"
	EthGetBlockByNumberMethod ethJsonRpcMethod = "eth_getBlockByNumber"
)
