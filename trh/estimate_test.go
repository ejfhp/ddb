package trh_test

import (
	"testing"

	"github.com/ejfhp/ddb/trh"
)

func TestEstimate_Estimate(t *testing.T) {
	file := "../testdata/image.png"
	fee, err := trh.Estimate(file, []string{"label1", "label2"}, "a lot of notes")
}
