package ddb_test

import (
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestLogText(t *testing.T) {
	log.SetWriter(os.Stdout)
	toAddress := "1PGh5YtRoohzcZF7WX8SJeZqm6wyaCte7X"
	fromKey := "L4ZaBkP1UTyxdEM7wysuPd1scHMLLf8sf8B2tcEcssUZ7ujrYWcQ"
	woc := ddb.NewWOC()
	utxo, err := woc.GetUTXOs("main", toAddress)
	if err != nil {
		t.Fatalf("failed to get UTXO: %v", err)
	}
	taal := ddb.NewTAAL()
	fees, err := taal.GetFee()
	if err != nil {
		t.Fatalf("failed to get fees: %v", err)
	}
	pretx, err := ddb.UTXOsToAddress(utxo, toAddress, fromKey, 0)
	if err != nil {
		t.Fatalf("failed to build TX: %v", err)
	}
	fee, err := ddb.CalculateFee(pretx.ToBytes(), fees)
	if err != nil {
		t.Fatalf("failed to calculate fee: %v", err)
	}
	tx, err := ddb.UTXOsToAddress(utxo, toAddress, fromKey, fee)
	if err != nil {
		t.Fatalf("failed to build TX: %v", err)
	}
	_, err = taal.SubmitTX(tx.ToString())
	if err != nil {
		t.Fatalf("failed to submit TX: %v", err)
	}
}
