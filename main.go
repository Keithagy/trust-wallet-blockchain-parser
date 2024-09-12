package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	ethclient "github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth/client"
	ethparser "github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth/parser"
	ethstore "github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth/store"
	ethsubs "github.com/keithagy/trust-wallet-blockchain-parser/pkg/eth/subscriptions"
)

const endpoint = "https://cloudflare-eth.com"

func main() {
	client := ethclient.New(endpoint)
	store := ethstore.NewInMemStore()
	subscriptions := ethsubs.New()
	parser := ethparser.New(subscriptions, client, store)
	setupRoutes(parser)

	var wg sync.WaitGroup

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case sig := <-sigs:
			fmt.Printf("Received signal: %v\n", sig)
			cancel() // Cancel the context
		case <-ctx.Done():
			// Context was canceled somewhere else
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		parser.Start(ctx)
	}()

	server := &http.Server{
		Addr:    ":8080",
		Handler: nil, // Use default ServeMux
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Server is shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	wg.Wait()
	log.Println("Server exiting")
}

func setupRoutes(parser *ethparser.Parser) {
	http.HandleFunc("/currentBlock", func(w http.ResponseWriter, r *http.Request) {
		block := parser.GetCurrentBlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatInt(block, 10)))
	})

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Address is required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodPost:
			if parser.Subscribe(address) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Subscribed successfully"))
			} else {
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte("Address already subscribed"))
			}
		case http.MethodDelete:
			if parser.Unsubscribe(address) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("Unsubscribed successfully"))
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Address not subscribed"))
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Address is required", http.StatusBadRequest)
			return
		}

		transactions := parser.GetTransactions(address)
		// Convert transactions to JSON and write to response
		w.WriteHeader(http.StatusOK)
		txnJson, err := json.Marshal(transactions)
		if err != nil {
			http.Error(w, "Error marshalling transactions", http.StatusInternalServerError)
		}
		w.Write(txnJson)
	})
}
