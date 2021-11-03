package trh

import (
	"fmt"
	"os"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
)

const (
	keystoreFile        = "keystore.trh"
	keystoreOldFile     = "keystore.trh.old"
	keystoreUnencrypted = "keystore_plain.trh"
)

func loadKeyStore(pin string) (*keys.KeyStore, error) {
	return keys.LoadKeyStore(keystoreFile, pin)
}

func saveKeyStore(k *keys.KeyStore, pin string) error {
	return k.Save(keystoreFile, pin)
}

func saveUnencryptedKeyStore(k *keys.KeyStore) error {
	return k.SaveUnencrypted(keystoreUnencrypted)
}

func loadKeyStoreUnencrypted() (*keys.KeyStore, error) {
	return keys.LoadKeyStoreUnencrypted(keystoreUnencrypted)

}

func updateKeyStore(k *keys.KeyStore, pin string) error {
	return k.Update(keystoreFile, keystoreOldFile, pin)
}

func KeystoreGenFromKey(bitcoinKey string, password string, pin string) error {
	keyStore, err := keys.NewKeystore(bitcoinKey, password)
	if err != nil {
		return fmt.Errorf("provided key %s has issues: %w", bitcoinKey, err)
	}
	saveKeyStore(keyStore, pin)
	if err != nil {
		return fmt.Errorf("error while saving keystore to local file %s: %w", keystoreFile, err)
	}
	showKeystore(keyStore)
	return nil
}

func KeystoreGenFromPhrase(phrase string, keygenID int, pin string) error {
	wif, password, err := keys.FromPassphrase(phrase, keygenID)
	if err != nil {
		return fmt.Errorf("error while generating key using passphrase: %w", err)
	}
	keyStore, err := keys.NewKeystore(wif, password)
	if err != nil {
		return fmt.Errorf("error while generating keystore from passphrase '%s' with keygen '%d': %w", phrase, keygenID, err)
	}
	saveKeyStore(keyStore, pin)
	if err != nil {
		return fmt.Errorf("error while saving keystore to local file %s: %w", keystoreFile, err)
	}
	showKeystore(keyStore)
	return nil
}

func KeystoreShow(pin string) error {
	keyStore, err := loadKeyStore(pin)
	if err != nil {
		return fmt.Errorf("error while loading keystore: %w", err)
	}
	showKeystore(keyStore)
	return nil
}

func KeystoreSaveUnencrypted(pin string) error {
	keyStore, err := loadKeyStore(pin)
	if err != nil {
		return fmt.Errorf("error while loading keystore: %w", err)
	}
	saveUnencryptedKeyStore(keyStore)
	return nil
}

func KeystoreRestoreFromUnencrypted(pin string) error {
	keyStore, err := loadKeyStoreUnencrypted()
	if err != nil {
		return fmt.Errorf("error while loading unenctrypted keystore: %w", err)
	}
	updateKeyStore(keyStore, pin)
	showKeystore(keyStore)
	return nil
}

func showKeystore(keystore *keys.KeyStore) {
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
