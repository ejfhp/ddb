package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

func TestEntryOfFile(t *testing.T) {
	log.SetWriter(os.Stdout)
	inputs := [][]string{
		{"testdata/test.txt", "text/plain; charset=utf-8"},
		{"testdata/image.png", "image/png"},
	}
	partsizes := [][]int{
		{100, 9},
		{1000, 4},
	}
	for in, fil := range inputs {
		bytes, err := ioutil.ReadFile(fil[0])
		hash := make([]byte, 64)
		sha := sha256.Sum256(bytes)
		hex.Encode(hash, sha[:])
		initLen := len(bytes)
		if err != nil {
			t.Fatalf("%d: error reading test file: %v", in, err)
		}
		entries, err := ddb.EntryOfFile(fil[0], bytes, partsizes[in][0])
		if err != nil {
			t.Fatalf("%d: error making entries of test file: %v", in, err)
		}
		if len(entries) != partsizes[in][1] {
			t.Fatalf("%d: wrong num of entries: %d != %d", in, len(entries), partsizes[in][1])
		}
		var data []byte
		for i, f := range entries {
			if data == nil {
				data = make([]byte, 0)
			}
			if f.Mime != fil[1] {
				t.Logf("wrong mime, should be %s but is %s ", fil[1], f.Mime)
				t.Fail()
			}
			if f.IdxPart != i {
				t.Logf("wrong idxpart, should be %d but is %d ", i, f.IdxPart)
				t.Fail()
			}
			if f.NumPart != partsizes[in][1] {
				t.Logf("wrong numpart, should be %d but is %d ", partsizes[in][1], f.NumPart)
				t.Fail()
			}
			if f.Hash != string(hash) {
				t.Logf("wrong hash, \nexp %s \ngot %s ", string(hash), f.Hash)
				t.Fail()
			}
			if len(f.Data) != f.Size {
				t.Logf("wrong size, %d but len data is %d ", f.Size, len(f.Data))
				t.Fail()

			}
			t.Logf("%d name:%s mime:%s part:%d/%d size:%d len(data):%d hash:%s\n", i, f.Name, f.Mime, f.IdxPart, f.NumPart, f.Size, len(f.Data), f.Hash)
			data = append(data, f.Data...)
		}
		endLen := len(data)
		if initLen != endLen {
			t.Logf("initLen != endLen => %d != %d ", initLen, endLen)
			t.Fail()

		}
		readhash := make([]byte, 64)
		readsha := sha256.Sum256(data)
		hex.Encode(readhash, readsha[:])
		if string(string(readhash)) != string(hash) {
			t.Logf("the read file has wrong hash \nexp %s\n got %s", string(hash), string(readhash))
			t.Fail()
		}
		ioutil.WriteFile(fmt.Sprintf("/tmp/%d.%s", in, fil[1]), data, 0664)
	}
}
