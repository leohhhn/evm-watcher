package watcher

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	usdcAddress  = common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
	transferSig  = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	// USDC has 6 decimals.
	usdcDecimals = big.NewFloat(1e6)
)

type Watcher struct {
	client *ethclient.Client
}

func New(ctx context.Context, rpcURL string) (*Watcher, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("get chain ID: %w", err)
	}
	log.Printf("connected to chain ID %s", chainID)

	return &Watcher{client: client}, nil
}

func (w *Watcher) Close() {
	w.client.Close()
}

func (w *Watcher) Start(ctx context.Context) error {
	logs := make(chan types.Log)
	header, err := w.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return fmt.Errorf("get latest block: %w", err)
	}

	query := ethereum.FilterQuery{
		FromBlock: header.Number,
		Addresses: []common.Address{usdcAddress},
		Topics:    [][]common.Hash{{transferSig}},
	}
	sub, err := w.client.SubscribeFilterLogs(ctx, query, logs)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	defer sub.Unsubscribe()

	log.Printf("subscribed to USDC Transfer events (%s)", usdcAddress.Hex())

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-sub.Err():
			return fmt.Errorf("subscription error: %w", err)
		case l := <-logs:
			printLog(l)
		}
	}
}

func printLog(l types.Log) {
	if len(l.Topics) < 3 {
		return
	}
	from := common.HexToAddress(l.Topics[1].Hex())
	to := common.HexToAddress(l.Topics[2].Hex())

	raw := new(big.Int).SetBytes(l.Data)
	amount := new(big.Float).Quo(new(big.Float).SetInt(raw), usdcDecimals)

	fmt.Printf("block=%-9d  tx=%s\n  from=%s\n  to  =%s\n  amount=%.2f USDC\n",
		l.BlockNumber, l.TxHash.Hex(), from.Hex(), to.Hex(), amount)
}
