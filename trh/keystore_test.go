package trh_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/trh"
)

func TestKeystore_KeystoreGenFromKey(t *testing.T) {
	address := "1H2KZJA9TjspsL7uPBUPdPzueeLbtvXs8R"
	key := "L5T6uSMcr9nkdSiPWpUDRfCKS8X6hSi16k4aqeJPMadVJJkYGf8h"
	password := "testpassword"
	pin := "0000"
	pathname := filepath.Join(os.TempDir(), "keystore.trh")
	keystore, err := trh.KeystoreGenFromKey(key, password, pin, pathname)
	if err != nil {
		t.Logf("keystore form key failed: %v", err)
		t.FailNow()
	}
	if keystore == nil {
		t.Logf("keystore form key failed: keystore is nil")
		t.FailNow()
	}
	if keystore.Key(keys.Main) != key {
		t.Logf("unexpected key: %s", keystore.Key(keys.Main))
		t.FailNow()
	}
	if keystore.Address(keys.Main) != address {
		t.Logf("unexpected address: %s", keystore.Address(keys.Main))
		t.FailNow()
	}
}
