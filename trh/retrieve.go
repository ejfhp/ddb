package trh

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ejfhp/ddb"
)

func (t *TRH) RetrieveFile(entryhash string, outFolder string, cacheOnly bool) (*ddb.Entry, error) {
	node, err := t.keystore.GetNode(entryhash)
	if err != nil {
		return nil, fmt.Errorf("error getting node of hash %s: %w", entryhash, err)
	}
	entry, err := t.btrunk.GetEntry(node, cacheOnly)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving entries file: %w", err)
	}
	err = ioutil.WriteFile(filepath.Join(outFolder, entry.Name), entry.Data, 0444)
	if err != nil {
		return nil, fmt.Errorf("error while saving entry: %w", err)
	}
	return entry, err
}

// func (cr *Retrieve) RetrieveAll() (int, error) {
// 	tr := trace.New().Source("retrieve.go", "Retrieve", "cmd")
// 	n, err := cr.diary.DowloadAll(flagOutputDir)
// 	if err != nil {
// 		trail.Println(trace.Info("error while downloadingAll").Append(tr).UTC().Add("address", cr.diary.BitcoinPublicAddress()).Add("ourFolder", cr.outfolder).Error(err))
// 		return 0, fmt.Errorf("error while downloadingAll files from address '%s' to folder '%s': %w", cr.diary.BitcoinPublicAddress(), cr.outfolder, err)
// 	}
// 	return n, nil
// }
