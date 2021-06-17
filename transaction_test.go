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
	utxo := &ddb.UTXO{
		TXPos:  1,
		TXHash: "e6706b900df5a46253b8788f691cbe1506c1e9b76766f1f9d6b3602e1458f055",
		Value:  0.000402740,
		ScriptPubKey: &ddb.ScriptPubKey{
			Asm:      "OP_DUP OP_HASH160 2f353ff06fe8c4d558b9f58dce952948252e5df7 OP_EQUALVERIFY OP_CHECKSIG",
			Hex:      "76a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac",
			ReqSigs:  1,
			Type:     "pubkeyhash",
			Adresses: []string{"15JcYsiTbhFXxU7RimJRyEgKWnUfbwttb3"},
		},
	}
	tx, err := ddb.BuildOPReturnTX(utxo, key, 170, payload)
	if err != nil {
		t.Fatalf("failed to create tx: %v", err)
	}
	t.Logf(tx.ToString())
	if len(tx.ToString()) < 300 {
		t.Logf("failed to create tx, hex too short: %d", len(tx.ToString()))
		t.Fail()
	}
}

// func TestSubmitRealTX(t *testing.T) {
// 	log.SetWriter(os.Stdout)
// 	toAddress := "15JcYsiTbhFXxU7RimJRyEgKWnUfbwttb3"
// 	fromKey := "L2Aoi3Zk9oQhiEBwH9tcqnTTRErh7J3bVWoxLDzYa8nw2bWktG6M"
// 	woc := ddb.NewWOC()
// 	utxo, err := woc.GetUTXOs("main", toAddress)
// 	if err != nil {
// 		t.Fatalf("failed to get UTXO: %v", err)
// 	}
// 	taal := ddb.NewTAAL()
// 	fees, err := taal.GetFee()
// 	if err != nil {
// 		t.Fatalf("failed to get fees: %v", err)
// 	}
// 	pretx, err := ddb.UTXOsToAddress(utxo, toAddress, fromKey, 0)
// 	if err != nil {
// 		t.Fatalf("failed to build TX: %v", err)
// 	}
// 	fee, err := ddb.CalculateFee(pretx.ToBytes(), fees)
// 	if err != nil {
// 		t.Fatalf("failed to calculate fee: %v", err)
// 	}
// 	tx, err := ddb.UTXOsToAddress(utxo, toAddress, fromKey, fee)
// 	if err != nil {
// 		t.Fatalf("failed to build TX: %v", err)
// 	}
// 	_, err = taal.SubmitTX(tx.ToString())
// 	if err != nil {
// 		t.Fatalf("failed to submit TX: %v", err)
// 	}
// }
