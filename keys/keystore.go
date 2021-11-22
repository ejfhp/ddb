package keys

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const (
	NodeMainTrunk     = "main"
	NodeDefaultBranch = "default"
)

type Node struct {
	Key      string   `json:"key"`
	Address  string   `json:"address"`
	Password [32]byte `json:"password"`
	PassName string   `json:"passname"`
}

type Keystore struct {
	nodes    map[string]*Node
	pin      string
	pathname string
}

func NewKeystore(mainKey string, password string) (*Keystore, error) {
	ks := Keystore{}
	ks.nodes = make(map[string]*Node)
	add, err := AddressOf(mainKey)
	if err != nil {
		return nil, fmt.Errorf("cannot generate address from key '%s': %w", mainKey, err)
	}
	node := Node{Key: mainKey, Address: add, PassName: password, Password: passwordToBytes(password)}
	ks.nodes[NodeMainTrunk] = &node
	defaultBranchNode, err := ks.NodeFromPassword(password)
	if err != nil {
		return nil, fmt.Errorf("cannot generate default node from password '%s': %w", password, err)
	}
	ks.nodes[NodeDefaultBranch] = defaultBranchNode
	return &ks, nil
}

func LoadKeystore(filepath string, pin string) (*Keystore, error) {
	tr := trace.New().Source("keys.go", "Keystore", "LoadKeystore")
	var pinpass = passwordToBytes(pin)
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
	ks, err := UnmarshalJSON(encoded)
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
	return UnmarshalJSON(read)
}

func (ks Keystore) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Nodes map[string]*Node `json:"nodes"`
	}{
		Nodes: ks.nodes,
	})
}

func UnmarshalJSON(data []byte) (*Keystore, error) {
	var un struct {
		Nodes map[string]*Node `json:"nodes"`
	}
	err := json.Unmarshal(data, &un)
	if err != nil {
		return nil, err
	}
	k := Keystore{nodes: un.Nodes}
	return &k, nil
}

func (ks *Keystore) PassNames() []string {
	pws := make([]string, 0, len(ks.nodes))
	for k, _ := range ks.nodes {
		pws = append(pws, k)
	}
	return pws
}

func (ks *Keystore) Nodes() []*Node {
	ns := make([]*Node, 0, len(ks.nodes))
	for _, n := range ks.nodes {
		ns = append(ns, n)
	}
	return ns
}

func (ks *Keystore) Password(name string) [32]byte {
	n, ok := ks.nodes[name]
	if ok {
		return n.Password
	}
	return [32]byte{}
}

func (ks *Keystore) Node(name string) (*Node, bool) {
	n, ok := ks.nodes[name]
	return n, ok
}

func (ks *Keystore) Passwords() map[string][32]byte {
	pwsCopy := make(map[string][32]byte)
	for p, n := range ks.nodes {
		pwsCopy[p] = n.Password
	}
	return pwsCopy
}

func (ks *Keystore) Key(passname string) string {
	n, ok := ks.nodes[passname]
	if ok {
		return n.Key
	}
	return ""
}
func (ks *Keystore) Address(passname string) string {
	n, ok := ks.nodes[passname]
	if ok {
		return n.Address
	}
	return ""
}

func (ks *Keystore) KeysAndAddresses() map[string][]string {
	ka := make(map[string][]string)
	for p, n := range ks.nodes {
		ka[p] = []string{n.Key, n.Address}
	}
	return ka
}

//NodeFromPassword returns a node containing key and an address derived from the main address and the password.
func (ks *Keystore) NodeFromPassword(password string) (*Node, error) {
	node, exists := ks.nodes[password]
	if exists {
		return node, nil
	}
	keySeed := []byte{}
	//The new key is a function of the main address and the current password
	keySeed = append(keySeed, []byte(ks.Address(NodeMainTrunk))...)
	keySeed = append(keySeed, password[:]...)
	keySeedHash := sha256.Sum256(keySeed)
	key, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), keySeedHash[:])
	fbwif, err := bsvutil.NewWIF(key, &chaincfg.MainNetParams, true)
	if err != nil {
		return nil, fmt.Errorf("error while generating key: %v", err)
	}
	fbWIF := fbwif.String()
	fbAdd, err := AddressOf(fbWIF)
	if err != nil {
		return nil, fmt.Errorf("error while generating address: %v", err)
	}
	pass := passwordToBytes(password)
	node = &Node{Key: fbWIF, Address: fbAdd, Password: pass, PassName: password}
	return node, nil

}

//StoreNode adds the node to the list if it's new. Return true if it's new and the keystore should be updated
func (ks *Keystore) StoreNode(node *Node) bool {
	_, exists := ks.nodes[node.PassName]
	if !exists {
		ks.nodes[node.PassName] = node
	}
	return !exists
}

func (ks *Keystore) Save(filepath string, pin string) error {
	tr := trace.New().Source("keys.go", "Keystore", "Save")
	encoded, err := json.Marshal(ks)
	if err != nil {
		trail.Println(trace.Alert("cannot encode Keystore").UTC().Error(err).Append(tr))
		return fmt.Errorf("cannot encode Keystore: %w", err)
	}
	var pinpass = passwordToBytes(pin)
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

func passwordToBytes(password string) [32]byte {
	var pass [32]byte
	copy(pass[:], []byte(password))
	return pass
}

func passwordToString(password [32]byte) string {
	return strings.TrimSpace(string(password[:]))
}
