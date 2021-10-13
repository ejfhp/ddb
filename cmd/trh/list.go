package main

import (
	"fmt"
	"os"

	"github.com/ejfhp/ddb"
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
	passwordAddress := map[string]string{}
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	switch opt {
	case "pin":
		keystore, err := loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading keystore: %w", err)
		}
		passwordAddress[""] = keystore.Address
		btrunk := &ddb.BTrunk{BitcoinWIF: keystore.Key, BitcoinAdd: keystore.Address, Blockchain: blockchain}
		for _, password := range keystore.Passwords {
			_, add, err := btrunk.GenerateKeyAndAddress(password)
			if err != nil {
				trail.Println(trace.Alert("error while generating address for keystore pasword").Append(tr).UTC().Error(err))
				return fmt.Errorf("error while generating address for keystore pasword %s: %w", string(password[:]), err)
			}
			passwordAddress[string(password[:])] = add
		}
	case "password":
		keystore, err := loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading keystore: %w", err)
		}
		btrunk := &ddb.BTrunk{BitcoinWIF: keystore.Key, BitcoinAdd: keystore.Address, Blockchain: blockchain}

		_, add, err := btrunk.GenerateKeyAndAddress(passwordtoBytes(flagPassword))
		if err != nil {
			trail.Println(trace.Alert("error while generating address for keystore pasword").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while generating address for pasword %s: %w", flagPassword, err)
		}
		passwordAddress[flagPassword] = add

	default:
		return fmt.Errorf("flag combination invalid")
	}
	for pwd, add := range passwordAddress {
		utxos, err := blockchain.GetUTXO(add)
		if err != nil {
			if err.Error() != "found no UTXO" {
				trail.Println(trace.Alert("error while retrieving unspent outputs (UTXO)").Append(tr).UTC().Error(err))
				return fmt.Errorf("error while retrieving unspend outputs (UTXO): %w", err)
			}
			utxos = []*ddb.UTXO{}
		}
		txs, err := blockchain.ListTXIDs(add, false)
		if err != nil {
			trail.Println(trace.Alert("error while retrieving existing transactions").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while retrieving existing transactions: %w", err)
		}
		if pwd == "" {
			fmt.Printf("Main address '%s' \n", add)
		} else {
			fmt.Printf("Address '%s' of password '%s'\n", add, pwd)
		}
		for _, u := range utxos {
			fmt.Printf(" Found UTXOS: %d satoshi in TX %s, %d\n", u.Value.Satoshi(), u.TXHash, u.TXPos)
		}
		for _, tx := range txs {
			fmt.Printf(" Found TX: %s\n", tx)
		}
	}

	return nil
}
