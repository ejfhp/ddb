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

func (bt *BTrunk) FBranch(password string) (*FBranch, error) {
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
	fb := FBranch{BitcoinAdd: branchAdd, CryptoKey: pass, Blockchain: bt.Blockchain}
	if bt.Dry {
		fb.Dry = true
	} else {
		fb.BitcoinWIF = branchWIF
		fb.Dry = false
	}
	return &fb, nil
}

func (bt *BTrunk) SameKeyFBranch(password string) *FBranch {
	tr := trace.New().Source("btrunk.go", "BTrunk", "SameKeyFBranch")
	trail.Println(trace.Debug("generating FBranch with the same key").UTC().Append(tr))
	pass := [32]byte{}
	copy(pass[:], []byte(password))
	fb := FBranch{BitcoinAdd: bt.BitcoinAdd, CryptoKey: pass, Blockchain: bt.Blockchain, Dry: bt.Dry, BitcoinWIF: bt.BitcoinWIF}
	return &fb
}

//BranchEntry generate a branch and store the entry on the blockchain
func (bt *BTrunk) BranchEntry(entry *Entry, simulate bool) (*Results, error) {
	return nil, nil
}

//CastEntry store the entry on the blockchain. This method is the concatenation of ProcessEntry and Submit. Returns the TXID of the transactions generated.
func (bt *BTrunk) CastEntry(entry *Entry) ([]string, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "CastEntry")
	trail.Println(trace.Info("casting entry to the blockcchain").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	txs, err := bt.ProcessEntry(entry)
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

//Submit push the transactions to the blockchain, returns the TXID of the transactions sent.
func (bt *BTrunk) Submit(txs []*DataTX) ([]string, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "Submit")
	trail.Println(trace.Info("submitting transactions to the blockcchain").Add("num txs", fmt.Sprintf("%d", len(txs))).UTC().Append(tr))
	ids, err := bt.Blockchain.Submit(txs)
	if err != nil {
		trail.Println(trace.Alert("error while sending transactions").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error while sending transactions: %w", err)
	}
	return ids, nil

}

//ProcessEntry prepares all the TXs required to store the entry on the blockchain.
func (bt *BTrunk) ProcessEntry(entry *Entry) ([]*DataTX, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "ProcessEntry")
	trail.Println(trace.Info("preparing file").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	encryptedEntryParts, err := bt.EncryptEntry(entry)
	if err != nil {
		trail.Println(trace.Alert("error encrypting entries of file").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error encrypting entries of file: %w", err)
	}
	utxo, err := bt.Blockchain.GetUTXO(bt.BitcoinAdd)
	if err != nil {
		trail.Println(trace.Alert("error getting UTXO").UTC().Add("address", bt.BitcoinAdd).Error(err).Append(tr))
		return nil, fmt.Errorf("error getting UTXO for address %s: %w", bt.BitcoinAdd, err)
	}
	fee, err := bt.Blockchain.GetDataFee()
	if err != nil {
		trail.Println(trace.Alert("error miner data fee").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error getting miner data fee: %w", err)
	}
	txs, err := PackData(VER_AES, bt.BitcoinWIF, encryptedEntryParts, utxo, fee)
	if err != nil {
		trail.Println(trace.Alert("error packing encrypted parts into DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error packing encrrypted parts into DataTXs: %w", err)
	}
	return txs, nil
}

//ProcessEntry prepares all the TXs required to store the entry on the blockchain.
func (bt *BTrunk) EncryptEntry(entry *Entry) ([][]byte, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "EncryptEntry")
	trail.Println(trace.Info("getting parts").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	parts, err := entry.Parts(bt.Blockchain.MaxDataSize())
	if err != nil {
		trail.Println(trace.Alert("error generating parts of file").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error generating parts of file: %w", err)
	}
	encryptedEntryParts := make([][]byte, 0, len(parts))
	for _, p := range parts {
		encodedp, err := p.Encode()
		if err != nil {
			trail.Println(trace.Alert("error encrypting entry part").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error encrypting entry part: %w", err)
		}
		cryptedp, err := AESEncrypt([32]byte{}, encodedp)
		if err != nil {
			trail.Println(trace.Alert("error encrypting entry part").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error encrypting entry part: %w", err)
		}
		encryptedEntryParts = append(encryptedEntryParts, cryptedp)
	}
	return encryptedEntryParts, nil
}

//RetrieveTXs retrieves the TXs with the given IDs.
func (bt *BTrunk) RetrieveTXs(txids []string) ([]*DataTX, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "RetrieveTXs")
	trail.Println(trace.Info("retrieving TXs from the blockchain").Add("len txids", fmt.Sprintf("%d", len(txids))).UTC().Append(tr))
	txs := make([]*DataTX, 0, len(txids))
	for _, txid := range txids {
		tx, err := bt.Blockchain.GetTX(txid)
		if err != nil {
			trail.Println(trace.Alert("error getting DataTX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error getting DataTX: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

//RetrieveAndExtractEntries retrieve the TX with the given IDs and extracts all the Entries fully contained.
func (bt *BTrunk) RetrieveAndExtractEntries(txids []string) ([]*Entry, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "RetrievingEntries")
	trail.Println(trace.Info("retrieving TXs from the blockchain and extracting entries").Add("len txids", fmt.Sprintf("%d", len(txids))).UTC().Append(tr))
	txs, err := bt.RetrieveTXs(txids)
	if err != nil {
		trail.Println(trace.Alert("error retrieving DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error retrieving DataTXs: %w", err)
	}
	entries, err := bt.ExtractEntries(txs)
	if err != nil {
		trail.Println(trace.Alert("error extracting Entries").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error extracting Entries: %w", err)
	}
	return entries, nil
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

//ExtractEntries rebuild all the Entries fully contained in the given TXs array.
func (bt *BTrunk) ExtractEntries(txs []*DataTX) ([]*Entry, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "ExtractEntries")
	trail.Println(trace.Info("reading entries from TXs").Add("len txs", fmt.Sprintf("%d", len(txs))).UTC().Append(tr))
	crypts, err := UnpackData(txs)
	if err != nil {
		trail.Println(trace.Alert("error unpacking data").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error unpacking data: %w", err)
	}
	entries, err := bt.DecryptEntries(crypts)
	if err != nil {
		trail.Println(trace.Warning("error decrypting EntryPart").UTC().Error(err).Append(tr))
	}
	return entries, nil
}

//DecryptEntries decrypts entries from the encrypted OP_RETURN data.
func (bt *BTrunk) DecryptEntries(crypts [][]byte) ([]*Entry, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "DecryptEntries")
	trail.Println(trace.Info("decrypting data").UTC().Append(tr))
	parts := make([]*EntryPart, 0, len(crypts))
	for _, cry := range crypts {
		enco, err := AESDecrypt([32]byte{}, cry)
		if err != nil {
			trail.Println(trace.Warning("error decrypting OP_RETURN data").UTC().Error(err).Append(tr))
			// return nil, fmt.Errorf("error decrypting OP_RETURN data: %w", err)
			continue
		}
		part, err := EntryPartFromEncodedData(enco)
		if err != nil {
			trail.Println(trace.Warning("error decoding EntryPart").UTC().Error(err).Append(tr))
		}
		parts = append(parts, part)
	}
	entries, err := EntriesFromParts(parts)
	if err != nil {
		trail.Println(trace.Warning("error while reassembling Entries").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error while reassembling Entries: %w", err)
	}
	return entries, nil
}

func (bt *BTrunk) ListHistory(address string) ([]string, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "ListHistory")
	trail.Println(trace.Info("listing TX history").UTC().Add("address", address).Append(tr))
	txids, err := bt.Blockchain.ListTXIDs(address)
	if err != nil {
		trail.Println(trace.Warning("error while listing TX history").UTC().Add("address", address).Error(err).Append(tr))
		return nil, fmt.Errorf("error while listing TX history: %w", err)
	}
	return txids, nil
}
