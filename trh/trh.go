package trh

import (
	"fmt"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
)

type TRH struct {
	miner      miner.Miner
	explorer   ddb.Explorer
	cache      *ddb.TXCache
	blockchain *ddb.Blockchain
	btrunk     *ddb.BTrunk
	keystore   *keys.Keystore
}

func New(keystore *keys.Keystore) (*TRH, error) {
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return nil, fmt.Errorf("cannot create TXCache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	btrunk := ddb.NewBTrunk(keystore.Source().Key(), keystore.Source().Address(), keystore.Source().Password(), blockchain)
	return &TRH{miner: taal, explorer: woc, cache: cache, blockchain: blockchain, btrunk: btrunk, keystore: keystore}, nil
}

func (t *TRH) SetKeystore(keystore *keys.Keystore) {
	t.keystore = keystore
}
