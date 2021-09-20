package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdKeystore(args []string) error {
	tr := trace.New().Source("setup.go", "", "cmdKeystore")
	flagset, options := newFlagset(commandKeystore)
	err := flagset.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	if flagHelp {
		printHelp(flagset)
		return nil
	}
	opt := areFlagConsistent(options)

	keyStore := ddb.NewKeystore()
	switch opt {
	case "key":
		keyStore.WIF = flagBitcoinKey
		keyStore.Address, err = ddb.AddressOf(flagBitcoinKey)
		if err != nil {
			trail.Println(trace.Alert("provided key has issues").Append(tr).UTC().Add("bitcoinKey", flagBitcoinKey).Error(err))
			return fmt.Errorf("provided key %s has issues: %w", flagBitcoinKey, err)
		}
		keyStore.Passwords["main"] = passwordtoBytes(flagPassword)
	case "phrase":
		passphrase, err = extractPassphrase(args)
		if err != nil {
			trail.Println(trace.Alert("error while reading passphrase").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while reading passphrase: %w", err)
		}

	}
	return nil
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
