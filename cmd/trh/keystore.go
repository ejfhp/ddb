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

const keystoreFile = "keystore.trh"
const keystoreOldFile = "keystore.trh.old"
const keystoreUnencrypted = "keystore_plain.trh"

func loadKeyStore() (*ddb.KeyStore, error) {
	return ddb.LoadKeyStore(keystoreFile, flagPIN)
}

func saveKeyStore(k *ddb.KeyStore) error {
	return k.Save(keystoreFile, flagPIN)
}

func saveUnencryptedKeyStore(k *ddb.KeyStore) error {
	return k.SaveUnencrypted(keystoreUnencrypted)
}

func loadKeyStoreUnencrypted() (*ddb.KeyStore, error) {
	return ddb.LoadKeyStoreUnencrypted(keystoreUnencrypted)

}

func updateKeyStore(k *ddb.KeyStore) error {
	return k.Update(keystoreFile, keystoreOldFile, flagPIN)
}

func cmdKeystore(args []string) error {
	tr := trace.New().Source("keystore.go", "", "cmdKeystore")
	flagset, options := newFlagset(keystoreCmd)
	err := flagset.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	if flagHelp {
		printHelp(keystoreCmd)
		return nil
	}
	opt, ok := areFlagConsistent(flagset, options)
	if !ok {
		return fmt.Errorf("flag combination invalid")
	}
	var keyStore *ddb.KeyStore
	toStore := false
	switch opt {
	case "genkey":
		keyStore = ddb.NewKeystore()
		keyStore.Key = flagBitcoinKey
		keyStore.Address, err = ddb.AddressOf(flagBitcoinKey)
		if err != nil {
			trail.Println(trace.Alert("provided key has issues").Append(tr).UTC().Add("bitcoinKey", flagBitcoinKey).Error(err))
			return fmt.Errorf("provided key %s has issues: %w", flagBitcoinKey, err)
		}
		keyStore.Passwords["main"] = passwordtoBytes(flagPassword)
		toStore = true
	case "genphrase":
		keyStore, err = keyStoreFromPassphrase(flagPhrase)
		if err != nil {
			trail.Println(trace.Alert("error while generating keystore from passphrase").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while generating keystore from passphrase: %w", err)
		}
		toStore = true
	case "show":
		keyStore, err = loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading keystore: %w", err)
		}
		updateKeyStore(keyStore)
	case "tounencrypted":
		keyStore, err = loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading keystore: %w", err)
		}
		saveUnencryptedKeyStore(keyStore)
	case "fromunencrypted":
		//TODO COMPLETE TO AND FROM UNENCRYPTED, PASSWORD ARE SAVED WITH 0000
		keyStore, err = loadKeyStoreUnencrypted()
		if err != nil {
			trail.Println(trace.Alert("error while loading unencrypted keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading unenctrypted keystore: %w", err)
		}
		updateKeyStore(keyStore)
	default:
		return fmt.Errorf("flag combination invalid")
	}
	if toStore {
		saveKeyStore(keyStore)
		if err != nil {
			trail.Println(trace.Alert("error while saving keystore to local file").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while saving keystore to local file %s: %w", keystoreFile, err)
		}
	}
	showKeystore(keyStore)
	return nil
}

func showKeystore(keystore *ddb.KeyStore) {
	fmt.Printf("*** KEYSTORE ***\n")
	fmt.Printf("\n")
	fmt.Printf("    KEY WIF\n")
	ddb.PrintQRCode(os.Stdout, keystore.Address)
	fmt.Printf("    %s\n", keystore.Key)
	fmt.Printf("\n")
	fmt.Printf("    ADDRESS\n")
	ddb.PrintQRCode(os.Stdout, keystore.Address)
	fmt.Printf("   %s\n", keystore.Address)
	fmt.Printf("\n")
	fmt.Printf("   Bitcoin Key (WIF): %s\n", keystore.Key)
	fmt.Printf("   Bitcoin Address: %s\n", keystore.Address)
	fmt.Printf("   Passwords:\n")
	for n, p := range keystore.Passwords {
		fmt.Printf("      %s: '%s' .  %d\n", n, string(p[:]), len(n))
	}
}

func keyStoreFromPassphrase(passphrase string) (*ddb.KeyStore, error) {
	keyStore := ddb.NewKeystore()
	wif, password, err := processPassphrase(passphrase, int(flagKeygenID))
	if err != nil {
		return nil, fmt.Errorf("error while generating key using passphrase: %w", err)
	}
	keyStore.Key = wif
	keyStore.Address, err = ddb.AddressOf(wif)
	if err != nil {
		return nil, fmt.Errorf("generated key '%s' has issues: %w", flagBitcoinKey, err)
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
	if keygenID == 1 {
		keygen, err = ddb.NewKeygen1(passnum, passphrase)
		if err != nil {
			return "", [32]byte{}, fmt.Errorf("error building Keygen2: %w", err)
		}
	} else {
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
