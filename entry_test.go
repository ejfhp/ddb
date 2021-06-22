package ddb_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestEntryOfFile(t *testing.T) {
	log.SetWriter(os.Stdout)
	bytes, err := ioutil.ReadFile("testdata/image.png")
	if err != nil {
		t.Fatalf("error reading test file: %v", err)
	}
	entries, err := ddb.EntryOfFile("image.png", bytes, 1000)
	if len(entries) < 4 {
		t.Fatalf("Incomplete entries: %d", len(entries))
	}
	for i, f := range entries {

		fmt.Printf("%d size: %s  %d\n", i, f.FeeType, f.MiningFee.Satoshis)
	}
}
