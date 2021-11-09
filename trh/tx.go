package trh

import (
	"fmt"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
)

func (t *TRH) ListAllTX(keystore *keys.Keystore) (map[string][]string, error) {
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
	allTXs := make(map[string][]string)
	for _, pwd := range keystore.PassNames() {
		txs, err := blockchain.ListTXIDs(keystore.Address(pwd), false)
		if err != nil {
			return nil, fmt.Errorf("error while retrieving existing transactions: %w", err)
		}
		allTXs[pwd] = txs
	}
	return allTXs, nil
}

func (t *TRH) ListSinglePasswordTX(keystore *keys.Keystore, password string) ([]string, error) {
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
	return txs, nil
}

func (t *TRH) ListUTXOs(keystore *keys.Keystore) error {
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
