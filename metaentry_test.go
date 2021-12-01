package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"testing"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
)

func TestMetaEntry_Encrypt_MetaEntryFromEncrypted(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	fil := "testdata/test.txt"
	bytes, err := ioutil.ReadFile(fil)
	if err != nil {
		t.Fatalf("error reading test file: %v", err)
	}
	eh := sha256.Sum256(bytes)
	ehash := hex.EncodeToString(eh[:])
	ent := &ddb.Entry{Name: fil, Data: bytes, DataHash: ehash}
	keystore, err := keys.NewKeystore(destinationKey, "mainpassword")
	if err != nil {
		t.Logf("failed to build keystore: %v", err)
		t.FailNow()
	}
	node, err := keystore.NewNode(fil, eh)
	if err != nil {
		t.Logf("failed to get new node: %v", err)
		t.FailNow()
	}
	mentry := ddb.NewMetaEntry(node, ent)
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
