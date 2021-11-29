package trh

import (
	"fmt"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/miner"
)

type TRH struct {
	miner      miner.Miner
	explorer   ddb.Explorer
	cache      *ddb.TXCache
	blockchain *ddb.Blockchain
}

func New() (*TRH, error) {
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return nil, fmt.Errorf("cannot create TXCache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	return &TRH{miner: taal, explorer: woc, cache: cache, blockchain: blockchain}, nil
}
