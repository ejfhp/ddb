package ddb

import (
	"fmt"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type EntryPart struct {
	Name    string `json:"n"`
	Hash    string `json:"h"`
	Mime    string `json:"m"`
	IdxPart int    `json:"i"`
	NumPart int    `json:"t"`
	Data    []byte `json:"d"`
}

func EntryOfFile(name string, data []byte, maxPartSize int) ([]*EntryPart, error) {
	t := trace.New().Source("entry.go", "Entry", "EntryOfFile")
	log.Println(trace.Debug("building entries").UTC().Add("maxPartSize", fmt.Sprintf("%d", maxPartSize)).Append(t))
	numPart := (len(data) / maxPartSize) + 1
	entries := make([]*EntryPart, 0, numPart)
	i := 0
	for i < len(data) {
		offset := i * maxPartSize
		e := EntryPart{Name: name, Hash: "fakehash", Mime: "image/png", IdxPart: i, NumPart: numPart, Data: data[offset:]}
		entries = append(entries, &e)
	}
	return entries, nil
}
