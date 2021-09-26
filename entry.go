package ddb

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"mime"
	"path/filepath"
	"time"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const (
	noncesize = 12
)

type MetaEntry struct {
	Name      string   `json:"n"`
	Labels    []string `json:"l"`
	Mime      string   `json:"m"`
	Hash      string   `json:"h"`
	Timestamp int64    `json:"e"`
	Notes     string   `json:"o,omitempty"`
}

func NewMetaEntry(entry *Entry) *MetaEntry {
	if entry == nil {
		return nil
	}
	requestTime := time.Now().Unix()
	meta := MetaEntry{Name: entry.Name, Labels: entry.Labels, Mime: entry.Mime, Hash: entry.Hash, Timestamp: requestTime, Notes: entry.Notes}
	return &meta
}

//Encrypt returns the EntryPart JSON encrypted.
func (me *MetaEntry) Encrypt(password [32]byte) ([]byte, error) {
	data, err := json.Marshal(me)
	if err != nil {
		return nil, fmt.Errorf("cannot encode MetaEntry to JSON: %w", err)
	}
	enc, err := AESEncrypt(password, data)
	if err != nil {
		return nil, fmt.Errorf("cannot encrypt MetaEntry: %w", err)
	}
	return enc, nil
}

type Entry struct {
	Name   string
	Labels []string
	Mime   string
	Hash   string
	Data   []byte
	Notes  string
}

func NewEntryFromFile(name string, file string, labels []string, notes string) (*Entry, error) {
	tr := trace.New().Source("entry.go", "Entry", "NewEntryFromFile")
	data, err := ioutil.ReadFile(file)
	if err != nil {
		trail.Println(trace.Alert("error while reading file").Append(tr).UTC().Add("file", file))
		return nil, fmt.Errorf("error while reading file %s: %w", file, err)
	}
	fm := mime.TypeByExtension(filepath.Ext(file))
	trail.Println(trace.Info("file read").Append(tr).UTC().Add("mime", fm).Add("size", fmt.Sprintf("%d", len(data))))
	return NewEntryFromData(name, fm, data, labels, notes), nil
}

func NewEntryFromData(name string, mime string, data []byte, labels []string, notes string) *Entry {
	sha := sha256.Sum256(data)
	hash := hex.EncodeToString(sha[:])
	ent := Entry{Name: name, Mime: mime, Hash: hash, Data: data, Labels: labels, Notes: notes}
	return &ent

}

//ToParts decompose the Entry in an array of EntryPart.
//The size of the encrypted EntryPart is guarantee to be less than maxPartSize
func (e *Entry) ToParts(password [32]byte, maxSize int) ([]*EntryPart, error) {
	tr := trace.New().Source("entry.go", "Entry", "ToParts")
	if maxSize < 300 {
		trail.Println(trace.Alert("maxSize excessively small (<300)").UTC().Append(tr).Add("maxSize", fmt.Sprintf("%d", maxSize)))
		return nil, fmt.Errorf("maxSize excessively small (<300): %d", maxSize)
	}
	//Find the correct size for entry data
	fits := false
	divisions := int(math.Ceil(float64(len(e.Data)) / float64(maxSize)))
	numParts := divisions + 1
	entryParts := make([]*EntryPart, 0, numParts)
	for !fits {
		partSize := len(e.Data) / divisions
		// fmt.Printf("Division: %d   PartSize: %d len(data): %d  MaxSize: %d\n", division, partSize, len(e.Data), maxSize)
		for i := 0; i < numParts; i++ {
			start := i * partSize
			end := start + partSize
			if end > len(e.Data)-1 {
				end = len(e.Data)
			}
			ep := EntryPart{Name: e.Name, Hash: e.Hash, Mime: e.Mime, IdxPart: i, NumPart: numParts, Size: (end - start), Data: e.Data[start:end]}
			encData, err := ep.Encrypt(password)
			if err != nil {
				trail.Println(trace.Alert("error while encrypting").UTC().Append(tr).Error(err))
				return nil, fmt.Errorf("error while encrypting: %w", err)
			}
			// fmt.Printf("encodedData: %d  maxSize: %d\n", len(encData), maxSize)
			if len(encData) > maxSize {
				fits = false
				divisions++
				numParts = divisions + 1
				entryParts = entryParts[:0]
				break
			}
			entryParts = append(entryParts, &ep)
			// fmt.Printf("appended entry: %d\n", ep.IdxPart)
			fits = true
		}
	}
	return entryParts, nil
}

func EntriesFromParts(parts []*EntryPart) ([]*Entry, error) {
	t := trace.New().Source("entry.go", "Entry", "EntriesFromPart")
	trail.Println(trace.Debug("getting entries from parts").UTC().Append(t))
	entries := make([]*Entry, 0)
	partsDict := make(map[string][]*EntryPart)
	for _, p := range parts {
		if _, ok := partsDict[p.Name+p.Hash]; !ok {
			partsDict[p.Name+p.Hash] = make([]*EntryPart, p.NumPart)
		}
		// fmt.Printf("entriesFromPart filling '%s' %d/%d\n", p.Name, p.IdxPart, p.NumPart)
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

//ToJSON returns the EntryPart JSON.
func (e *EntryPart) ToJSON() ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("cannot encode to json: %w", err)
	}
	return data, nil
}

//Encrypt returns the EntryPart JSON encrypted.
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
