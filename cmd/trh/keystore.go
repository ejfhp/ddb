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

	switch opt {
	case "key":
		keyStore := ddb.NewKeystore()
		keyStore.WIF = flagBitcoinKey
		keyStore.Address, err = ddb.AddressOf(flagBitcoinKey)
		if err != nil {
			trail.Println(trace.Alert("provided key has issues").Append(tr).UTC().Add("bitcoinKey", flagBitcoinKey).Error(err))
			return fmt.Errorf("provided key %s has issues: %w", flagBitcoinKey, err)
		}
		keyStore.Passwords["main"] = passwordtoBytes(flagPassword)
		showKeystore(keyStore)
	case "phrase":
		keyStore, err := keyStoreFromPassphrase(args)
		if err != nil {
			trail.Println(trace.Alert("error while generating keystore from passphrase").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while generating keystore from passphrase: %w", err)
		}
		showKeystore(keyStore)
	case "actionphrase":
		keystore, err := keyStoreFromPassphrase(args)
		if err != nil {
			trail.Println(trace.Alert("error while generating keystore from passphrase").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while generating keystore from passphrase: %w", err)
		}
		switch flagAction {
		case "generate":
			keyStore.Save("keystore.trh", flagPIN)
		}
	}
	return nil
}

func showKeystore(keystore *ddb.KeyStore) {
	fmt.Printf("KEYSTORE")
	fmt.Printf("Key: %s\n", keystore.WIF)
	fmt.Printf("Address: %s\n", keystore.Address)
	fmt.Printf("Passwords:\n")
	for n, p := range keystore.Passwords {
		fmt.Printf("%s: %s\n", n, string(p[:]))
	}
}

func keyStoreFromPassphrase(args []string) (*ddb.KeyStore, error) {
	keyStore := ddb.NewKeystore()
	passphrase, err := extractPassphrase(args)
	if err != nil {
		return nil, fmt.Errorf("error while reading passphrase: %w", err)
	}
	wif, password, err := processPassphrase(passphrase, int(flagKeygenID))
	if err != nil {
		return nil, fmt.Errorf("error while generating key using passphrase: %w", err)
	}
	keyStore.WIF = wif
	keyStore.Address, err = ddb.AddressOf(flagBitcoinKey)
	if err != nil {
		return nil, fmt.Errorf("generated key %s has issues: %w", flagBitcoinKey, err)
	}
	keyStore.Passwords["main"] = password
	return keyStore, nil
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
