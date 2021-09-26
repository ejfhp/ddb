package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdStore(args []string) error {
	tr := trace.New().Source("store.go", "", "cmdStore")
	flagset, options := newFlagset(commands["store"])
	fmt.Printf("cmdStore flags: %v\n", args[2:])
	err := flagset.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	if flagHelp {
		printHelp(flagset)
		return nil
	}
	opt := areFlagConsistent(flagset, options)
	keystore, err := loadKeyStore()
	if err != nil {
		trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
		return fmt.Errorf("error while loading keystore: %w", err)
	}
	switch opt {
	case "file":
		woc := ddb.NewWOC()
		taal := ddb.NewTAAL()
		blockchain := ddb.NewBlockchain(taal, woc, nil)
		btrunk := &ddb.BTrunk{BitcoinWIF: keystore.WIF, BitcoinAdd: keystore.Address, Blockchain: blockchain}
		lff := strings.Split(flagLabels, ",")
		labels := []string{}
		for _, l := range lff {
			labels = append(labels, strings.TrimSpace(strings.ToLower(l)))
		}
		ent, err := ddb.NewEntryFromFile(filepath.Base(flagFile), flagFile, labels, flagNotes)
		if err != nil {
			trail.Println(trace.Alert("failed to generate entry from file").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to generate entry from file: %w", err)
		}
		password := passwordtoBytes(flagPassword)
		bWIF, bAdd, err := btrunk.GenerateKeyAndAddress(password)
		txs, err := btrunk.TXOfBranchedEntry(bWIF, bAdd, password, ent, defaultHeader, 100000, false)
		if err != nil {
			trail.Println(trace.Alert("failed to store fee for file").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to store fee for file: %w", err)
		}
		fmt.Printf("Stored Fee: %d satoshi\n", fee)
	default:
		return fmt.Errorf("flag combination invalid")
	}
	return nil
}
