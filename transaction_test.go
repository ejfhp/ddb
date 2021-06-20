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
		Value:  ddb.FromBitcoin(0.000402740),
		ScriptPubKey: &ddb.ScriptPubKey{
			Asm:      "OP_DUP OP_HASH160 2f353ff06fe8c4d558b9f58dce952948252e5df7 OP_EQUALVERIFY OP_CHECKSIG",
			Hex:      "76a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac",
			ReqSigs:  1,
			Type:     "pubkeyhash",
			Adresses: []string{"15JcYsiTbhFXxU7RimJRyEgKWnUfbwttb3"},
		},
	}
	tx, err := ddb.BuildOPReturnBytesTX(utxo, key, ddb.FromSatoshis(170), payload)
	if err != nil {
		t.Fatalf("failed to create tx: %v", err)
	}
	t.Logf("TX len: %d", len(tx))
	if len(tx) < 100 {
		t.Logf("failed to create tx, []byte too short: %d", len(tx))
		t.Fail()
	}
}

func TestBitcoinToSatoshi(t *testing.T) {
	bitsat := map[uint64]float64{
		1:                0.00000001,
		211337:           0.00211337,
		211338:           0.00211338,
		211336:           0.00211336,
		100211337:        1.00211337,
		100000000037:     1000.00000037,
		2100000000000001: 21000000.00000001,
		10:               0.00000010,
		11:               0.00000011,
		12:               0.00000012,
		13:               0.00000013,
		14:               0.00000014,
		15:               0.00000015,
		16:               0.00000016,
		17:               0.00000017,
		18:               0.00000018,
		19:               0.00000019,
	}
	for s, f := range bitsat {
		b := ddb.FromBitcoin(f)
		sat := b.Satoshis()
		if sat != s {
			t.Logf("Amount are different! %d != %d", s, sat)
			t.Fail()
		}
		bit := ddb.FromSatoshis(sat)
		if bit.Value() != b.Value() {
			t.Logf("Amount are different! %0.30f != %0.30f", b.Value(), bit.Value())
			t.Fail()
		}
	}
}

func TestSumBitcoin(t *testing.T) {
	bitsat := [][]float64{
		{1, 0.00000001, 1.00000001},
		{21000000, 0.00000001, 21000000.00000001},
		{0.00000001, 0.00000001, 0.00000002},
		{0.00000016, 0.00000001, 0.00000017},
		{0.10000016, 0.10000001, 0.20000017},
	}
	for i, v := range bitsat {
		res := ddb.FromBitcoin(v[0]).Add(ddb.FromBitcoin(v[1]))
		if res.Value() != v[2] {
			t.Logf("%d: sum is not correct! %0.8f + %0.8f = %0.30f != %0.30f", i, v[0], v[1], res.Value(), v[2])
			t.Fail()
		}
	}
}
