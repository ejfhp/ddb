package trh

import (
	"fmt"
	"path/filepath"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
	"github.com/ejfhp/ddb/satoshi"
)

func (t *TRH) Store(keystore *keys.Keystore, password string, pathfile string, labels []string, notes string, txheader string, maxSpend uint64) error {
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		return fmt.Errorf("cannot open cache")
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	btrunk := &ddb.BTrunk{MainKey: keystore.Key(keys.Main), MainAddress: keystore.Address(keys.Main), Blockchain: blockchain}
	ent, err := ddb.NewEntryFromFile(filepath.Base(pathfile), pathfile, labels, notes)
	if err != nil {
		return fmt.Errorf("failed to generate entry from file: %w", err)
	}
	node, err := keystore.NodeFromPassword(password)
	if err != nil {
		return fmt.Errorf("failed to generate branch key and address: %w", err)
	}
	//Storing password in Keystore before to try to store the file on the blockchain
	err = keystore.Update()
	if err != nil {
		return fmt.Errorf("failed to save the current password in the keystore: %w", err)
	}
	txs, err := btrunk.TXOfBranchedEntry(node.Key, node.Address, node.Password, ent, txheader, satoshi.Satoshi(maxSpend), false)
	if err != nil {
		return fmt.Errorf("failed to generate txs for entry: %w", err)
	}
	totFee := satoshi.Satoshi(0)
	for i, t := range txs {
		_, _, fee, err := t.TotInOutFee()
		if err != nil {
			return fmt.Errorf("failed to get fee from tx num %d: %w", i, err)
		}
		totFee = totFee.Add(fee)
	}
	ids, err := btrunk.Blockchain.Submit(txs)
	if err != nil {
		return fmt.Errorf("failed to submit txs: %w", err)
	}
	for id, success := range ids {
		fmt.Printf("%s: %s\n", id, success)
	}

	return nil
}
