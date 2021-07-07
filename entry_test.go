package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
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
		entry := ddb.Entry{Name: fil[0], Data: bytes}
		parts, err := entry.Parts(partsizes[in][0])
		if err != nil {
			t.Fatalf("%d: error making entries of test file: %v", in, err)
		}
		if len(parts) != partsizes[in][1] {
			t.Fatalf("%d: wrong num of entries: %d != %d", in, len(parts), partsizes[in][1])
		}
		var data []byte
		for i, f := range parts {
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
			// fmt.Printf("'%s'\n", hex.EncodeToString(f.Data))
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

func TestEntriesFromParts(t *testing.T) {
	log.SetWriter(os.Stdout)
	inputs := []string{
		"testdata/test.txt",
		"testdata/image.png",
	}
	partLen := 133
	parts := make([]*ddb.EntryPart, 0)
	for in, fil := range inputs {
		bytes, err := ioutil.ReadFile(fil)
		if err != nil {
			t.Fatalf("%d: error reading test file: %v", in, err)
		}
		ent := ddb.Entry{Name: fil, Data: bytes}
		pts, err := ent.Parts(partLen)
		if err != nil {
			t.Logf("cannot get entry parts: %v", err)
			t.Fail()
		}
		parts = append(parts, pts...)
	}
	//Normal case
	entries1, err := ddb.EntriesFromParts(parts)
	if err != nil {
		t.Logf("cannot rebuild entries: %v", err)
		t.Fail()
	}
	if len(entries1) != 2 {
		t.Logf("wrong num of entries: %d", len(entries1))
		t.Fail()
	}

	//Incomplete case
	entries2, err := ddb.EntriesFromParts(parts[1:])
	if err != nil {
		t.Logf("cannot rebuild entries: %v", err)
		t.Fail()
	}
	if len(entries2) != 1 {
		t.Logf("wrong num of entries: %d", len(entries2))
		t.Fail()
	}

	//Mixed case
	parts3 := append(parts[4:13], append(parts[:23], parts[18:]...)...)
	entries3, err := ddb.EntriesFromParts(parts3)
	if err != nil {
		t.Logf("cannot rebuild entries: %v", err)
		t.Fail()
	}
	if len(entries3) != 2 {
		t.Logf("wrong num of entries: %d", len(entries3))
		t.Fail()
	}
}

func TestEncodeDecodeEntryPart(t *testing.T) {
	log.SetWriter(os.Stdout)
	name := "test.txt"
	hash := "hhhhhhhhhhhhhhhh"
	mime := "text/plain"
	idxPart := 1
	numPart := 3
	size := 30
	data := []byte("this is the data")
	e := ddb.EntryPart{Name: name, Hash: hash, Mime: mime, IdxPart: idxPart, NumPart: numPart, Size: size, Data: data}
	encoded, err := e.Encode()
	if err != nil {
		t.Logf("Entry encoding failed: %v", err)
		t.Fail()
	}
	if len(encoded) < 20 {
		t.Logf("Encoded entry too short: %s", string(encoded))
		t.Fail()
	}
	de, err := ddb.EntryPartFromEncodedData(encoded)
	if err != nil {
		t.Logf("Entry decoding failed: %v", err)
		t.Fail()
	}
	if de == nil {
		t.Fatalf("Decoded entry is nil")
	}
	if de.Name != name || de.Hash != hash || de.Mime != mime || de.IdxPart != idxPart || de.NumPart != numPart || de.Size != size || de.Data == nil {
		t.Logf("Entry decoding failed, some fields doesn't match.")
		t.Fail()

	}
}

func TestEncodeDecodeSingleEntry(t *testing.T) {
	log.SetWriter(os.Stdout)
	name := "image.png"
	file := "testdata/image.png"
	image, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("error reading test file %s: %v", file, err)
	}
	imageSha := sha256.Sum256(image)
	imageHash := hex.EncodeToString(imageSha[:])
	mime := mime.TypeByExtension(filepath.Ext(name))
	e := ddb.Entry{Name: name, Hash: imageHash, Mime: mime, Data: image}
	parts, err := e.Parts(1000)
	if err != nil {
		t.Logf("Entry to parts failed: %v", err)
		t.Fail()
	}
	encParts := make([][]byte, 0, len(parts))
	for _, p := range parts {
		ep, err := p.Encode()
		if err != nil {
			t.Logf("EntryPart encoding failed: %v", err)
			t.Fail()
		}
		encParts = append(encParts, ep)
	}
	decEntryParts := make([]*ddb.EntryPart, 0, len(encParts))
	for _, encp := range encParts {
		de, err := ddb.EntryPartFromEncodedData(encp)
		if err != nil {
			t.Logf("Entry decoding failed: %v", err)
			t.Fail()
		}
		decEntryParts = append(decEntryParts, de)
	}
	out, err := ddb.EntriesFromParts(decEntryParts)
	if err != nil {
		t.Logf("Entry composition failed: %v", err)
		t.Fail()
	}
	if len(out) != 1 {
		t.Logf("Entry decoded should be 1: %d", len(out))
		t.Fail()
	}
	outSha := sha256.Sum256(out[0].Data)
	outHash := hex.EncodeToString(outSha[:])
	if outHash != imageHash {
		t.Logf("Entry decoded hash different")
		t.Fail()

	}

}
