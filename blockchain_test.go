package ddb_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/miner"
)

func TestBlockchain_EstimateDataTXFee(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	miner := miner.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl, nil)
	//Check consistency
	for i := 0; i < 10; i++ {
		fee, err := blk.EstimateDataTXFee(1, []byte{}, "header000")
		if err != nil {
			t.Logf("%d - cannot estimate fee: %v", i, err)
			t.FailNow()
		}
		if fee < 125 {
			t.Logf("%d - fee estimation failed: %d", i, fee.Satoshi())
			t.FailNow()
		}
	}
}

func TestBlockchain_EstimateStandardTXFee(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	miner := miner.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl, nil)
	//Check consistency
	for i := 0; i < 10; i++ {
		fee, err := blk.EstimateStandardTXFee(1)
		if err != nil {
			t.Logf("%d - cannot estimate fee: %v", i, err)
			t.FailNow()
		}
		if fee != 115 {
			t.Logf("%d - fee estimation failed: %d", i, fee.Satoshi())
			t.FailNow()
		}
	}
}

func TestBlockchain_Submit(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	t.SkipNow()
	he := "010000000155f058142e60b3d6f9f16667b7e9c10615be1c698f78b85362a4f50d906b70e6010000006a473044022076b0dd878a1d7b6919c5c8becc3f2d993436dac616bc5f055273eda570c9d59502204ab86e52cb85ff425f16001633774618a5f3b0a7108ab9a73638e3ef3fa25a684121032f8bdd0bdb654616c362a427a01cf7abafa0b61831297c09211998ede8b99b45ffffffff02a89c0000000000001976a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac00000000000000001e006a1b646462202d2052656d696e64204d792e2e2e20627920656a66687000000000"
	tx, err := ddb.DataTXFromHex(he)
	if err != nil {
		t.Logf("cannot decode hex: %v", err)
		t.FailNow()
	}
	txs := []*ddb.DataTX{tx}
	miner := miner.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl, nil)
	ids, err := blk.Submit(txs)
	if len(ids) != len(txs) {
		t.Logf("unexpected num of returned id: %d  exp: %d", len(ids), len(txs))
		t.Fail()
	}
	if err != nil {
		t.Logf("sumbit failed: %v", err)
	}
}

func TestBlockchain_ListTXIDs(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	miner := miner.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl, nil)
	txids, err := blk.ListTXIDs(destinationAddress, false)
	if err != nil {
		t.Logf("failed to list TXs: %v", err)
		t.Fail()
	}
	if len(txids) < 20 {
		t.Logf("unexpected number of TXIDs: %d", len(txids))
		t.Fail()
	}
}

func TestBlockchain_GetTX(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	txid := "afbdf4a215f5e7dc3beca36e1625f3597995afa5906b2bbfee6a572d87764426"
	miner := miner.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(miner, expl, nil)
	dataTx, err := blk.GetTX(txid, false)
	if err != nil {
		t.Logf("failed to get data TXs: %v", err)
		t.Fail()
	}
	if dataTx.GetTxID() != txid {
		t.Logf("unexpected ID: %s", dataTx.GetTxID())
		t.Fail()
	}
}
func TestBlockchain_GetTX_OnlyCache(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	cache, _ := os.UserCacheDir()
	cacheDir := filepath.Join(cache, "trh")
	txCache, err := ddb.NewTXCache(cacheDir)
	if err != nil {
		t.Logf("failed to get data TX Cache: %v", err)
		t.Fail()
	}
	tx := Helper_FakeTX(t)
	txid := tx.GetTxID()
	txCache.StoreTX(txid, tx.ToBytes())
	mir := miner.NewTAAL()
	expl := ddb.NewWOC()
	blk := ddb.NewBlockchain(mir, expl, txCache)
	dataTx, err := blk.GetTX(txid, true)
	if err != nil {
		t.Logf("failed to get data TXs: %v", err)
		t.Fail()
	}
	if dataTx.GetTxID() != txid {
		t.Logf("unexpected ID: %s", dataTx.GetTxID())
		t.Fail()
	}
	dataTx, err = blk.GetTX(txid, true)
	if err != nil {
		t.Logf("failed to get data TXs: %v", err)
		t.Fail()
	}
	if dataTx.GetTxID() != txid {
		t.Logf("unexpected ID: %s", dataTx.GetTxID())
		t.Fail()
	}
}

// func TestBlockchain_ListTXHistoryBackward(t *testing.T) {
// 	trail.SetWriter(os.Stdout)
// 	txid := "8c4e50050f1a82e14765f4a79b2bdac700967e592486dcaa9eedb4f8bf441d16"
// 	miner := ddb.NewTAAL()
// 	expl := ddb.NewWOC()
// 	blk := ddb.NewBlockchain(miner, expl, nil)
// 	limit := 23
// 	path, err := blk.ListTXHistoryBackward(txid, address, limit)
// 	if err != nil {
// 		t.Logf("search backward failed: %v", err)
// 		t.FailNow()
// 	}
// 	if len(path) != limit {
// 		t.Logf("Unexpected path len: %d", len(path))
// 		t.Fail()
// 	}
// 	if path[limit-1] != "4f438cf8954a475684f460461b3a66955e9ced065dbd74c00deae4dd12f7843d" {
// 		t.Logf("Unexpected first TXID")
// 		t.Fail()
// 	}
// 	for i, p := range path {
// 		t.Logf("%d: %s", i, p)
// 	}
// }
