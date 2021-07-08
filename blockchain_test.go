package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestSubmit(t *testing.T) {
	t.Skip("This always fail because it's a fake TX")
	he := "010000000155f058142e60b3d6f9f16667b7e9c10615be1c698f78b85362a4f50d906b70e6010000006a473044022076b0dd878a1d7b6919c5c8becc3f2d993436dac616bc5f055273eda570c9d59502204ab86e52cb85ff425f16001633774618a5f3b0a7108ab9a73638e3ef3fa25a684121032f8bdd0bdb654616c362a427a01cf7abafa0b61831297c09211998ede8b99b45ffffffff02a89c0000000000001976a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac00000000000000001e006a1b646462202d2052656d696e64204d792e2e2e20627920656a66687000000000"
	tx, err := ddb.DataTXFromHex(he)
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
	testData := [][]byte{
		[]byte("not encrypted"),
		[]byte("really not encrypted"),
		[]byte("I dont't care if they are not encrypted"),
	}
	miner := ddb.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl)
	txs, err := blk.PackData(ddb.VER_AES, key, testData)
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

func TestPackUnpackText(t *testing.T) {
	testData := [][]byte{
		[]byte("I just want to write something a bit longer, even if I know it doesn't matter 1"),
		[]byte("I just want to write something a bit longer, even if I know it doesn't matter 2"),
		[]byte("I just want to write something a bit longer, even if I know it doesn't matter 3"),
		[]byte("I just want to write something a bit longer, even if I know it doesn't matter 4"),
	}
	miner := ddb.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl)
	txs, err := blk.PackData(ddb.VER_AES, key, testData)
	if err != nil {
		t.Logf("failed to prepare TXs: %v", err)
		t.Fail()
	}
	if len(txs) != len(testData) {
		t.Logf("wrong number of TX: %d  exp: %d", len(txs), len(testData))
	}
	exData, err := blk.UnpackData(txs)
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

func TestPackUnpackFile(t *testing.T) {
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
	miner := ddb.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl)
	txs, err := blk.PackData(ddb.VER_AES, key, testData)
	if err != nil {
		t.Logf("failed to prepare TXs: %v", err)
		t.Fail()
	}
	if len(txs) != len(testData) {
		t.Logf("wrong number of TX: %d  exp: %d", len(txs), len(testData))
	}
	exData, err := blk.UnpackData(txs)
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
