package ddb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
)

func TestGetUTXOs(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	unsTx, err := woc.GetUTXOs("1K2HC5AQQniJ2zcWSyjjtkKZgKMkZ1CGNr")
	fmt.Printf("spub: %s\n", unsTx[0].ScriptPubKeyHex)
	fmt.Printf("txhs: %s\n", unsTx[0].TXHash)
	fmt.Printf("valu: %v\n", unsTx[0].Value)
	fmt.Printf("posi: %d\n", unsTx[0].TXPos)
	if err != nil {
		t.Logf("error: %v", err)
		t.Fail()
	}
	t.Logf("UTXO count: %d", len(unsTx))
	if float64(unsTx[0].Value) != 135.70301 {
		t.Logf("wrong value in bitcoin: %.8f", unsTx[0].Value)
		t.Fail()
	}
	if uint64(unsTx[0].Value.Satoshi()) != 13570301000 {
		t.Logf("wrong value in bitcoin")
		t.Fail()
	}
	t.Logf("Unspent bitcoin: %.8f\n", unsTx[0].Value.Bitcoin())
	t.Logf("Unspent satoshi: %d\n", unsTx[0].Value.Satoshi())
}

func TestWOCGetTX(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	// txid := "afbdf4a215f5e7dc3beca36e1625f3597995afa5906b2bbfee6a572d87764426"
	txid := "d715807cf35de1663d9413b0b0863aae83876c81a78206cedf4fd60bb0a986b7"
	tx, err := woc.GetTX(txid)
	if err != nil {
		t.Logf("error: %v", err)
		t.Fail()
	}
	fmt.Println(tx)
	if tx.ID != txid {
		t.Logf("unexpected ID: '%s'", tx.ID)
		t.Fail()
	}
}

func TestWOCGetRAWTXHEX(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	// txid := "afbdf4a215f5e7dc3beca36e1625f3597995afa5906b2bbfee6a572d87764426"
	txid := "d715807cf35de1663d9413b0b0863aae83876c81a78206cedf4fd60bb0a986b7"
	tx, err := woc.GetRAWTXHEX(txid)
	if err != nil {
		t.Logf("error: %v", err)
		t.Fail()
	}
	if len(tx) < 100 {
		t.Logf("raw hex too short: %d", len(tx))
		t.Fail()
	}
}

func TestWOCGetTXIDs(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	txids, err := woc.GetTXIDs(destinationAddress)
	if err != nil {
		t.Logf("error: %v", err)
		t.Fail()
	}
	if len(txids) < 23 {
		t.Logf("tx history too short: %d", len(txids))
		t.Fail()
	}
}
