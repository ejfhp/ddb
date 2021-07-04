package ddb_test

import (
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestBuildOPReturnTX(t *testing.T) {
	log.SetWriter(os.Stdout)
	key := "L2Aoi3Zk9oQhiEBwH9tcqnTTRErh7J3bVWoxLDzYa8nw2bWktG6M"
	address := "15JcYsiTbhFXxU7RimJRyEgKWnUfbwttb3"
	txid := "e6706b900df5a46253b8788f691cbe1506c1e9b76766f1f9d6b3602e1458f055"
	scriptHex := "76a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac"
	payload := []byte("ddb - Remind My... by ejfhp")
	bsv := ddb.Bitcoin(0.000402740).Satoshi()
	fee := ddb.Satoshi(170)
	datatx, err := ddb.BuildOPReturnTX(address, txid, bsv, 1, scriptHex, key, fee, payload)
	if err != nil {
		t.Fatalf("failed to create tx: %v", err)
	}
	t.Logf("TX ID: %s len: %d", datatx.GetTxID(), len(datatx.ToString()))
	if txid == "" {
		t.Logf("failed to create tx, ID is empty")
		t.Fail()
	}
	if len(datatx.Outputs) != 2 {
		t.Logf("wrong number of output: %d", len(datatx.Outputs))
		t.Fail()
	}
	if datatx.Outputs[0].Satoshis <= 0 {
		t.Logf("output num 0 should be the change but has no output value: %d", datatx.Outputs[0].Satoshis)
		t.Fail()
	}
	if len(datatx.ToBytes()) < 100 {
		t.Logf("failed to create tx, []byte too short: %d", len(datatx.ToBytes()))
		t.Fail()
	}
}

func TestOpReturn(t *testing.T) {
	he := "010000000155f058142e60b3d6f9f16667b7e9c10615be1c698f78b85362a4f50d906b70e6010000006a473044022076b0dd878a1d7b6919c5c8becc3f2d993436dac616bc5f055273eda570c9d59502204ab86e52cb85ff425f16001633774618a5f3b0a7108ab9a73638e3ef3fa25a684121032f8bdd0bdb654616c362a427a01cf7abafa0b61831297c09211998ede8b99b45ffffffff02a89c0000000000001976a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac00000000000000001e006a1b646462202d2052656d696e64204d792e2e2e20627920656a66687000000000"
	tx, err := ddb.TransactionFromHex(he)
	if err != nil {
		t.Logf("failed to create tx: %v", err)
		t.Fail()
	}
	if err != nil {
		t.Logf("failed to get opreturn: %v", err)
		t.Fail()
	}
	opr, err := tx.OpReturn()
	if string(opr) != "ddb - Remind My... by ejfhp" {
		t.Logf("opreturn is not correct: %v", string(opr))
		t.Fail()

	}

}
