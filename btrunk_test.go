package ddb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/miner"
	"github.com/ejfhp/ddb/satoshi"
	"github.com/ejfhp/trail"
)

const (
	destinationAddress = "1PGh5YtRoohzcZF7WX8SJeZqm6wyaCte7X"
	destinationKey     = "L4ZaBkP1UTyxdEM7wysuPd1scHMLLf8sf8B2tcEcssUZ7ujrYWcQ"
)

func TestBTrunk_TXOfBranchedEntry(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	//EMPTY TEST ADDRESS
	// changeAddress := "1EpFjTzJoNAFyJKVGATzxhgqXigUWLNWM6"
	// changeKey := "L2mk9qzXebT1gfwUuALMJrbqBtrJxGUN5JnVeqQTGRXytqpXsPr8"
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	keystore, err := keys.NewKeystore(destinationKey, "mainpassword")
	if err != nil {
		t.Logf("failed to build keystore: %v", err)
		t.FailNow()
	}
	passwords := [][32]byte{
		{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'},
		{'c', 'i', 'a', 'o', 'm', 'a', 'm', 'm', 'a', 'g', 'u', 'a', 'r', 'd', 'a', 'c', 'o', 'm', 'e', 'm', 'i', 'd', 'i', 'v', 'e', 'r', 't', 'o', '.', '.', '.'},
	}
	files := map[string]string{
		"test.txt":  "testdata/test.txt",
		"image.png": "testdata/image.png",
	}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	for i, p := range passwords {
		btrunk := ddb.NewBTrunk(destinationKey, destinationAddress, keystore.Source().Password(), blockchain)
		for n, f := range files {
			node, err := keystore.NewNode(n, p)
			if err != nil {
				t.Logf("%d - failed to generate node: %v", i, err)
				t.FailNow()
			}
			maxToSpend := satoshi.Satoshi(10000)
			entry, err := ddb.NewEntryFromFile(n, f, []string{"label1", "label2"}, "notes")
			if err != nil {
				t.Logf("%d - failed to generate entry: %v", i, err)
				t.FailNow()
			}
			txs, err := btrunk.TXOfBranchedEntry(node, entry, "test01234", maxToSpend, true)
			if err != nil {
				t.Logf("%d - failed to generate branched entry TXs: %v", i, err)
				t.FailNow()
			}
			if len(txs) != 4 {
				t.Logf("%d - unexpected number of branched TXs: %d", i, len(txs))
				t.FailNow()
			}
			for i, tx := range txs {
				in := tx.Inputs
				if len(in) == 0 {
					t.Logf("%d - unexpected number of input: %d", i, len(in))
					t.FailNow()
				}
				out := tx.Outputs
				if len(out) < 2 && i < len(txs)-1 {
					t.Logf("%d - unexpected number of output: %d", i, len(in))
					t.FailNow()
				}
			}
			totFee := satoshi.Satoshi(0)
			firstIn, _, _, _ := txs[1].TotInOutFee()
			_, lastOut, _, _ := txs[len(txs)-1].TotInOutFee()
			firstFee := satoshi.Satoshi(0)
			for i, tx := range txs {
				_, _, tfe, err := tx.TotInOutFee()
				if err != nil {
					t.Logf("%d - failed to get TX fee: %v", i, err)
					t.FailNow()
				}
				if i == 0 {
					firstFee = tfe
				}
				totFee = totFee.Add(tfe.Satoshi())
				// fmt.Printf("BTRUNK_T %d IN %d OUT %d FEE: %d  TOT: %d\n\n", i, in, out, tfe, totFee)
			}
			if totFee > maxToSpend {
				t.Logf("fee greater than limit (%d): %d", maxToSpend, totFee)
				t.FailNow()
			}
			entTotFee, err := firstIn.Sub(lastOut)
			if err != nil {
				t.Logf("error calculating fee")
				t.FailNow()
			}

			if totFee != entTotFee.Add(firstFee) {
				t.Logf("in, out, fee, don't match: %d, %d, %d but fees are: %d", firstIn, lastOut, totFee, entTotFee.Add(firstFee).Satoshi())
				t.FailNow()
			}
		}
	}
}

func TestBTrunk_ListEntries(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	passwords := map[string][32]byte{
		"1": {'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'},
		"2": {'c', 'i', 'a', 'o', 'm', 'a', 'm', 'm', 'a', 'g', 'u', 'a', 'r', 'd', 'a', 'c', 'o', 'm', 'e', 'm', 'i', 'd', 'i', 'v', 'e', 'r', 't', 'o', '.', '.', '.'},
	}
	cache, err := ddb.NewUserTXCache()
	if err != nil {
		t.Logf("cannot open cache")
		t.FailNow()
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	btrunk := ddb.NewBTrunk(destinationKey, destinationAddress, "mainpassword", blockchain)
	list, err := btrunk.ListEntries(passwords, true)
	if err != nil {
		t.Logf("failed to list btrunk transactions: %v", err)
		t.FailNow()
	}
	for i, me := range list {
		fmt.Printf("Entry %s\n: %v", i, me)
	}
}
