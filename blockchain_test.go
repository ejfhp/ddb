package ddb_test

import (
	"fmt"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestAddHeader(t *testing.T) {
	data := "rosso di sera bel tempo si spera"
	payload := ddb.AddHeader(ddb.APP_NAME, ddb.VER_AES, []byte(data))
	if len(payload) != 9+len(data) {
		t.Logf("wrong header len: %d", len(payload))
	}
	fmt.Printf("%s\n", payload)
	if string(payload) != "ddb;0001;"+data {
		t.Logf("wrong header: '%s'", payload)
	}
}
