package trh

import (
	"fmt"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
)

//ListAll returns an array of all the entries found
func (t *TRH) ListAll(keystore *keys.Keystore) (map[string][]*ddb.MetaEntry, error) {
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return nil, fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	var mEntries map[string][]*ddb.MetaEntry
	btrunk := &ddb.BTrunk{MainKey: keystore.Key(keys.Main), MainAddress: keystore.Address(keys.Main), Blockchain: blockchain}
	mEntries, err = btrunk.ListEntries(keystore.Passwords(), false)
	if err != nil {
		return nil, fmt.Errorf("error while listing MetaEntry for password: %w", err)
	}
	return mEntries, nil
}

func (t *TRH) ListSinglePassword(keystore *keys.Keystore, password string) (map[string][]*ddb.MetaEntry, error) {
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return nil, fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	var mEntries map[string][]*ddb.MetaEntry
	btrunk := &ddb.BTrunk{MainKey: keystore.Key(keys.Main), MainAddress: keystore.Address(keys.Main), Blockchain: blockchain}
	passmap := map[string][32]byte{password: keystore.Password(password)}
	mEntries, err = btrunk.ListEntries(passmap, false)
	if err != nil {
		return nil, fmt.Errorf("error while listing MetaEntries: %w", err)
	}
	return mEntries, nil
}
