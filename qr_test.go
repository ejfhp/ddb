package ddb_test

import (
	"os"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestQRCode_Print(t *testing.T) {
	ddb.PrintQRCode(os.Stdout, destinationKey)
	ddb.PrintQRCode(os.Stdout, destinationAddress)

}
