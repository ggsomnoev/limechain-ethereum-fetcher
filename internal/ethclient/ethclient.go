package ethclient

import (
	"context"
	"encoding/hex"
	"ethfetcher/internal/ethereumfetcher/api"
	"ethfetcher/internal/logger"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	gethclient "github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

type EthereumClient struct {
	client *gethclient.Client
}

func NewEthereumClientWithRetry(
	ctx context.Context,
	ethNodeURL string,
	maxRetries int,
	baseDelay time.Duration,
) (*EthereumClient, error) {
	var (
		client *gethclient.Client
		err    error
	)

	for i := 0; i < maxRetries; i++ {
		client, err = gethclient.DialContext(ctx, ethNodeURL)
		if err == nil {
			return &EthereumClient{client: client}, nil
		}

		logger.GetLogger().Info(fmt.Sprintf("Retrying ethereum node connection...(%d)", i))

		backoff := baseDelay * (1 << i)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during ethereum client dial: %w", ctx.Err())
		}
	}

	return nil, fmt.Errorf("failed to connect to ethereum node after %d attempts: %w", maxRetries, err)
}

func (ec *EthereumClient) FetchTransactionByHash(ctx context.Context, hash string) (api.Transaction, error) {
	var result api.Transaction

	txHash := common.HexToHash(hash)
	tx, isPending, err := ec.client.TransactionByHash(ctx, txHash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return result, nil
		}
		return result, fmt.Errorf("failed to fetch transaction: %w", err)
	}
	if isPending {
		logger.GetLogger().Info(fmt.Printf("transaction %s is still pending", hash))
	}

	receipt, err := ec.client.TransactionReceipt(ctx, txHash)
	if err != nil {
		return result, fmt.Errorf("failed to get receipt: %w", err)
	}

	signer := types.LatestSignerForChainID(tx.ChainId())
	from, err := types.Sender(signer, tx)
	if err != nil {
		return result, fmt.Errorf("failed to get sender: %w", err)
	}

	result = api.Transaction{
		TransactionHash:   tx.Hash().Hex(),
		TransactionStatus: int(receipt.Status),
		BlockHash:         receipt.BlockHash.Hex(),
		BlockNumber:       receipt.BlockNumber.Uint64(),
		From:              from.Hex(),
		To:                toAddress(tx.To()),
		ContractAddress:   contractAddress(receipt),
		LogsCount:         len(receipt.Logs),
		Input:             fmt.Sprintf("0x%x", tx.Data()),
		Value:             tx.Value().String(),
	}

	return result, nil
}

func toAddress(addr *common.Address) string {
	if addr == nil {
		return ""
	}
	return addr.Hex()
}

func contractAddress(receipt *types.Receipt) string {
	if receipt.ContractAddress == (common.Address{}) {
		return ""
	}
	return receipt.ContractAddress.Hex()
}

func RlpHexToHashList(rlpHex string) ([]string, error) {
	rlpHex = strings.TrimPrefix(rlpHex, "0x")
	rlpBytes, err := hex.DecodeString(rlpHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %v", err)
	}
	var txHashes [][]byte
	err = rlp.DecodeBytes(rlpBytes, &txHashes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode RLP: %v", err)
	}
	var hashStrings []string

	for _, hash := range txHashes {
		hashString := hex.EncodeToString(hash)
		if !strings.HasPrefix(hashString, "0x") {
			hashString = "0x" + hashString
		}
		hashStrings = append(hashStrings, hashString)
	}
	return hashStrings, nil
}
