package ddb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Keygen interface {
	WIF() (string, error)
	Password() [32]byte
}

func DecodeWIF(wifkey string) (*bsvec.PrivateKey, error) {
	t := trace.New().Source("keys.go", "", "DecodeWIF")
	wif, err := bsvutil.DecodeWIF(wifkey)
	if err != nil {
		trail.Println(trace.Alert("cannot decode WIF").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot decode WIF: %w", err)
	}
	priv := wif.PrivKey
	return priv, nil
}

type KeyStore struct {
	WIF       string              `json:"wif"`
	Address   string              `json:"address"`
	Passwords map[string][32]byte `json:"password"`
}

func NewKeystore() *KeyStore {
	ks := KeyStore{}
	ks.Passwords = make(map[string][32]byte)
	return &ks
}

func LoadKeyStore(filepath string, pin string) (*KeyStore, error) {
	tr := trace.New().Source("keys.go", "KeyStore", "LoadKeyStore")
	var pinpass = PINPassFromPIN(pin)
	file, err := os.Open(filepath)
	if err != nil {
		trail.Println(trace.Alert("cannot open KeyStore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot open KeyStore file: %w", err)
	}
	defer file.Close()
	read, err := ioutil.ReadAll(file)
	if err != nil || len(read) == 0 {
		trail.Println(trace.Alert("error while reading KeyStore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("error while reading KeyStore file: %w", err)
	}
	encoded, err := AESDecrypt(pinpass, read)
	if err != nil {
		trail.Println(trace.Alert("cannot decrypt KeyStore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot decrypt KeyStore file: %w", err)
	}
	var k KeyStore
	err = json.Unmarshal(encoded, &k)
	if err != nil {
		trail.Println(trace.Alert("cannot decode KeyStore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot decode KeyStore file: %w", err)
	}
	return &k, nil
}

func (ks *KeyStore) Password(label string) string {
	pwd, ok := ks.Passwords["main"]
	if !ok {
		return ""
	}
	return string(pwd[:])
}

func (ks *KeyStore) Save(filepath string, pin string) error {
	tr := trace.New().Source("keys.go", "KeyStore", "Save")
	encoded, err := json.Marshal(ks)
	if err != nil {
		trail.Println(trace.Alert("cannot encode KeyStore").UTC().Error(err).Append(tr))
		return fmt.Errorf("cannot encode KeyStore: %w", err)
	}
	var pinpass = PINPassFromPIN(pin)
	encrypted, err := AESEncrypt(pinpass, encoded)
	if err != nil {
		trail.Println(trace.Alert("cannot encrypt KeyStore").UTC().Error(err).Append(tr))
		return fmt.Errorf("cannot encrypt KeyStore: %w", err)
	}
	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("KeyStore already exsist")
	}
	file, err := os.Create(filepath)
	if err != nil {
		trail.Println(trace.Alert("cannot create KeyStore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return fmt.Errorf("cannot create KeyStore file: %w", err)
	}
	defer file.Close()
	n, err := file.Write(encrypted)
	if err != nil || n != len(encrypted) {
		trail.Println(trace.Alert("error while writing KeyStore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return fmt.Errorf("error while writing KeyStore file: %w", err)
	}
	return nil
}

func AddressOf(wifkey string) (string, error) {
	tr := trace.New().Source("keys.go", "", "AddressOf")
	w, err := bsvutil.DecodeWIF(wifkey)
	if err != nil {
		trail.Println(trace.Alert("cannot decode WIF").UTC().Add("wif", wifkey).Error(err).Append(tr))
		return "", fmt.Errorf("cannot decode WIF: %w", err)
	}
	_, err = bsvec.ParsePubKey(w.SerializePubKey(), bsvec.S256())
	if err != nil {
		trail.Println(trace.Alert("cannot parse").UTC().Add("wif", wifkey).Error(err).Append(tr))
		return "", err
	}
	// fmt.Printf("pubk: %s\n", string(pubk.SerializeCompressed()))
	// fmt.Printf("pubk ser: %s\n", string(w.SerializePubKey()))
	add, err := bsvutil.NewAddressPubKey(w.SerializePubKey(), &chaincfg.MainNetParams)
	if err != nil {
		trail.Println(trace.Alert("cannot generate address from WIF").UTC().Add("wif", wifkey).Error(err).Append(tr))
		return "", fmt.Errorf("cannot generate address from WIF: %w", err)
	}
	return add.EncodeAddress(), nil

}

func PINPassFromPIN(PIN string) [32]byte {
	var pinpass = [32]byte{}
	for i := 0; i < len(pinpass); i++ {
		pinpass[i] = '#'
	}
	copy(pinpass[:], PIN[:])
	return pinpass
}
