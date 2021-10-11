package ddb_test

import (
	"testing"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/miner"
	"github.com/ejfhp/ddb/satoshi"
)

func TestBTrunk_GenerateKeyAndAddress(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
	passwords := [][32]byte{
		{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'},
		{'c', 'i', 'a', 'o', 'm', 'a', 'm', 'm', 'a', 'g', 'u', 'a', 'r', 'd', 'a', 'c', 'o', 'm', 'e', 'm', 'i', 'd', 'i', 'v', 'e', 'r', 't', 'o', '.', '.', '.'},
	}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	btrunk := &ddb.BTrunk{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Blockchain: blockchain}
	for i, v := range passwords {
		bWIF, bAdd, err := btrunk.GenerateKeyAndAddress(v)
		if err != nil {
			t.Logf("%d - failed to generate key and add: %v", i, err)
			t.FailNow()
		}
		b2Add, err := ddb.AddressOf(bWIF)
		if err != nil {
			t.Logf("%d - failed to generate address from generated WIF: %v", i, err)
			t.FailNow()
		}
		if b2Add != bAdd {
			t.Logf("%d - something is wrong in the key-add generation", i)
			t.FailNow()
		}
	}
}

func TestBTrunk_TXOfBranchedEntry(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := miner.NewTAAL()
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
		btrunk := &ddb.BTrunk{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Blockchain: blockchain}
		bWIF, bAdd, err := btrunk.GenerateKeyAndAddress(p)
		if err != nil {
			t.Logf("%d - failed to generate key and add: %v", i, err)
			t.FailNow()
		}
		for n, f := range files {
			maxToSpend := satoshi.Satoshi(10000)
			entry, err := ddb.NewEntryFromFile(n, f, []string{"label1", "label2"}, "notes")
			if err != nil {
				t.Logf("%d - failed to generate entry: %v", i, err)
				t.FailNow()
			}
			txs, err := btrunk.TXOfBranchedEntry(bWIF, bAdd, p, entry, "test01234", maxToSpend, true)
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
				if len(out) < 2 {
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
