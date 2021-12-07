package trh

import (
	"fmt"
	"path/filepath"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/satoshi"
)

func (t *TRH) Store(name string, pathfile string, labels []string, notes string, txheader string, maxSpend uint64) ([]string, error) {
	ent, err := ddb.NewEntryFromFile(filepath.Base(pathfile), pathfile, labels, notes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate entry from file: %w", err)
	}
	node, err := t.keystore.NewNode(name, ent.HashOfEntry())
	if err != nil {
		return nil, fmt.Errorf("failed to generate new node: %w", err)
	}
	t.keystore.Update()
	if err != nil {
		return nil, fmt.Errorf("failed to generate new node: %w", err)
	}
	txs, err := t.btrunk.TXOfBranchedEntry(node, ent, txheader, satoshi.Satoshi(maxSpend), false)
	if err != nil {
		return nil, fmt.Errorf("failed to update keystore: %w", err)
	}
	totFee := satoshi.Satoshi(0)
	for i, t := range txs {
		_, _, fee, err := t.TotInOutFee()
		if err != nil {
			return nil, fmt.Errorf("failed to get fee from tx num %d: %w", i, err)
		}
		totFee = totFee.Add(fee)
	}
	txres, err := t.blockchain.Submit(txs)
	if err != nil {
		return nil, fmt.Errorf("failed to submit txs: %w", err)
	}
	ids := make([]string, len(txres))
	for i, tx := range txres {
		ids[i] = tx[0]
	}
	return ids, nil
}

//TODO add store with TX from simulate as input
