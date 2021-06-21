package ddb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestGetUTXOs(t *testing.T) {
	log.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	unsTx, err := woc.GetUTXOs("1K2HC5AQQniJ2zcWSyjjtkKZgKMkZ1CGNr")
	if err != nil {
		t.Logf("error: %v", err)
		t.Fail()
	}
	t.Logf("UTXO count: %d", len(unsTx))
	if float64(*unsTx[0].Value) != 135.70301 {
		t.Logf("wrong value in bitcoin: %.8f", *unsTx[0].Value)
		t.Fail()
	}
	if uint64(unsTx[0].Value.Satoshi()) != 13570301000 {
		t.Logf("wrong value in bitcoin")
		t.Fail()
	}
	t.Logf("Unspent bitcoin: %.8f\n", unsTx[0].Value.Bitcoin())
	t.Logf("Unspent satoshi: %d\n", unsTx[0].Value.Satoshi())
}

func TestGetTX(t *testing.T) {
	log.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	tx, err := woc.GetTX("d715807cf35de1663d9413b0b0863aae83876c81a78206cedf4fd60bb0a986b7")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	fmt.Printf("TX ID: \n%v\n", tx.ID)
}
