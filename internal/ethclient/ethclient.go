package ethclient

import (
	"context"
	"encoding/hex"
	"errors"
	"ethfetcher/internal/ethereumfetcher/api"
	"ethfetcher/internal/logger"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

//counterfeiter:generate . GEthClient
type GEthClient interface {
	TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error)
	TransactionReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error)
}
type EthereumClient struct {
	client GEthClient
}

func NewEthereumClient(client GEthClient) *EthereumClient {
	return &EthereumClient{client: client}
}

func (ec *EthereumClient) FetchTransactionByHash(ctx context.Context, hash string) (api.Transaction, error) {
	var result api.Transaction

	txHash := common.HexToHash(hash)
	tx, isPending, err := ec.client.TransactionByHash(ctx, txHash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
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

func (ec *EthereumClient) RlpHexToHashList(rlpHex string) ([]string, error) {
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
