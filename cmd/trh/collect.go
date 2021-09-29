package main

import (
	"fmt"
	"os"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdCollect(args []string) error {
	tr := trace.New().Source("collect.go", "", "cmdCollect")
	flagset, options := newFlagset(commands["collect"])
	fmt.Printf("cmdCollect flags: %v\n", args[2:])
	err := flagset.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	if flagHelp {
		printHelp("collect")
		return nil
	}
	opt, ok := areFlagConsistent(flagset, options)
	if !ok {
		return fmt.Errorf("flag combination invalid")
	}
	var pass [32]byte
	var keystore *ddb.KeyStore
	switch opt {
	case "pin":
		keystore, err = loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while opening keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while opening keystore: %w", err)
		}
		pass = keystore.Passwords["main"]
	case "password":
		keystore, err = loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while opening keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while opening keystore: %w", err)
		}
		pass = passwordtoBytes(flagPassword)
	default:
		return fmt.Errorf("flag combination invalid")
	}
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	btrunk := &ddb.BTrunk{BitcoinWIF: keystore.WIF, BitcoinAdd: keystore.Address, Blockchain: blockchain}
	wif, address, err := btrunk.GenerateKeyAndAddress(pass)
	if err != nil {
		trail.Println(trace.Alert("error while generating branched key").Append(tr).UTC().Error(err))
		return fmt.Errorf("error while generating branched key): %w", err)
	}
	collectingTX, err := btrunk.TXOfCollectedFunds(wif, address, keystore.Address)
	if err != nil {
		trail.Println(trace.Alert("error while building collecting TX").Append(tr).UTC().Error(err))
		return fmt.Errorf("error while building collecting TX: %w", err)
	}
	ids, err := btrunk.Blockchain.Submit(collectingTX)
	if err != nil {
		trail.Println(trace.Alert("error submitting collecting TX").Append(tr).UTC().Error(err))
		return fmt.Errorf("error submitting collecting TX: %w", err)
	}
	for _, tx := range ids {
		fmt.Printf(" Submitted TX: %s\n", tx)
	}
	return nil
}
