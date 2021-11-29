package keys

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const (
	NodeDefaultBranch = "default"
)

type Source struct {
	phrase   string
	keygenID int
	key      string
	address  string
	password string
}

func (so Source) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Phrase   string `json:"phrase,omitempty"`
		KeygenID int    `json:"keygenid,omitempty"`
		Key      string `json:"key"`
		Address  string `json:"address"`
		Password string `json:"password,omitempty"`
	}{
		Phrase:   so.phrase,
		KeygenID: so.keygenID,
		Key:      so.key,
		Address:  so.address,
		Password: so.password,
	})
}

func (so *Source) UnmarshalJSON(data []byte) error {
	var un struct {
		Phrase   string `json:"phrase,omitempty"`
		KeygenID int    `json:"keygenid,omitempty"`
		Key      string `json:"key"`
		Address  string `json:"address"`
		Password string `json:"password,omitempty"`
	}
	err := json.Unmarshal(data, &un)
	if err != nil {
		return err
	}
	*so = Source{
		phrase:   un.Phrase,
		keygenID: un.KeygenID,
		key:      un.Key,
		address:  un.Address,
		password: un.Password,
	}
	return nil
}

func (so *Source) Phrase() string {
	return so.phrase
}

func (so *Source) KeygenID() int {
	return so.keygenID
}

func (so *Source) Key() string {
	return so.key
}

func (so *Source) Address() string {
	return so.address
}

func (so *Source) Password() string {
	return so.password
}

type Node struct {
	name      string
	key       string
	address   string
	password  [32]byte
	hashHEX   string
	timestamp int64
}

func (no Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Name      string   `json:"name"`
		Key       string   `json:"key"`
		Address   string   `json:"address"`
		Password  [32]byte `json:"password"`
		HashHEX   string   `json:"hashhex"`
		Timestamp int64    `json:"timestamp"`
	}{
		Name:      no.name,
		Key:       no.key,
		Address:   no.address,
		Password:  no.password,
		HashHEX:   no.hashHEX,
		Timestamp: no.timestamp,
	})
}

func (no *Node) UnmarshalJSON(data []byte) error {
	var un struct {
		Name      string   `json:"name"`
		Key       string   `json:"key"`
		Address   string   `json:"address"`
		Password  [32]byte `json:"password"`
		HashHEX   string   `json:"hashhex"`
		Timestamp int64    `json:"timestamp"`
	}
	err := json.Unmarshal(data, &un)
	if err != nil {
		return err
	}
	*no = Node{
		name:      un.Name,
		key:       un.Key,
		address:   un.Address,
		password:  un.Password,
		hashHEX:   un.HashHEX,
		timestamp: un.Timestamp,
	}
	return nil
}

func (no *Node) Name() string {
	return no.name
}

func (no *Node) Key() string {
	return no.key
}

func (no *Node) Address() string {
	return no.address
}

func (no *Node) Password() [32]byte {
	return no.password
}

func (no *Node) HashHEX() string {
	return no.hashHEX
}

func (no *Node) Timestamp() time.Time {
	return time.Unix(no.timestamp, 0)
}

type Keystore struct {
	source   *Source
	nodes    []*Node
	pin      string
	pathname string
}

func NewKeystore(mainKey string, password string) (*Keystore, error) {
	ks := Keystore{}
	ks.nodes = []*Node{}
	add, err := AddressOf(mainKey)
	if err != nil {
		return nil, fmt.Errorf("cannot generate address from key '%s': %w", mainKey, err)
	}
	source := Source{key: mainKey, address: add, password: password}
	ks.source = &source
	return &ks, nil
}

func LoadKeystore(filepath string, pin string) (*Keystore, error) {
	tr := trace.New().Source("keys.go", "Keystore", "LoadKeystore")
	var pinpass = StringToPassword(pin)
	file, err := os.Open(filepath)
	if err != nil {
		trail.Println(trace.Alert("cannot open Keystore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot open Keystore file: %w", err)
	}
	defer file.Close()
	read, err := ioutil.ReadAll(file)
	if err != nil || len(read) == 0 {
		trail.Println(trace.Alert("error while reading Keystore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("error while reading Keystore file: %w", err)
	}
	encoded, err := AESDecrypt(pinpass, read)
	if err != nil {
		trail.Println(trace.Alert("cannot decrypt Keystore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot decrypt Keystore file: %w", err)
	}
	ks := &Keystore{}
	err = ks.UnmarshalJSON(encoded)
	if err != nil {
		return nil, err
	}
	ks.pin = pin
	ks.pathname = filepath
	return ks, nil
}

func LoadKeystoreUnencrypted(filepath string) (*Keystore, error) {
	tr := trace.New().Source("keys.go", "Keystore", "LoadKeystoreUnencrypted")
	file, err := os.Open(filepath)
	if err != nil {
		trail.Println(trace.Alert("cannot open Keystore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot open Keystore file: %w", err)
	}
	defer file.Close()
	read, err := ioutil.ReadAll(file)
	if err != nil || len(read) == 0 {
		trail.Println(trace.Alert("error while reading Keystore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return nil, fmt.Errorf("error while reading Keystore file: %w", err)
	}
	k := &Keystore{}
	err = k.UnmarshalJSON(read)
	return k, err
}

func (ks Keystore) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Nodes  []*Node `json:"nodes"`
		Source *Source `json:"source"`
	}{
		Nodes:  ks.nodes,
		Source: ks.source,
	})
}

func (ks *Keystore) UnmarshalJSON(data []byte) error {
	var un struct {
		Nodes  []*Node `json:"nodes"`
		Source *Source `json:"source"`
	}
	err := json.Unmarshal(data, &un)
	if err != nil {
		return err
	}
	*ks = Keystore{nodes: un.Nodes, source: un.Source}
	return nil
}

func (ks *Keystore) SetPhrase(passphrase string, keygenID int) {
	ks.source.phrase = passphrase
	ks.source.keygenID = keygenID
}

//NewNode returns a node containing key, address and password derived from the source and the given hash.
func (ks *Keystore) NewNode(entityname string, hash [32]byte) (*Node, error) {
	if len(entityname) == 0 {
		return nil, fmt.Errorf("entity name cannot be empty")
	}
	keySeed := []byte{}
	//The new key is a function of the main key and the given hash
	keySeed = append(keySeed, []byte(ks.source.key)...)
	keySeed = append(keySeed, hash[:]...)
	keySeedHash := sha256.Sum256(keySeed)
	key, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), keySeedHash[:])
	nwif, err := bsvutil.NewWIF(key, &chaincfg.MainNetParams, true)
	if err != nil {
		return nil, fmt.Errorf("error while generating key: %v", err)
	}
	nWIF := nwif.String()
	nAdd, err := AddressOf(nWIF)
	if err != nil {
		return nil, fmt.Errorf("error while generating address: %v", err)
	}

	pwdSeed := []byte{}
	//The new key is a function of the main password and the given hash
	pwdSeed = append(pwdSeed, []byte(ks.source.password)...)
	pwdSeed = append(pwdSeed, hash[:]...)
	pwdSeedHash := sha256.Sum256(pwdSeed)

	for _, n := range ks.nodes {
		if n.key == nWIF {
			return n, nil
		}
	}
	hashhex := hex.EncodeToString(hash[:])
	node := &Node{
		name:      entityname,
		timestamp: time.Now().Unix(),
		key:       nWIF,
		address:   nAdd,
		password:  pwdSeedHash,
		hashHEX:   hashhex,
	}
	ks.nodes = append(ks.nodes, node)
	return node, nil
}

func (ks *Keystore) Source() *Source {
	return ks.source
}

func (ks *Keystore) Nodes() []*Node {
	nodes := make([]*Node, len(ks.nodes))
	for i, n := range ks.nodes {
		nodes[i] = n
	}
	return nodes
}

func (ks *Keystore) Save(filepath string, pin string) error {
	tr := trace.New().Source("keys.go", "Keystore", "Save")
	encoded, err := json.Marshal(ks)
	if err != nil {
		trail.Println(trace.Alert("cannot encode Keystore").UTC().Error(err).Append(tr))
		return fmt.Errorf("cannot encode Keystore: %w", err)
	}
	var pinpass = StringToPassword(pin)
	encrypted, err := AESEncrypt(pinpass, encoded)
	if err != nil {
		trail.Println(trace.Alert("cannot encrypt Keystore").UTC().Error(err).Append(tr))
		return fmt.Errorf("cannot encrypt Keystore: %w", err)
	}
	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("Keystore already exsist")
	}
	file, err := os.Create(filepath)
	if err != nil {
		trail.Println(trace.Alert("cannot create Keystore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return fmt.Errorf("cannot create Keystore file: %w", err)
	}
	defer file.Close()
	n, err := file.Write(encrypted)
	if err != nil || n != len(encrypted) {
		trail.Println(trace.Alert("error while writing Keystore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return fmt.Errorf("error while writing Keystore file: %w", err)
	}
	ks.pathname = filepath
	ks.pin = pin
	return nil
}

func (ks *Keystore) SaveUnencrypted(filepath string) error {
	tr := trace.New().Source("keys.go", "Keystore", "SaveUnencrypted")
	encoded, err := json.Marshal(ks)
	if err != nil {
		trail.Println(trace.Alert("cannot encode Keystore").UTC().Error(err).Append(tr))
		return fmt.Errorf("cannot encode Keystore: %w", err)
	}
	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("Keystore unencrypted already exsist")
	}
	file, err := os.Create(filepath)
	if err != nil {
		trail.Println(trace.Alert("cannot create Keystore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return fmt.Errorf("cannot create Keystore file: %w", err)
	}
	defer file.Close()
	n, err := file.Write(encoded)
	if err != nil || n != len(encoded) {
		trail.Println(trace.Alert("error while writing Keystore file").UTC().Add("filepath", filepath).Error(err).Append(tr))
		return fmt.Errorf("error while writing Keystore file: %w", err)
	}
	return nil
}

func (ks *Keystore) Update() error {
	tr := trace.New().Source("keys.go", "Keystore", "Update")
	if ks.pathname == "" || ks.pin == "" {
		return fmt.Errorf("PIN of Filenamen in Keystore missing")
	}
	oldFile := ks.pathname + ".old"
	copy, err := ioutil.ReadFile(ks.pathname)
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
	err = os.Remove(ks.pathname)
	if err != nil {
		trail.Println(trace.Alert("cannot delete old keystore file").Append(tr).UTC().Error(err))
		return fmt.Errorf("cannot delete old keystore file: %w", err)
	}
	return ks.Save(ks.pathname, ks.pin)
}

func StringToPassword(password string) [32]byte {
	var pass [32]byte
	copy(pass[:], []byte(password))
	return pass
}

func PasswordToString(password [32]byte) string {
	return string(bytes.Trim(password[:], "\x00"))
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
