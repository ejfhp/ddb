package trh

import (
	"fmt"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/satoshi"
)

func (t *TRH) Simulate(name string, pathfile string, labels []string, notes string, txheader string, maxSpend uint64) ([]*ddb.DataTX, uint64, error) {
	ent, err := ddb.NewEntryFromFile(name, pathfile, labels, notes)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to generate entry from file: %w", err)
	}
	node, err := t.keystore.NewNode(name, ent.HashOfEntry())
	if err != nil {
		return nil, 0, fmt.Errorf("failed to generate new node: %w", err)
	}
	txs, err := t.btrunk.TXOfBranchedEntry(node, ent, txheader, satoshi.Satoshi(maxSpend), true)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to generate txs for entry: %w", err)
	}
	totFee := satoshi.Satoshi(0)
	for i, t := range txs {
		_, _, fee, err := t.TotInOutFee()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get fee from tx num %d: %w", i, err)
		}
		totFee = totFee.Add(fee)
	}
	return txs, uint64(totFee), nil
}
