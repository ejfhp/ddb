package trh

import (
	"fmt"
	"os"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
)

func (t *TRH) KeystoreGenFromKey(pin string, bitcoinKey string, password string, pathname string) (*keys.Keystore, error) {
	keystore, err := keys.NewKeystore(bitcoinKey, password)
	if err != nil {
		return nil, fmt.Errorf("provided key %s has issues: %w", bitcoinKey, err)
	}
	err = keystore.Save(pathname, pin)
	if err != nil {
		return nil, fmt.Errorf("error while saving keystore to file %s: %w", pathname, err)
	}
	showKeystore(keystore)
	return keystore, nil
}

func (t *TRH) KeystoreGenFromPhrase(pin string, phrase string, keygenID int, pathname string) (*keys.Keystore, error) {
	wif, password, err := keys.FromPassphrase(phrase, keygenID)
	if err != nil {
		return nil, fmt.Errorf("error while generating key using passphrase: %w", err)
	}
	keystore, err := keys.NewKeystore(wif, password)
	if err != nil {
		return nil, fmt.Errorf("error while generating keystore from passphrase '%s' with keygen '%d': %w", phrase, keygenID, err)
	}
	err = keystore.Save(pathname, pin)
	if err != nil {
		return nil, fmt.Errorf("error while saving keystore to local file %s: %w", pathname, err)
	}
	showKeystore(keystore)
	return keystore, nil
}

func (t *TRH) KeystoreShow(pin string, pathname string) error {
	keystore, err := keys.LoadKeystore(pathname, pin)
	if err != nil {
		return fmt.Errorf("error while loading keystore: %w", err)
	}
	showKeystore(keystore)
	return nil
}

func (t *TRH) KeystoreSaveUnencrypted(pin string, pathname string, pathplain string) error {
	keystore, err := keys.LoadKeystore(pathname, pin)
	if err != nil {
		return fmt.Errorf("error while loading keystore: %w", err)
	}
	err = keystore.SaveUnencrypted(pathplain)
	if err != nil {
		return fmt.Errorf("error saving unencrypted keystore to %s: %w", pathplain, err)
	}
	return nil
}

func (t *TRH) KeystoreRestoreFromUnencrypted(pin string, pathplain string, pathname string) error {
	keystore, err := keys.LoadKeystoreUnencrypted(pathplain)
	if err != nil {
		return fmt.Errorf("error while loading unenctrypted keystore: %w", err)
	}
	err = keystore.Save(pathname, pin)
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
	ddb.PrintQRCode(os.Stdout, keystore.Key(keys.NodeMainTrunk))
	fmt.Printf("    %s\n", keystore.Key(keys.NodeMainTrunk))
	fmt.Printf("\n")
	fmt.Printf("    ADDRESS\n")
	ddb.PrintQRCode(os.Stdout, keystore.Address(keys.NodeMainTrunk))
	fmt.Printf("   %s\n", keystore.Address(keys.NodeMainTrunk))
	fmt.Printf("\n")
	fmt.Printf("   Bitcoin Key (WIF): %s\n", keystore.Key(keys.NodeMainTrunk))
	fmt.Printf("   Bitcoin Address: %s\n", keystore.Address(keys.NodeMainTrunk))
	fmt.Printf("   Passwords:\n")
	for n, p := range keystore.Passwords() {
		fmt.Printf("      %s: '%s' .  %d\n", n, p, len(p))
	}
}
