package ddb

import (
	"os"
	"path/filepath"
)

type TXCache struct {
	Path string
}

func NewTXCache(path string) error {
	os.MkdirAll(path, "700")
}