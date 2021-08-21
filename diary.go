package ddb

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Diary struct {
	bitcoinWif string
	bitcoinAdd string
	cryptoKey  [32]byte
	Blockchain *Blockchain
}

func NewDiary(wif string, password [32]byte, blockchain *Blockchain) (*Diary, error) {
	tr := trace.New().Source("diary.go", "Diary", "NewDiary")
	address, err := AddressOf(wif)
	if err != nil {
		trail.Println(trace.Alert("cannot get address of key").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get address of key: %w", err)
	}
	d := Diary{bitcoinWif: wif, bitcoinAdd: address, cryptoKey: password, Blockchain: blockchain}
	trail.Println(trace.Debug("new Diary built").UTC().Append(tr).Add("wif", wif).Add("address", address).Add("password", d.EncodingPassword()))
	return &d, nil
}

func NewDiaryRO(address string, password [32]byte, blockchain *Blockchain) (*Diary, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "NewDiaryRO")
	d := Diary{bitcoinWif: "", bitcoinAdd: address, cryptoKey: password, Blockchain: blockchain}
	trail.Println(trace.Debug("new readonly Diary").UTC().Append(tr).Add("address", address).Add("password", d.EncodingPassword()))
	return &d, nil
}

func (l *Diary) BitcoinPrivateKey() string {
	return l.bitcoinWif
}

func (l *Diary) BitcoinPublicAddress() string {
	return l.bitcoinAdd
}

func (l *Diary) EncodingPassword() string {
	return string(l.cryptoKey[:])
}

func (l *Diary) ReadOnly() bool {
	return l.bitcoinWif == ""
}

//CastEntry store the entry on the blockchain. This method is the concatenation of ProcessEntry and Submit. Returns the TXID of the transactions generated.
func (l *Diary) CastEntry(entry *Entry) ([]string, error) {
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

//EstimateFee returns an estimation in satoshi of the fee necessary to store the entry on the blockchain.
func (l *Diary) EstimateFee(entry *Entry) (Satoshi, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "EstimateFee")
	trail.Println(trace.Info("estimate fee to cast entry").UTC().Append(tr).Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))))
	encryptedEntryParts, err := l.EncryptEntry(entry)
	if err != nil {
		trail.Println(trace.Alert("error encrypting entries of file").UTC().Error(err).Append(tr))
		return Satoshi(0), fmt.Errorf("error encrypting entries of file: %w", err)
	}
	fakeUtxos := []*UTXO{{TXPos: 0, TXHash: "72124e293287ab0ca20a723edb61b58d6ef89aba05508b92198bd948bfb6da40", Value: 100000000000, ScriptPubKeyHex: "76a914330a97979931a961d1e5f05d3c7ace4217fc7adc88ac"}}
	dataFee, err := l.Blockchain.Miner.GetDataFee()
	if err != nil {
		trail.Println(trace.Alert("error getting miner data fee").UTC().Error(err).Append(tr))
		return Satoshi(0), fmt.Errorf("error getting miner data fee: %w", err)
	}
	txs, err := PackData(VER_AES, l.bitcoinWif, encryptedEntryParts, fakeUtxos, dataFee)
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
func (l *Diary) Submit(txs []*DataTX) ([]string, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "Submit")
	trail.Println(trace.Info("submitting transactions to the blockcchain").Add("num txs", fmt.Sprintf("%d", len(txs))).UTC().Append(tr))
	ids, err := l.Blockchain.Submit(txs)
	if err != nil {
		trail.Println(trace.Alert("error while sending transactions").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error while sending transactions: %w", err)
	}
	return ids, nil

}

//ProcessEntry prepares all the TXs required to store the entry on the blockchain.
func (l *Diary) ProcessEntry(entry *Entry) ([]*DataTX, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "ProcessEntry")
	trail.Println(trace.Info("preparing file").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	encryptedEntryParts, err := l.EncryptEntry(entry)
	if err != nil {
		trail.Println(trace.Alert("error encrypting entries of file").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error encrypting entries of file: %w", err)
	}
	utxo, err := l.Blockchain.GetUTXO(l.bitcoinAdd)
	if err != nil {
		trail.Println(trace.Alert("error getting UTXO").UTC().Add("address", l.bitcoinAdd).Error(err).Append(tr))
		return nil, fmt.Errorf("error getting UTXO for address %s: %w", l.bitcoinAdd, err)
	}
	fee, err := l.Blockchain.Miner.GetDataFee()
	if err != nil {
		trail.Println(trace.Alert("error miner data fee").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error getting miner data fee: %w", err)
	}
	txs, err := PackData(VER_AES, l.bitcoinWif, encryptedEntryParts, utxo, fee)
	if err != nil {
		trail.Println(trace.Alert("error packing encrypted parts into DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error packing encrrypted parts into DataTXs: %w", err)
	}
	return txs, nil
}

//ProcessEntry prepares all the TXs required to store the entry on the blockchain.
func (l *Diary) EncryptEntry(entry *Entry) ([][]byte, error) {
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
func (l *Diary) RetrieveTXs(txids []string) ([]*DataTX, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "RetrieveTXs")
	trail.Println(trace.Info("retrieving TXs from the blockchain").Add("len txids", fmt.Sprintf("%d", len(txids))).UTC().Append(tr))
	txs := make([]*DataTX, 0, len(txids))
	for _, txid := range txids {
		tx, err := l.Blockchain.GetTX(txid)
		if err != nil {
			trail.Println(trace.Alert("error getting DataTX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error getting DataTX: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

//RetrieveAndExtractEntries retrieve the TX with the given IDs and extracts all the Entries fully contained.
func (l *Diary) RetrieveAndExtractEntries(txids []string) ([]*Entry, error) {
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
func (l *Diary) DowloadAll(outPath string) (int, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "DownloadAll")
	trail.Println(trace.Info("download all entries locally").Add("outpath", outPath).UTC().Append(tr))
	history, err := l.Blockchain.Explorer.GetTXIDs(l.bitcoinAdd)
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
func (l *Diary) ExtractEntries(txs []*DataTX) ([]*Entry, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "ExtractEntries")
	trail.Println(trace.Info("reading entries from TXs").Add("len txs", fmt.Sprintf("%d", len(txs))).UTC().Append(tr))
	crypts, err := UnpackData(txs)
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
func (l *Diary) DecryptEntries(crypts [][]byte) ([]*Entry, error) {
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

func (l *Diary) ListHistory(address string) ([]string, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "ListHistory")
	trail.Println(trace.Info("listing TX history").UTC().Add("address", address).Append(tr))
	txids, err := l.Blockchain.Explorer.GetTXIDs(address)
	if err != nil {
		trail.Println(trace.Warning("error while listing TX history").UTC().Add("address", address).Error(err).Append(tr))
		return nil, fmt.Errorf("error while listing TX history: %w", err)
	}
	return txids, nil
}

func (l *Diary) MaxDataSize() int {
	//9 is header size and must never be changed
	avai := l.Blockchain.Miner.MaxOpReturn() - 9
	cryptFactor := 0.5
	disp := float64(avai) * cryptFactor
	return int(disp)
}
