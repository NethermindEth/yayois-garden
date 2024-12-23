package indexer

import (
	"fmt"

	contractYayoiFactory "github.com/NethermindEth/yayois-garden/pkg/bindings/YayoiFactory"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type CollectionCache struct {
	cache   map[common.Address]bool
	factory *contractYayoiFactory.ContractYayoiFactory
}

func NewCollectionCache(factory *contractYayoiFactory.ContractYayoiFactory) *CollectionCache {
	return &CollectionCache{
		cache:   make(map[common.Address]bool),
		factory: factory,
	}
}

func (c *CollectionCache) IsCollectionRegistered(collectionAddress common.Address) (bool, error) {
	if _, ok := c.cache[collectionAddress]; ok {
		return true, nil
	}

	isRegistered, err := c.factory.IsRegisteredCollection(&bind.CallOpts{}, collectionAddress)
	if err != nil {
		return false, fmt.Errorf("failed to check if collection is registered: %v", err)
	}

	if isRegistered {
		c.cache[collectionAddress] = true
	}

	return isRegistered, nil
}
