package ddb_test

import (
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestUTXOToAddress(t *testing.T) {
	log.SetWriter(os.Stdout)
	key := "L2Aoi3Zk9oQhiEBwH9tcqnTTRErh7J3bVWoxLDzYa8nw2bWktG6M"
	payload := []byte("ddb - Remind My... by ejfhp")
	bsv := ddb.Bitcoin(0.000402740)
	fee := ddb.Satoshi(170)
	utxo := &ddb.UTXO{
		TXPos:  1,
		TXHash: "e6706b900df5a46253b8788f691cbe1506c1e9b76766f1f9d6b3602e1458f055",
		Value:  &bsv,
		ScriptPubKey: &ddb.ScriptPubKey{
			Asm:      "OP_DUP OP_HASH160 2f353ff06fe8c4d558b9f58dce952948252e5df7 OP_EQUALVERIFY OP_CHECKSIG",
			Hex:      "76a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac",
			ReqSigs:  1,
			Type:     "pubkeyhash",
			Adresses: []string{"15JcYsiTbhFXxU7RimJRyEgKWnUfbwttb3"},
		},
	}
	tx, err := ddb.BuildOPReturnBytesTX(utxo, key, &fee, payload)
	if err != nil {
		t.Fatalf("failed to create tx: %v", err)
	}
	t.Logf("TX len: %d", len(tx))
	if len(tx) < 100 {
		t.Logf("failed to create tx, []byte too short: %d", len(tx))
		t.Fail()
	}
}
