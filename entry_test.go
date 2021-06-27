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
	// bytes, err := ioutil.ReadFile("testdata/image.png")
	bytes, err := ioutil.ReadFile("testdata/test.txt")
	initLen := len(bytes)
	if err != nil {
		t.Fatalf("error reading test file: %v", err)
	}
	entries, err := ddb.EntryOfFile("test.txt", bytes, 1000)
	if err != nil {
		t.Fatalf("error making entries of test file: %v", err)
	}
	// if len(entries) != 4 {
	// 	t.Fatalf("Wronf num of entries: %d", len(entries))
	// }
	var data []byte
	var entityHash string
	for i, f := range entries {
		if data == nil {
			data = make([]byte, 0)
		}
		t.Logf("%d size: %s  %d   hash: %s\n", i, f.Mime, len(f.Data), f.Hash)
		data = append(data, f.Data...)
		entityHash = f.Hash
	}
	endLen := len(data)
	if initLen != endLen {
		t.Logf("initLen != endLen => %d != %d ", initLen, endLen)
		t.Fail()

	}
	hash := make([]byte, 64)
	sha := sha256.Sum256(data)
	hex.Encode(hash, sha[:])
	if entityHash != string(hash) {
		t.Logf("the read file is wrong =>\n   %s\n!= %s", hash, entityHash)
		t.Fail()
	}
	ioutil.WriteFile("readtest.txt", data, 0664)
}
