package ddb_test

import (
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestBuildDataTX(t *testing.T) {
	log.SetWriter(os.Stdout)
	txid := "e6706b900df5a46253b8788f691cbe1506c1e9b76766f1f9d6b3602e1458f055"
	scriptHex := "76a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac"
	payload := []byte("ddb - Remind My... by ejfhp")
	bsv := ddb.Bitcoin(0.000402740)
	fee := ddb.Satoshi(170)
	version := "test"
	utxos := []*ddb.UTXO{{TXHash: txid, TXPos: 1, Value: bsv, ScriptPubKeyHex: scriptHex}}
	datatx, err := ddb.BuildDataTX(address, utxos, key, fee, payload, version)
	if err != nil {
		t.Fatalf("failed to create tx: %v", err)
	}
	t.Logf("TX ID: %s len: %d", datatx.GetTxID(), len(datatx.ToString()))
	//fmt.Printf("DataTX hex: '%s'", datatx.ToString())
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

func TestData(t *testing.T) {
	he := "010000000155f058142e60b3d6f9f16667b7e9c10615be1c698f78b85362a4f50d906b70e6010000006b483045022100f729300b6b8b253d412b232d847f088f394321f785ff16f967303514acc6ad7b02203f49f2a8405bd1a0f419d8808d44ef68f1bb323e7608ab5fd326f567e84014684121032f8bdd0bdb654616c362a427a01cf7abafa0b61831297c09211998ede8b99b45ffffffff02a89c0000000000001976a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac000000000000000027006a246464623b746573743b646462202d2052656d696e64204d792e2e2e20627920656a66687000000000"
	tx, err := ddb.DataTXFromHex(he)
	if err != nil {
		t.Logf("failed to create tx: %v", err)
		t.Fail()
	}
	opr, version, err := tx.Data()
	if err != nil {
		t.Logf("failed to get data: %v", err)
		t.Fail()
	}
	if version != "test" {
		t.Logf("version is not correct: %v", version)
		t.Fail()
	}
	if string(opr) != "ddb - Remind My... by ejfhp" {
		t.Logf("opreturn is not correct: %v", string(opr))
		t.Fail()
	}

}
