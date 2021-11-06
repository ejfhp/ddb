package trh

import (
	"fmt"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
)

func ListAllTX(keystore *keys.Keystore) ([]string, error) {
	passwordAddress := map[string]string{}
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return nil, fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	if err != nil {
		return nil, fmt.Errorf("error while loading keystore: %w", err)
	}
	for _, pn := range keystore.PassNames() {
		passwordAddress[pn] = keystore.Address(pn)
	}
	allTXs := []string{}
	for pwd, add := range passwordAddress {
		txs, err := blockchain.ListTXIDs(add, false)
		if err != nil {
			return nil, fmt.Errorf("error while retrieving existing transactions: %w", err)
		}
		fmt.Printf("Address '%s' of password '%s'\n", add, pwd)
		for _, tx := range txs {
			fmt.Printf(" Found TX: %s\n", tx)
			allTXs = append(allTXs, tx)
		}
	}
	return allTXs, nil
}

func ListSinglePasswordTX(keystore *keys.Keystore, password string) ([]string, error) {
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return nil, fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	address := keystore.Address(password)
	txs, err := blockchain.ListTXIDs(address, false)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving existing transactions: %w", err)
	}
	fmt.Printf("Address '%s' of password '%s'\n", address, password)
	for _, tx := range txs {
		fmt.Printf(" Found TX: %s\n", tx)
	}
	return txs, nil
}

func ListUTXOs(keystore *keys.Keystore) error {
	passwordAddress := map[string]string{}
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	for _, ka := range keystore.PassNames() {
		passwordAddress[ka] = keystore.Address(ka)
	}
	for pwd, add := range passwordAddress {
		utxos, err := blockchain.GetUTXO(add)
		if err != nil {
			if err.Error() != "found no UTXO" {
				return fmt.Errorf("error while retrieving unspend outputs (UTXO): %w", err)
			}
			utxos = []*ddb.UTXO{}
		}
		fmt.Printf("Address '%s' of password '%s'\n", add, pwd)
		for _, u := range utxos {
			fmt.Printf(" Found UTXOS: %d satoshi in TX %s, %d\n", u.Value.Satoshi(), u.TXHash, u.TXPos)
		}
	}
	return nil
}
