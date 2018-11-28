package client

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math/big"
	"time"

	"github.com/tenderly/tenderly-trace/ethereum"
	"github.com/tenderly/tenderly-trace/ethereum/geth"
	"github.com/tenderly/tenderly-trace/ethereum/parity"

	"github.com/tenderly/tenderly-trace/jsonrpc2"
)

// Client represents an implementation agnostic interface to the Ethereum node.
// It is able connect to both different protocols (http, ws) and implementations (geth, parity).
type Client struct {
	rpc    *jsonrpc2.Client
	schema ethereum.Schema

	openChannels []chan int64
}

func Dial(target string) (*Client, error) {
	rpcClient, err := jsonrpc2.DiscoverAndDial(target)
	if err != nil {
		return nil, fmt.Errorf("dial ethereum rpc: %s", err)
	}

	nodeType := "geth"

	req, resp := parity.DefaultSchema.Parity().VersionInfo()
	if err = rpcClient.CallRequest(resp, req); err == nil {
		nodeType = "parity"
	}

	var schema ethereum.Schema
	switch nodeType {
	case "geth":
		schema = &geth.DefaultSchema
	case "parity":
		schema = &parity.DefaultSchema
	default:
		return nil, fmt.Errorf("unsupported node type: %s", err)
	}

	return &Client{
		rpc:    rpcClient,
		schema: schema,
	}, nil
}

func (c *Client) Call(message *jsonrpc2.Message) error {
	var params []interface{}
	if message.Params != nil {
		err := json.Unmarshal(message.Params, &params)
		if err != nil {
			return err
		}
	}

	req := jsonrpc2.NewRequest(message.Method, params...)

	resMsg, err := c.rpc.SendRawRequest(req)
	if err != nil {
		return fmt.Errorf("proxy calling failed method: [%s], parameters [%s], error: %s",
			req.Method,
			req.Params,
			err)
	}

	message.Result = resMsg.Result
	message.Error = resMsg.Error
	return nil
}

func (c *Client) CurrentBlockNumber() (int64, error) {
	req, resp := c.schema.Eth().BlockNumber()

	err := c.rpc.CallRequest(resp, req)
	if err != nil {
		return 0, fmt.Errorf("current block number: %s", err)
	}

	return resp.Value(), nil
}

func (c *Client) GetBlock(number int64) (ethereum.Block, error) {
	req, resp := c.schema.Eth().GetBlockByNumber(ethereum.Number(number))

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return nil, fmt.Errorf("get block by number [%d]: %s", number, err)
	}

	return resp, nil
}

func (c *Client) GetBlockByHash(hash string) (ethereum.BlockHeader, error) {
	req, resp := c.schema.Eth().GetBlockByHash(hash)

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return nil, fmt.Errorf("get block by hash [%s]: %s", hash, err)
	}

	return resp, nil
}

func (c *Client) GetTransaction(hash string) (ethereum.Transaction, error) {
	req, resp := c.schema.Eth().GetTransaction(hash)

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return nil, fmt.Errorf("get transaction [%s]: %s", hash, err)
	}

	return resp, nil
}

func (c *Client) GetTransactionReceipt(hash string) (ethereum.TransactionReceipt, error) {
	req, resp := c.schema.Eth().GetTransactionReceipt(hash)

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return nil, fmt.Errorf("get transaction receipt [%s]: %s", hash, err)
	}

	return resp, nil
}

func (c *Client) GetNetworkID() (string, error) {
	req, resp := c.schema.Net().Version()

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return "", fmt.Errorf("get network ID: %s", err)
	}

	return *resp, nil
}

func (c *Client) GetTransactionVMTrace(tx ethereum.Transaction) (ethereum.TransactionStates, error) {
	req, resp := c.schema.Trace().VMTrace(tx.Hash().String())

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return nil, fmt.Errorf("get transaction trace [%s]: %s", tx.Hash().String(), err)
	}

	resp.ProcessTrace(tx)

	return resp, nil
}

func (c *Client) GetTransactionCallTrace(hash string) (ethereum.CallTraces, error) {
	req, resp := c.schema.Trace().CallTrace(hash)

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return nil, fmt.Errorf("get transaction pretty trace [%s]: %s", hash, err)
	}

	return resp, nil
}

func (c *Client) GetBalance(address string, block ethereum.Number) (*big.Int, error) {
	req, resp := c.schema.Eth().GetBalance(address, block)

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return nil, fmt.Errorf("get code [%s]: %s", address, err)
	}

	return resp.ToInt(), nil
}

func (c *Client) GetCode(address string, block ethereum.Number) (*string, error) {
	req, resp := c.schema.Eth().GetCode(address, block)

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return nil, fmt.Errorf("get code [%s]: %s", address, err)
	}

	return resp, nil
}

func (c *Client) GetStorageAt(hash string, offset common.Hash, block ethereum.Number) (*common.Hash, error) {
	req, resp := c.schema.Eth().GetStorage(hash, offset, block)

	if err := c.rpc.CallRequest(resp, req); err != nil {
		return nil, fmt.Errorf("get transaction receipt [%s]: %s", hash, err)
	}

	return resp, nil
}

func (c *Client) Subscribe(forcePoll bool) (chan int64, error) {
	if forcePoll {
		log.Printf("Forcing polling subscription...")
		return c.subscribeViaPoll()
	}

	//@TODO: Manage closing of the subscription.
	req, subscription := c.schema.PubSub().Subscribe()
	err := c.rpc.CallRequest(subscription, req)
	if err != nil {
		//@TODO: Do specific check if subscription not supported.
		log.Printf("Subscription not supported, falling back to polling")
		return c.subscribeViaPoll()
	}

	return c.subscribe(subscription)
}

func (c *Client) subscribeViaPoll() (chan int64, error) {
	outCh := make(chan int64)

	go func() {
		var lastBlock int64

		for {
			blockNumber, err := c.CurrentBlockNumber()
			if err != nil {
				log.Printf("failed pollig for last block number: %s", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if lastBlock == 0 {
				lastBlock = blockNumber
				continue
			}

			for lastBlock < blockNumber {
				outCh <- blockNumber

				lastBlock++
			}

			time.Sleep(200 * time.Millisecond)
		}
	}()

	return outCh, nil
}

func (c *Client) subscribe(id *ethereum.SubscriptionID) (chan int64, error) {
	outCh := make(chan int64)

	inCh, err := c.rpc.Subscribe(id.String())
	if err != nil {
		return nil, fmt.Errorf("listen for subscriptions: %s", err)
	}

	go func() {
		for msg := range inCh {
			var resp geth.SubscriptionResult
			err = json.Unmarshal(msg.Params, &resp)
			if err != nil {
				log.Printf("failed reading notification: %s", err)
				continue
			}

			outCh <- resp.Result.Number().Value()
		}

		close(outCh)
	}()

	return outCh, nil
}

func (c *Client) Close() error {
	c.rpc.Close()

	return nil
}
