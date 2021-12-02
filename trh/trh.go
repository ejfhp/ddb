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

func NewWithoutKeystore() *TRH {
	return &TRH{}
}

func (t *TRH) SetKeystore(keystore *keys.Keystore) error {
	t.explorer = ddb.NewWOC()
	t.miner = miner.NewTAAL()
	var err error
	t.cache, err = ddb.NewUserTXCache()
	if err != nil {
		return fmt.Errorf("cannot create cache")
	}
	t.blockchain = ddb.NewBlockchain(t.miner, t.explorer, t.cache)
	t.keystore = keystore
	t.btrunk = ddb.NewBTrunk(keystore.Source().Key(), keystore.Source().Address(), keystore.Source().Password(), t.blockchain)
	return nil
}
