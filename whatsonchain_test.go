package remy_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/ejfhp/remy"
	log "github.com/ejfhp/trail"
)

func TestGetUTXOs(t *testing.T) {
	log.SetWriter(os.Stdout)
	woc := remy.NewWOC()
	unsTx, err := woc.GetUTXOs("main", "1K2HC5AQQniJ2zcWSyjjtkKZgKMkZ1CGNr")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	fmt.Printf("Unspent: \n%v\n", unsTx[0].Value)
}

func TestGetTX(t *testing.T) {
	log.SetWriter(os.Stdout)
	woc := remy.NewWOC()
	tx, err := woc.GetTX("main", "d715807cf35de1663d9413b0b0863aae83876c81a78206cedf4fd60bb0a986b7")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	fmt.Printf("TX ID: \n%v\n", tx.ID)
}
