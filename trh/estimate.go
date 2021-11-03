package trh

import (
	"fmt"
	"path/filepath"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/miner"
	"github.com/ejfhp/ddb/satoshi"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const (
	fakeKey      = "L2vWab1E8FhsDvrPF59CEoB2Txnqkn8XwH3BPgirnmnGnoByCw82"
	fakeAddress  = "1NG2BMqAbNgsFWkRATFSWWz6JzPbWLV5SP"
	fakeHeader   = "fakeheader"
	fakeValue    = 10000000
	fakePassword = "B3QGlJVqH7ZmhLo_oT8WcElm9OzOLxM5"
)

func Estimate(file string, labels []string, notes string) (satoshi.Satoshi, error) {
	tr := trace.New().Source("estimate.go", "", "cmdEstimate")
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	btrunk := &ddb.BTrunk{MainKey: fakeKey, MainAddress: fakeAddress, Blockchain: blockchain}
	ent, err := ddb.NewEntryFromFile(filepath.Base(file), file, labels, notes)
	if err != nil {
		trail.Println(trace.Alert("failed to generate entry from file").Append(tr).UTC().Error(err))
		return 0, fmt.Errorf("failed to generate entry from file: %w", err)
	}

	// pwd := (*[32]byte)([]byte(fakePassword))
	pwd := [32]byte{}
	copy(pwd[:], []byte(fakePassword))
	txs, err := btrunk.TXOfBranchedEntry(fakeKey, fakeAddress, pwd, ent, fakeHeader, satoshi.Bitcoin(fakeValue).Satoshi(), true)
	if err != nil {
		trail.Println(trace.Alert("failed to generate txs for entry").Append(tr).UTC().Error(err))
		return 0, fmt.Errorf("failed to generate txs for entry: %w", err)
	}
	totFee := satoshi.Satoshi(0)
	for i, t := range txs {
		_, _, fee, err := t.TotInOutFee()
		if err != nil {
			trail.Println(trace.Alert("failed to get fee from tx").Append(tr).UTC().Error(err))
			return 0, fmt.Errorf("failed to get fee from tx num %d: %w", i, err)
		}
		totFee = totFee.Add(fee)
	}
	fmt.Printf("Estimated fee: %d satoshi\n", totFee)
	fmt.Printf("Estimated traffic: %d tx\n", len(txs))
	return totFee, nil
}
