package main

import (
	"os"
	"strings"
	"testing"

	"github.com/ejfhp/trail"
)

func TestPassphrase(t *testing.T) {
	trail.SetWriter(os.Stdout)
	passphrases := [][]string{
		strings.Split("+ Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF", " "),
		strings.Split("+ This is the passphrase used in the TRH help, the 24th of July, 2021.", " "),
	}
	for i, pp := range passphrases {
		err := cmdDescribe(pp)
		if err != nil {
			t.Logf("%d passphrase failed '%s': %v", i, pp, err)
			t.Fail()
		}

	}
}
