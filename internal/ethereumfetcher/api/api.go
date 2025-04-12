package api

type Transaction struct {
	TransactionHash   string `json:"transactionHash"`
	TransactionStatus int    `json:"transactionStatus"`
	BlockHash         string `json:"blockHash"`
	BlockNumber       uint64 `json:"blockNumber"`
	From              string `json:"from"`
	To                string `json:"to"`
	ContractAddress   string `json:"contractAddress"`
	LogsCount         int    `json:"logsCount"`
	Input             string `json:"input"`
	Value             string `json:"value"`
}
