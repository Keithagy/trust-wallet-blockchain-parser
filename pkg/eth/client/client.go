package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth"
)

type Client struct {
	url string
}

func New(url string) *Client {
	return &Client{url: url}
}

type ethJsonRpcMethod string

const (
	EthBlockNumberMethod      ethJsonRpcMethod = "eth_blockNumber"
	EthGetBlockByNumberMethod ethJsonRpcMethod = "eth_getBlockByNumber"
)

// GetLatestBlock returns the latest block number, parsed as an int64.
func (c *Client) GetLatestBlock() (int64, error) {
	response, err := c.call(EthBlockNumberMethod, []any{})
	if err != nil {
		return 0, fmt.Errorf("client.GetLatestBlock():: EthBlockNumberMethod RPC call error: %w", err)
	}
	var result string
	err = json.Unmarshal(response, &result)
	if err != nil {
		return 0, fmt.Errorf("client.GetLatestBlock():: EthBlockNumberMethod RPC result unmarshal error: %w", err)
	}
	return hexToInt64(result)
}

// GetBlockTransactions returns the [eth.Txn]s for a given block number.
func (c *Client) GetBlockTransactions(blockNumber int64) ([]eth.Txn, error) {
	params := []any{fmt.Sprintf("0x%x", blockNumber), true}
	response, err := c.call(EthGetBlockByNumberMethod, params)
	if err != nil {
		return nil, fmt.Errorf("client.GetBlockTransactions():: EthGetBlockByNumberMethod RPC call error: %w", err)
	}

	var block eth.Block
	err = json.Unmarshal(response, &block)
	if err != nil {
		return nil, fmt.Errorf("client.GetBlockTransactions():: EthGetBlockByNumberMethod RPC result unmarshal error: %w", err)
	}

	return block.Transactions, nil
}

type RpcCallError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *RpcCallError) Error() string {
	return fmt.Sprintf("[%v] %v", e.Code, e.Message)
}

func (c *Client) call(method ethJsonRpcMethod, params []any) (json.RawMessage, error) {
	payload, err := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	})
	if err != nil {
		return nil, fmt.Errorf("client.call():: error marshalling payload: %w", err)
	}

	resp, err := http.Post(c.url, "application/json", strings.NewReader(string(payload)))
	if err != nil {
		return nil, fmt.Errorf("client.call():: error invoking RPC: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Result json.RawMessage `json:"result"`
		Error  *RpcCallError   `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("client.call():: error decoding RPC response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("client.call():: RPC error response: %w", result.Error)
	}

	return result.Result, nil
}

func hexToInt64(hex string) (int64, error) {
	i, err := strconv.ParseInt(hex, 0, 64)
	if err != nil {
		return 0, fmt.Errorf("hexToInt64: string parse error %w", err)
	}
	return i, nil
}
