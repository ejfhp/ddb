package trh

import (
	"fmt"
	"os"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdList(args []string) error {
	tr := trace.New().Source("list.go", "", "cmdList")
	flagset, options := newFlagset(listCmd)
	err := flagset.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	if flagHelp {
		printHelp(txCmd)
		return nil
	}
	opt, ok := areFlagConsistent(flagset, options)
	if !ok {
		return fmt.Errorf("flag combination invalid")
	}
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	keystore, err := loadKeyStore()
	if err != nil {
		trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
		return fmt.Errorf("error while loading keystore: %w", err)
	}
	var mEntries map[string][]*ddb.MetaEntry
	btrunk := &ddb.BTrunk{MainKey: keystore.Key(keys.Main), MainAddress: keystore.Address(keys.Main), Blockchain: blockchain}
	switch opt {
	case "pin":
		mEntries, err = btrunk.ListEntries(keystore.Passwords(), false)
		if err != nil {
			trail.Println(trace.Alert("error while listing MetaEntry").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while listing MetaEntry for password: %w", err)
		}
	case "password":
		passmap := map[string][32]byte{flagPassword: keystore.Password(flagPassword)}
		mEntries, err = btrunk.ListEntries(passmap, false)
		if err != nil {
			trail.Println(trace.Alert("error while listing MetaEntry").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while listing MetaEntry for password: %w", err)
		}

	default:
		return fmt.Errorf("flag combination invalid")
	}
	for pass, mes := range mEntries {
		fmt.Printf("Entry for password '%s':\n", pass)
		for i, me := range mes {
			fmt.Printf("%d found entry: %s\t%s\n", i, me.Name, me.Hash)
		}

	}
	return nil
}
