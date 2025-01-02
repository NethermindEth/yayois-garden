package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/alitto/pond/v2"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/sync/singleflight"

	contractYayoiCollection "github.com/NethermindEth/yayois-garden/pkg/bindings/YayoiCollection"
	contractYayoiFactory "github.com/NethermindEth/yayois-garden/pkg/bindings/YayoiFactory"
)

const (
	indexingLogChunkSize         = 10000
	initializeCollectionPoolSize = 100
)

type CollectionInfo struct {
	NextAuctionIdInitialized bool
	MetadataInitialized      bool

	CreationTimestamp uint64
	CollectionAddress common.Address
	AuctionDuration   uint64
	NextAuctionId     uint64
}

func (c *CollectionInfo) Initialized() bool {
	return c.NextAuctionIdInitialized && c.MetadataInitialized
}

type AuctionEnd struct {
	AuctionId         uint64
	CollectionAddress common.Address
	Winner            common.Address
	Prompt            string
}

type IndexerEthClient interface {
	bind.ContractBackend
	ethereum.LogFilterer
	ethereum.BlockNumberReader
}

type IndexerClock interface {
	Now() time.Time
}

type IndexerConfig struct {
	EthClient              IndexerEthClient
	FactoryAddress         common.Address
	EventPollingInterval   time.Duration
	AuctionPollingInterval time.Duration
	Clock                  IndexerClock
}

type Indexer struct {
	group                    singleflight.Group
	initializeCollectionPool pond.Pool

	cache map[common.Address]*CollectionInfo

	factoryAbi    *abi.ABI
	collectionAbi *abi.ABI

	factoryAddress common.Address
	factory        *contractYayoiFactory.ContractYayoiFactory

	provider IndexerEthClient

	lastIndexedBlock       uint64
	eventPollingInterval   time.Duration
	auctionPollingInterval time.Duration
	clock                  IndexerClock
}

func NewIndexer(opts IndexerConfig) (*Indexer, error) {
	slog.Info("creating new indexer", "factoryAddress", opts.FactoryAddress, "eventPollingInterval", opts.EventPollingInterval, "auctionPollingInterval", opts.AuctionPollingInterval)

	factory, err := contractYayoiFactory.NewContractYayoiFactory(opts.FactoryAddress, opts.EthClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get factory: %v", err)
	}

	factoryAbi, err := contractYayoiFactory.ContractYayoiFactoryMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get factory ABI: %v", err)
	}

	collectionAbi, err := contractYayoiCollection.ContractYayoiCollectionMetaData.GetAbi()
	if err != nil {
		return nil, fmt.Errorf("failed to get collection ABI: %v", err)
	}

	indexer := &Indexer{
		group:                    singleflight.Group{},
		initializeCollectionPool: pond.NewPool(initializeCollectionPoolSize),

		cache: make(map[common.Address]*CollectionInfo),

		factoryAbi:    factoryAbi,
		collectionAbi: collectionAbi,

		factoryAddress: opts.FactoryAddress,
		factory:        factory,

		lastIndexedBlock: 0,
		provider:         opts.EthClient,

		eventPollingInterval:   opts.EventPollingInterval,
		auctionPollingInterval: opts.AuctionPollingInterval,
		clock:                  opts.Clock,
	}

	slog.Info("indexer created successfully")
	return indexer, nil
}

func (i *Indexer) Start(ctx context.Context, auctionEndChan chan<- AuctionEnd) {
	slog.Info("starting indexer")
	i.indexEvents(ctx)

	go i.indexEventsTask(ctx)
	go i.monitorAuctionsTask(ctx, auctionEndChan)
	slog.Info("indexer tasks started")
}

func (i *Indexer) monitorAuctionsTask(ctx context.Context, auctionEndChan chan<- AuctionEnd) {
	slog.Info("starting auction monitor task")
	ticker := time.NewTicker(i.auctionPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := uint64(i.clock.Now().Unix())

			for addr, info := range i.cache {
				slog.Info("monitoring auction", "collection", addr, "info", info, "now", now)

				if !info.Initialized() {
					continue
				}

				auctionEnd := info.CreationTimestamp + (info.NextAuctionId * info.AuctionDuration)
				for ; auctionEnd <= now; auctionEnd += info.AuctionDuration {
					currentAuctionId := info.NextAuctionId - 1
					slog.Info("auction ended", "collection", addr, "auctionId", currentAuctionId)

					info.NextAuctionId++

					go func() {
						collection, err := contractYayoiCollection.NewContractYayoiCollection(addr, i.provider)
						if err != nil {
							slog.Error("failed to get collection", "error", err)
							return
						}

						auction, err := collection.GetAuction(&bind.CallOpts{Context: ctx}, big.NewInt(int64(currentAuctionId)))
						if err != nil {
							slog.Error("failed to get auction", "error", err)
							return
						}

						slog.Info("processing auction end",
							"collection", addr,
							"auctionId", currentAuctionId,
							"winner", auction.HighestBidder,
							"prompt", auction.Prompt,
						)

						if auction.HighestBidder != (common.Address{}) {
							auctionEndChan <- AuctionEnd{
								CollectionAddress: addr,
								AuctionId:         currentAuctionId,
								Prompt:            auction.Prompt,
								Winner:            auction.HighestBidder,
							}
						}
					}()
				}
			}
		case <-ctx.Done():
			slog.Info("auction monitor task stopping")
			return
		}
	}
}

func (i *Indexer) indexEventsTask(ctx context.Context) {
	slog.Info("starting event indexing task")
	ticker := time.NewTicker(i.eventPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := i.indexEvents(ctx)
			if err != nil {
				slog.Error("failed to index events", "error", err)
			}
		case <-ctx.Done():
			slog.Info("event indexing task stopping")
			return
		}
	}
}

func (i *Indexer) indexEvents(ctx context.Context) error {
	targetBlock, err := i.provider.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current block: %v", err)
	}

	slog.Info("indexing events", "fromBlock", i.lastIndexedBlock, "toBlock", targetBlock)

	collectionCreatedId := i.factoryAbi.Events["CollectionCreated"].ID
	promptAuctionFinishedId := i.collectionAbi.Events["PromptAuctionFinished"].ID

	fromBlockBI := new(big.Int)
	toBlockBI := new(big.Int)

	discoveredCollections := []common.Address{}

	fromBlock := i.lastIndexedBlock
	for fromBlock <= targetBlock {
		toBlock := fromBlock + indexingLogChunkSize
		if toBlock > targetBlock {
			toBlock = targetBlock
		}

		fromBlockBI.SetUint64(fromBlock)
		toBlockBI.SetUint64(toBlock)

		logs, err := i.provider.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: fromBlockBI,
			ToBlock:   toBlockBI,
			Topics: [][]common.Hash{{
				collectionCreatedId,
				promptAuctionFinishedId,
			}},
		})
		if err != nil {
			return fmt.Errorf("failed to filter logs: %v", err)
		}

		slog.Info("processing logs", "count", len(logs), "fromBlock", fromBlock, "toBlock", toBlock)

		for _, log := range logs {
			if log.Topics[0] == collectionCreatedId && log.Address == i.factoryAddress {
				var event contractYayoiFactory.ContractYayoiFactoryCollectionCreated
				err = unpackLog(i.factoryAbi, &event, "CollectionCreated", log)
				if err != nil {
					slog.Error("failed to unpack CollectionCreated event", "error", err)
					continue
				}

				slog.Info("new collection created", "collection", event.Collection)
				i.cacheCollectionKey(event.Collection)
				discoveredCollections = append(discoveredCollections, event.Collection)

				i.initializeCollectionPool.Submit(func() {
					err = i.initializeCollection(ctx, event.Collection)
					if err != nil {
						slog.Error("failed to initialize collection", "collection", event.Collection.String(), "error", err)
					}
				})
			} else if log.Topics[0] == promptAuctionFinishedId {
				if !i.isCollectionKeyCached(log.Address) {
					slog.Warn("collection is not cached", "collection", log.Address.String())
					continue
				}

				var event contractYayoiCollection.ContractYayoiCollectionPromptAuctionFinished
				err = unpackLog(i.collectionAbi, &event, "PromptAuctionFinished", log)
				if err != nil {
					slog.Error("failed to unpack PromptAuctionFinished event", "error", err)
					continue
				}

				slog.Info("prompt auction finished", "collection", log.Address, "auctionId", event.AuctionId)

				info := i.getCollectionInfo(log.Address)
				if !info.NextAuctionIdInitialized {
					info.NextAuctionId = event.AuctionId.Uint64() + 1
				}
			}
		}

		fromBlock = toBlock + 1
	}

	for _, collection := range discoveredCollections {
		info := i.getCollectionInfo(collection)
		info.NextAuctionIdInitialized = true
		slog.Info("initialized next auction ID", "collection", collection)
	}

	i.lastIndexedBlock = targetBlock
	slog.Info("finished indexing events", "lastIndexedBlock", targetBlock)

	return nil
}

func (i *Indexer) initializeCollection(ctx context.Context, collectionAddress common.Address) error {
	slog.Info("initializing collection", "collection", collectionAddress)
	info := i.getCollectionInfo(collectionAddress)

	collection, err := contractYayoiCollection.NewContractYayoiCollection(collectionAddress, i.provider)
	if err != nil {
		return fmt.Errorf("failed to get collection: %v", err)
	}

	creationTimestamp, err := collection.CreationTimestamp(&bind.CallOpts{Context: ctx})
	if err != nil {
		return fmt.Errorf("failed to get creation timestamp: %v", err)
	}

	auctionDuration, err := collection.AuctionDuration(&bind.CallOpts{Context: ctx})
	if err != nil {
		return fmt.Errorf("failed to get auction duration: %v", err)
	}

	info.MetadataInitialized = true
	info.CollectionAddress = collectionAddress
	info.CreationTimestamp = creationTimestamp
	info.AuctionDuration = auctionDuration

	slog.Info("collection initialized",
		"collection", collectionAddress,
		"creationTimestamp", creationTimestamp,
		"auctionDuration", auctionDuration)

	return nil
}

func (i *Indexer) getCollectionInfo(collectionAddress common.Address) *CollectionInfo {
	info, _, _ := i.group.Do(collectionAddress.String(), func() (interface{}, error) {
		info, ok := i.cache[collectionAddress]
		if !ok {
			info = &CollectionInfo{}
			i.cache[collectionAddress] = info
			slog.Info("created new collection info", "collection", collectionAddress)
		}

		return info, nil
	})

	return info.(*CollectionInfo)
}

func (i *Indexer) cacheCollectionKey(collectionAddress common.Address) {
	_ = i.getCollectionInfo(collectionAddress)
}

func (i *Indexer) isCollectionKeyCached(collectionAddress common.Address) bool {
	_, ok := i.cache[collectionAddress]
	return ok
}

func unpackLog(contractAbi *abi.ABI, out interface{}, event string, log types.Log) error {
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
