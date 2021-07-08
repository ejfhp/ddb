package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
	"testing"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

var address = "1PGh5YtRoohzcZF7WX8SJeZqm6wyaCte7X"
var key = "L4ZaBkP1UTyxdEM7wysuPd1scHMLLf8sf8B2tcEcssUZ7ujrYWcQ"

func TestProcessEntry(t *testing.T) {
	log.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc)
	logbook, err := ddb.NewLogbook(key, password, blockchain)
	if err != nil {
		t.Logf("failed to create new Logbook: %v", err)
		t.Fail()
	}
	filename := "Inferno"
	file := `Nel mezzo del cammin di nostra vita
		mi ritrovai per una selva oscura,
		ché la diritta via era smarrita.
		
		Ahi quanto a dir qual era è cosa dura
		esta selva selvaggia e aspra e forte
		che nel pensier rinova la paura!`
	entry := ddb.Entry{Name: filename, Data: []byte(file)}
	txs, err := logbook.ProcessEntry(&entry)
	if err != nil {
		t.Logf("failed to process entry: %v", err)
		t.Fail()
	}
	for _, tx := range txs {
		opr, ver, err := tx.Data()
		if err != nil {
			t.Logf("failed to get OP_RETURN: %v", err)
			t.FailNow()
		}
		if ver != ddb.VER_AES {
			t.Logf("wrong version: %s", ver)
			t.FailNow()
		}
		decrypt, err := ddb.AESDecrypt(password, opr)
		if err != nil {
			t.Logf("failed to decrypt: %v", err)
			t.FailNow()
		}
		ep, err := ddb.EntryPartFromEncodedData(decrypt)
		if err != nil {
			t.Logf("failed to decode EntryPart: %v", err)
			t.FailNow()
		}
		if ep.Name != filename {
			t.Logf("unexpected name: %s != %s", ep.Name, filename)
			t.FailNow()

		}
	}
}

func TestCastEntry(t *testing.T) {
	log.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc)
	logbook, err := ddb.NewLogbook(key, password, blockchain)
	if err != nil {
		t.Logf("failed to create new Logbook: %v", err)
		t.Fail()
	}
	filename := "Inferno.txt"
	file := `Nel mezzo del cammin di nostra vita
		mi ritrovai per una selva oscura,
		ché la diritta via era smarrita.
		
		Ahi quanto a dir qual era è cosa dura
		esta selva selvaggia e aspra e forte
		che nel pensier rinova la paura!`
	entry := ddb.Entry{Name: filename, Data: []byte(file), Mime: ""}
	txs, err := logbook.ProcessEntry(&entry)
	if err != nil {
		t.Logf("failed to process entry: %v", err)
		t.Fail()
	}
	for _, tx := range txs {
		opr, ver, err := tx.Data()
		if err != nil {
			t.Logf("failed to get OP_RETURN: %v", err)
			t.FailNow()
		}
		if ver != ddb.VER_AES {
			t.Logf("wrong version: %s", ver)
			t.FailNow()
		}
		decrypt, err := ddb.AESDecrypt(password, opr)
		if err != nil {
			t.Logf("failed to decrypt: %v", err)
			t.FailNow()
		}
		ep, err := ddb.EntryPartFromEncodedData(decrypt)
		if err != nil {
			t.Logf("failed to decode EntryPart: %v", err)
			t.FailNow()
		}
		if ep.Name != filename {
			t.Logf("unexpected name: %s != %s", ep.Name, filename)
			t.FailNow()

		}
	}
}

func TestLogbookEntryFullCycleText(t *testing.T) {
	log.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc)
	logbook, err := ddb.NewLogbook(key, password, blockchain)
	name := "test.txt"
	fm := mime.TypeByExtension(filepath.Ext(name))
	bytes := []byte("just a test")
	sha := sha256.Sum256(bytes)
	hash := hex.EncodeToString(sha[:])
	entry := &ddb.Entry{Name: name, Mime: fm, Hash: hash, Data: bytes}
	txs, err := logbook.ProcessEntry(entry)
	t.Logf("txs len: %d", len(txs))
	if err != nil {
		t.Logf("txs preparation failed")
		t.Fail()
	}
	// here data should be cast to blockchain and then
	// retrieved trough a blockchain explorer
	ents, err := logbook.ExtractEntries(txs)
	if err != nil {
		t.Logf("entry extraction failed")
		t.Fail()
	}
	if len(ents) != 1 {
		t.Logf("len(entries) should be 1: %d", len(ents))
		t.FailNow()
	}
	if ents[0].Name != name {
		t.Logf("unexpected name: %s != %s", name, ents[0].Name)
		t.Fail()
	}
	if ents[0].Mime != fm {
		t.Logf("unexpected mime: %s != %s", fm, ents[0].Mime)
		t.Fail()
	}
	if ents[0].Hash != hash {
		t.Logf("unexpected hash: %s != %s", hash, ents[0].Hash)
		t.Fail()
	}
	shaOut := sha256.Sum256(ents[0].Data)
	hashOut := hex.EncodeToString(shaOut[:])
	if hashOut != hash {
		t.Logf("unexpected hash of extracted data: %s != %s", hash, hashOut)
		t.Fail()
	}
}

func TestLogbookEntryFullCycleImage(t *testing.T) {
	log.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc)
	logbook, err := ddb.NewLogbook(key, password, blockchain)
	name := "image.png"
	file := "testdata/image.png"
	image, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("error reading test file %s: %v", file, err)
	}
	imageSha := sha256.Sum256(image)
	imageHash := hex.EncodeToString(imageSha[:])
	fm := mime.TypeByExtension(filepath.Ext(name))
	entry := &ddb.Entry{Name: name, Mime: fm, Hash: imageHash, Data: image}
	txs, err := logbook.ProcessEntry(entry)
	t.Logf("txs len: %d len(data):%d  maxDataSize:%d", len(txs), len(image), logbook.MaxDataSize())
	if err != nil {
		t.Logf("txs preparation failed")
		t.Fail()
	}
	// here data should be cast to blockchain and then
	// retrieved trough a blockchain explorer
	ents, err := logbook.ExtractEntries(txs)
	if err != nil {
		t.Logf("entry extraction failed")
		t.Fail()
	}
	if len(ents) != 1 {
		t.Logf("len(entries) should be 1: %d", len(ents))
		t.FailNow()
	}
	if ents[0].Name != name {
		t.Logf("unexpected name: %s != %s", name, ents[0].Name)
		t.Fail()
	}
	if ents[0].Mime != fm {
		t.Logf("unexpected mime: %s != %s", fm, ents[0].Mime)
		t.Fail()
	}
	if ents[0].Hash != imageHash {
		t.Logf("unexpected hash: %s != %s", imageHash, ents[0].Hash)
		t.Fail()
	}
	shaOut := sha256.Sum256(ents[0].Data)
	hashOut := hex.EncodeToString(shaOut[:])
	if hashOut != imageHash {
		t.Logf("unexpected hash of extracted data: %s != %s", imageHash, hashOut)
		t.Fail()
	}
}
