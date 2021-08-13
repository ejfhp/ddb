package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ejfhp/ddb"
)

type Environment struct {
	Diary *ddb.Diary
}

func extractPassphrase(args []string) (string, error) {
	startidx := -1
	for i, t := range args {
		if t == "+" {
			startidx = i + 1
		}
	}
	if startidx < 0 || startidx >= len(args) {
		return "", fmt.Errorf("passphrase is missing")
	}
	passphrase := strings.Join(args[startidx:], " ")
	return passphrase, nil
}

func processPassphrase(passphrase string, keygenID int) (string, [32]byte, error) {
	var passnum int
	reg, err := regexp.Compile("[^0-9 ]+")
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("error compiling regexp: %w", err)
	}
	phrasenum := reg.ReplaceAllString(passphrase, "")
	for _, n := range strings.Split(phrasenum, " ") {
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
		return "", [32]byte{}, fmt.Errorf("passphrase must contain a number")
	}
	var keygen ddb.Keygen
	if keygenID == 2 {
		keygen, err = ddb.NewKeygen2(passnum, passphrase)
		if err != nil {
			return "", [32]byte{}, fmt.Errorf("error building Keygen2: %w", err)
		}
	}
	wif, err := keygen.WIF()
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("error while generating bitcoin key: %w", err)
	}
	password := keygen.Password()
	return wif, password, nil
}

func prepareDiary(passphrase string, keygenID int, privateKey string, address string, password string, enableCache bool, cachePath string) (*ddb.Diary, error) {
	var passEncrypt [32]byte
	var bitcoinKey string
	var err error
	if passphrase != "" {
		bitcoinKey, passEncrypt, err = processPassphrase(passphrase, keygenID)
		if err != nil {
			return nil, fmt.Errorf("error while processing passphrase: %w", err)
		}
	}
	if privateKey != "" {
		bitcoinKey = privateKey
	}
	if password != "" {
		copy(passEncrypt[:], []byte(password))
	}
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	var cache *ddb.TXCache
	if enableCache {
		if cachePath != "" {
			cache, err = ddb.NewTXCache(cachePath)
		} else {
			usercache, _ := os.UserCacheDir()
			cache, err = ddb.NewTXCache(filepath.Join(usercache, "trh"))
		}
		if err != nil {
			return nil, fmt.Errorf("error while creating Cache: %w", err)
		}
	}
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	diary, err := ddb.NewDiary(bitcoinKey, passEncrypt, blockchain)
	if err != nil {
		return nil, fmt.Errorf("error while creating a new Diary: %w", err)
	}
	return diary, nil
}
