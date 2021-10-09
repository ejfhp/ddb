package satoshi_test

import (
	"errors"
	"testing"

	"github.com/ejfhp/ddb/satoshi"
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
		b := satoshi.Bitcoin(f)
		sat := b.Satoshi()
		if sat != satoshi.Satoshi(s) {
			t.Logf("Amount are different! %d != %d", s, sat)
			t.Fail()
		}
		bit := satoshi.Satoshi(sat)
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
		res := satoshi.Bitcoin(v[0]).Add(satoshi.Bitcoin(v[1]))
		if float64(res.Bitcoin()) != v[2] {
			t.Logf("%d: sum is not correct! %0.8f + %0.8f = %0.30f != %0.30f", i, v[0], v[1], res.Bitcoin(), v[2])
			t.Fail()
		}
	}
}

func TestSubBitcoin(t *testing.T) {
	bitsat := [][]float64{
		{1, 0.00000001, 0.99999999},
		{21000000, 0.00000001, 20999999.99999999},
		{0.00000001, 0.00000001, 0.00000000},
		{0.00000016, 0.00000001, 0.00000015},
		{0.10000016, 0.10000001, 0.00000015},
		{0.10000016, 0.10000021, 9999},
	}
	for i, v := range bitsat {
		res, err := satoshi.Bitcoin(v[0]).Sub(satoshi.Bitcoin(v[1]))
		if err != nil {
			if !errors.Is(err, satoshi.ErrNegativeAmount) {
				t.Logf("Wrong kind of error")
				t.Fail()
			}
		}
		if v[2] == 9999 && err == nil {
			t.Logf("negative result must return error")
			t.Fail()
		}
		if v[2] != 9999 && float64(res.Bitcoin()) != v[2] {
			t.Logf("%d: sum is not correct! %0.8f - %0.8f = %0.30f != %0.30f", i, v[0], v[1], res.Bitcoin(), v[2])
			t.Fail()
		}
	}
}

func TestSubSatoshi(t *testing.T) {
	bitsat := [][]uint64{
		{100000000, 1, 99999999},
		{2100000000000000, 1, 2099999999999999},
		{1, 1, 0},
		{16, 1, 15},
		{10000016, 10000001, 15},
		{10000016, 10000021, 9999},
	}
	for i, v := range bitsat {
		res, err := satoshi.Satoshi(v[0]).Sub(satoshi.Satoshi(v[1]))
		if err != nil {
			if !errors.Is(err, satoshi.ErrNegativeAmount) {
				t.Logf("Wrong kind of error")
				t.Fail()
			}
		}
		if v[2] == 9999 && err == nil {
			t.Logf("negative result must return error")
			t.Fail()
		}
		if v[2] != 9999 && uint64(res.Satoshi()) != v[2] {
			t.Logf("%d: sum is not correct! %d - %d = %d != %d", i, v[0], v[1], res.Satoshi(), v[2])
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
		res := satoshi.Satoshi(v[0]).Add(satoshi.Bitcoin(v[1]))
		if float64(res.Bitcoin()) != v[2] {
			t.Logf("%d: sum is not correct! %0.8f + %0.8f = %0.30f != %0.30f", i, v[0], v[1], res.Bitcoin(), v[2])
			t.Fail()
		}
	}
}
