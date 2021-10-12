package main

import (
	"fmt"
	"os"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/miner"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdCollect(args []string) error {
	tr := trace.New().Source("collect.go", "", "cmdCollect")
	flagset, options := newFlagset(collectCmd)
	err := flagset.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	if flagHelp {
		printHelp(collectCmd)
		return nil
	}
	opt, ok := areFlagConsistent(flagset, options)
	if !ok {
		return fmt.Errorf("flag combination invalid")
	}
	var keystore *ddb.KeyStore
	switch opt {
	case "pin":
		keystore, err = loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while opening keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while opening keystore: %w", err)
		}
	default:
		return fmt.Errorf("flag combination invalid")
	}
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	btrunk := &ddb.BTrunk{BitcoinWIF: keystore.WIF, BitcoinAdd: keystore.Address, Blockchain: blockchain}

	utxos := make(map[string][]*ddb.UTXO)
	for _, pwd := range keystore.Passwords {
		k, a, err := btrunk.GenerateKeyAndAddress(pwd)
		if err != nil {
			trail.Println(trace.Alert("error while generating key/address from password").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while generating key/address from password: %w", err)
		}
		u, err := blockchain.GetUTXO(a)
		if err != nil && err.Error() != "found no UTXO" {
			trail.Println(trace.Alert("error while retrieving UTXO").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while retrieving UTXO for address %s: %w", a, err)
		}
		if len(u) > 0 {
			utxos[k] = u
		}
	}
	if len(utxos) > 0 {
		fee, err := blockchain.EstimateStandardTXFee(len(utxos))
		if err != nil {
			trail.Println(trace.Alert("error while estimating collecting tx fee").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while estimating collecting tx fee: %w", err)
		}
		collectingTX, err := ddb.NewMultiInputTX(keystore.Address, utxos, fee)
		if err != nil {
			trail.Println(trace.Alert("error while building collecting TX").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while building collecting TX: %w", err)
		}
		fmt.Printf("Collecting TX :\n%s\n", collectingTX.ToString())
		ids, err := btrunk.Blockchain.Submit([]*ddb.DataTX{collectingTX})
		if err != nil {
			trail.Println(trace.Alert("error submitting collecting TX").Append(tr).UTC().Error(err))
			return fmt.Errorf("error submitting collecting TX: %w", err)
		}
		for _, tx := range ids {
			fmt.Printf(" Submitted TX: %s\n", tx)
		}
	}
	return nil
}
