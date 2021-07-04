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
		entry := ddb.Entry{Name: fil[0], Data: bytes}
		entries, err := entry.Parts(partsizes[in][0])
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

func TestEncodeDecodeEntry(t *testing.T) {
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
	de, err := ddb.Decode(encoded)
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

func TestEncryptDecrypt(t *testing.T) {
	key := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2}
	tests := []string{
		"tanto va la gatta al lardo che ci lascia lo zampino",
		`Nel mezzo del cammin di nostra vita
		mi ritrovai per una selva oscura,
		ché la diritta via era smarrita.
		
		Ahi quanto a dir qual era è cosa dura
		esta selva selvaggia e aspra e forte
		che nel pensier rinova la paura!
		
		Tant’è amara che poco è più morte;
		ma per trattar del ben ch’i’ vi trovai,
		dirò de l’altre cose ch’i’ v’ ho scorte.
		
		Io non so ben ridir com’i’ v’intrai,
		tant’era pien di sonno a quel punto
		che la verace via abbandonai.
		
		Ma poi ch’i’ fui al piè d’un colle giunto,
		là dove terminava quella valle
		che m’avea di paura il cor compunto,
		
		guardai in alto e vidi le sue spalle
		vestite già de’ raggi del pianeta
		che mena dritto altrui per ogne calle.`,
	}
	for i, txt := range tests {
		crypted1, err := ddb.AESEncrypt(key, []byte(txt))
		if err != nil {
			t.Logf("first encryption has failed: %v", err)
			t.Fail()
		}
		//Second encoding to encode random bytes and not text
		crypted2, err := ddb.AESEncrypt(key, []byte(crypted1))
		if err != nil {
			t.Logf("second encryption has failed: %v", err)
			t.Fail()
		}
		decrypted1, err := ddb.AESDecrypt(key, crypted2)
		if err != nil {
			t.Logf("first decryption has failed: %v", err)
			t.Fail()
		}
		decrypted2, err := ddb.AESDecrypt(key, decrypted1)
		if err != nil {
			t.Logf("second decryption has failed: %v", err)
			t.Fail()
		}
		if string(decrypted2) != txt {
			t.Logf("%d: encryption/decription failed '%s' != '%s'", i, string(decrypted2), txt)
			t.Fail()

		}
	}
}
