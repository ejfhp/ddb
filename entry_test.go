package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"mime"
	"path/filepath"
	"testing"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/satoshi"
)

func TestEntry_NewEntryFromFile(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	inputs := [][]string{
		{"testdata/test.txt", ""},
		{"testdata/image.png", ""},
		{"testdata", "error"},
		{"testdata/", "error"},
		{"/", "error"},
	}
	for i, f := range inputs {
		e, err := ddb.NewEntryFromFile(filepath.Base(f[0]), f[0], []string{"label1", "label2"}, "notes")
		if err != nil && f[1] != "error" {
			t.Logf("%d NewEntry failed: %s", i, f)
		}
		if err == nil && f[1] == "error" {
			t.Logf("%d NewEntry failed: %s", i, f)
		}
		if e == nil {
			t.Logf("%d NewEntry failed: %s", i, f)
		}
	}

}
func TestEntry_ToParts(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	inputs := [][]string{
		{"testdata/test.txt", "text/plain; charset=utf-8"},
		{"testdata/image.png", "image/png"},
	}
	partsizes := [][]int{
		{500, 5},
		{2000, 4},
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
		entry, err := ddb.NewEntryFromFile(fil[0], fil[0], []string{"label1", "label2"}, "notes")
		if err != nil {
			t.Fatalf("%d: error making entry of test file: %v", in, err)
		}
		parts, err := entry.ToParts([32]byte{}, partsizes[in][0])
		if err != nil {
			t.Fatalf("%d: error making entryParts of test file: %v", in, err)
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
				t.Logf("%d - wrong mime, should be %s but is %s ", i, fil[1], f.Mime)
				t.Fail()
			}
			if f.IdxPart != i {
				t.Logf("%d - wrong idxpart, should be %d but is %d ", i, i, f.IdxPart)
				t.Fail()
			}
			if f.NumPart != partsizes[in][1] {
				t.Logf("%d - wrong numpart, should be %d but is %d ", i, partsizes[in][1], f.NumPart)
				t.Fail()
			}
			if f.Hash != string(hash) {
				t.Logf("%d - wrong hash, \nexp %s \ngot %s ", i, string(hash), f.Hash)
				t.Fail()
			}
			if len(f.Data) != f.Size {
				t.Logf("%d - wrong size, %d but len data is %d ", i, f.Size, len(f.Data))
				t.Fail()

			}
			//fmt.Printf("'%s'\n", hex.EncodeToString(f.Data))
			// t.Logf("%d name:%s mime:%s part:%d/%d size:%d len(data):%d hash:%s\n", i, f.Name, f.Mime, f.IdxPart, f.NumPart, f.Size, len(f.Data), f.Hash)
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
		// fmt.Println(string(data))
		if string(string(readhash)) != string(hash) {
			t.Logf("the read file has wrong hash \nexp %s\n got %s", string(hash), string(readhash))
			t.Fail()
		}
	}
}

func TestEntry_EntriesFromParts(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	inputs := []string{
		"testdata/test.txt",
		"testdata/image.png",
	}
	partLen := 433
	parts := make([]*ddb.EntryPart, 0)
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	for in, fil := range inputs {
		bytes, err := ioutil.ReadFile(fil)
		if err != nil {
			t.Fatalf("%d: error reading test file: %v", in, err)
		}
		eh := sha256.Sum256(bytes)
		ehash := hex.EncodeToString(eh[:])
		ent := ddb.Entry{Name: fil, Data: bytes, Hash: ehash}
		pts, err := ent.ToParts(password, partLen)
		if err != nil {
			t.Logf("cannot get entry parts: %v", err)
			t.FailNow()
		}
		parts = append(parts, pts...)
	}
	//Normal case
	entries1, err := ddb.EntriesFromParts(parts)
	if err != nil {
		t.Logf("normal - cannot rebuild entries: %v", err)
		t.FailNow()
	}
	if len(entries1) != 2 {
		t.Logf("normal - wrong num of entries: %d", len(entries1))
		t.FailNow()
	}

	//Incomplete case
	entries2, err := ddb.EntriesFromParts(parts[1:])
	if err != nil {
		t.Logf("incomplete - cannot rebuild entries: %v", err)
		t.FailNow()
	}
	if len(entries2) != 1 {
		t.Logf("incomplete - wrong num of entries: %d", len(entries2))
		t.FailNow()
	}

	//Mixed case
	parts3 := append(parts[4:13], append(parts[:23], parts[18:]...)...)
	entries3, err := ddb.EntriesFromParts(parts3)
	if err != nil {
		t.Logf("mixed - cannot rebuild entries: %v", err)
		t.FailNow()
	}
	if len(entries3) != 2 {
		t.Logf("mixed - wrong num of entries: %d", len(entries3))
		t.FailNow()
	}
}

func TestMetaEntry_Encrypt_MetaEntryFromEncrypted(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	fil := "testdata/test.txt"
	bytes, err := ioutil.ReadFile(fil)
	if err != nil {
		t.Fatalf("error reading test file: %v", err)
	}
	eh := sha256.Sum256(bytes)
	ehash := hex.EncodeToString(eh[:])
	ent := &ddb.Entry{Name: fil, Data: bytes, Hash: ehash}
	mentry := ddb.NewMetaEntry(ent)
	mentry.Labels = []string{"test"}
	mentry.Notes = "Notes"
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	enc, err := mentry.Encrypt(password)
	if err != nil {
		t.Logf("Entry encrypt failed: %v", err)
		t.Fail()
	}
	de, err := ddb.MetaEntryFromEncrypted(password, enc)
	if err != nil {
		t.Logf("Entry decrypt failed: %v", err)
		t.Fail()
	}
	if de == nil {
		t.Fatalf("Decoded entry is nil")
	}
	if de.Name != ent.Name || de.Labels[0] != mentry.Labels[0] || de.Notes != mentry.Notes {
		t.Logf("MetaEntry encrypt/decrypt failed, some fields doesn't match: %v", de)
		t.Fail()
	}
}

func TestEntry_ToJSON_EntryPartFromJSON(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	name := "test.txt"
	hash := "hhhhhhhhhhhhhhhh"
	mime := "text/plain"
	idxPart := 1
	numPart := 3
	size := 30
	data := []byte("this is the data")
	e := ddb.EntryPart{Name: name, Hash: hash, Mime: mime, IdxPart: idxPart, NumPart: numPart, Size: size, Data: data}
	encoded, err := e.ToJSON()
	if err != nil {
		t.Logf("Entry encoding failed: %v", err)
		t.Fail()
	}
	if len(encoded) < 20 {
		t.Logf("Encoded entry too short: %s", string(encoded))
		t.Fail()
	}
	de, err := ddb.EntryPartFromJSON(encoded)
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

func TestEntry_EncodedToAndReadFromDataTX(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
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
	parts, err := e.ToParts(password, 1000)
	if err != nil {
		t.Logf("Entry to parts failed: %v", err)
		t.Fail()
	}
	//encrypting
	cryParts := make([][]byte, 0, len(parts))
	for _, p := range parts {
		ep, err := p.Encrypt(password)
		if err != nil {
			t.Logf("EntryPart encrypting failed: %v", err)
			t.Fail()
		}
		cryParts = append(cryParts, ep)
	}
	//pack
	txid := "e6706b900df5a46253b8788f691cbe1506c1e9b76766f1f9d6b3602e1458f055"
	scriptHex := "76a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac"
	utxos := []*ddb.UTXO{{TXPos: 1, TXHash: txid, Value: satoshi.Bitcoin(1), ScriptPubKeyHex: scriptHex}}
	txs := make([]*ddb.DataTX, 0, len(cryParts))
	for _, p := range cryParts {
		tx, err := ddb.NewDataTX(destinationKey, destinationAddress, changeAddress, utxos, satoshi.Satoshi(10), satoshi.Satoshi(200), p, "123456789")
		if err != nil {
			t.Logf("EntryPart packing failed: %v", err)
			t.Fail()
		}
		txs = append(txs, tx)
	}
	//tx to hex
	hextxs := make([]string, 0, len(txs))
	for _, p := range txs {
		ep := p.ToString()
		hextxs = append(hextxs, ep)
	}

	/////////////////////////////////////////////////////////
	//hex to tx
	datatxs := make([]*ddb.DataTX, 0, len(hextxs))
	for _, p := range hextxs {
		ep, err := ddb.DataTXFromHex(p)
		if err != nil {
			t.Logf("DataTX from hex failed: %v", err)
			t.Fail()
		}
		datatxs = append(datatxs, ep)
	}
	//unpacking
	oprets := make([][]byte, 0, len(datatxs))
	for _, tx := range datatxs {
		de, header, err := tx.Data()
		if err != nil {
			t.Logf("OP_RETURN retrieval failed: %v", err)
			t.Fail()
		}
		if len(header) != 9 {
			t.Logf("Wrong headerl len(9): %s", header)
			t.Fail()
		}
		oprets = append(oprets, de)
	}
	//decrypting
	decEntryParts := make([]*ddb.EntryPart, 0, len(oprets))
	for _, p := range oprets {
		de, err := ddb.EntryPartFromEncrypted(password, p)
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
