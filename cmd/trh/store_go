package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
	"github.com/ejfhp/ddb/satoshi"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdStore(args []string) error {
	tr := trace.New().Source("store.go", "", "cmdStore")
	flagset, options := newFlagset(storeCmd)
	err := flagset.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	if flagHelp {
		printHelp(storeCmd)
		return nil
	}
	opt, ok := areFlagConsistent(flagset, options)
	if !ok {
		return fmt.Errorf("flag combination invalid")
	}
	keystore, err := loadKeyStore()
	if err != nil {
		trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
		return fmt.Errorf("error while loading keystore: %w", err)
	}
	switch opt {
	case "file":
		woc := ddb.NewWOC()
		taal := miner.NewTAAL()
		cache, err := ddb.NewUserTXCache()
		if err != nil {
			return fmt.Errorf("cannot open cache")
		}
		blockchain := ddb.NewBlockchain(taal, woc, cache)
		btrunk := &ddb.BTrunk{MainKey: keystore.Key(keys.Main), MainAddress: keystore.Address(keys.Main), Blockchain: blockchain}
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
		node, err := keystore.NodeFromPassword(flagPassword)
		// password := passwordtoBytes(flagPassword)
		// bWIF, bAdd, err := keystore.AddNewKeyAndAddress(password)
		if err != nil {
			trail.Println(trace.Alert("failed to generate branch key and address").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to generate branch key and address: %w", err)
		}
		err = updateKeyStore(keystore)
		if err != nil {
			trail.Println(trace.Alert("failed to save current password in the keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to save the current password in the keystore: %w", err)
		}
		txs, err := btrunk.TXOfBranchedEntry(node.Key, node.Address, node.Password, ent, defaultHeader, satoshi.Satoshi(flagMaxSpend), false)
		if err != nil {
			trail.Println(trace.Alert("failed to generate txs for entry").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to generate txs for entry: %w", err)
		}
		totFee := satoshi.Satoshi(0)
		for i, t := range txs {
			_, _, fee, err := t.TotInOutFee()
			if err != nil {
				trail.Println(trace.Alert("failed to get fee from tx").Append(tr).UTC().Error(err))
				return fmt.Errorf("failed to get fee from tx num %d: %w", i, err)
			}
			totFee = totFee.Add(fee)
		}
		// for i, t := range txs {
		// 	fmt.Printf("\n%d:\n%s\n", i, t.ToString())
		// }
		ids, err := btrunk.Blockchain.Submit(txs)
		if err != nil {
			trail.Println(trace.Alert("failed to submit txs").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to submit txs: %w", err)
		}
		for id, success := range ids {
			fmt.Printf("%s: %s\n", id, success)
		}

	default:
		return fmt.Errorf("flag combination invalid")
	}
	return nil
}
