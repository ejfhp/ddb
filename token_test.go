package ddb_test

import (
	"testing"

	"github.com/ejfhp/ddb"
)

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
		b := ddb.Bitcoin(f)
		sat := b.Satoshi()
		if sat != ddb.Satoshi(s) {
			t.Logf("Amount are different! %d != %d", s, sat)
			t.Fail()
		}
		bit := ddb.Satoshi(sat)
		if bit != b.Satoshi() {
			t.Logf("Amount are different! %0.30f != %0.30f", b, bit.Bitcoin())
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
		res := ddb.Bitcoin(v[0]).Add(ddb.Bitcoin(v[1]))
		if float64(res.Bitcoin()) != v[2] {
			t.Logf("%d: sum is not correct! %0.8f + %0.8f = %0.30f != %0.30f", i, v[0], v[1], res.Bitcoin(), v[2])
			t.Fail()
		}
	}
}

func TestSumBitcoinSatoshi(t *testing.T) {
	bitsat := [][]float64{
		{1, 0.00000001, 0.00000002},
		{21000000, 0.00000001, 0.21000001},
		{10, 0.00000001, 0.00000011},
	}
	for i, v := range bitsat {
		res := ddb.Satoshi(v[0]).Add(ddb.Bitcoin(v[1]))
		if float64(res.Bitcoin()) != v[2] {
			t.Logf("%d: sum is not correct! %0.8f + %0.8f = %0.30f != %0.30f", i, v[0], v[1], res.Bitcoin(), v[2])
			t.Fail()
		}
	}
}
