package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"mime"
	"path/filepath"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestFBranch_ProcessEntry(t *testing.T) {
	// trail.SetWriter(os.Stdout)
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
		fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: v, Blockchain: blockchain}
		entry := ddb.Entry{Name: filename, Data: []byte(file)}
		txs, err := fbranch.ProcessEntry("123456789", &entry, Helper_FakeTX(t).UTXOs())
		if err != nil {
			t.Logf("%d failed to process entry: %v", i, err)
			t.Fail()
		}
		if len(txs) < 2 {
			t.Logf("%d unexpected number of transactions: %d", i, len(txs))
			t.FailNow()
		}
		for _, tx := range txs {
			opr, header, err := tx.Data()
			if err != nil {
				t.Logf("failed to get OP_RETURN: %v", err)
				t.FailNow()
			}
			if len(opr) == 0 {
				t.Logf("unexpected len of OP_RETURN: %d", len(opr))
				t.FailNow()
			}
			if len(header) != 9 {
				t.Logf("unexpected len of header: %d", len(header))
				t.FailNow()
			}
		}
	}
}

func TestFBranch_EstimateEntryFee(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: password, Blockchain: blockchain}
	name := "image.png"
	filename := "testdata/image.png"
	entry, err := ddb.NewEntryFromFile(name, filename, []string{"label1", "label2"}, "notes")
	if err != nil {
		t.Logf("failed to create Entry: %v", err)
		t.FailNow()
	}
	for i := 0; i < 10; i++ {
		fee, err := fbranch.EstimateEntryFee("123456789", entry)
		if err != nil {
			t.Logf("%d - failed to estimate required fee to cast entry: %v", i, err)
			t.FailNow()
		}
		if fee != 2849 {
			t.Logf("%d - fee seems to be different from the past: %d", i, fee)
			t.FailNow()
		}
	}
}

// func TestFBranch_CastEntry_CheckingFee(t *testing.T) {
// 	trail.SetWriter(os.Stdout)
// 	woc := ddb.NewWOC()
// 	taal := ddb.NewTAAL()
// 	passwords := [][32]byte{
// 		{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'},
// 	}
// 	blockchain := ddb.NewBlockchain(taal, woc, nil)
// 	filename := "Inferno"
// 	file := `Nel mezzo del cammin di nostra vita
// 		mi ritrovai per una selva oscura,
// 		ché la diritta via era smarrita.

// 		Ahi quanto a dir qual era è cosa dura
// 		esta selva selvaggia e aspra e forte
// 		che nel pensier rinova la paura!`
// 	for i, v := range passwords {
// 		fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: v, Blockchain: blockchain}
// 		entry := ddb.Entry{Name: filename, Data: []byte(file)}
// 		result, err := fbranch.CastEntry("123456789", &entry, ddb.Satoshi(10000), true)
// 		if err != nil {
// 			t.Logf("%d failed to estimate fee: %v", i, err)
// 			t.Fail()
// 		}
// 		if result.Cost.Satoshi() < 300 {
// 			t.Logf("%d fee too cheap: %d", i, result.Cost.Satoshi())
// 			t.Fail()
// 		}
// 	}
// }

// func TestFBranch_CastEntry_SpendingLimit(t *testing.T) {
// 	trail.SetWriter(os.Stdout)
// 	woc := ddb.NewWOC()
// 	taal := ddb.NewTAAL()
// 	passwords := [][32]byte{
// 		{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'},
// 	}
// 	blockchain := ddb.NewBlockchain(taal, woc, nil)
// 	filename := "Inferno"
// 	file := `Nel mezzo del cammin di nostra vita
// 		mi ritrovai per una selva oscura,
// 		ché la diritta via era smarrita.

// 		Ahi quanto a dir qual era è cosa dura
// 		esta selva selvaggia e aspra e forte
// 		che nel pensier rinova la paura!`
// 	for i, v := range passwords {
// 		fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: v, Blockchain: blockchain}
// 		entry := ddb.Entry{Name: filename, Data: []byte(file)}
// 		_, err := fbranch.CastEntry("123456789", &entry, ddb.Satoshi(100), true)
// 		if err == nil {
// 			t.Logf("%d expected error for exceeding spending limit is nil", i)
// 			t.FailNow()
// 		}
// 	}
// }

// func TestFBranch_CastEntry_Text(t *testing.T) {
// 	t.SkipNow()
// 	trail.SetWriter(os.Stdout)
// 	woc := ddb.NewWOC()
// 	taal := ddb.NewTAAL()
// 	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
// 	blockchain := ddb.NewBlockchain(taal, woc, nil)
// 	fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: password, Blockchain: blockchain}
// 	filename := "Inferno.txt"
// 	file := `Nel mezzo del cammin di nostra vita
// 		mi ritrovai per una selva oscura,
// 		ché la diritta via era smarrita.

// 		Ahi quanto a dir qual era è cosa dura
// 		esta selva selvaggia e aspra e forte
// 		che nel pensier rinova la paura!`
// 	entry := ddb.NewEntryFromData(filename, mime.TypeByExtension(".txt"), []byte(file), []string{"label1", "label2"}, "notes")
// 	res, err := fbranch.CastEntry("123456789", entry, 300, true)
// 	if err != nil {
// 		t.Logf("failed to process entry: %v", err)
// 		t.Fail()
// 	}
// 	if len(res.TXIDs) != 1 {
// 		t.Logf("unexpected number of TX ID: %d", len(res.TXIDs))
// 		t.Fail()
// 	}
// 	for _, id := range res.TXIDs {
// 		t.Logf("TX ID: %s\n", id)
// 	}
// }

// func TestFBranch_CastEntry_Image(t *testing.T) {
// 	t.SkipNow()
// 	trail.SetWriter(os.Stdout)
// 	woc := ddb.NewWOC()
// 	taal := ddb.NewTAAL()
// 	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
// 	blockchain := ddb.NewBlockchain(taal, woc, nil)
// 	fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: password, Blockchain: blockchain}
// 	name := "image.png"
// 	filename := "testdata/image.png"
// 	entry, err := ddb.NewEntryFromFile(name, filename, []string{"label1", "label2"}, "notes")
// 	if err != nil {
// 		t.Logf("failed to create Entry: %v", err)
// 		t.Fail()
// 	}
// 	res, err := fbranch.CastEntry("123456789", entry, 300, true)
// 	if err != nil {
// 		t.Logf("failed to process entry: %v", err)
// 		t.Fail()
// 	}
// 	if len(res.TXIDs) != 1 {
// 		t.Logf("unexpected number of TX ID: %d", len(res.TXIDs))
// 		t.Fail()
// 	}
// 	for _, id := range res.TXIDs {
// 		t.Logf("TX ID: %s\n", id)
// 	}
// }

func TestFBranch_GetEntryFromTXID_Text(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	txid := "afbdf4a215f5e7dc3beca36e1625f3597995afa5906b2bbfee6a572d87764426"
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: password, Blockchain: blockchain}
	entries, err := fbranch.GetEntriesFromTXID([]string{txid}, false)
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
	expEntry := ddb.NewEntryFromData(filename, mime.TypeByExtension(".txt"), []byte(file), []string{"label1", "label2"}, "notes")
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

func TestFBranch_GetEntryFromTXID_Image(t *testing.T) {
	// trail.SetWriter(os.Stdout)
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
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, nil)
	fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: password, Blockchain: blockchain}
	entries, err := fbranch.GetEntriesFromTXID(txids, false)
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
	expEntry, err := ddb.NewEntryFromFile(name, filename, []string{"label1", "label2"}, "notes")
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

	// err = ioutil.WriteFile("imagefromblockchain.png", entry.Data, 0644)
	// if err != nil {
	// 	t.Logf("failed to create file: %v", err)
	// 	t.Fail()
	// }
}

func TestFBranch_ProcessAndGetEntry_Text(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	cache, err := ddb.NewTXCache("./.trhcache")
	if err != nil {
		t.Logf("cache preparation failed: %v", err)
		t.FailNow()
	}
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: password, Blockchain: blockchain}
	name := "test.txt"
	fm := mime.TypeByExtension(filepath.Ext(name))
	bytes := []byte("just a test")
	sha := sha256.Sum256(bytes)
	hash := hex.EncodeToString(sha[:])
	entry := &ddb.Entry{Name: name, Mime: fm, Hash: hash, Data: bytes}
	txs, err := fbranch.ProcessEntry("123456789", entry, Helper_FakeTX(t).UTXOs())
	t.Logf("txs len: %d", len(txs))
	if err != nil {
		t.Logf("txs preparation failed")
		t.FailNow()
	}
	txids := make([]string, len(txs))
	for i, t := range txs {
		txids[i] = t.GetTxID()
		fbranch.Blockchain.Cache.StoreTX(t.GetTxID(), t.ToBytes())
	}
	// here data should be cast to blockchain and then
	// retrieved trough a blockchain explorer
	ents, err := fbranch.GetEntriesFromTXID(txids, true)
	if err != nil {
		t.Logf("entry extraction failed: %v", err)
		t.FailNow()
	}
	if len(ents) != 1 {
		t.Logf("len(entries) should be 1: %d", len(ents))
		t.FailNow()
	}
	if ents[0].Name != name {
		t.Logf("unexpected name: %s != %s", name, ents[0].Name)
		t.FailNow()
	}
	if ents[0].Mime != fm {
		t.Logf("unexpected mime: %s != %s", fm, ents[0].Mime)
		t.FailNow()
	}
	if ents[0].Hash != hash {
		t.Logf("unexpected hash: %s != %s", hash, ents[0].Hash)
		t.FailNow()
	}
	shaOut := sha256.Sum256(ents[0].Data)
	hashOut := hex.EncodeToString(shaOut[:])
	if hashOut != hash {
		t.Logf("unexpected hash of extracted data: %s != %s", hash, hashOut)
		t.FailNow()
	}
}

func TestFBranch_ProcessAndGetEntry_Image(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	cache, err := ddb.NewTXCache("./.trhcache")
	if err != nil {
		t.Logf("cache preparation failed: %v", err)
		t.FailNow()
	}
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: password, Blockchain: blockchain}
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
	txs, err := fbranch.ProcessEntry("123456789", entry, Helper_FakeTX(t).UTXOs())
	txids := make([]string, len(txs))
	for i, t := range txs {
		txids[i] = t.GetTxID()
		fbranch.Blockchain.Cache.StoreTX(t.GetTxID(), t.ToBytes())
	}
	t.Logf("txs len: %d len(data):%d  maxDataSize:%d", len(txs), len(image), fbranch.Blockchain.MaxDataSize())
	if err != nil {
		t.Logf("txs preparation failed")
		t.Fail()
	}
	// here data should be cast to blockchain and then
	// retrieved trough a blockchain explorer
	ents, err := fbranch.GetEntriesFromTXID(txids, false)
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

func TestFBranch_DownloadAll(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	// cache, err := ddb.NewTXCache("/tmp")
	cache, err := ddb.NewTXCache("./.trhcache")
	if err != nil {
		t.Logf("cache preparation failed: %v", err)
		t.FailNow()
	}
	password := [32]byte{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	fbranch := &ddb.FBranch{BitcoinWIF: destinationKey, BitcoinAdd: destinationAddress, Password: password, Blockchain: blockchain}
	output := "download"
	n, err := fbranch.DowloadAll(output, false)
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
