package trh

import (
	"fmt"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
)

func ListAll(keystore keys.KeyStore) error {
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	var mEntries map[string][]*ddb.MetaEntry
	btrunk := &ddb.BTrunk{MainKey: keystore.Key(keys.Main), MainAddress: keystore.Address(keys.Main), Blockchain: blockchain}
	mEntries, err = btrunk.ListEntries(keystore.Passwords(), false)
	if err != nil {
		return fmt.Errorf("error while listing MetaEntry for password: %w", err)
	}
	for pass, mes := range mEntries {
		fmt.Printf("Entry for password '%s':\n", pass)
		for i, me := range mes {
			fmt.Printf("%d found entry: %s\t%s\n", i, me.Name, me.Hash)
		}
	}
	return nil
}

func ListSinglePassword(keystore keys.KeyStore, password string) error {
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	var mEntries map[string][]*ddb.MetaEntry
	btrunk := &ddb.BTrunk{MainKey: keystore.Key(keys.Main), MainAddress: keystore.Address(keys.Main), Blockchain: blockchain}
	passmap := map[string][32]byte{password: keystore.Password(password)}
	mEntries, err = btrunk.ListEntries(passmap, false)
	if err != nil {
		return fmt.Errorf("error while listing MetaEntry for password: %w", err)
	}
	for pass, mes := range mEntries {
		fmt.Printf("Entry for password '%s':\n", pass)
		for i, me := range mes {
			fmt.Printf("%d found entry: %s\t%s\n", i, me.Name, me.Hash)
		}
	}
	return nil
}
