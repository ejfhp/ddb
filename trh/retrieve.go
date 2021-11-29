package trh

// func (t *TRH) RetrieveEntry(address string, password [32]byte, hash string) (map[string][]*ddb.MetaEntry, error) {
// 	fb := ddb.FBranch{BitcoinAdd: address, Password: password, Blockchain: t.blockchain}
// 	fb.DowloadAll()
// 	var mEntries map[string][]*ddb.MetaEntry
// 	btrunk := &ddb.BTrunk{MainKey: keystore.Source().Key, MainAddress: keystore.Source().Address, Blockchain: blockchain}
// 	passmap := map[string][32]byte{password: keystore.Password(password)}

// 	mEntries, err = btrunk.ListEntries(passmap, false)
// 	if err != nil {
// 		return nil, fmt.Errorf("error while listing MetaEntries: %w", err)
// 	}
// 	return mEntries, nil

// }

// func (cr *Retrieve) RetrieveAll() (int, error) {
// 	tr := trace.New().Source("retrieve.go", "Retrieve", "cmd")
// 	n, err := cr.diary.DowloadAll(flagOutputDir)
// 	if err != nil {
// 		trail.Println(trace.Info("error while downloadingAll").Append(tr).UTC().Add("address", cr.diary.BitcoinPublicAddress()).Add("ourFolder", cr.outfolder).Error(err))
// 		return 0, fmt.Errorf("error while downloadingAll files from address '%s' to folder '%s': %w", cr.diary.BitcoinPublicAddress(), cr.outfolder, err)
// 	}
// 	return n, nil
// }
