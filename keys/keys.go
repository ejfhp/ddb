package keys

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const (
	Main = "main"
)

type KeyStore struct {
	passwords map[string][32]byte
	pKeyAdd   map[string][]string
}

func NewKeystore(mainKey string, password string) (*KeyStore, error) {
	ks := KeyStore{}
	ks.passwords = make(map[string][32]byte)
	ks.pKeyAdd = make(map[string][]string)
	add, err := AddressOf(mainKey)
	if err != nil {
		return nil, fmt.Errorf("cannot generate address from key '%s': %w", mainKey, err)
	}
	ks.pKeyAdd[Main] = []string{mainKey, add}
	return &ks, nil
}

func LoadKeyStore(filepath string, pin string) (*KeyStore, error) {
	tr := trace.New().Source("keys.go", "KeyStore", "LoadKeyStore")
	var pinpass = PasswordFromString(pin)
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

func LoadKeyStoreUnencrypted(filepath string) (*KeyStore, error) {
	tr := trace.New().Source("keys.go", "KeyStore", "LoadKeyStoreUnencrypted")
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
	var k KeyStore
	err = json.Unmarshal(read, &k)
	if err != nil {
		trail.Println(trace.Alert("cannot decode KeyStore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot decode KeyStore file: %w", err)
	}
	return &k, nil
}

func (ks *KeyStore) PassNames() []string {
	pws := make([]string, 0, len(ks.passwords))
	for k, _ := range ks.passwords {
		pws = append(pws, k)
	}
	return pws
}

func (ks *KeyStore) Password(name string) [32]byte {
	return ks.passwords[name]
}

func (ks *KeyStore) Key(name string) string {
	ka, ok := ks.pKeyAdd[name]
	if ok {
		return ka[0]
	}
	return ""
}
func (ks *KeyStore) Address(name string) string {
	ka, ok := ks.pKeyAdd[name]
	if ok {
		return ka[1]
	}
	return ""
}

func (ks KeyStore) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Passwords map[string][32]byte `json:"password"`
		PKeyAdd   map[string][]string `json:"pwifadd"`
	}{
		Passwords: ks.passwords,
		PKeyAdd:   ks.pKeyAdd,
	})
}

//AddNewKeyAndAddress returns key and address to be used when branching an entry. Password, key and address are stored in the Keystore.
func (ks *KeyStore) AddNewKeyAndAddress(password [32]byte) (string, string, error) {
	keySeed := []byte{}
	keySeed = append(keySeed, []byte(ks.Address(Main))...)
	keySeed = append(keySeed, password[:]...)
	keySeedHash := sha256.Sum256(keySeed)
	key, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), keySeedHash[:])
	fbwif, err := bsvutil.NewWIF(key, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", "", fmt.Errorf("error while generating key: %v", err)
	}
	fbWIF := fbwif.String()
	fbAdd, err := AddressOf(fbWIF)
	if err != nil {
		return "", "", fmt.Errorf("error while generating address: %v", err)
	}
	pass := PasswordToString(password)
	ks.passwords[pass] = password
	ks.pKeyAdd[pass] = []string{fbWIF, fbAdd}
	return fbWIF, fbAdd, nil

}

func (ks *KeyStore) Save(filepath string, pin string) error {
	tr := trace.New().Source("keys.go", "KeyStore", "Save")
	encoded, err := json.Marshal(ks)
	if err != nil {
		trail.Println(trace.Alert("cannot encode KeyStore").UTC().Error(err).Append(tr))
		return fmt.Errorf("cannot encode KeyStore: %w", err)
	}
	var pinpass = PasswordFromString(pin)
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

func (ks *KeyStore) SaveUnencrypted(filepath string) error {
	tr := trace.New().Source("keys.go", "KeyStore", "SaveUnencrypted")
	encoded, err := json.Marshal(ks)
	if err != nil {
		trail.Println(trace.Alert("cannot encode KeyStore").UTC().Error(err).Append(tr))
		return fmt.Errorf("cannot encode KeyStore: %w", err)
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
	n, err := file.Write(encoded)
	if err != nil || n != len(encoded) {
		trail.Println(trace.Alert("error while writing KeyStore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return fmt.Errorf("error while writing KeyStore file: %w", err)
	}
	return nil
}

func (ks *KeyStore) Update(filepath string, oldFile string, pin string) error {
	tr := trace.New().Source("keys.go", "KeyStore", "Update")
	copy, err := ioutil.ReadFile(filepath)
	if err != nil {
		trail.Println(trace.Alert("cannot read current keystore file").Append(tr).UTC().Error(err))
		return fmt.Errorf("cannot read current keystore file: %w", err)
	}
	err = os.Remove(oldFile)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			trail.Println(trace.Alert("cannot delete old keystore file").Append(tr).UTC().Error(err))
			return fmt.Errorf("cannot delete old keystore file: %w", err)
		}
	}
	err = ioutil.WriteFile(oldFile, copy, 0644)
	if err != nil {
		trail.Println(trace.Alert("cannot duplicate current keystore file").Append(tr).UTC().Error(err))
		return fmt.Errorf("cannot duplicate current keystore file: %w", err)
	}
	ck, err := LoadKeyStore(filepath, pin)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			trail.Println(trace.Alert("keystore doesn't exist in the current directory").Append(tr).UTC().Error(err))
			return fmt.Errorf("keystore doesn't exist in the current directory: %w", err)
		} else {
			trail.Println(trace.Alert("cannot load current keystore file").Append(tr).UTC().Error(err))
			return fmt.Errorf("cannot load current keystore file: %w", err)
		}
	}
	if ck.Key(Main) != ks.Key(Main) {
		trail.Println(trace.Alert("in memory keystore and saved keystore have a different key").Append(tr).UTC())
		return fmt.Errorf("in memory keystore and saved keystore have a different key")
	}
	err = os.Remove(filepath)
	if err != nil {
		trail.Println(trace.Alert("cannot delete old keystore file").Append(tr).UTC().Error(err))
		return fmt.Errorf("cannot delete old keystore file: %w", err)
	}
	return ks.Save(filepath, pin)
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
