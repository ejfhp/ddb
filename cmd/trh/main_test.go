package main

import (
	"fmt"
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

func TestEnvironment(t *testing.T) {
	clis := []map[string]string{
		{
			"cli":        "trh store -log -file logo.png -password logo + this passphrase is only for test1ng purpose, nothing is stored",
			"command":    "store",
			"password":   "logo",
			"passphrase": "this passphrase is only for test1ng purpose, nothing is stored",
		},
	}
	for i, conf := range clis {
		flagset := newFlagset(conf["command"])
		args := strings.Split(conf["cli"], " ")
		for i, a := range args {
			fmt.Printf("%d: %s\n", i, a)
		}
		env, err := prepareEnvironment(args, flagset)
		if err != nil {
			t.Logf("%d, prepare environment failed: %v", i, err)
			t.FailNow()
		}
		for k, v := range conf {
			if k == "cli" || k == "command" {
				continue
			}
			if k == "password" && env.passwordString() != v {
				t.Logf("%d, wrong %s: %s (%d) != %s", 1, k, env.passwordString(), len(env.passwordString()), v)
				t.Fail()
			}
			if k == "passphrase" && env.passphrase != v {
				t.Logf("%d, wrong %s: %s (%d) != %s", 1, k, env.passphrase, len(env.passphrase), v)
				t.Fail()
			}
		}

	}

}
