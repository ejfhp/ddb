package ddb

import (
	"crypto/sha256"
	"fmt"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Results struct {
	Cost  Satoshi
	TXIDs []string
}

type BTrunk struct {
	BitcoinWIF string
	BitcoinAdd string
	Blockchain *Blockchain
}

func (bt *BTrunk) getGreenBud(simulate bool) ([]*UTXO, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "getGreenBud")
	trail.Println(trace.Info("getting green bud UTXO of the btrunk").Append(tr).UTC())
	var utxo []*UTXO
	var err error
	if simulate {
		utxo = bt.Blockchain.GetFakeUTXO()
	} else {
		utxo, err = bt.Blockchain.GetUTXO(bt.BitcoinAdd)
		if err != nil {
			trail.Println(trace.Alert("error getting UTXO").UTC().Add("address", bt.BitcoinAdd).Error(err).Append(tr))
			return nil, fmt.Errorf("error getting UTXO for address %s: %w", bt.BitcoinAdd, err)
		}
	}
	return utxo, nil
}

func (bt *BTrunk) newFBranch(password string) (*FBranch, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "FBranch")
	trail.Println(trace.Debug("generating FBranch with a derived key").UTC().Append(tr))
	pass := [32]byte{}
	copy(pass[:], []byte(password))
	keySeed := []byte{}
	keySeed = append(keySeed, []byte(bt.BitcoinAdd)...)
	keySeed = append(keySeed, pass[:]...)
	keySeedHash := sha256.Sum256(keySeed)
	branchKey, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), keySeedHash[:])
	fbwif, err := bsvutil.NewWIF(branchKey, &chaincfg.MainNetParams, true)
	if err != nil {
		return nil, fmt.Errorf("error while generating FBranch WIF: %v", err)
	}
	branchWIF := fbwif.String()
	branchAdd, err := AddressOf(branchWIF)
	if err != nil {
		return nil, fmt.Errorf("error while generating FBranch address: %v", err)
	}
	trail.Println(trace.Debug("derived key generated").UTC().Append(tr).Add("address", branchAdd))
	fb := FBranch{BitcoinWIF: branchWIF, BitcoinAdd: branchAdd, Password: pass, Blockchain: bt.Blockchain}
	return &fb, nil
}

func (bt *BTrunk) newSameKeyFBranch(password string) *FBranch {
	tr := trace.New().Source("btrunk.go", "BTrunk", "SameKeyFBranch")
	trail.Println(trace.Debug("generating FBranch with the same key").UTC().Append(tr))
	pass := [32]byte{}
	copy(pass[:], []byte(password))
	fb := FBranch{BitcoinWIF: bt.BitcoinWIF, BitcoinAdd: bt.BitcoinAdd, Password: pass, Blockchain: bt.Blockchain}
	return &fb
}

func (bt *BTrunk) prepareFBranchBud(fBranch *FBranch, entry *Entry, spendingLimit Satoshi, simulate bool) (*UTXO, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "prepareBud")
	trail.Println(trace.Debug("generating UTXO bud for a new entry").UTC().Append(tr))

	// mentry := NewMetaEntry(entry)
	return nil, nil
}

//BranchEntry store the entry on the blockchain encrypted with the given password.
// func (bt *BTrunk) BranchEntry(entry *Entry, password string, sameKey bool, spendLimit Satoshi, simulate bool) (*Results, error) {
// 	tr := trace.New().Source("btrunk.go", "BTrunk", "BranchEntry")
// 	trail.Println(trace.Info("casting entry to the blockcchain").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
// 	var fBranch *FBranch
// 	var err error
// 	if sameKey {
// 		fBranch = bt.newSameKeyFBranch(password)
// 	} else {
// 		fBranch, err = bt.newFBranch(password)
// 		if err != nil {
// 			trail.Println(trace.Alert("error while creating a new branch").UTC().Error(err).Append(tr))
// 			return nil, fmt.Errorf("error while creating a new branch: %w", err)
// 		}
// 	}
// 	mentry := NewMetaEntry(entry)
// 	res, err := bt.castMetaEntry(mentry, fBranch, spendLimit, simulate)
// 	fmt.Println(res)
// 	return nil, nil
// }

// func (bt *BTrunk) castMetaEntry(mentry *MetaEntry, fBranch *FBranch, spendLimit Satoshi, simulate bool) (*BResult, error) {
// 	tr := trace.New().Source("btrunk.go", "BTrunk", "CastMetaEntry")
// 	trail.Println(trace.Info("casting entry to the blockcchain").Append(tr).Add("file", mentry.Name).UTC())
// 	tx, err := bt.processMetaEntry(mentry, fBranch, simulate)
// 	if err != nil {
// 		trail.Println(trace.Alert("error during TXs preparation").Append(tr).UTC().Add("file", mentry.Name).Error(err))
// 		return nil, fmt.Errorf("error during TXs preparation: %w", err)
// 	}
// 	fee, err := tx.Fee()
// 	if err != nil {
// 		trail.Println(trace.Alert("error getting fee from DataTXs").Append(tr).UTC().Error(err))
// 		return nil, fmt.Errorf("error getting fee DataTXs: %w", err)
// 	}
// 	if fee.Satoshi() > spendLimit {
// 		trail.Println(trace.Alert("total cost of transaction exceeds the spending limit").Append(tr).UTC())
// 		return nil, fmt.Errorf("total cost of transaction exceeds the spending limit: %w", err)

// 	}
// 	if simulate {
// 		trail.Println(trace.Info("simulation mode is on").UTC().Append(tr).Error(err).Append(tr))
// 		res := BResult{Cost: fee.Satoshi()}
// 		return &res, nil
// 	}
// 	ids, err := bt.Blockchain.Submit([]*DataTX{tx})
// 	if err != nil {
// 		trail.Println(trace.Alert("error while sending transactions").UTC().Append(tr).Add("file", mentry.Name).Error(err))
// 		return nil, fmt.Errorf("error while sending transactions: %w", err)
// 	}
// 	res := BResult{TXIDs: ids}
// 	return &res, nil
// }

// func (bt *BTrunk) processMetaEntry(mentry *MetaEntry, fBranch *FBranch, simulate bool) (*DataTX, error) {
// 	tr := trace.New().Source("btrunk.go", "BTrunk", "processMetaEntry")
// 	trail.Println(trace.Info("preparing meta entry").Add("file", mentry.Name).UTC().Append(tr))
// 	utxo, err := bt.getGreenBud(simulate)
// 	if err != nil {
// 		trail.Println(trace.Alert("error getting UTXO of BTrunk").UTC().Append(tr).Error(err))
// 		return nil, fmt.Errorf("error getting UTXO of BTrunk: %w", err)
// 	}
// 	fee, err := bt.Blockchain.GetDataFee()
// 	if err != nil {
// 		trail.Println(trace.Alert("error miner data fee").UTC().Error(err).Append(tr))
// 		return nil, fmt.Errorf("error getting miner data fee: %w", err)
// 	}
// 	tx, err := bt.packMetaEntry(VER_AES, mentry, fBranch, utxo, fee)
// 	if err != nil {
// 		trail.Println(trace.Alert("error packing encrypted parts into DataTXs").UTC().Error(err).Append(tr))
// 		return nil, fmt.Errorf("error packing encrrypted parts into DataTXs: %w", err)
// 	}
// 	return tx, nil
// }

func (bt *BTrunk) packMetaEntry(version string, mentry *MetaEntry, fBranch *FBranch, utxos []*UTXO, dataFee *Fee, amount Satoshi) (*DataTX, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "packMetaEntry")
	trail.Println(trace.Info("packing bytes in an array of DataTX").UTC().Append(tr))
	encbytes, err := mentry.Encrypt(fBranch.Password)
	if err != nil {
		trail.Println(trace.Alert("error while encrypting entry part").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error while encrypting entry part: %w", err)
	}
	tempTx, err := NewDataTX(bt.BitcoinWIF, fBranch.BitcoinAdd, bt.BitcoinAdd, utxos, Bitcoin(0), amount, encbytes, version)
	if err != nil {
		trail.Println(trace.Alert("cannot build 0 fee DataTX").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build 0 fee DataTX: %w", err)
	}
	fee, err := tempTx.Fee()
	if err != nil {
		trail.Println(trace.Alert("cannot get fee of DataTX").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get fee of DataTX: %w", err)
	}
	dataTX, err := NewDataTX(bt.BitcoinWIF, fBranch.BitcoinAdd, bt.BitcoinAdd, utxos, fee, amount, encbytes, version)
	if err != nil {
		trail.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build TX: %w", err)
	}
	trail.Println(trace.Info("got estimated fee").UTC().Add("fee", fmt.Sprintf("%0.8f", fee.Bitcoin())).Add("txid", dataTX.GetTxID()).Append(tr))
	//UTXO in TX built by BuildDataTX is in position 0
	inPos := 0
	utxos = []*UTXO{{TXPos: 0, TXHash: dataTX.GetTxID(), Value: Satoshi(dataTX.Outputs[inPos].Satoshis).Bitcoin(), ScriptPubKeyHex: dataTX.Outputs[inPos].GetLockingScriptHexString()}}
	return dataTX, nil
}

//DownloadAll saves locally all the files connected to the address. Return the number of entries saved.
// func (bt *BTrunk) DowloadAll(outPath string) (int, error) {
// 	tr := trace.New().Source("btrunk.go", "BTrunk", "DownloadAll")
// 	trail.Println(trace.Info("download all entries locally").Add("outpath", outPath).UTC().Append(tr))
// 	history, err := bt.ListHistory(bt.BitcoinAdd)
// 	if err != nil {
// 		trail.Println(trace.Alert("error getting address history").UTC().Error(err).Append(tr))
// 		return 0, fmt.Errorf("error getting address history: %w", err)
// 	}
// 	entries, err := bt.RetrieveAndExtractEntries(history)
// 	if err != nil {
// 		trail.Println(trace.Alert("error retrieving entries").UTC().Error(err).Append(tr))
// 		return 0, fmt.Errorf("error retrieving entries: %w", err)
// 	}
// 	for _, e := range entries {
// 		ioutil.WriteFile(filepath.Join(outPath, e.Name), e.Data, 0444)
// 	}
// 	return len(entries), nil

// }