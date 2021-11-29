package ddb

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ejfhp/ddb/keys"
)

type MetaEntry struct {
	Name      string   `json:"n"`
	Password  [32]byte `json:"p"`
	Key       string   `json:"k"`
	Address   string   `json:"a"`
	Labels    []string `json:"l"`
	Mime      string   `json:"m"`
	Hash      string   `json:"h"`
	Timestamp int64    `json:"e"`
	Notes     string   `json:"o,omitempty"`
	Size      int      `json:"s"`
}

func NewMetaEntry(node *keys.Node, entry *Entry) *MetaEntry {
	if entry == nil {
		return nil
	}
	requestTime := time.Now().Unix()
	meta := MetaEntry{
		Name:      entry.Name,
		Password:  node.Password(),
		Key:       node.Key(),
		Address:   node.Address(),
		Labels:    entry.Labels,
		Mime:      entry.Mime,
		Hash:      entry.Hash,
		Timestamp: requestTime,
		Notes:     entry.Notes,
		Size:      entry.Size}
	return &meta
}

func MetaEntryFromEncrypted(password [32]byte, encrypted []byte) (*MetaEntry, error) {
	encoded, err := keys.AESDecrypt(password, encrypted)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt data: %w", err)
	}
	var mentry MetaEntry
	err = json.Unmarshal(encoded, &mentry)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal data: %w", err)
	}
	return &mentry, nil
}

//Encrypt returns the EntryPart JSON encrypted.
func (me *MetaEntry) Encrypt(password [32]byte) ([]byte, error) {
	data, err := json.Marshal(me)
	if err != nil {
		return nil, fmt.Errorf("cannot encode MetaEntry to JSON: %w", err)
	}
	enc, err := keys.AESEncrypt(password, data)
	if err != nil {
		return nil, fmt.Errorf("cannot encrypt MetaEntry: %w", err)
	}
	return enc, nil
}
