package ddb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type TXCache struct {
	path string
}

type AddressInfo struct {
	Address      string          `json:"address"`
	SourceOutput []*SourceOutput `json:"outputs"`
	TXIDs        []string        `json:"txids"`
}

var ErrNotCached error = fmt.Errorf("entry not in cache")

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

func (c *TXCache) DirPath() string {
	return c.path
}

func (c *TXCache) Path(base string) string {
	return path.Join(c.path, base+".trh")
}

func (c *TXCache) StoreTX(id string, tx []byte) error {
	tr := trace.New().Source("cache.go", "TXCache", "Store")
	trail.Println(trace.Debug("storing TX").UTC().Add("path", c.path).Add("id", id).Append(tr))
	txpath := c.Path(id)
	err := ioutil.WriteFile(txpath, tx, 0600)
	if err != nil {
		trail.Println(trace.Alert("error storing tx to cache").UTC().Add("path", c.path).Add("id", id).Error(err).Append(tr))
		return fmt.Errorf("error storing tx '%s' to cache dir '%s': %w", id, c.path, err)
	}
	return nil
}

func (c *TXCache) RetrieveTX(id string) ([]byte, error) {
	tr := trace.New().Source("cache.go", "TXCache", "Retrieve")
	trail.Println(trace.Debug("retrieving TX").UTC().Add("id", id).Append(tr))
	txpath := c.Path(id)
	tx, err := ioutil.ReadFile(txpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			trail.Println(trace.Alert("tx not in cache").UTC().Add("path", c.path).Add("id", id).Error(err).Append(tr))
			return nil, ErrNotCached
		}
		trail.Println(trace.Alert("error retrieving tx from cache").UTC().Add("path", c.path).Add("id", id).Error(err).Append(tr))
		return nil, fmt.Errorf("error retrieving tx '%s' from cache dir '%s': %w", id, c.path, err)
	}
	return tx, nil
}

func (c *TXCache) StoreSourceOutput(address string, sourceOutput *SourceOutput) error {
	tr := trace.New().Source("cache.go", "TXCache", "StoreSourceOutput")
	trail.Println(trace.Debug("storing SourceOutput").UTC().Add("path", c.path).Add("address", address).Append(tr))
	addinfo, err := c.retrieveAddressInfo(address)
	if err != nil {
		if err == ErrNotCached {
			addinfo = &AddressInfo{SourceOutput: []*SourceOutput{}, TXIDs: []string{}, Address: address}
		} else {
			trail.Println(trace.Alert("error storing sourceoutput to cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
			return fmt.Errorf("error storing sourceoutput of address '%s' to cache dir '%s': %w", address, c.path, err)
		}
	}
	for _, so := range addinfo.SourceOutput {
		if so.Equals(sourceOutput) {
			return nil
		}
	}
	addinfo.SourceOutput = append(addinfo.SourceOutput, sourceOutput)
	err = c.storeAddressInfo(address, addinfo)
	if err != nil {
		trail.Println(trace.Alert("error storing source output to cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
		return fmt.Errorf("error storing source output for address '%s' to cache dir '%s': %w", address, c.path, err)
	}
	return nil
}

func (c *TXCache) RetrieveSourceOutput(address string) ([]*SourceOutput, error) {
	tr := trace.New().Source("cache.go", "TXCache", "RetrieveSourceOutput")
	trail.Println(trace.Debug("retrieving SourceOutput").UTC().Add("path", c.path).Add("address", address).Append(tr))
	addinfo, err := c.retrieveAddressInfo(address)
	if err != nil {
		if err == ErrNotCached {
			trail.Println(trace.Alert("address not in cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
			return nil, err
		} else {
			trail.Println(trace.Alert("error retrieving address from cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
			return nil, fmt.Errorf("error retrieving address '%s' from cache dir '%s': %w", address, c.path, err)
		}
	}
	if addinfo.SourceOutput == nil || len(addinfo.SourceOutput) == 0 {
		trail.Println(trace.Alert("sourceoutput not found in cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
		return nil, ErrNotCached
	}
	return addinfo.SourceOutput, nil
}

func (c *TXCache) StoreTXID(address string, txid string) error {
	tr := trace.New().Source("cache.go", "TXCache", "StoreTXID")
	trail.Println(trace.Debug("storing TXID").UTC().Add("path", c.path).Add("address", address).Append(tr))
	addinfo, err := c.retrieveAddressInfo(address)
	if err != nil {
		if err == ErrNotCached {
			addinfo = &AddressInfo{SourceOutput: []*SourceOutput{}, TXIDs: []string{}, Address: address}
		} else {
			trail.Println(trace.Alert("error storing txid to cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
			return fmt.Errorf("error storing txid of address '%s' to cache dir '%s': %w", address, c.path, err)
		}
	}
	addinfo.TXIDs = append(addinfo.TXIDs, txid)
	err = c.storeAddressInfo(address, addinfo)
	if err != nil {
		trail.Println(trace.Alert("error storing txid to cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
		return fmt.Errorf("error storing txid for address '%s' to cache dir '%s': %w", address, c.path, err)
	}
	return nil
}

func (c *TXCache) RetrieveTXID(address string) ([]string, error) {
	tr := trace.New().Source("cache.go", "TXCache", "RetrieveTXID")
	trail.Println(trace.Debug("retrieving TXID").UTC().Add("path", c.path).Add("address", address).Append(tr))
	addinfo, err := c.retrieveAddressInfo(address)
	if err != nil {
		if err == ErrNotCached {
			trail.Println(trace.Alert("address not in cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
			return nil, err
		} else {
			trail.Println(trace.Alert("error retrieving address from cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
			return nil, fmt.Errorf("error retrieving address '%s' from cache dir '%s': %w", address, c.path, err)
		}
	}
	if addinfo.TXIDs == nil || len(addinfo.TXIDs) == 0 {
		trail.Println(trace.Alert("txid not found in cache").UTC().Add("path", c.path).Add("address", address).Error(err).Append(tr))
		return nil, ErrNotCached
	}
	return addinfo.TXIDs, nil
}

func (c *TXCache) retrieveAddressInfo(address string) (*AddressInfo, error) {
	txpath := c.Path(address)
	bytes, err := ioutil.ReadFile(txpath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotCached
		}
		return nil, fmt.Errorf("error retrieving sourceoutput of address '%s' from cache dir '%s': %w", address, c.path, err)
	}
	var addinfo AddressInfo
	json.Unmarshal(bytes, &addinfo)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling address info for '%s': %w", address, err)
	}
	return &addinfo, nil
}

func (c *TXCache) storeAddressInfo(address string, addressinfo *AddressInfo) error {
	aipath := c.Path(address)
	bytes, err := json.Marshal(addressinfo)
	if err != nil {
		return fmt.Errorf("error marshaling address info for '%s': %w", address, err)
	}
	err = ioutil.WriteFile(aipath, bytes, 0600)
	if err != nil {
		return fmt.Errorf("error storing address info for address '%s' to cache dir '%s': %w", address, c.path, err)
	}
	return nil
}

func (c *TXCache) Size() (int, error) {
	tr := trace.New().Source("cache.go", "TXCache", "Size")
	trail.Println(trace.Debug("getting cache cardinality").UTC().Add("dir", c.path).Append(tr))
	dir, err := os.Open(c.path)
	if err != nil {
		trail.Println(trace.Alert("error opening cache dir").UTC().Add("dir", c.path).Error(err).Append(tr))
		return -1, fmt.Errorf("error opening cache dir '%s': %w", c.path, err)
	}
	defer dir.Close()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		trail.Println(trace.Debug("error listing files in cache dir").UTC().Add("dir", c.path).Error(err).Append(tr))
		return -1, fmt.Errorf("error listiing files in cache dir '%s': %w", c.path, err)
	}
	return len(names), nil
}

func (c *TXCache) Clear() error {
	tr := trace.New().Source("cache.go", "TXCache", "Clear")
	trail.Println(trace.Debug("clearing cache").UTC().Add("dir", c.path).Append(tr))
	dir, err := os.Open(c.path)
	if err != nil {
		trail.Println(trace.Debug("error opening cache dir").UTC().Add("dir", c.path).Error(err).Append(tr))
		return fmt.Errorf("error opening cache dir '%s': %w", c.path, err)
	}
	defer func() {
		dir.Close()
		err := os.Remove(c.path)
		if err != nil {
			trail.Println(trace.Warning("error deleting cache dir").UTC().Add("dir", c.path).Error(err).Append(tr))
		}
	}()
	names, err := dir.Readdirnames(-1)
	if err != nil {
		trail.Println(trace.Debug("error listing files in cache dir").UTC().Add("dir", c.path).Error(err).Append(tr))
		return fmt.Errorf("error listiing files in cache dir '%s': %w", c.path, err)
	}
	for _, name := range names {
		if strings.HasSuffix(name, ".trh") {
			err = os.Remove(path.Join(c.path, name))
			if err != nil {
				trail.Println(trace.Warning("error deleting file in cache dir").UTC().Add("dir", c.path).Add("name", name).Error(err).Append(tr))
			}
		}
	}
	return nil
}
