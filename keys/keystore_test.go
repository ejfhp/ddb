package keys_test

import (
	"os"
	"testing"

	"github.com/ejfhp/ddb/keys"
)

//EMPTY TEST ADDRESS
var destinationAddress string = "1PGh5YtRoohzcZF7WX8SJeZqm6wyaCte7X"
var destinationKey string = "L4ZaBkP1UTyxdEM7wysuPd1scHMLLf8sf8B2tcEcssUZ7ujrYWcQ"
var changeAddress string = "1EpFjTzJoNAFyJKVGATzxhgqXigUWLNWM6"
var changeKey string = "L2mk9qzXebT1gfwUuALMJrbqBtrJxGUN5JnVeqQTGRXytqpXsPr8"

func TestDecodeWIF(t *testing.T) {
	wif := "L2Aoi3Zk9oQhiEBwH9tcqnTTRErh7J3bVWoxLDzYa8nw2bWktG6M"
	k, err := keys.DecodeWIF(wif)
	if err != nil {
		t.Fatalf("WIF decoding failed: %v", err)
	}
	if k == nil {
		t.Fatalf("WIF decoded key is nil")
	}
}

func TestAddressOf(t *testing.T) {
	mapkeys := map[string]string{
		"1GB5MLgNF4zDVQc65BdrXKac1GJK8K59Ck": "KxdpCLdUFVuY9KCLaRVGfsSKQWnFobegqVjn8tM8oPo3UBbzgraF",
		"17cM2c5ybSidHThYa5rBykMEJ5dANkJWVW": "L3MB8BnEVH1gM4oGADEqXLWLpVXvbXP5pf7ezZaSoWi37sig3ZA6",
		"1KiMqNRH98WJGosedCZyw3nzJQG8w3iN54": "KzQiUaeAx9vfDSdMaFseaNzgvkXYzDPLJEiTxFHT4oQKgT4zLowf",
		"1JLqtRfMf77vbeE8ASPP3hWLduLBow5fQP": "Kzx2g5x4tDavJfRX7fhewQvjtR2kg5EkF47y2NPnN6vxux4Ag7pT",
		"1KKK563UqCR5nz5figdRekp4BUzvCQ7S3B": "L3HcLioKSRRsCjafvRfZ7Yre2UqSU5cQ6C8W6zi849RjzJ9cN3Wh",
		"1Nme23uK2iFW3MX8UEguLQeNNHwdEr23TL": "L1htQ8AePB3t4PxUr4wAojWYZmB9RiQCDBRB5C97GynkmvPKXrGn",
		"17HKJKar5dh3HbzGpB4Hoy7WHHo5totazd": "L4yumTCmLnJQBvDrKy1gTS1fbGrw792uWBftZaShg67uT7CapX7T",
		"1BXbpQ9ffsXRr9uyUCy1X4mXDnz7iHY7Qs": "L49YYrcxJWDG8emWPGrdTisSCsq1HYLRnqP2rzXHrHcCgNZ6khG7",
		"1ADi6SNG6LqX3PmdANhBAZY8oGbZbDFtAb": "L12fQB2YPC6rXZB2f8y2j6c2dzjiMQA58vuuBNJXYbNtiiL7yKq1",
		"1BRiuijd9zSsybGdQqoC5G67oXQLgMTojg": "KxGcDN28hBLfEDF6wPfB9c4ftVFm4nddMB2AoSDFVwz4sTw9CMmQ",
	}
	for add, key := range mapkeys {
		a, err := keys.AddressOf(key)
		if err != nil {
			t.Error(err)
		}
		if a != add {
			t.Fatalf("geerated %s != expected  %s", a, add)
		}
	}
}

func TestKeyStore_Save_LoadKeystore(t *testing.T) {
	keyfile := "/tmp/keystore.trhk"
	os.RemoveAll(keyfile)
	pin := "trh"
	passwordt := "tantovalagattaallardochecilascia"
	password := [32]byte{}
	copy(password[:], []byte(passwordt)[:])
	ks, err := keys.NewKeystore(destinationKey, string(passwordt))
	if err != nil {
		t.Logf("failed to create keystore: %v", err)
		t.FailNow()
	}
	ks.AddNewKeyAndAddress(password)
	err = ks.Save(keyfile, pin)
	if err != nil {
		t.Logf("failed to save keystore: %v", err)
		t.FailNow()
	}
	ks2, err := keys.LoadKeyStore(keyfile, pin)
	if err != nil {
		t.Logf("failed to load keystore: %v", err)
		t.FailNow()
	}
	if ks2.Address(keys.Main) != destinationAddress {
		t.Logf("load keystore has wrong address: %s", ks2.Address(keys.Main))
		t.FailNow()
	}
	if ks2.Key(keys.Main) != destinationKey {
		t.Logf("load keystore has wrong key: %s", ks2.Key(keys.Main))
		t.FailNow()
	}
	p2 := ks2.Password("one")
	if string(p2[:]) != passwordt {
		t.Logf("load keystore has wrong password: %s", string(p2[:]))
		t.FailNow()
	}
}

func TestKeyStore_GenerateKeyAndAddress(t *testing.T) {
	// trail.SetWriter(os.Stdout)
	passwords := [][32]byte{
		{'a', ' ', '3', '2', ' ', 'b', 'y', 't', 'e', ' ', 'p', 'a', 's', 's', 'w', 'o', 'r', 'd', ' ', 'i', 's', ' ', 'v', 'e', 'r', 'y', ' ', 'l', 'o', 'n', 'g'},
		{'c', 'i', 'a', 'o', 'm', 'a', 'm', 'm', 'a', 'g', 'u', 'a', 'r', 'd', 'a', 'c', 'o', 'm', 'e', 'm', 'i', 'd', 'i', 'v', 'e', 'r', 't', 'o', '.', '.', '.'},
	}
	for i, v := range passwords {
		ks, err := keys.NewKeystore(destinationKey, "mainpassword")
		if err != nil {
			t.Logf("%d - failed to generate ney keystore: %v", i, err)
			t.FailNow()
		}
		bWIF, bAdd, err := ks.AddNewKeyAndAddress(v)
		if err != nil {
			t.Logf("%d - failed to generate key and add: %v", i, err)
			t.FailNow()
		}
		b2Add, err := keys.AddressOf(bWIF)
		if err != nil {
			t.Logf("%d - failed to generate address from generated WIF: %v", i, err)
			t.FailNow()
		}
		if b2Add != bAdd {
			t.Logf("%d - something is wrong in the key-add generation", i)
			t.FailNow()
		}
	}
}
