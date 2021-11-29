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
	for add := range keystore.AddressesAndKeys() {
		txs, err := blockchain.ListTXIDs(add, false)
		if err != nil {
			return nil, fmt.Errorf("error while retrieving existing transactions: %w", err)
		}
		allTXs[add] = txs
	}
	return allTXs, nil
}

func (t *TRH) ListUTXOs(keystore *keys.Keystore) (map[string][]*ddb.UTXO, error) {
	passwordAddress := map[string]string{}
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return nil, fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	for _, no := range keystore.Nodes() {
		passwordAddress[no.Name] = no.Address
	}
	passwordAddress["Source"] = keystore.Source().Address
	res := make(map[string][]*ddb.UTXO)
	for _, add := range passwordAddress {
		utxos, err := blockchain.GetUTXO(add)
		if err != nil {
			if err.Error() != "found no UTXO" {
				return nil, fmt.Errorf("error while retrieving unspend outputs (UTXO): %w", err)
			}
			utxos = []*ddb.UTXO{}
		}
		res[add] = utxos
	}
	return res, nil
}
