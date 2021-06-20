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
	if unsTx[0].Value.Value() != 135.70301 {
		t.Logf("wrong value in bitcoin")
		t.Fail()
	}
	if unsTx[0].Value.Satoshis() != 13570301000 {
		t.Logf("wrong value in bitcoin")
		t.Fail()
	}
	t.Logf("Unspent bitcoin: %v\n", unsTx[0].Value.Value())
	t.Logf("Unspent satoshi: %v\n", unsTx[0].Value.Satoshis())
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
