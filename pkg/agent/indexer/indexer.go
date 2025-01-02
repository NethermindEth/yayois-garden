package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	contractYayoiCollection "github.com/NethermindEth/yayois-garden/pkg/bindings/YayoiCollection"
	contractYayoiFactory "github.com/NethermindEth/yayois-garden/pkg/bindings/YayoiFactory"
)

type Indexer struct {
	ethClient       IndexerEthClient
	collectionCache *CollectionCache

	contractAddress common.Address
	pollingInterval time.Duration
}

type IndexerEthClient interface {
	bind.ContractBackend
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainIDReader
}

type IndexerOptions struct {
	EthClient       IndexerEthClient
	ContractAddress common.Address
	PollingInterval time.Duration
}

type PromptSuggestion struct {
	Log    types.Log
	Sender common.Address
	Prompt string
}

type PromptAuctionFinished struct {
	Log    types.Log
	Winner common.Address
	Prompt string
}

func NewIndexer(opts IndexerOptions) (*Indexer, error) {
	factory, err := contractYayoiFactory.NewContractYayoiFactory(opts.ContractAddress, opts.EthClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create factory: %v", err)
	}

	if opts.PollingInterval == 0 {
		opts.PollingInterval = 15 * time.Second
	}

	return &Indexer{
		ethClient:       opts.EthClient,
		contractAddress: opts.ContractAddress,
		pollingInterval: opts.PollingInterval,
		collectionCache: NewCollectionCache(factory),
	}, nil
}

func (i *Indexer) GetContractAddress() common.Address {
	return i.contractAddress
}

func (i *Indexer) IndexEvents(ctx context.Context, promptSuggestedChan chan<- PromptSuggestion, promptAuctionFinishedChan chan<- PromptAuctionFinished) error {
	defer close(promptSuggestedChan)
	defer close(promptAuctionFinishedChan)

	latestBlock := uint64(0)
	fromBlock := new(big.Int)
	toBlock := new(big.Int)

	promptSuggestedLog := new(contractYayoiCollection.ContractYayoiCollectionPromptSuggested)
	promptAuctionFinishedLog := new(contractYayoiCollection.ContractYayoiCollectionPromptAuctionFinished)

	yayoiCollectionAbi, err := contractYayoiCollection.ContractYayoiCollectionMetaData.GetAbi()
	if err != nil {
		return fmt.Errorf("failed to get yayoi collection abi: %v", err)
	}

	for {
		select {
		case <-time.After(i.pollingInterval):
			currentBlock, err := i.ethClient.BlockNumber(ctx)
			if err != nil {
				return fmt.Errorf("failed to get current block number: %v", err)
			}

			// Skip if no new blocks
			if currentBlock <= latestBlock {
				slog.Debug("no new blocks", "latestBlock", latestBlock, "currentBlock", currentBlock)
				continue
			}

			fromBlock.SetUint64(latestBlock + 1)
			toBlock.SetUint64(currentBlock)

			// Filter for both PromptSuggested and PromptAuctionFinished events
			logs, err := i.ethClient.FilterLogs(ctx, ethereum.FilterQuery{
				FromBlock: fromBlock,
				ToBlock:   toBlock,
				Topics: [][]common.Hash{{
					yayoiCollectionAbi.Events["PromptSuggested"].ID,
					yayoiCollectionAbi.Events["PromptAuctionFinished"].ID,
				}},
			})
			if err != nil {
				return fmt.Errorf("failed to filter events: %v", err)
			}

			for _, event := range logs {
				selector := event.Topics[0]

				if promptSuggestedChan != nil && selector == yayoiCollectionAbi.Events["PromptSuggested"].ID {
					if err := unpackPromptSuggested(yayoiCollectionAbi, promptSuggestedLog, event); err != nil {
						slog.Warn("failed to unpack PromptSuggested event", "error", err)
						continue
					}

					promptSuggestedChan <- PromptSuggestion{
						Log:    event,
						Sender: promptSuggestedLog.Sender,
						Prompt: promptSuggestedLog.Prompt,
					}
				} else if promptAuctionFinishedChan != nil && selector == yayoiCollectionAbi.Events["PromptAuctionFinished"].ID {
					if err := unpackPromptAuctionFinished(yayoiCollectionAbi, promptAuctionFinishedLog, event); err != nil {
						slog.Warn("failed to unpack PromptAuctionFinished event", "error", err)
						continue
					}

					promptAuctionFinishedChan <- PromptAuctionFinished{
						Log:    event,
						Winner: promptAuctionFinishedLog.Winner,
						Prompt: promptAuctionFinishedLog.Prompt,
					}
				}
			}

			latestBlock = currentBlock
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func unpackPromptSuggested(contractAbi *abi.ABI, out *contractYayoiCollection.ContractYayoiCollectionPromptSuggested, log types.Log) error {
	out.Raw = log
	return unpackEvent(contractAbi, out, "PromptSuggested", log)
}

func unpackPromptAuctionFinished(contractAbi *abi.ABI, out *contractYayoiCollection.ContractYayoiCollectionPromptAuctionFinished, log types.Log) error {
	out.Raw = log
	return unpackEvent(contractAbi, out, "PromptAuctionFinished", log)
}

func unpackEvent(contractAbi *abi.ABI, out interface{}, event string, log types.Log) error {
	if len(log.Topics) == 0 {
		return fmt.Errorf("no event signature")
	}
	if log.Topics[0] != contractAbi.Events[event].ID {
		return fmt.Errorf("event signature mismatch")
	}
	if len(log.Data) > 0 {
		if err := contractAbi.UnpackIntoInterface(out, event, log.Data); err != nil {
			return fmt.Errorf("failed to unpack event: %v", err)
		}
	}
	var indexed abi.Arguments
	for _, arg := range contractAbi.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}
