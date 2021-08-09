package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestPacker_PackEntities(t *testing.T) {
	log.SetWriter(os.Stdout)
	testData := [][]byte{
		[]byte("not encrypted"),
		[]byte("really not encrypted"),
		[]byte("I dont't care if they are not encrypted"),
	}
	utxos := []*ddb.UTXO{{TXPos: 0, TXHash: "00ac11165e83888a749c0540aeddc44e99b6a95ffba845454eef355db6d4440e", Value: 70301, ScriptPubKeyHex: "76a914c5b0320f5005f3a3f2897570f90b48ddc95f981588ac"}}
	satoshi500 := ddb.Satoshi(500)
	fee := &ddb.Fee{
		FeeType: "data",
		MiningFee: ddb.FeeUnit{
			Satoshis: &satoshi500,
			Bytes:    1000,
		},
		RelayFee: ddb.FeeUnit{
			Satoshis: &satoshi500,
			Bytes:    1000,
		},
	}
	txs, err := ddb.PackData(ddb.VER_AES, key, testData, utxos, fee)
	if err != nil {
		t.Logf("failed to prepare TXs: %v", err)
		t.Fail()
	}
	if len(txs) != len(testData) {
		t.Logf("wrong number of TX: %d  exp: %d", len(txs), len(testData))
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
		opr, ver, err := tx.Data()
		if err != nil {
			t.Logf("failed to get OP_RETURN: %v", err)
			t.Fail()
		}
		if ver != ddb.VER_AES {
			t.Logf("unexpected version: %s", ver)
			t.Fail()
		}
		t.Logf("%d,OP_RETURN data length: opr:%d   orig:%d", i, len(opr), len(testData[i]))
		if len(opr) != len(testData[i]) {
			t.Logf("%d, unexpected OP_RETURN data length: %d != %d", i, len(opr), len(testData[i]))
			t.Fail()
		}
		for j := 0; j < len(opr); j++ {
			if opr[j] != testData[i][j] {
				t.Logf("OP_RETURN data is wrong")
				t.FailNow()
			}
		}

		prevID = tx.GetTxID()
	}
}

func TestPacker_PackUnpackText(t *testing.T) {
	log.SetWriter(os.Stdout)
	testData := [][]byte{
		[]byte("I just want to write something a bit longer, even if I know it doesn't matter 1"),
		[]byte("I just want to write something a bit longer, even if I know it doesn't matter 2"),
		[]byte("I just want to write something a bit longer, even if I know it doesn't matter 3"),
		[]byte("I just want to write something a bit longer, even if I know it doesn't matter 4"),
	}
	utxos := []*ddb.UTXO{{TXPos: 0, TXHash: "00ac11165e83888a749c0540aeddc44e99b6a95ffba845454eef355db6d4440e", Value: 70301, ScriptPubKeyHex: "76a914c5b0320f5005f3a3f2897570f90b48ddc95f981588ac"}}
	satoshi500 := ddb.Satoshi(500)
	fee := &ddb.Fee{
		FeeType: "data",
		MiningFee: ddb.FeeUnit{
			Satoshis: &satoshi500,
			Bytes:    1000,
		},
		RelayFee: ddb.FeeUnit{
			Satoshis: &satoshi500,
			Bytes:    1000,
		},
	}
	txs, err := ddb.PackData(ddb.VER_AES, key, testData, utxos, fee)
	if err != nil {
		t.Logf("failed to prepare TXs: %v", err)
		t.Fail()
	}
	if len(txs) != len(testData) {
		t.Logf("wrong number of TX: %d  exp: %d", len(txs), len(testData))
	}
	exData, err := ddb.UnpackData(txs)
	if err != nil {
		t.Logf("failed to unpack data TXs: %v", err)
		t.Fail()
	}
	for i, dt := range exData {
		if len(dt) != len(testData[i]) {
			t.Logf("%d, unexpected data length: %d != %d", i, len(dt), len(testData[i]))
			t.Fail()
		}
		if string(dt) != string(testData[i]) {
			t.Logf("OP_RETURN data is wrong: %s != %s", string(dt), string(testData[i]))
			t.FailNow()
		}
	}
}

func TestPacker_PackUnpackFile(t *testing.T) {
	log.SetWriter(os.Stdout)
	file := "testdata/image.png"
	image, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("error reading test file %s: %v", file, err)
	}
	testData := [][]byte{
		image,
	}
	imageSha := sha256.Sum256(image)
	imageHash := hex.EncodeToString(imageSha[:])
	utxos := []*ddb.UTXO{{TXPos: 0, TXHash: "00ac11165e83888a749c0540aeddc44e99b6a95ffba845454eef355db6d4440e", Value: 70301, ScriptPubKeyHex: "76a914c5b0320f5005f3a3f2897570f90b48ddc95f981588ac"}}
	satoshi500 := ddb.Satoshi(500)
	fee := &ddb.Fee{
		FeeType: "data",
		MiningFee: ddb.FeeUnit{
			Satoshis: &satoshi500,
			Bytes:    1000,
		},
		RelayFee: ddb.FeeUnit{
			Satoshis: &satoshi500,
			Bytes:    1000,
		},
	}
	txs, err := ddb.PackData(ddb.VER_AES, key, testData, utxos, fee)
	if err != nil {
		t.Logf("failed to prepare TXs: %v", err)
		t.Fail()
	}
	if len(txs) != len(testData) {
		t.Logf("wrong number of TX: %d  exp: %d", len(txs), len(testData))
	}
	exData, err := ddb.UnpackData(txs)
	if err != nil {
		t.Logf("failed to unpack data TXs: %v", err)
		t.Fail()
	}
	for i, dt := range exData {
		if len(dt) != len(testData[i]) {
			t.Logf("%d, unexpected data length: %d != %d", i, len(dt), len(testData[i]))
			t.Fail()
		}
		dataSha := sha256.Sum256(image)
		dataHash := hex.EncodeToString(dataSha[:])
		if dataHash != imageHash {
			t.Logf("OP_RETURN data is wrong: %s != %s", dataHash, imageHash)
		}
	}
}
