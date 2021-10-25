package ddb_test

import (
	"os"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestQRCode_Print(t *testing.T) {
	destinationAddress := "1PGh5YtRoohzcZF7WX8SJeZqm6wyaCte7X"
	destinationKey := "L4ZaBkP1UTyxdEM7wysuPd1scHMLLf8sf8B2tcEcssUZ7ujrYWcQ"
	ddb.PrintQRCode(os.Stdout, destinationKey)
	ddb.PrintQRCode(os.Stdout, destinationAddress)

}
