package ddb

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type FBranch struct {
	BitcoinWIF string
	BitcoinAdd string
	Password   [32]byte
	SpendLimit Satoshi
	Blockchain *Blockchain
}

type BResult struct {
	Cost  Satoshi
	TXIDs []string
}

func (fb *FBranch) EncodingPassword() string {
	return string(fb.Password[:])
}

//CastEntry store the entry on the blockchain. This method is the concatenation of ProcessEntry and Submit. Returns the TXID of the transactions generated.
func (fb *FBranch) CastEntry(entry *Entry, spendLimit Satoshi, simulate bool) (*BResult, error) {
	tr := trace.New().Source("fbranch.go", "FBranch", "CastEntry")
	trail.Println(trace.Info("casting entry to the blockcchain").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	txs, err := fb.ProcessEntry(entry, simulate)
	if err != nil {
		trail.Println(trace.Alert("error during TXs preparation").UTC().Add("file", entry.Name).Error(err).Append(tr))
		return nil, fmt.Errorf("error during TXs preparation: %w", err)
	}
	fee := Satoshi(0)
	for _, tx := range txs {
		f, err := tx.Fee()
		if err != nil {
			trail.Println(trace.Alert("error getting fee from DataTXs").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error getting fee DataTXs: %w", err)
		}
		fee = fee.Add(f)
	}
	if fee > spendLimit {
		trail.Println(trace.Alert("total cost of transaction exceeds the spending limit").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("total cost of transaction exceeds the spending limit: %w", err)

	}
	if simulate {
		trail.Println(trace.Info("simulation mode is on").UTC().Error(err).Append(tr))
		res := BResult{Cost: fee}
		return &res, nil
	}
	ids, err := fb.Blockchain.Submit(txs)
	if err != nil {
		trail.Println(trace.Alert("error while sending transactions").UTC().Add("file", entry.Name).Error(err).Append(tr))
		return nil, fmt.Errorf("error while sending transactions: %w", err)
	}
	res := BResult{TXIDs: ids}
	return &res, nil
}

//ProcessEntry prepares all the TXs required to store the entry on the blockchain.
func (fb *FBranch) ProcessEntry(entry *Entry, simulate bool) ([]*DataTX, error) {
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

//PackEncryptedEntriesPart writes each []data on a single TX chained with the others, returns the TXIDs and the hex encoded TXs
func (fb *FBranch) packEntryParts(version string, parts []*EntryPart, utxos []*UTXO, dataFee *Fee) ([]*DataTX, error) {
	tr := trace.New().Source("fbranch.go", "", "packEntryParts")
	trail.Println(trace.Info("packing bytes in an array of DataTX").UTC().Append(tr))
	dataTXs := make([]*DataTX, len(parts))
	for i, ep := range parts {
		encbytes, err := ep.Encrypt(fb.Password)
		if err != nil {
			trail.Println(trace.Alert("error while encrypting entry part").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error while encrypting entry part: %w", err)
		}
		// tempTx, err := NewDataTX(fb.BitcoinWIF, fb.BitcoinAdd, utxos, Bitcoin(0), encbytes, version)
		// if err != nil {
		// 	trail.Println(trace.Alert("cannot build 0 fee DataTX").UTC().Error(err).Append(tr))
		// 	return nil, fmt.Errorf("cannot build 0 fee DataTX: %w", err)
		// }
		fee := dataFee.CalculateFee(len(encbytes) + 150) //150 bytes for header
		dataTx, err := NewDataTX(fb.BitcoinWIF, fb.BitcoinAdd, fb.BitcoinAdd, utxos, fee, Bitcoin(-1), encbytes, version)
		if err != nil {
			trail.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("cannot build TX: %w", err)
		}
		trail.Println(trace.Info("got estimated fee").UTC().Add("fee", fmt.Sprintf("%0.8f", fee.Bitcoin())).Add("txid", dataTx.GetTxID()).Append(tr))
		//UTXO in TX built by BuildDataTX is in position 0
		inPos := 0
		utxos = []*UTXO{{TXPos: 0, TXHash: dataTx.GetTxID(), Value: Satoshi(dataTx.Outputs[inPos].Satoshis).Bitcoin(), ScriptPubKeyHex: dataTx.Outputs[inPos].GetLockingScriptHexString()}}
		dataTXs[i] = dataTx
	}
	return dataTXs, nil
}

//GetEntriesFromTXID retrieve all the Entries fully contained in the transactions with the given IDs.
func (fb *FBranch) GetEntriesFromTXID(txids []string, cacheOnly bool) ([]*Entry, error) {
	tr := trace.New().Source("fbranch.go", "FBranch", "RetrievingEntries")
	trail.Println(trace.Info("retrieving TXs from the blockchain and extracting entries").Add("len txids", fmt.Sprintf("%d", len(txids))).UTC().Append(tr))
	txs, err := fb.Blockchain.GetTXs(txids, cacheOnly)
	if err != nil {
		trail.Println(trace.Alert("error retrieving DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error retrieving DataTXs: %w", err)
	}
	parts, err := fb.unpackEntryParts(txs)
	if err != nil {
		trail.Println(trace.Warning("error while unpacking entry parts").UTC().Error(err).Append(tr))
	}
	entries, err := EntriesFromParts(parts)
	if err != nil {
		trail.Println(trace.Warning("error while reassembling Entries").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error while reassembling Entries: %w", err)
	}
	return entries, nil
}

//DownloadAll saves fb.cally all the files connected to the address. Return the number of entries saved.
func (fb *FBranch) DowloadAll(outPath string, cacheOnly bool) (int, error) {
	tr := trace.New().Source("fbranch.go", "FBranch", "DownloadAll")
	trail.Println(trace.Info("download all entries fb.cally").Add("outpath", outPath).UTC().Append(tr))
	history, err := fb.Blockchain.ListTXIDs(fb.BitcoinAdd, cacheOnly)
	if err != nil {
		trail.Println(trace.Alert("error getting address history").UTC().Error(err).Append(tr))
		return 0, fmt.Errorf("error getting address history: %w", err)
	}
	entries, err := fb.GetEntriesFromTXID(history, cacheOnly)
	if err != nil {
		trail.Println(trace.Alert("error retrieving entries").UTC().Error(err).Append(tr))
		return 0, fmt.Errorf("error retrieving entries: %w", err)
	}
	for _, e := range entries {
		ioutil.WriteFile(filepath.Join(outPath, e.Name), e.Data, 0444)
	}
	return len(entries), nil

}

//UnpackEntryParts builds EntryPart from the OP_RETURN encrypted data of the given transactions
func (fb *FBranch) unpackEntryParts(txs []*DataTX) ([]*EntryPart, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "UnpackEncryptedEntriesPart")
	trail.Println(trace.Info("opening TXs").UTC().Append(tr))
	parts := make([]*EntryPart, 0, len(txs))
	for _, tx := range txs {
		opr, ver, err := tx.Data()
		// trail.Println(trace.Info("DataTX version").Add("version", ver).UTC().Error(err).Append(tr))
		if err != nil {
			trail.Println(trace.Alert("error while getting OpReturn data from DataTX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error while getting OpReturn data from DataTX ver%s: %w", ver, err)
		}
		ep, err := EntryPartFromEncrypted(fb.Password, opr)
		if err != nil {
			trail.Println(trace.Alert("error while making entry part from encrypted bytes").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error while making entry part from encrypted bytes: %w", err)
		}
		parts = append(parts, ep)
	}
	return parts, nil
}
