package ddb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type TXCache struct {
	path string
}

var ErrTXNotExist error = fmt.Errorf("TX not exists")

func NewTXCache(path string) (*TXCache, error) {
	tr := trace.New().Source("cache.go", "TXCache", "NewTXCache")
	trail.Println(trace.Debug("new TXCache").UTC().Add("path", path).Append(tr))
	err := os.MkdirAll(path, 0700)
	if err != nil {
		trail.Println(trace.Alert("error creating cache dir").UTC().Add("path", path).Error(err).Append(tr))
		return nil, fmt.Errorf("error creating cache dir '%s': %w", path, err)
	}
	cache := TXCache{path: path}
	return &cache, nil
}

func (c *TXCache) Path() string {
	return c.path
}

func (c *TXCache) Store(id string, tx []byte) error {
	tr := trace.New().Source("cache.go", "TXCache", "Store")
	trail.Println(trace.Debug("storing TX").UTC().Add("path", c.path).Add("id", id).Append(tr))
	txpath := path.Join(c.path, id)
	err := ioutil.WriteFile(txpath, tx, 0600)
	if err != nil {
		trail.Println(trace.Alert("error storing tx to cache").UTC().Add("path", c.path).Add("id", id).Error(err).Append(tr))
		return fmt.Errorf("error storing tx '%s' to cache dir '%s': %w", id, c.path, err)
	}
	return nil
}

func (c *TXCache) Retrieve(id string) ([]byte, error) {
	tr := trace.New().Source("cache.go", "TXCache", "Retrieve")
	trail.Println(trace.Debug("retrieving TX").UTC().Add("id", id).Append(tr))
	txpath := path.Join(c.path, id)
	tx, err := ioutil.ReadFile(txpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			trail.Println(trace.Alert("tx not in cache").UTC().Add("path", c.path).Add("id", id).Error(err).Append(tr))
			return nil, ErrTXNotExist
		}
		trail.Println(trace.Alert("error retrieving tx from cache").UTC().Add("path", c.path).Add("id", id).Error(err).Append(tr))
		return nil, fmt.Errorf("error retrieving tx '%s' from cache dir '%s': %w", id, c.path, err)
	}
	return tx, nil
}
