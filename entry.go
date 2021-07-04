package ddb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"path/filepath"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const (
	noncesize = 12
)

type Entry struct {
	Name string
	Data []byte
}

func EntryFromParts(parts []*EntryPart) (*Entry, error) {
	return nil, nil
}

//EntryPart is the payload of a single transaction, it can contains an entire file or be a single part of a multi entry file.
type EntryPart struct {
	Name    string `json:"n"` //name of file
	Hash    string `json:"h"` //hash of file
	Mime    string `json:"m"` //mime type of file
	IdxPart int    `json:"i"` //index of part idx of numpart
	NumPart int    `json:"t"` //total number of parts that compose the entire file
	Size    int    `json:"s"` //size of data
	Data    []byte `json:"d"` //data part of the file
}

//EntriesOfFile returns an array of entries for the given file.
func (e *Entry) Parts(maxPartSize int) ([]*EntryPart, error) {
	t := trace.New().Source("entry.go", "Entry", "EntryOfFile")
	log.Println(trace.Debug("building entries").UTC().Add("maxPartSize", fmt.Sprintf("%d", maxPartSize)).Append(t))
	fm := mime.TypeByExtension(filepath.Ext(e.Name))
	fmt.Printf("ext: %s   mime: %s\n", filepath.Ext(e.Name), fm)
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
		e := EntryPart{Name: e.Name, Hash: string(hash), Mime: fm, IdxPart: i, NumPart: numPart, Size: (end - start), Data: e.Data[start:end]}
		entries = append(entries, &e)
	}
	return entries, nil
}

//Encode return the json encoded form of rhe EntryPart
func (e *EntryPart) Encode() ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("cannot encode to json: %w", err)
	}
	return data, nil
}

//Decode return the EntryPart decoded from the given json
func Decode(encoded []byte) (*EntryPart, error) {
	var entry EntryPart
	err := json.Unmarshal(encoded, &entry)
	return &entry, err
}

// Key should be 32 bytes (AES-256).
func AESEncrypt(key [32]byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cannot initialize key: %w", err)
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, noncesize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("cannot generate nonce: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize encryption: %w", err)
	}

	encoded := make([]byte, 0)
	encoded = append(encoded, nonce...)
	encoded = append(encoded, aesgcm.Seal(nil, nonce, data, nil)...)
	return encoded, nil
}

// Key should be 32 bytes (AES-256).
func AESDecrypt(key [32]byte, encrypted []byte) ([]byte, error) {
	nonce := encrypted[:noncesize]
	cipherdata := encrypted[noncesize:]

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cannot initialize key: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize encryption: %w", err)
	}

	plaindata, err := aesgcm.Open(nil, nonce, cipherdata, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt: %w", err)
	}

	return plaindata, err
}
