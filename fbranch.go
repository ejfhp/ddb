package ddb

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ejfhp/ddb/satoshi"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type FBranch struct {
	BitcoinWIF string
	BitcoinAdd string
	Password   [32]byte
	Blockchain *Blockchain
}

type BResult struct {
	Cost  satoshi.Satoshi
	TXIDs []string
}

func (fb *FBranch) EncodingPassword() string {
	return string(fb.Password[:])
}

//ProcessEntry prepares all the TXs required to store the entry on the blockchain.
func (fb *FBranch) ProcessEntry(entry *Entry, utxo []*UTXO, header string) ([]*DataTX, error) {
	tr := trace.New().Source("fbranch.go", "FBranch", "ProcessEntry")
	trail.Println(trace.Info("preparing file").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	// entryParts, err := fb.EncryptEntry(entry)
	entryParts, err := entry.ToParts(fb.Password, fb.Blockchain.miner.MaxOpReturn())
	if err != nil {
		trail.Println(trace.Alert("error making parts of entry").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error making parts of entry: %w", err)
	}
	txs, err := fb.packEntryParts(header, entryParts, utxo)
	if err != nil {
		trail.Println(trace.Alert("error packing encrypted parts into DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error packing encrypted parts into DataTXs: %w", err)
	}
	return txs, nil
}

func (fb *FBranch) EstimateEntryFee(header string, entry *Entry) (satoshi.Satoshi, error) {
	tr := trace.New().Source("fbranch.go", "FBranch", "EstimateEntryFee")
	entryParts, err := entry.ToParts(fb.Password, fb.Blockchain.miner.MaxOpReturn())
	if err != nil {
		trail.Println(trace.Alert("error making parts of entry").UTC().Error(err).Append(tr))
		return 0, fmt.Errorf("error making parts of entry: %w", err)
	}
	utxo := fb.Blockchain.GetFakeUTXO()
	txs, err := fb.packEntryParts(header, entryParts, utxo)
	if err != nil {
		trail.Println(trace.Alert("error packing encrypted parts into DataTXs").UTC().Error(err).Append(tr))
		return 0, fmt.Errorf("error packing encrypted parts into DataTXs: %w", err)
	}
	fee := satoshi.Satoshi(0)
	for _, t := range txs {
		_, _, f, err := t.TotInOutFee()
		if err != nil {
			trail.Println(trace.Alert("error getting fee of TX").UTC().Error(err).Append(tr))
			return 0, fmt.Errorf("error getting fee of TX: %w", err)
		}
		fee = fee.Add(f)
	}
	return fee, nil

}

//PackEncryptedEntriesPart writes each []data on a single TX chained with the others, returns the TXIDs and the hex encoded TXs
func (fb *FBranch) packEntryParts(header string, parts []*EntryPart, utxos []*UTXO) ([]*DataTX, error) {
	tr := trace.New().Source("fbranch.go", "FBranch", "packEntryParts")
	dataTXs := make([]*DataTX, len(parts))
	trail.Println(trace.Info("packing []EntryPart").UTC().Append(tr).Add("len parts", fmt.Sprintf("%d", len(parts))))
	for i, ep := range parts {
		encbytes, err := ep.Encrypt(fb.Password)
		if err != nil {
			trail.Println(trace.Alert("error while encrypting entry part").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error while encrypting entry part: %w", err)
		}
		fee, err := fb.Blockchain.EstimateDataTXFee(len(utxos), encbytes, header)
		if err != nil {
			trail.Println(trace.Alert("cannot calculate DataTX fee").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("cannot calculate DataTX fee: %w", err)
		}
		dataTx, err := NewDataTX(fb.BitcoinWIF, fb.BitcoinAdd, fb.BitcoinAdd, utxos, satoshi.EmptyWallet, fee, encbytes, header)
		if err != nil {
			trail.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("cannot build TX: %w", err)
		}
		trail.Println(trace.Info("DataTX built").UTC().Add("fee", fmt.Sprintf("%0.8f", fee.Bitcoin())).Add("txid", dataTx.GetTxID()).Append(tr))
		//UTXO in TX built by BuildDataTX is in position 0
		inPos := 0
		utxos = []*UTXO{{TXPos: uint32(inPos), TXHash: dataTx.GetTxID(), Value: satoshi.Satoshi(dataTx.Outputs[inPos].Satoshis).Bitcoin(), ScriptPubKeyHex: dataTx.Outputs[inPos].GetLockingScriptHexString()}}
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
	trail.Println(trace.Info("download all entries").Add("outpath", outPath).UTC().Append(tr))
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
	tr := trace.New().Source("fbranch.go", "FBranch", "unpackEntryParts")
	trail.Println(trace.Info("opening TXs").UTC().Append(tr))
	parts := make([]*EntryPart, 0, len(txs))
	for _, tx := range txs {
		opr, header, err := tx.Data()
		if err != nil {
			trail.Println(trace.Warning("error while getting OpReturn data from DataTX, probably is not a TRH transaction").Append(tr).UTC().Add("TXID", tx.GetTxID()).Error(err))
			continue
		}
		ep, err := EntryPartFromEncrypted(fb.Password, opr)
		if err != nil {
			trail.Println(trace.Warning("error while exctracting entry part from encrypted bytes, probably the encrypting password was different").Append(tr).UTC().Add("header", header).Add("TXID", tx.GetTxID()).Error(err))
			continue
		}
		parts = append(parts, ep)
	}
	return parts, nil
}
