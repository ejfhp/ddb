package main

import (
	"os"
	"testing"

	"github.com/ejfhp/trail"
)

func TestPassphrase(t *testing.T) {
	trail.SetWriter(os.Stdout)
	passphrases := []string{
		"Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF",
		"This is the passphrase used in the TRH help, the 24th of July, 2021.",
	}
}
