package trh_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ejfhp/ddb/trh"
)

func TestKeystore_KeystoreGenFromKey(t *testing.T) {
	address := "1H2KZJA9TjspsL7uPBUPdPzueeLbtvXs8R"
	key := "L5T6uSMcr9nkdSiPWpUDRfCKS8X6hSi16k4aqeJPMadVJJkYGf8h"
	password := "testpassword"
	pin := "0000"
	pathname := filepath.Join(os.TempDir(), "keystore.trh")
	th := &trh.TRH{}
	keystore, err := th.KeystoreGenFromKey(pin, key, password, pathname)
	if err != nil {
		t.Logf("keystore form key failed: %v", err)
		t.FailNow()
	}
	if keystore == nil {
		t.Logf("keystore form key failed: keystore is nil")
		t.FailNow()
	}
	if keystore.Source().Key() != key {
		t.Logf("unexpected key: %s", keystore.Source().Key())
		t.FailNow()
	}
	if keystore.Source().Address() != address {
		t.Logf("unexpected address: %s", keystore.Source().Address())
		t.FailNow()
	}
	if keystore.Source().Password() != password {
		t.Logf("unexpected password: %s", keystore.Source().Password())
		t.FailNow()
	}
}
