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
	var data []byte
	var entityHash string
	for i, f := range entries {
		if data == nil {
			data = make([]byte, f.Size)
		}
		t.Logf("%d size: %s  %d   hash: %s\n", i, f.Mime, len(f.Data), f.Hash)
		data = append(data, f.Data...)
		entityHash = f.Hash
	}
	hash := make([]byte, 64)
	sha := sha256.Sum256(data)
	hex.Encode(hash, sha[:])
	if entityHash != string(hash) {
		t.Logf("the read file is wrong =>\n   %s\n!= %s", hash, entityHash)
		t.Fail()
	}
	ioutil.WriteFile("readimage.png", data, 664)
}
