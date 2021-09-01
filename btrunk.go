package ddb

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"path/filepath"

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
	Dry        bool
}

func NewBTrunk(wif string, blockchain *Blockchain) (*BTrunk, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "NewBTrunk")
	address, err := AddressOf(wif)
	if err != nil {
		trail.Println(trace.Alert("cannot get address of key").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get address of key: %w", err)
	}
	d := BTrunk{BitcoinWIF: wif, BitcoinAdd: address, Blockchain: blockchain}
	trail.Println(trace.Debug("new BTrunk built").UTC().Append(tr).Add("wif", wif).Add("address", address))
	return &d, nil
}

func NewDryBTrunk(address string, blockchain *Blockchain) (*BTrunk, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "NewDryBTrunk")
	d := BTrunk{BitcoinWIF: "", BitcoinAdd: address, Blockchain: blockchain}
	trail.Println(trace.Debug("new dry BTrunk").UTC().Append(tr).Add("address", address))
	return &d, nil
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
	fb := FBranch{BitcoinAdd: branchAdd, Password: pass, Blockchain: bt.Blockchain}
	if bt.Dry {
		fb.Dry = true
	} else {
		fb.BitcoinWIF = branchWIF
		fb.Dry = false
	}
	return &fb, nil
}

func (bt *BTrunk) newSameKeyFBranch(password string) *FBranch {
	tr := trace.New().Source("btrunk.go", "BTrunk", "SameKeyFBranch")
	trail.Println(trace.Debug("generating FBranch with the same key").UTC().Append(tr))
	pass := [32]byte{}
	copy(pass[:], []byte(password))
	fb := FBranch{BitcoinAdd: bt.BitcoinAdd, Password: pass, Blockchain: bt.Blockchain, Dry: bt.Dry, BitcoinWIF: bt.BitcoinWIF}
	return &fb
}

//BranchEntry store the entry on the blockchain encrypted with the given password.
func (bt *BTrunk) BranchEntry(entry *Entry, password string, sameKey bool, simulate bool) (*Results, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "BranchEntry")
	trail.Println(trace.Info("casting entry to the blockcchain").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	var fBranch *FBranch
	var err error
	if sameKey {
		fBranch = bt.newSameKeyFBranch(password)
	} else {
		fBranch, err = bt.newFBranch(password)
		if err != nil {
			trail.Println(trace.Alert("error while creating a new branch").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error while creating a new branch: %w", err)
		}
	}
	mentry := NewMetaEntry(entry)
	res, err := bt.castMetaEntry(mentry, simulate)
	fmt.Println(res)
	return nil, nil
}

//CastEntry store the entry on the blockchain. Returns the TXID of the transactions generated.
func (bt *BTrunk) castMetaEntry(mentry *MetaEntry, fbranch FBranch, simulate bool) (*Results, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "CastEntry")
	trail.Println(trace.Info("casting meta entry to the blockcchain").Add("file", mentry.Name).UTC().Append(tr))
	encData, err := mentry.Encrypt(fbranch.Password)

	txs, err := bt.ProcessMetaEntry(entry)
	if err != nil {
		trail.Println(trace.Alert("error during TXs preparation").UTC().Add("file", entry.Name).Error(err).Append(tr))
		return nil, fmt.Errorf("error during TXs preparation: %w", err)
	}
	ids, err := bt.Submit(txs)
	if err != nil {
		trail.Println(trace.Alert("error while sending transactions").UTC().Add("file", entry.Name).Error(err).Append(tr))
		return nil, fmt.Errorf("error while sending transactions: %w", err)
	}
	return ids, nil
}

func (fb *BTrunk) ProcessEntry(entry *Entry, simulate bool) ([]*DataTX, error) {
	tr := trace.New().Source("fbranch.go", "FBranch", "ProcessEntry")
	trail.Println(trace.Info("preparing file").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	// entryParts, err := fb.EncryptEntry(entry)
	entryParts, err := entry.ToParts(fb.Password, fb.Blockchain.miner.MaxOpReturn())
	if err != nil {
		trail.Println(trace.Alert("error making parts of entry").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error making parts of entry: %w", err)
	}
	var utxo []*UTXO
	if simulate {
		utxo = fb.Blockchain.GetFakeUTXO()
	} else {
		utxo, err = fb.Blockchain.GetUTXO(fb.BitcoinAdd)
		if err != nil {
			trail.Println(trace.Alert("error getting UTXO").UTC().Add("address", fb.BitcoinAdd).Error(err).Append(tr))
			return nil, fmt.Errorf("error getting UTXO for address %s: %w", fb.BitcoinAdd, err)
		}
	}
	fee, err := fb.Blockchain.GetDataFee()
	if err != nil {
		trail.Println(trace.Alert("error miner data fee").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error getting miner data fee: %w", err)
	}
	txs, err := fb.packEntryParts(VER_AES, entryParts, utxo, fee)
	if err != nil {
		trail.Println(trace.Alert("error packing encrypted parts into DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error packing encrrypted parts into DataTXs: %w", err)
	}
	return txs, nil
}

//EstimateFee returns an estimation in satoshi of the fee necessary to store the entry on the blockchain.
func (bt *BTrunk) EstimateFee(entry *Entry) (Satoshi, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "EstimateFee")
	trail.Println(trace.Info("estimate fee to cast entry").UTC().Append(tr).Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))))
	encryptedEntryParts, err := bt.EncryptEntry(entry)
	if err != nil {
		trail.Println(trace.Alert("error encrypting entries of file").UTC().Error(err).Append(tr))
		return Satoshi(0), fmt.Errorf("error encrypting entries of file: %w", err)
	}
	fakeUtxos := []*UTXO{{TXPos: 0, TXHash: "72124e293287ab0ca20a723edb61b58d6ef89aba05508b92198bd948bfb6da40", Value: 100000000000, ScriptPubKeyHex: "76a914330a97979931a961d1e5f05d3c7ace4217fc7adc88ac"}}
	dataFee, err := bt.Blockchain.GetDataFee()
	if err != nil {
		trail.Println(trace.Alert("error getting miner data fee").UTC().Error(err).Append(tr))
		return Satoshi(0), fmt.Errorf("error getting miner data fee: %w", err)
	}
	txs, err := PackData(VER_AES, bt.BitcoinWIF, encryptedEntryParts, fakeUtxos, dataFee)
	if err != nil {
		trail.Println(trace.Alert("error packing encrypted parts into DataTXs").UTC().Error(err).Append(tr))
		return Satoshi(0), fmt.Errorf("error packing encrrypted parts into DataTXs: %w", err)
	}
	fee := Satoshi(0)
	for _, tx := range txs {
		f, err := tx.Fee()
		if err != nil {
			trail.Println(trace.Alert("error getting fee from DataTXs").UTC().Error(err).Append(tr))
			return Satoshi(0), fmt.Errorf("error getting fee DataTXs: %w", err)
		}
		fee = fee.Add(f)
	}
	return fee, nil
}

//DownloadAll saves locally all the files connected to the address. Return the number of entries saved.
func (bt *BTrunk) DowloadAll(outPath string) (int, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "DownloadAll")
	trail.Println(trace.Info("download all entries locally").Add("outpath", outPath).UTC().Append(tr))
	history, err := bt.ListHistory(bt.BitcoinAdd)
	if err != nil {
		trail.Println(trace.Alert("error getting address history").UTC().Error(err).Append(tr))
		return 0, fmt.Errorf("error getting address history: %w", err)
	}
	entries, err := bt.RetrieveAndExtractEntries(history)
	if err != nil {
		trail.Println(trace.Alert("error retrieving entries").UTC().Error(err).Append(tr))
		return 0, fmt.Errorf("error retrieving entries: %w", err)
	}
	for _, e := range entries {
		ioutil.WriteFile(filepath.Join(outPath, e.Name), e.Data, 0444)
	}
	return len(entries), nil

}
