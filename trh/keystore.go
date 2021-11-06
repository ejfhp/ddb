package trh

import (
	"fmt"
	"os"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
)

func KeystoreGenFromKey(bitcoinKey string, password string, pin string, pathname string) (*keys.Keystore, error) {
	keyStore, err := keys.NewKeystore(bitcoinKey, password)
	if err != nil {
		return nil, fmt.Errorf("provided key %s has issues: %w", bitcoinKey, err)
	}
	err = keyStore.Save(pathname, pin)
	if err != nil {
		return nil, fmt.Errorf("error while saving keystore to file %s: %w", pathname, err)
	}
	showKeystore(keyStore)
	return keyStore, nil
}

func KeystoreGenFromPhrase(phrase string, keygenID int, pin string, pathname string) error {
	wif, password, err := keys.FromPassphrase(phrase, keygenID)
	if err != nil {
		return fmt.Errorf("error while generating key using passphrase: %w", err)
	}
	keyStore, err := keys.NewKeystore(wif, password)
	if err != nil {
		return fmt.Errorf("error while generating keystore from passphrase '%s' with keygen '%d': %w", phrase, keygenID, err)
	}
	err = keyStore.Save(pathname, pin)
	if err != nil {
		return fmt.Errorf("error while saving keystore to local file %s: %w", pathname, err)
	}
	showKeystore(keyStore)
	return nil
}

func KeystoreShow(pin string, pathname string) error {
	keystore, err := keys.LoadKeystore(pathname, pin)
	if err != nil {
		return fmt.Errorf("error while loading keystore: %w", err)
	}
	showKeystore(keystore)
	return nil
}

func KeystoreSaveUnencrypted(pin string, pathname string) error {
	keystore, err := keys.LoadKeystore(pathname, pin)
	if err != nil {
		return fmt.Errorf("error while loading keystore: %w", err)
	}
	err = keystore.SaveUnencrypted(pathname)
	if err != nil {
		return fmt.Errorf("error saving unencrypted keystore to %s: %w", pathname, err)
	}
	return nil
}

func KeystoreRestoreFromUnencrypted(pin string, pathname string) error {
	keystore, err := keys.LoadKeystoreUnencrypted(pathname)
	if err != nil {
		return fmt.Errorf("error while loading unenctrypted keystore: %w", err)
	}
	err = keystore.Update()
	if err != nil {
		return fmt.Errorf("error while updating keystore: %w", err)
	}
	showKeystore(keystore)
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
