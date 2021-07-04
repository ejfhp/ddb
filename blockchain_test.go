package ddb_test

import (
	"fmt"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestAddHeader(t *testing.T) {
	blk := ddb.NewBlockchain(nil, nil)
	data := "rosso di sera bel tempo si spera"
	payload := blk.AddHeader(ddb.APP_NAME, ddb.VER_AES, []byte(data))
	if len(payload) != 9+len(data) {
		t.Logf("wrong header len: %d", len(payload))
	}
	fmt.Printf("%s\n", payload)
	if string(payload) != "ddb;0001;"+data {
		t.Logf("wrong header: '%s'", payload)
	}
}

func TestSubmit(t *testing.T) {
	t.Skip("This always fail because it's a fake TX")
	he := "010000000155f058142e60b3d6f9f16667b7e9c10615be1c698f78b85362a4f50d906b70e6010000006a473044022076b0dd878a1d7b6919c5c8becc3f2d993436dac616bc5f055273eda570c9d59502204ab86e52cb85ff425f16001633774618a5f3b0a7108ab9a73638e3ef3fa25a684121032f8bdd0bdb654616c362a427a01cf7abafa0b61831297c09211998ede8b99b45ffffffff02a89c0000000000001976a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac00000000000000001e006a1b646462202d2052656d696e64204d792e2e2e20627920656a66687000000000"
	tx, err := ddb.TransactionFromHex(he)
	if err != nil {
		t.Logf("cannot decode hex: %v", err)
		t.FailNow()
	}
	txs := []*ddb.DataTX{tx}
	miner := ddb.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl)
	ids, err := blk.Submit(txs)
	if len(ids) != len(txs) {
		t.Logf("unexpected num of returned id: %d  exp: %d", len(ids), len(txs))
		t.Fail()
	}
	if err != nil {
		t.Logf("sumbit failed: %v", err)
	}

}

func TestPackEntities(t *testing.T) {
	key := "L2Aoi3Zk9oQhiEBwH9tcqnTTRErh7J3bVWoxLDzYa8nw2bWktG6M"
	testEncData := [][]byte{
		[]byte("not encrypted"),
		[]byte("really not encrypted"),
		[]byte("I dont't care if they are not encrypted"),
	}
	miner := ddb.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl)
	txs, err := blk.PackEncryptedEntriesPart(ddb.VER_AES, key, testEncData)
	if err != nil {
		t.Logf("failed to prepare TXs: %v", err)
		t.Fail()
	}
	if len(txs) != len(testEncData) {
		t.Logf("wrong number of TX: %d  exp: %d", len(txs), len(testEncData))
	}
	prevID := ""
	for i, tx := range txs {
		if i > 0 {
			inTXID := tx.Inputs[0].PreviousTxID
			if inTXID != prevID {
				t.Logf("txs are not chained, TXIDs are different: %s != %s", prevID, inTXID)
				t.Fail()
			}
		}
		prevID = tx.GetTxID()
	}
}
