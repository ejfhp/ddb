package ddb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime"
	"path/filepath"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type EntryPart struct {
	Name    string `json:"n"`
	Hash    string `json:"h"`
	Mime    string `json:"m"`
	IdxPart int    `json:"i"`
	NumPart int    `json:"t"`
	Size    int    `json:"s"`
	Data    []byte `json:"d"`
}

func EntryOfFile(filename string, data []byte, maxPartSize int) ([]*EntryPart, error) {
	t := trace.New().Source("entry.go", "Entry", "EntryOfFile")
	log.Println(trace.Debug("building entries").UTC().Add("maxPartSize", fmt.Sprintf("%d", maxPartSize)).Append(t))
	fm := mime.TypeByExtension(filepath.Ext(filename))
	fmt.Printf("ext: %s   mime: %s\n", filepath.Ext(filename), fm)
	hash := make([]byte, 64)
	sha := sha256.Sum256(data)
	hex.Encode(hash, sha[:])
	numPart := (len(data) / maxPartSize) + 1
	entries := make([]*EntryPart, 0, numPart)
	for i := 0; i < numPart; i++ {
		start := i * maxPartSize
		end := start + maxPartSize
		if end > len(data)-1 {
			end = len(data)
		}
		e := EntryPart{Name: filename, Hash: string(hash), Mime: fm, IdxPart: i, NumPart: numPart, Size: (end - start), Data: data[start:end]}
		entries = append(entries, &e)
	}
	return entries, nil
}
