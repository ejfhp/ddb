package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const keystoreFile = "keystore.trh"
const keystoreOldFile = "keystore.trh.old"
const keystoreUnencrypted = "keystore_plain.trh"

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
	var keyStore *keys.Keystore
	toStore := false
	switch opt {
	case "genkey":
		keyStore, err = keys.NewKeystore(flagBitcoinKey, flagPassword)
		if err != nil {
			trail.Println(trace.Alert("provided key has issues").Append(tr).UTC().Add("bitcoinKey", flagBitcoinKey).Error(err))
			return fmt.Errorf("provided key %s has issues: %w", flagBitcoinKey, err)
		}
		toStore = true
	case "genphrase":
		keyStore, err = keyStoreFromPassphrase(flagPhrase, int(flagKeygenID))
		if err != nil {
			trail.Println(trace.Alert("error while generating keystore from passphrase").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while generating keystore from passphrase: %w", err)
		}
		toStore = true
	case "show":
		keyStore, err = keys.LoadKeystore(keystoreFile)
		if err != nil {
			trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading keystore: %w", err)
		}
		keyStore.Update()
	case "tounencrypted":
		keyStore, err = loadKeyStore()
		if err != nil {
			trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
			return fmt.Errorf("error while loading keystore: %w", err)
		}
		saveUnencryptedKeyStore(keyStore)
	case "fromunencrypted":
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

func showKeystore(keystore *keys.Keystore) {
	fmt.Printf("*** KEYSTORE ***\n")
	fmt.Printf("\n")
	fmt.Printf("    KEY WIF\n")
	ddb.PrintQRCode(os.Stdout, keystore.Key(keys.Main))
	fmt.Printf("    %s\n", keystore.Key(keys.Main))
	fmt.Printf("\n")
	fmt.Printf("    ADDRESS\n")
	ddb.PrintQRCode(os.Stdout, keystore.Address(keys.Main))
	fmt.Printf("   %s\n", keystore.Address(keys.Main))
	fmt.Printf("\n")
	fmt.Printf("   Bitcoin Key (WIF): %s\n", keystore.Key(keys.Main))
	fmt.Printf("   Bitcoin Address: %s\n", keystore.Address(keys.Main))
	fmt.Printf("   Passwords:\n")
	for n, p := range keystore.Passwords() {
		fmt.Printf("      %s: '%s' .  %d\n", n, p, len(p))
	}
}

func keyStoreFromPassphrase(passphrase string, keygenID int) (*keys.Keystore, error) {
	wif, password, err := keys.FromPassphrase(passphrase, keygenID)
	if err != nil {
		return nil, fmt.Errorf("error while generating key using passphrase: %w", err)
	}
	keystore, err := keys.NewKeystore(wif, password)
	if err != nil {
		return nil, fmt.Errorf("error while generating keystore from passphrase '%s' with keygen '%d': %w", passphrase, keygenID, err)
	}
	return keystore, nil
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
