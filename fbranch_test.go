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
	"github.com/ejfhp/trail"
)

//TO BE SET IF REAL ONCHAIN TEST ARE GOING TO BE EXECUTED
var address string
var key string

func TestFBranch_ProcessEntry(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	passwords := [][32]byte{
		{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'},
	}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	filename := "Inferno"
	file := `Nel mezzo del cammin di nostra vita
		mi ritrovai per una selva oscura,
		ché la diritta via era smarrita.
		
		Ahi quanto a dir qual era è cosa dura
		esta selva selvaggia e aspra e forte
		che nel pensier rinova la paura!`
	for i, v := range passwords {
		fbranch, err := ddb.NewFBranch(key, v, blockchain)
		if err != nil {
			t.Logf("%d failed to create new FBranch: %v", i, err)
			t.Fail()
		}
		entry := ddb.Entry{Name: filename, Data: []byte(file)}
		txs, err := fbranch.ProcessEntry(&entry)
		if err != nil {
			t.Logf("%d failed to process entry: %v", i, err)
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
			decrypt, err := ddb.AESDecrypt(v, opr)
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
}

func TestFBranch_EstimateFee(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	passwords := [][32]byte{
		{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'},
	}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	filename := "Inferno"
	file := `Nel mezzo del cammin di nostra vita
		mi ritrovai per una selva oscura,
		ché la diritta via era smarrita.
		
		Ahi quanto a dir qual era è cosa dura
		esta selva selvaggia e aspra e forte
		che nel pensier rinova la paura!`
	for i, v := range passwords {
		fbranch, err := ddb.NewFBranch(key, v, blockchain)
		if err != nil {
			t.Logf("%d failed to create new FBranch: %v", i, err)
			t.Fail()
		}
		entry := ddb.Entry{Name: filename, Data: []byte(file)}
		fee, err := fbranch.EstimateFee(&entry)
		if err != nil {
			t.Logf("%d failed to estimate fee: %v", i, err)
			t.Fail()
		}
		if fee < 300 {
			t.Logf("%d fee too cheap: %d", i, fee.Satoshi())
			t.Fail()
		}
	}
}

func TestFBranch_EncryptDecryptEntry(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	phrases := []string{
		"tanto va la gatta al lardo che ci lascia lo zampino",
		"ciao",
	}
	nums := []int{5, 5}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	filename := "Inferno"
	file := `Nel mezzo del cammin di nostra vita
		mi ritrovai per una selva oscura,
		ché la diritta via era smarrita.
		
		Ahi quanto a dir qual era è cosa dura
		esta selva selvaggia e aspra e forte
		che nel pensier rinova la paura!`
	for i, phrase := range phrases {
		k, err := ddb.NewKeygen2(nums[i], phrase)
		if err != nil {
			t.Logf("cannot generate Keygen2: %v", err)
			t.Fail()
		}
		wif, err := k.WIF()
		if err != nil {
			t.Logf("cannot generate WIF: %v", err)
			t.Fail()
		}
		pass := k.Password()
		fbranch, err := ddb.NewFBranch(wif, pass, blockchain)
		if err != nil {
			t.Logf("%d failed to create new FBranch: %v", i, err)
			t.Fail()
		}
		entry := ddb.Entry{Name: filename, Data: []byte(file)}
		encs, err := fbranch.EncryptEntry(&entry)
		if err != nil {
			t.Logf("%d failed to process entry: %v", i, err)
			t.Fail()
		}
		entries, err := fbranch.DecryptEntries(encs)
		if err != nil {
			t.Logf("%d failed to decrypt entry: %v", i, err)
			t.Fail()
		}
		if len(entries) != 1 {
			t.Logf("%d unexpected nuber of entries %d", i, len(entries))
			t.Fail()

		}
	}
}

func NO_TestFBranch_CastEntry(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch, err := ddb.NewFBranch(key, password, blockchain)
	if err != nil {
		t.Logf("failed to create new FBranch: %v", err)
		t.Fail()
	}
	filename := "Inferno.txt"
	file := `Nel mezzo del cammin di nostra vita
		mi ritrovai per una selva oscura,
		ché la diritta via era smarrita.
		
		Ahi quanto a dir qual era è cosa dura
		esta selva selvaggia e aspra e forte
		che nel pensier rinova la paura!`
	entry := ddb.NewEntryFromData(filename, mime.TypeByExtension(".txt"), []byte(file))
	ids, err := fbranch.CastEntry(entry)
	if err != nil {
		t.Logf("failed to process entry: %v", err)
		t.Fail()
	}
	if len(ids) != 1 {
		t.Logf("unexpected number of TX ID: %d", len(ids))
		t.Fail()
	}
	for _, id := range ids {
		t.Logf("TX ID: %s\n", id)
	}
}

func NO_TestFBranch_CastImageEntry(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch, err := ddb.NewFBranch(key, password, blockchain)
	if err != nil {
		t.Logf("failed to create new FBranch: %v", err)
		t.Fail()
	}
	name := "image.png"
	filename := "testdata/image.png"
	entry, err := ddb.NewEntryFromFile(name, filename)
	if err != nil {
		t.Logf("failed to create Entry: %v", err)
		t.Fail()
	}
	ids, err := fbranch.CastEntry(entry)
	if err != nil {
		t.Logf("failed to process entry: %v", err)
		t.Fail()
	}
	if len(ids) != 1 {
		t.Logf("unexpected number of TX ID: %d", len(ids))
		t.Fail()
	}
	for _, id := range ids {
		t.Logf("TX ID: %s\n", id)
	}
}

func TestFBranch_RetrieveAndExtractEntries(t *testing.T) {
	txid := "afbdf4a215f5e7dc3beca36e1625f3597995afa5906b2bbfee6a572d87764426"
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch, err := ddb.NewFBranch(key, password, blockchain)
	if err != nil {
		t.Logf("failed to create new FBranch: %v", err)
		t.Fail()
	}
	entries, err := fbranch.RetrieveAndExtractEntries([]string{txid})
	if err != nil {
		t.Logf("failed to retrieve entry: %v", err)
		t.Fail()
	}
	if len(entries) != 1 {
		t.Logf("unexpected number of entries: %d", len(entries))
		t.Fail()
	}
	entry := entries[0]
	filename := "Inferno.txt"
	file := `Nel mezzo del cammin di nostra vita
		mi ritrovai per una selva oscura,
		ché la diritta via era smarrita.
		
		Ahi quanto a dir qual era è cosa dura
		esta selva selvaggia e aspra e forte
		che nel pensier rinova la paura!`
	expEntry := ddb.NewEntryFromData(filename, mime.TypeByExtension(".txt"), []byte(file))
	if entry.Name != expEntry.Name {
		t.Logf("unexpected name for retrieved entry: %s", entry.Name)
		t.Fail()
	}
	if entry.Hash != expEntry.Hash {
		t.Logf("unexpected hash for retrieved entry: %s", entry.Hash)
		t.Fail()
	}
	if entry.Mime != expEntry.Mime {
		t.Logf("unexpected mime for retrieved entry: %s", entry.Mime)
		t.Fail()
	}
	if string(entry.Data) != string(expEntry.Data) {
		t.Logf("unexpected data for retrieved entry: %s", string(entry.Data))
		t.Fail()
	}
}

func TestFBranch_RetrieveAndExtractImageEntry(t *testing.T) {
	txids := []string{
		"afbdf4a215f5e7dc3beca36e1625f3597995afa5906b2bbfee6a572d87764426", //EXTRA TX
		"33c5339f5f942793867898d92c72cdab8fc5ff464f77970fc6fd0cf8dd99f271",
		"c33492a97f30156ba725acd7f38ef201459adb19fe9be8eefc7578c81535c032",
		"74e905acee78a53bc858afb3ae44ce3bb016424df088712ce94071adc0f1b7fb",
		"7d09ecbae40b4d07e3dc62d9ab4639428571190fea263fb7d3614adae89d6d21",
		"bdf47620b761c6e3b46422d05b88d5c29ea10b39488208cb921f8e60242032a6",
		"fcdd1ecb2703e9025ac60caddabadf44cbcae341465a8c18ee4f9b3ec1a4580f",
		"ee4fd4a05f45c09c5717321a3e21c871494a80d50c886c1bfda51d16d0c84cf1", //EXTRA TX
		"4d4f9f1a737e7eae37cadcd4289b436b2bcf39bdf5f374152420196ab14b0b65",
		"1668afdd6978ef2cd594aa15c96138736e86d22abc3aba2b8428b96400dd2f87",
	}
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch, err := ddb.NewFBranch(key, password, blockchain)
	if err != nil {
		t.Logf("failed to create new FBranch: %v", err)
		t.Fail()
	}
	entries, err := fbranch.RetrieveAndExtractEntries(txids)
	if err != nil {
		t.Logf("failed to retrieve entry: %v", err)
		t.Fail()
	}
	if len(entries) != 2 {
		t.Logf("unexpected number of entries: %d", len(entries))
		t.Fail()
	}
	//entries[0] is "Inferno.txt"
	entry := entries[1]
	name := "image.png"
	filename := "testdata/image.png"
	expEntry, err := ddb.NewEntryFromFile(name, filename)
	if err != nil {
		t.Logf("failed to build entry: %v", err)
		t.Fail()
	}
	if entry.Name != expEntry.Name {
		t.Logf("unexpected name for retrieved entry: %s", entry.Name)
		t.Fail()
	}
	if entry.Hash != expEntry.Hash {
		t.Logf("unexpected hash for retrieved entry: %s", entry.Hash)
		t.Fail()
	}
	if entry.Mime != expEntry.Mime {
		t.Logf("unexpected mime for retrieved entry: %s", entry.Mime)
		t.Fail()
	}
	sha := sha256.Sum256(entry.Data)
	hash := hex.EncodeToString(sha[:])
	if entry.Hash != hash {
		t.Logf("retrieved entry hash doesn't match with data hash: %s", hash)
		t.Fail()
	}

	err = ioutil.WriteFile("imagefromblockchain.png", entry.Data, 0644)
	if err != nil {
		t.Logf("failed to create file: %v", err)
		t.Fail()
	}
}

func TestFBranch_EntryFullCycleText(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch, err := ddb.NewFBranch(key, password, blockchain)
	if err != nil {
		t.Logf("failed to create fbranch: %v", err)
		t.Fail()
	}
	name := "test.txt"
	fm := mime.TypeByExtension(filepath.Ext(name))
	bytes := []byte("just a test")
	sha := sha256.Sum256(bytes)
	hash := hex.EncodeToString(sha[:])
	entry := &ddb.Entry{Name: name, Mime: fm, Hash: hash, Data: bytes}
	txs, err := fbranch.ProcessEntry(entry)
	t.Logf("txs len: %d", len(txs))
	if err != nil {
		t.Logf("txs preparation failed")
		t.Fail()
	}
	// here data should be cast to blockchain and then
	// retrieved trough a blockchain explorer
	ents, err := fbranch.ExtractEntries(txs)
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

func TestFBranch_EntryFullCycleImage(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch, err := ddb.NewFBranch(key, password, blockchain)
	if err != nil {
		t.Fatalf("error building FBranch: %v", err)
	}
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
	txs, err := fbranch.ProcessEntry(entry)
	t.Logf("txs len: %d len(data):%d  maxDataSize:%d", len(txs), len(image), fbranch.Blockchain.MaxDataSize())
	if err != nil {
		t.Logf("txs preparation failed")
		t.Fail()
	}
	// here data should be cast to blockchain and then
	// retrieved trough a blockchain explorer
	ents, err := fbranch.ExtractEntries(txs)
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

func TestFBranch_RetrieveTXs(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch, err := ddb.NewFBranch(key, password, blockchain)
	if err != nil {
		t.Logf("error building FBranch: %v", err)
		t.FailNow()
	}
	txids := []string{
		"8686df3af289968bf286023190a0e2aa0cd9fd12bce9e4e7f9763cc16219a114",
		"4286b420ce6d33da881342697a2ebf19a475817f0bb41547768fe61070e5a42b",
	}
	txs, err := fbranch.RetrieveTXs(txids)
	if err != nil {
		t.Logf("error retrieving TXs: %v", err)
		t.FailNow()
	}
	if len(txs) != len(txids) {
		t.Logf("unexpected number of txs: %d", len(txs))
		t.Fail()
	}
}

func TestFBranch_DownloadAll(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch, err := ddb.NewFBranch(key, password, blockchain)
	if err != nil {
		t.Logf("error building FBranch: %v", err)
		t.FailNow()
	}
	output := "download"
	n, err := fbranch.DowloadAll(output)
	if err != nil {
		t.Logf("failed to download all: %v", err)
		t.FailNow()
	}
	if n < 2 {
		t.Logf("unexpected value for n: %d", n)
		t.FailNow()
	}
	t.Logf("downloaded entries: %d", n)
}

func TestFBranch_ListHistory(t *testing.T) {
	trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch, err := ddb.NewFBranch(key, password, blockchain)
	if err != nil {
		t.Logf("error building FBranch: %v", err)
		t.FailNow()
	}
	txids, err := fbranch.ListHistory(address)
	if err != nil {
		t.Logf("error: %v", err)
		t.Fail()
	}
	if len(txids) < 23 {
		t.Logf("tx history too short: %d", len(txids))
		t.Fail()
	}
}
