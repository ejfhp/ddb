package trh

import (
	"fmt"
	"os"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/miner"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdTx(args []string) error {
	tr := trace.New().Source("tx.go", "", "cmdTx")
	flagset, options := newFlagset(txCmd)
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
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	switch opt {
	case "pin":
		keystore, err := loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading keystore: %w", err)
		}
		for pwd, ka := range keystore.PassNames() {
			fmt.Printf("PWD: %s\n", pwd)
			passwordAddress[ka] = keystore.Address(ka)
		}
	case "password":
		keystore, err := loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading keystore: %w", err)
		}
		ka := keystore.Address(flagPassword)
		if ka != "" {
			passwordAddress[flagPassword] = ka
		}

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
		fmt.Printf("Address '%s' of password '%s'\n", add, pwd)
		for _, u := range utxos {
			fmt.Printf(" Found UTXOS: %d satoshi in TX %s, %d\n", u.Value.Satoshi(), u.TXHash, u.TXPos)
		}
		for _, tx := range txs {
			fmt.Printf(" Found TX: %s\n", tx)
		}
	}

	return nil
}
