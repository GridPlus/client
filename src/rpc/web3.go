// Functions and structs for interacting with the web3 API
package rpc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

const (
	JSON_MEDIA_TYPE = "application/json"
)

type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	ID      int64         `json:"id"`
	Params  []interface{} `json:"params"`
}

// ToJSON marshals a JSONRPCRequest into JSON
func (req *JSONRPCRequest) ToJSON() ([]byte, error) {
	s, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return s, nil
}

type ResponseBase struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id"`
}

type BlockNumberResponse struct {
	ResponseBase
	Result string `json:"result"`
}

type CallResponse struct {
	ResponseBase
	Result string `json:"result"`
}

type NewFilterResponse struct {
	ResponseBase
	Result string `json:"result"`
}

type GetFilterChangesResponse struct {
	ResponseBase
	Result []string `json:"result"`
}

type BlockResponse struct {
	ResponseBase
	Result BlockResult `json:"result"`
}

type TransactionResponse struct {
	ResponseBase
	Result TransactionResult `json:"result"`
}

type BlockResult struct {
	Author           string              `json:"author"` // Parity only
	Difficulty       string              `json:"difficulty"`
	ExtraData        string              `json:"extraData"`
	GasLimit         string              `json:"gasLimit"`
	GasUsed          string              `json:"gasUsed"`
	Hash             string              `json:"hash"`
	LogsBloom        string              `json:"logsBloom"`
	Miner            string              `json:"miner"`
	MixHash          string              `json:"mixHash"`
	Nonce            string              `json:"nonce"`
	Number           string              `json:"number"`
	ParentHash       string              `json:"parentHash"`
	ReceiptsRoot     string              `json:"receiptsRoot"`
	SealFields       []string            `json:"sealFields"` // Parity only
	SHA3Uncles       string              `json:"sha3Uncles"`
	Size             string              `json:"size"`
	StateRoot        string              `json:"stateRoot"`
	Timestamp        string              `json:"timestamp"`
	TotalDifficulty  string              `json:"totalDifficulty"`
	Transactions     []TransactionResult `json:"transactions"`
	TransactionsRoot string              `json:"transactionsRoot"`
	Uncles           []string            `json:"uncles"`
}

// ToBlock converts a BlockResult to a Block
func (blockResult *BlockResult) ToBlock() (*Block, error) {
	// string-to-integer conversions
	difficulty, err := strconv.ParseInt(blockResult.Difficulty, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("ToBlock Difficulty: %v", err)
	}

	gasLimit, err := strconv.ParseInt(blockResult.GasLimit, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToBlock GasLimit: %v", err)
	}

	gasUsed, err := strconv.ParseInt(blockResult.GasUsed, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToBlock GasUsed: %v", err)
	}

	nonce := new(big.Int)
	nonce.SetString(blockResult.Nonce, 0)

	number, err := strconv.ParseInt(blockResult.Number, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToBlock Number: %v", err)
	}

	size, err := strconv.ParseInt(blockResult.Size, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToBlock Size: %v", err)
	}

	timestamp, err := strconv.ParseInt(blockResult.Timestamp, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToBlock Timestamp: %v", err)
	}

	totalDifficulty := new(big.Int)
	totalDifficulty.SetString(blockResult.TotalDifficulty, 0)

	block := Block{
		Author:          blockResult.Author,
		Difficulty:      difficulty,
		ExtraData:       blockResult.ExtraData,
		GasLimit:        int(gasLimit),
		GasUsed:         int(gasUsed),
		Hash:            blockResult.Hash,
		LogsBloom:       blockResult.LogsBloom,
		Miner:           blockResult.Miner,
		MixHash:         blockResult.MixHash,
		Nonce:           nonce,
		Number:          int(number),
		ParentHash:      blockResult.ParentHash,
		ReceiptsRoot:    blockResult.ReceiptsRoot,
		SealFields:      blockResult.SealFields,
		SHA3Uncles:      blockResult.SHA3Uncles,
		Size:            int(size),
		StateRoot:       blockResult.StateRoot,
		Timestamp:       int(timestamp),
		TotalDifficulty: totalDifficulty,
		// Transactions
		TransactionsRoot: blockResult.TransactionsRoot,
		Uncles:           blockResult.Uncles,
	}

	// populate the transactions in the block
	for _, resultTx := range blockResult.Transactions {
		tx, err := resultTx.ToTransaction()
		if err != nil {
			return nil, err
		}
		block.Transactions = append(block.Transactions, *tx)
	}

	return &block, nil
}

type TransactionResult struct {
	BlockHash        *string `json:"blockHash"`   // null for pending tx
	BlockNumber      *string `json:"blockNumber"` // null for pending tx
	Creates          *string `json:"creates"`     // Parity only; null when not creating contract
	From             string  `json:"from"`
	Gas              string  `json:"gas"`
	GasPrice         string  `json:"gasPrice"`
	Hash             string  `json:"hash"`
	Input            string  `json:"input"`
	NetworkId        *int    `json:"networkId"` // Parity only
	Nonce            string  `json:"nonce"`
	PublicKey        string  `json:"publicKey"` // Parity only
	R                string  `json:"r"`
	Raw              string  `json:"raw"` // Parity only
	S                string  `json:"s"`
	StandardV        string  `json:"standardV"`        // Parity only
	To               *string `json:"to"`               // null when creating contract
	TransactionIndex *string `json:"transactionIndex"` // null for pending tx
	V                string  `json:"v"`
	Value            string  `json:"value"`
}

// ToTransaction converts a TransactionResult to a Transaction
func (txResult *TransactionResult) ToTransaction() (*Transaction, error) {
	blockNumber, err := strconv.ParseInt(*txResult.BlockNumber, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToTransaction BlockNumber: %v", err)
	}
	blockNumberInt := int(blockNumber)

	gas, err := strconv.ParseInt(txResult.Gas, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToTransaction Gas: %v", err)
	}

	gasPrice := new(big.Int)
	gasPrice.SetString(txResult.GasPrice, 0)

	nonce, err := strconv.ParseInt(txResult.Nonce, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToTransaction Nonce: %v", err)
	}

	standardV, err := strconv.ParseInt(txResult.StandardV, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToTransaction StandardV: %v", err)
	}

	transactionIndex, err := strconv.ParseInt(*txResult.TransactionIndex, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToTransaction TransactionIndex: %v", err)
	}
	transactionIndexInt := int(transactionIndex)

	v, err := strconv.ParseInt(txResult.V, 0, 32)
	if err != nil {
		return nil, fmt.Errorf("ToTransaction V: %v", err)
	}

	value := new(big.Int)
	value.SetString(txResult.Value, 0)

	tx := Transaction{
		BlockHash:        txResult.BlockHash,
		BlockNumber:      &blockNumberInt,
		Creates:          txResult.Creates,
		From:             txResult.From,
		Gas:              int(gas),
		GasPrice:         gasPrice,
		Hash:             txResult.Hash,
		Input:            txResult.Input,
		NetworkId:        txResult.NetworkId,
		Nonce:            int(nonce),
		PublicKey:        txResult.PublicKey,
		R:                txResult.R,
		Raw:              txResult.Raw,
		S:                txResult.S,
		StandardV:        int(standardV),
		To:               txResult.To,
		TransactionIndex: &transactionIndexInt,
		V:                int(v),
		Value:            value,
	}
	return &tx, nil
}

type Block struct {
	Author           string        `json:"author"`
	Difficulty       int64         `json:"difficulty"`
	ExtraData        string        `json:"extra_data"`
	GasLimit         int           `json:"gas_limit"`
	GasUsed          int           `json:"gas_used"`
	Hash             string        `json:"hash"`
	LogsBloom        string        `json:"logs_bloom"`
	Miner            string        `json:"miner"`
	MixHash          string        `json:"mix_hash"`
	Nonce            *big.Int      `json:"nonce"`
	Number           int           `json:"number"`
	ParentHash       string        `json:"parent_hash"`
	ReceiptsRoot     string        `json:"receipts_root"`
	SealFields       []string      `json:"seal_fields"`
	SHA3Uncles       string        `json:"sha3_uncles"`
	Size             int           `json:"size"`
	StateRoot        string        `json:"state_root"`
	Timestamp        int           `json:"timestamp"`
	TotalDifficulty  *big.Int      `json:"total_difficulty"`
	Transactions     []Transaction `json:"transactions"`
	TransactionsRoot string        `json:"transactions_root"`
	Uncles           []string      `json:"uncles"`
}

// ToJSON marshals a Block into JSON
func (block *Block) ToJSON() ([]byte, error) {
	s, err := json.Marshal(block)
	if err != nil {
		return nil, err
	}
	return s, nil
}

type Transaction struct {
	BlockHash        *string  `json:"block_hash"`
	BlockNumber      *int     `json:"block_number"`
	Creates          *string  `json:"creates"`
	From             string   `json:"from"`
	Gas              int      `json:"gas"`
	GasPrice         *big.Int `json:"gas_price"`
	Hash             string   `json:"hash"`
	Input            string   `json:"input"`
	NetworkId        *int     `json:"network_id"`
	Nonce            int      `json:"nonce"`
	PublicKey        string   `json:"public_key"`
	R                string   `json:"r"`
	Raw              string   `json:"raw"`
	S                string   `json:"s"`
	StandardV        int      `json:"standard_v"`
	To               *string  `json:"to"`
	TransactionIndex *int     `json:"transaction_index"`
	V                int      `json:"v"`
	Value            *big.Int `json:"value"`
}

// ToJSON marshals a Transaction into JSON
func (tx *Transaction) ToJSON() ([]byte, error) {
	s, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

type EthereumClient struct {
	URL string
}

// issueRequest issues the JSON-RPC request
func (client *EthereumClient) issueRequest(reqBody *JSONRPCRequest) ([]byte, error) {

	payload, err := reqBody.ToJSON()
	if err != nil {
		return nil, err
	}

	reader := strings.NewReader(string(payload))
	resp, err := http.Post(client.URL, JSON_MEDIA_TYPE, reader)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Eth_newBlockFilter calls the eth_newBlockFilter JSON-RPC method
func (client *EthereumClient) Eth_newBlockFilter() (string, error) {

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_newBlockFilter",
		Params:  nil,
	}

	body, err := client.issueRequest(&reqBody)
	if err != nil {
		return "", err
	}

	var clientResp NewFilterResponse
	err = json.Unmarshal(body, &clientResp)
	if err != nil {
		return "", err
	}

	return clientResp.Result, nil
}

// Eth_newPendingTransactionFilter calls the eth_newPendingTransactionFilter JSON-RPC method
func (client *EthereumClient) Eth_newPendingTransactionFilter() (string, error) {

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_newPendingTransactionFilter",
		Params:  nil,
	}

	body, err := client.issueRequest(&reqBody)
	if err != nil {
		return "", err
	}

	var clientResp NewFilterResponse
	err = json.Unmarshal(body, &clientResp)
	if err != nil {
		return "", err
	}

	return clientResp.Result, nil
}

// Eth_getFilterChanges calls the eth_getFilterChanges JSON-RPC method
func (client *EthereumClient) Eth_getFilterChanges(filterID string) ([]string, error) {

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_getFilterChanges",
		Params:  []interface{}{filterID},
	}

	body, err := client.issueRequest(&reqBody)
	if err != nil {
		return nil, err
	}

	var clientResp GetFilterChangesResponse
	err = json.Unmarshal(body, &clientResp)
	if err != nil {
		return nil, err
	}

	return clientResp.Result, nil
}

// Eth_getBlockByHash calls the eth_getBlockByHash JSON-RPC method
func (client *EthereumClient) Eth_getBlockByHash(blockHash string, full bool) (*Block, error) {

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_getBlockByHash",
		Params:  []interface{}{blockHash, full},
	}

	body, err := client.issueRequest(&reqBody)
	if err != nil {
		return nil, err
	}

	var clientResp BlockResponse
	err = json.Unmarshal(body, &clientResp)
	if err != nil {
		return nil, err
	}

	block, err := clientResp.Result.ToBlock()
	if err != nil {
		return nil, err
	}

	return block, nil
}

// Eth_getTransactionByHash calls the eth_getTransactionByHash JSON-RPC method
func (client *EthereumClient) Eth_getTransactionByHash(txHash string) (*Transaction, error) {

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_getTransactionByHash",
		Params:  []interface{}{txHash},
	}

	body, err := client.issueRequest(&reqBody)
	if err != nil {
		return nil, err
	}

	var clientResp TransactionResponse
	err = json.Unmarshal(body, &clientResp)
	if err != nil {
		return nil, err
	}

	tx, err := clientResp.Result.ToTransaction()
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// Eth_getBlockByNumber calls the eth_getBlockByNumber JSON-RPC method
func (client *EthereumClient) Eth_getBlockByNumber(blockNumber int, full bool) (*Block, error) {

	blockNumberHex := "0x" + strconv.FormatInt(int64(blockNumber), 16)

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_getBlockByNumber",
		Params:  []interface{}{blockNumberHex, full},
	}

	body, err := client.issueRequest(&reqBody)
	if err != nil {
		return nil, err
	}

	var clientResp BlockResponse
	err = json.Unmarshal(body, &clientResp)
	if err != nil {
		return nil, err
	}

	block, err := clientResp.Result.ToBlock()
	if err != nil {
		return nil, err
	}

	return block, nil
}

// Eth_blockNumber calls the eth_blockNumber JSON-RPC method
func (client *EthereumClient) Eth_blockNumber() (int, error) {

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_blockNumber",
		Params:  []interface{}{},
	}

	body, err := client.issueRequest(&reqBody)
	if err != nil {
		return 0, err
	}

	var clientResp BlockNumberResponse
	err = json.Unmarshal(body, &clientResp)
	if err != nil {
		return 0, err
	}

	blockNumber, err := strconv.ParseInt(clientResp.Result, 0, 32)
	if err != nil {
		return 0, err
	}

	return int(blockNumber), nil
}

// Get the network version (a.k.a. chainId)
func (client *EthereumClient) NetVersion() (int64, error) {

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "net_version",
		Params:  []interface{}{},
	}

	body, err := client.issueRequest(&reqBody)
	if err != nil {
		return 0, err
	}

	var clientResp BlockNumberResponse
	err = json.Unmarshal(body, &clientResp)
	if err != nil {
		return 0, err
	}
	version, err := strconv.ParseInt(clientResp.Result, 10, 64)
	// if err != nil {
	// 	return 0, err
	// }

	return int64(version), nil
}

// Use this to make calls to a contract
type Call struct {
	From string `json:"from"`
	To string `json:"to"`
	Data string `json:"data"`
}

// Make a call to the blockchain
func (client *EthereumClient) Eth_call(_call Call) (string, error) {
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_call",
		Params:  []interface{}{_call},
	}

	res, err := client.issueRequest(&reqBody)
	if err != nil {
		return "", err
	}

	var clientResp CallResponse
	err = json.Unmarshal(res, &clientResp)
	if err != nil {
		return "", err
	}
	return clientResp.Result, nil
}

// Get the transaction count (nonce) for an account, which must be 0x prefixed
func (client *EthereumClient) Eth_getTransactionCount(addr string) (string, error) {
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_getTransactionCount",
		Params:  []interface{}{addr, "latest"},
	}

	res, err := client.issueRequest(&reqBody)
	if err != nil {
		return "", err
	}

	var clientResp CallResponse // We can reuse this struct
	err = json.Unmarshal(res, &clientResp)
	if err != nil {
		return "", err
	}
	return clientResp.Result, nil
}

type TxResponse struct {
	ResponseBase
	Result string `json:"result"`
}

func (client *EthereumClient) Eth_sendRawTransaction(data string) (string, error) {
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_sendRawTransaction",
		Params:  []interface{}{data},
	}
	res2, err := client.issueRequest(&reqBody)
	if err != nil {
		return "", err
	}

	var clientResp TxResponse // We can reuse this struct
	err = json.Unmarshal(res2, &clientResp)
	if err != nil {
		return "", err
	}
	return clientResp.Result, nil
}

type Receipt struct {
	TransactionHash string `json:"transactionHash"`
	TransactionIndex string `json:"transactionIndex"`
	BlockNumber string `json:"blockNumber"`
	BlockHash string `json:"blockHash"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	ContractAddress string `json:"contractAddress"`
}

type ReceiptResponse struct {
	ResponseBase
	Result Receipt
}

// Get the cumulative gas used from a receipt given a txhash
// This is a good way to check if your transaction threw.
// If cumulativeGasUsed = max gas allowed, your contract probably threw
func (client *EthereumClient) Eth_gasUsed(txhash string) (string, error) {
	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_getTransactionReceipt",
		Params:  []interface{}{txhash},
	}
	res2, err := client.issueRequest(&reqBody)
	if err != nil {
		return "", err
	}

	var clientResp ReceiptResponse
	err = json.Unmarshal(res2, &clientResp)
	if err != nil {
		return "", err
	}
	// if clientResp.Result.CumulativeGasUsed == "" {
	// 	return "", nil
	// }

	return clientResp.Result.CumulativeGasUsed, nil
}
