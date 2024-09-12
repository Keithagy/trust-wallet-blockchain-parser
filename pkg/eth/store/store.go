package store

import (
	"github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth"
)

type InMemStore struct {
	lastBlock    int64
	transactions map[string][]eth.Txn
}

func NewInMemStore() *InMemStore {
	return &InMemStore{
		transactions: make(map[string][]eth.Txn),
	}
}

func (s *InMemStore) SaveBlock(blockNumber int64) error {
	s.lastBlock = blockNumber
	return nil
}

func (s *InMemStore) GetLastBlock() (int64, error) {
	return s.lastBlock, nil
}

func (s *InMemStore) SaveTransaction(tx eth.Txn) error {
	s.transactions[tx.From] = append(s.transactions[tx.From], tx)
	s.transactions[tx.To] = append(s.transactions[tx.To], tx)
	return nil
}

func (s *InMemStore) GetTransactions(address string) ([]eth.Txn, error) {
	return s.transactions[address], nil
}
