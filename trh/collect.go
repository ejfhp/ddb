package trh

import (
	"fmt"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

//Collect gather all the UTXO left in branches to the main address. Return the txids.
func (t *TRH) Collect(keystore *keys.Keystore) ([]string, error) {
	tr := trace.New().Source("collect.go", "", "cmdCollect")
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return nil, fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)

	utxos := make(map[string][]*ddb.UTXO)
	for _, n := range keystore.Nodes() {
		u, err := blockchain.GetUTXO(n.Address())
		if err != nil && err.Error() != "found no UTXO" {
			return nil, fmt.Errorf("error while retrieving UTXO for address %s: %w", n.Address, err)
		}
		if len(u) > 0 {
			utxos[n.Key()] = u
		}
	}
	var ids []string
	if len(utxos) > 0 {
		fee, err := blockchain.EstimateStandardTXFee(len(utxos))
		if err != nil {
			return nil, fmt.Errorf("error while estimating collecting tx fee: %w", err)
		}
		//TODO COLLECT HAS ISSUES
		collectingTX, err := ddb.NewMultiInputTX(keystore.Source().Address(), utxos, fee)
		if err != nil {
			return nil, fmt.Errorf("error while building collecting TX: %w", err)
		}
		txres, err := blockchain.Submit([]*ddb.DataTX{collectingTX})
		if err != nil {
			trail.Println(trace.Alert("error submitting collecting TX").Append(tr).UTC().Error(err))
			return nil, fmt.Errorf("error submitting collecting TX: %w", err)
		}
		ids = make([]string, len(txres))
		for i, tx := range txres {
			ids[i] = tx[0]
		}

	}
	return ids, nil
}
