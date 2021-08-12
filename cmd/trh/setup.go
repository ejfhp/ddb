package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ejfhp/ddb"
)

type Environment struct {
	Diary *ddb.Diary
}

func checkPassphrase(args []string) (string, int, error) {
	startidx := -1
	for i, t := range args {
		if t == "+" {
			startidx = i + 1
		}
	}
	if startidx < 0 || startidx >= len(args) {
		return "", 0, fmt.Errorf("passphrase is missing")
	}
	passphrase := strings.Join(args[startidx:], " ")
	passnum := 0
	reg, err := regexp.Compile("[^0-9 ]+")
	if err != nil {
		return "", 0, fmt.Errorf("error compiling regexp: %w", err)
	}
	phnum := reg.ReplaceAllString(passphrase, "")
	for _, n := range strings.Split(phnum, " ") {
		num, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			continue
		}
		if num < 0 {
			num = num * -1
		}
		passnum = int(num)
	}
	if passnum == 0 {
		return "", 0, fmt.Errorf("passphrase must contain a number")
	}
	return passphrase, passnum, nil
}

func keyGen(passphrase string, passnum int) (string, [32]byte, error) {
	keygen, err := ddb.NewKeygen2(passnum, passphrase)
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("error while building Keygen: %w", err)
	}
	wif, err := keygen.WIF()
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("error while generating bitcoin key: %w", err)
	}
	password := keygen.Password()
	return wif, password, nil
}

func newDiary(passphrase string, passnum int, cache *ddb.TXCache) (*ddb.Diary, error) {
	wif, password, err := keyGen(passphrase, passnum)
	if err != nil {
		return nil, fmt.Errorf("error while generating the Bitcoin private key: %w", err)
	}
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	diary, err := ddb.NewDiary(wif, password, blockchain)
	if err != nil {
		return nil, fmt.Errorf("error while creating a new Diary: %w", err)
	}
	return diary, nil
}
