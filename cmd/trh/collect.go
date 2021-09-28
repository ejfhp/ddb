package main

import (
	"fmt"
	"os"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdShow(args []string) error {
	tr := trace.New().Source("show.go", "", "cmdShow")
	flagset, options := newFlagset(commands["show"])
	fmt.Printf("cmdShow flags: %v\n", args[2:])
	err := flagset.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	if flagHelp {
		printHelp("show")
		return nil
	}
	opt := areFlagConsistent(flagset, options)
	switch opt {
	case "pin":
		woc := ddb.NewWOC()
		taal := ddb.NewTAAL()
		blockchain := ddb.NewBlockchain(taal, woc, nil)
		keystore, err := loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading keystore: %w", err)
		}
		utxos, err := blockchain.GetUTXO(keystore.Address)
		if err != nil {
			trail.Println(trace.Alert("error while retrieving unspent outputs (UTXO)").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while retrieving unspend outputs (UTXO): %w", err)
		}
		txs, err := blockchain.ListTXIDs(keystore.Address, false)
		if err != nil {
			trail.Println(trace.Alert("error while retrieving existing transactions").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while retrieving existing transactions: %w", err)
		}
		for _, u := range utxos {
			fmt.Printf(" Found UTXOS: %d satoshi in TX %s, %d\n", u.Value.Satoshi(), u.TXHash, u.TXPos)
		}
		for _, tx := range txs {
			fmt.Printf(" Found TX: %s\n", tx)
		}

	default:
		return fmt.Errorf("flag combination invalid")
	}
	return nil
}
