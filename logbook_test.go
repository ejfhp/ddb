package ddb_test

import (
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestLogText(t *testing.T) {
	log.SetWriter(os.Stdout)
	// toAddress := "1PGh5YtRoohzcZF7WX8SJeZqm6wyaCte7X"
	fromKey := "L4ZaBkP1UTyxdEM7wysuPd1scHMLLf8sf8B2tcEcssUZ7ujrYWcQ"
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	logbook := ddb.Logbook{Key: fromKey, Miner: taal, Explorer: woc}
	txid, err := logbook.LogText("ddb - diario di bordo - logbook - test")
	if err != nil {
		t.Logf("failed to log text: %v", err)
		t.Fail()
	}
	if len(txid) < 10 {
		t.Logf("txid is too short: %s", txid)
		t.Fail()
	}
}
