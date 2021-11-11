package trh_test

import (
	"testing"

	"github.com/ejfhp/ddb/trh"
)

func TestEstimate_Estimate(t *testing.T) {
	file := "../testdata/image.png"
	th := &trh.TRH{}
	txs, fee, err := th.Estimate(file, []string{"label1", "label2"}, "a lot of notes")
	if err != nil {
		t.Logf("estimate returns error: %v", err)
		t.FailNow()
	}
	if fee < 1 {
		t.Logf("Unexpected fee: %d", fee)
		t.FailNow()
	}
	if len(txs) < 1 {
		t.Logf("Unexpected num of tx: %d", len(txs))
		t.FailNow()

	}
}
