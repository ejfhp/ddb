package main

import (
	"os"
	"strings"
	"testing"

	"github.com/ejfhp/trail"
)

func TestPassphrase(t *testing.T) {
	trail.SetWriter(os.Stdout)
	clis := [][]string{
		strings.Split("+ Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF", " "),
		strings.Split("+ This is the passphrase used in the TRH help, the 24th of July, 2021.", " "),
		strings.Split("+ ciao 2", " "),
	}
	keys := []string{
		"L1XZMDYzVwkPUMnNgsKD8ysUMFzvvPAA1SKbeo5cjMuWSSPATQ6v",
		"L1XkuhuE2vGfs2XSnVwr9nb4nzLcV9B2HehxUN2GiHsyAMNCpZRG",
		"KykxKKgmT6b8N2abJACuRHiHievFgzwWLLYjvU7AikBruhCur4Qu",
	}
	pwds := [][32]byte{
		{'B', 'i', 't', 'c', 'o', 'i', 'n', ':', ' ', 'A', ' ', 'P', 'e', 'e', 'r', '-', 't', 'o', '-', 'P', 'e', 'e', 'r', ' ', 'E', 'l', 'e', 'c', 't', 'r', 'o', 'n'},
		{'T', 'h', 'i', 's', ' ', 'i', 's', ' ', 't', 'h', 'e', ' ', 'p', 'a', 's', 's', 'p', 'h', 'r', 'a', 's', 'e', ' ', 'u', 's', 'e', 'd', ' ', 'i', 'n', ' ', 't'},
		{'c', 'i', 'a', 'o', ' ', '2', 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	for i, args := range clis {
		passp, err := extractPassphrase(args)
		if err != nil {
			t.Logf("%d extractPassphrase failed '%v': %v", i, args, err)
			t.Fail()
		}
		key, pwd, err := processPassphrase(passp, 2)
		if err != nil {
			t.Logf("%d keygen failed, passphrase '%s': %v", i, passp, err)
			t.Fail()
		}
		if key != keys[i] {
			t.Logf("%d key doesn't match '%s': %s", i, key, keys[i])
			t.Fail()
		}
		for j, c := range pwd {
			if c != pwds[i][j] {
				t.Logf("%d,%d passwords doesn't match '%v': %v", i, j, pwd, pwds[i])
				t.Fail()

			}
		}

	}
}
