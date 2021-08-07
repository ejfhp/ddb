package ddb

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Logbook struct {
	bitcoinWif string
	bitcoinAdd string
	cryptoKey  [32]byte
	blockchain *Blockchain
}

func NewLogbook(wif string, password [32]byte, blockchain *Blockchain) (*Logbook, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "NewLogbook")
	trail.Println(trace.Debug("new Logbook").UTC().Append(tr))
	address, err := AddressOf(wif)
	if err != nil {
		trail.Println(trace.Alert("cannot get address of key").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get address of key: %w", err)
	}
	return &Logbook{bitcoinWif: wif, bitcoinAdd: address, cryptoKey: password, blockchain: blockchain}, nil

}

func (l *Logbook) BitcoinPrivateKey() string {
	return l.bitcoinWif
}

func (l *Logbook) BitcoinPublicAddress() string {
	return l.bitcoinAdd
}

func (l *Logbook) EncodingPassword() string {
	return string(l.cryptoKey[:])
}

//CastEntry store the entry on the blockchain. This method is the concatenation of ProcessEntry and Submit. Returns the TXID of the transactions generated.
func (l *Logbook) CastEntry(entry *Entry) ([]string, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "CastEntry")
	trail.Println(trace.Info("casting entry to the blockcchain").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	txs, err := l.ProcessEntry(entry)
	if err != nil {
		trail.Println(trace.Alert("error during TXs preparation").UTC().Add("file", entry.Name).Error(err).Append(tr))
		return nil, fmt.Errorf("error during TXs preparation: %w", err)
	}
	ids, err := l.Submit(txs)
	if err != nil {
		trail.Println(trace.Alert("error while sending transactions").UTC().Add("file", entry.Name).Error(err).Append(tr))
		return nil, fmt.Errorf("error while sending transactions: %w", err)
	}
	return ids, nil

}

//Submit push the transactions to the blockchain, returns the TXID of the transactions sent.
func (l *Logbook) Submit(txs []*DataTX) ([]string, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "Submit")
	trail.Println(trace.Info("submitting transactions to the blockcchain").Add("num txs", fmt.Sprintf("%d", len(txs))).UTC().Append(tr))
	ids, err := l.blockchain.Submit(txs)
	if err != nil {
		trail.Println(trace.Alert("error while sending transactions").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error while sending transactions: %w", err)
	}
	return ids, nil

}

//ProcessEntry prepares all the TXs required to store the entry on the blockchain.
func (l *Logbook) ProcessEntry(entry *Entry) ([]*DataTX, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "ProcessEntry")
	trail.Println(trace.Info("preparing file").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	encryptedEntryParts, err := l.EncryptEntry(entry)
	if err != nil {
		trail.Println(trace.Alert("error encrypting entries of file").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error encrypting entries of file: %w", err)
	}
	txs, err := l.blockchain.PackData(VER_AES, l.bitcoinWif, encryptedEntryParts)
	if err != nil {
		trail.Println(trace.Alert("error packing encrypted parts into DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error packing encrrypted parts into DataTXs: %w", err)
	}
	return txs, nil
}

//ProcessEntry prepares all the TXs required to store the entry on the blockchain.
func (l *Logbook) EncryptEntry(entry *Entry) ([][]byte, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "EncryptEntry")
	trail.Println(trace.Info("getting parts").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	parts, err := entry.Parts(l.MaxDataSize())
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
		cryptedp, err := AESEncrypt(l.cryptoKey, encodedp)
		if err != nil {
			trail.Println(trace.Alert("error encrypting entry part").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error encrypting entry part: %w", err)
		}
		encryptedEntryParts = append(encryptedEntryParts, cryptedp)
	}
	return encryptedEntryParts, nil
}

//RetrieveTXs retrieves the TXs with the given IDs.
func (l *Logbook) RetrieveTXs(txids []string) ([]*DataTX, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "RetrieveTXs")
	trail.Println(trace.Info("retrieving TXs from the blockchain").Add("len txids", fmt.Sprintf("%d", len(txids))).UTC().Append(tr))
	txs := make([]*DataTX, 0, len(txids))
	for _, txid := range txids {
		tx, err := l.blockchain.GetTX(txid)
		if err != nil {
			trail.Println(trace.Alert("error getting DataTX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error getting DataTX: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

//RetrieveAndExtractEntries retrieve the TX with the given IDs and extracts all the Entries fully contained.
func (l *Logbook) RetrieveAndExtractEntries(txids []string) ([]*Entry, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "RetrievingEntries")
	trail.Println(trace.Info("retrieving TXs from the blockchain and extracting entries").Add("len txids", fmt.Sprintf("%d", len(txids))).UTC().Append(tr))
	txs, err := l.RetrieveTXs(txids)
	if err != nil {
		trail.Println(trace.Alert("error retrieving DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error retrieving DataTXs: %w", err)
	}
	entries, err := l.ExtractEntries(txs)
	if err != nil {
		trail.Println(trace.Alert("error extracting Entries").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error extracting Entries: %w", err)
	}
	return entries, nil
}

//DownloadAll saves locally all the files connected to the address. Return the number of entries saved.
func (l *Logbook) DowloadAll(outPath string) (int, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "DownloadAll")
	trail.Println(trace.Info("download all entries locally").Add("outpath", outPath).UTC().Append(tr))
	history, err := l.blockchain.explorer.GetTXIDs(l.bitcoinAdd)
	if err != nil {
		trail.Println(trace.Alert("error getting address history").UTC().Error(err).Append(tr))
		return 0, fmt.Errorf("error getting address history: %w", err)
	}
	entries, err := l.RetrieveAndExtractEntries(history)
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
func (l *Logbook) ExtractEntries(txs []*DataTX) ([]*Entry, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "ExtractEntries")
	trail.Println(trace.Info("reading entries from TXs").Add("len txs", fmt.Sprintf("%d", len(txs))).UTC().Append(tr))
	crypts, err := l.blockchain.UnpackData(txs)
	if err != nil {
		trail.Println(trace.Alert("error unpacking data").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error unpacking data: %w", err)
	}
	entries, err := l.DecryptEntries(crypts)
	if err != nil {
		trail.Println(trace.Warning("error decrypting EntryPart").UTC().Error(err).Append(tr))
	}
	return entries, nil
}

//DecryptEntries decrypts entries from the encrypted OP_RETURN data.
func (l *Logbook) DecryptEntries(crypts [][]byte) ([]*Entry, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "DecryptEntries")
	trail.Println(trace.Info("decrypting data").UTC().Append(tr))
	parts := make([]*EntryPart, 0, len(crypts))
	for _, cry := range crypts {
		enco, err := AESDecrypt(l.cryptoKey, cry)
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

func (l *Logbook) ListHistory(address string) ([]string, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "ListHistory")
	trail.Println(trace.Info("listing TX history").UTC().Add("address", address).Append(tr))
	txids, err := l.blockchain.explorer.GetTXIDs(address)
	if err != nil {
		trail.Println(trace.Warning("error while listing TX history").UTC().Add("address", address).Error(err).Append(tr))
		return nil, fmt.Errorf("error while listing TX history: %w", err)
	}
	return txids, nil
}

func (l *Logbook) MaxDataSize() int {
	//9 is header size and must never be changed
	avai := l.blockchain.miner.MaxOpReturn() - 9
	cryptFactor := 0.5
	disp := float64(avai) * cryptFactor
	return int(disp)
}
