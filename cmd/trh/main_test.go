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
	nums := []int{2008, 2021}
	keys := []string{
		"",
		"",
	}
	pwd := [][32]byte{
		{'B', 'i', 't', 'c', 'o', 'i', 'n', ':', ' ', 'A', ' ', 'P', 'e', 'e', 'r', '-', 't', 'o', '-', 'P', 'e', 'e', 'r', ' ', 'E', 'l', 'e', 'c', 't', 'r', 'o', 'n'},
	}
	for i, pp := range passphrases {
		pass, num, err := checkPassphrase(pp)
		if err != nil {
			t.Logf("%d passphrase failed '%s': %v", i, pp, err)
			t.Fail()
		}
		if num != nums[i] {
			t.Logf("%d num dows't match '%d': %d", i, num, nums[i])
			t.Fail()
		}
		key, pwd, err := keyGen(pass, num)
		if key != keys[i][0] {
			t.Logf("%d key dows't match '%s': %s", i, key, keys[i])
			t.Fail()
		}

	}
}
