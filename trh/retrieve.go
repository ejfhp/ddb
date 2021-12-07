package trh

import (
	"fmt"
)

func (t *TRH) ExportFile(hash string, outfolder string, cacheOnly bool) error {
	node, err := t.keystore.GetNode(hash)
	if err != nil {
		return err
	}
	err = t.btrunk.ExportFile(node, hash, outfolder, cacheOnly)
	if err != nil {
		return fmt.Errorf("error while exporting file: %w", err)
	}
	return nil
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
