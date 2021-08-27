package ddb

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"path/filepath"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const (
	noncesize = 12
)

type MetaEntry struct {
	Name   string   `json:"n"`
	Labels []string `json:"l"`
	Mime   string   `json:"m"`
	Hash   string   `json:"h"`
	Date   string   `json:"e"`
	Notes  string   `json:"o,omitempty"`
}

type Entry struct {
	Name   string
	Labels []string
	Mime   string
	Hash   string
	Data   []byte
}

func NewEntryFromFile(name string, file string) (*Entry, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("error while reading file %s: %w", file, err)
	}
	fm := mime.TypeByExtension(filepath.Ext(file))
	return NewEntryFromData(name, fm, data), nil
}

func NewEntryFromData(name string, mime string, data []byte) *Entry {
	sha := sha256.Sum256(data)
	hash := hex.EncodeToString(sha[:])
	ent := Entry{Name: name, Mime: mime, Hash: hash, Data: data}
	return &ent

}

func EntriesFromParts(parts []*EntryPart) ([]*Entry, error) {
	t := trace.New().Source("entry.go", "Entry", "EntriesFromPart")
	trail.Println(trace.Debug("getting entries from parts").UTC().Append(t))
	entries := make([]*Entry, 0)
	partsDict := make(map[string][]*EntryPart)
	for _, p := range parts {
		if _, ok := partsDict[p.Name+p.Hash]; ok == false {
			partsDict[p.Name+p.Hash] = make([]*EntryPart, p.NumPart)
		}
		//fmt.Printf("entriesFromPart filling '%s' %d/%d\n", p.Name, p.IdxPart+1, p.NumPart)
		partsDict[p.Name+p.Hash][p.IdxPart] = p
	}
	for _, pa := range partsDict {
		if pa[0] == nil {
			trail.Println(trace.Warning("missing part").UTC().Add("part", "0").Append(t))
			continue
		}
		numPart := pa[0].NumPart
		entry := Entry{Name: pa[0].Name, Mime: pa[0].Mime, Hash: pa[0].Hash}
		data := make([]byte, 0)
		for i := 0; i < numPart; i++ {
			if pa[i] == nil {
				trail.Println(trace.Warning("missing part").UTC().Add("part", fmt.Sprintf("%d", i)).Append(t))
				continue
			}
			data = append(data, pa[i].Data...)
		}
		nh := sha256.Sum256(data)
		nhash := hex.EncodeToString(nh[:])
		if nhash != entry.Hash {
			trail.Println(trace.Alert("hash of decoded entry doesn't match").UTC().Add("new hash", nhash).Add("hash", entry.Hash))
			return nil, fmt.Errorf("hash of decoded entry doesn't match stored:%s  new:%s", entry.Hash, nhash)
		}
		entry.Data = data
		entries = append(entries, &entry)
	}
	return entries, nil
}

//EntriesOfFile returns an array of entries for the given file.
func (e *Entry) Parts(maxPartSize int) ([]*EntryPart, error) {
	t := trace.New().Source("entry.go", "Entry", "EntryOfFile")
	trail.Println(trace.Debug("cutting the entry in an array of EntryPart").UTC().Add("maxPartSize", fmt.Sprintf("%d", maxPartSize)).Append(t))
	hash := make([]byte, 64)
	sha := sha256.Sum256(e.Data)
	hex.Encode(hash, sha[:])
	numPart := (len(e.Data) / maxPartSize) + 1
	entries := make([]*EntryPart, 0, numPart)
	for i := 0; i < numPart; i++ {
		start := i * maxPartSize
		end := start + maxPartSize
		if end > len(e.Data)-1 {
			end = len(e.Data)
		}
		e := EntryPart{Name: e.Name, Hash: string(hash), Mime: e.Mime, IdxPart: i, NumPart: numPart, Size: (end - start), Data: e.Data[start:end]}
		entries = append(entries, &e)
	}
	return entries, nil
}

//EncryptedParts decompose the Entry in an array of EntryPart.
//The size of the encrypted EntryPart is guarantee to be less than maxPartSize
func (e *Entry) ToParts(password [32]byte, maxSize int) ([]*EntryPart, error) {
	t := trace.New().Source("entry.go", "Entry", "EntryOfFile")
	trail.Println(trace.Debug("cutting the entry in an array of EntryPart").UTC().Add("maxSize when encrypted", fmt.Sprintf("%d", maxSize)).Append(t))

	fits := false
	numPart := (len(e.Data) / maxSize)
	for fits == false {
		for i := 0; i < numPart; i++ {
			start := i * maxSize
			end := start + maxSize
			if end > len(e.Data)-1 {
				end = len(e.Data)
			}
			e := EntryPart{Name: e.Name, Hash: e.Hash, Mime: e.Mime, IdxPart: i, NumPart: numPart, Size: (end - start), Data: e.Data[start:end]}
			encdata, err := e.Encrypt(password)
			if err != nil {
				return nil, fmt.Errorf("error while encrypting EntryPart: %w", err)
			}
			if len(encdata) > maxSize {
				fits = false
				numPart++
				break
			}
			fits = true
		}
	}
	entries := make([]*EntryPart, 0, numPart)
	for i := 0; i < numPart; i++ {
		start := i * maxSize
		end := start + maxSize
		if end > len(e.Data)-1 {
			end = len(e.Data)
		}
		e := EntryPart{Name: e.Name, Hash: e.Hash, Mime: e.Mime, IdxPart: i, NumPart: numPart, Size: (end - start), Data: e.Data[start:end]}
		entries = append(entries, &e)
	}
	return entries, nil
}

//EntryPart is the payload of a single transaction, it can contains an entire file or be a single part of a multi entry file.
type EntryPart struct {
	Name    string   `json:"n"` //name of file
	Labels  []string `json:"l"` //labels
	Hash    string   `json:"h"` //hash of file
	Mime    string   `json:"m"` //mime type of file
	IdxPart int      `json:"i"` //index of part idx of numpart
	NumPart int      `json:"t"` //total number of parts that compose the entire file
	Size    int      `json:"s"` //size of data
	Data    []byte   `json:"d"` //data part of the file
}

//EntryPartFromEncodedData return the EntryPart decoded from the given json
func EntryPartFromJSON(encoded []byte) (*EntryPart, error) {
	var entry EntryPart
	err := json.Unmarshal(encoded, &entry)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal data: %w", err)
	}
	return &entry, nil
}

func EntryPartFromEncrypted(password [32]byte, encrypted []byte) (*EntryPart, error) {
	encoded, err := AESDecrypt(password, encrypted)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt data: %w", err)
	}
	var entry EntryPart
	err = json.Unmarshal(encoded, &entry)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal data: %w", err)
	}
	return &entry, nil
}

//Encode return the json encoded form of rhe EntryPart
func (e *EntryPart) ToJSON() ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("cannot encode to json: %w", err)
	}
	return data, nil
}

func (e *EntryPart) Encrypt(password [32]byte) ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("cannot encode EntryPart to JSON: %w", err)
	}
	enc, err := AESEncrypt(password, data)
	if err != nil {
		return nil, fmt.Errorf("cannot encrypt EntryPart: %w", err)
	}
	return enc, nil
}
