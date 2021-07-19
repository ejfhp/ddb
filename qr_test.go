package ddb_test

import (
	"os"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestPrintQRCode(t *testing.T) {
	ddb.PrintQRCode(os.Stdout, key)
	ddb.PrintQRCode(os.Stdout, address)

}
