package ddb

import (
	"fmt"

	log "github.com/ejfhp/trail"
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
	log.Println(trace.Debug("new Logbook").UTC().Append(tr))
	address, err := AddressOf(wif)
	if err != nil {
		log.Println(trace.Alert("cannot get address of key").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get address of key: %w", err)
	}
	return &Logbook{bitcoinWif: wif, bitcoinAdd: address, cryptoKey: password, blockchain: blockchain}, nil

}

//CastEntry store the entry on the blockchain, returns the TXID of the transactions generated.
func (l *Logbook) CastEntry(entry *Entry) ([]string, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "CastEntry")
	log.Println(trace.Info("casting entry to the blockcchain").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	txs, err := l.ProcessEntry(entry)
	if err != nil {
		log.Println(trace.Alert("error during TXs preparation").UTC().Add("file", entry.Name).Error(err).Append(tr))
		return nil, fmt.Errorf("error during TXs preparation: %w", err)
	}
	ids, err := l.blockchain.Submit(txs)
	if err != nil {
		log.Println(trace.Alert("error while sending transactions").UTC().Add("file", entry.Name).Error(err).Append(tr))
		return nil, fmt.Errorf("error while sending transactions: %w", err)
	}
	return ids, nil

}

//ProcessEntry prepares all the TXs required to store the entry on the blockchain.
func (l *Logbook) ProcessEntry(entry *Entry) ([]*DataTX, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "ProcessEntry")
	log.Println(trace.Info("preparing file").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(tr))
	parts, err := entry.Parts(l.MaxDataSize())
	if err != nil {
		log.Println(trace.Alert("error generating entries of file").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error generating entries of file: %w", err)
	}
	encryptedEntryParts := make([][]byte, 0, len(parts))
	for _, p := range parts {
		encodedp, err := p.Encode()
		if err != nil {
			log.Println(trace.Alert("error encoding entry part").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error encoding entry part: %w", err)
		}
		cryptedp, err := AESEncrypt(l.cryptoKey, encodedp)
		if err != nil {
			log.Println(trace.Alert("error encrypting entry part").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error encrypting entry part: %w", err)
		}
		encryptedEntryParts = append(encryptedEntryParts, cryptedp)
	}
	txs, err := l.blockchain.PackData(VER_AES, l.bitcoinWif, encryptedEntryParts)
	if err != nil {
		log.Println(trace.Alert("error packing encrypted parts into DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error packing encrrypted parts into DataTXs: %w", err)
	}
	return txs, nil
}

//RetrieveTXs retrieves the TXs with the given IDs.
func (l *Logbook) RetrieveTXs(txids []string) ([]*DataTX, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "RetrieveTXs")
	log.Println(trace.Info("retrieving TXs from the blockchain").Add("len txids", fmt.Sprintf("%d", len(txids))).UTC().Append(tr))
	txs := make([]*DataTX, 0, len(txids))
	for _, txid := range txids {
		tx, err := l.blockchain.GetTX(txid)
		if err != nil {
			log.Println(trace.Alert("error getting DataTX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error getting DataTX: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

//RetrieveAndExtractEntries retrieve the TX with the given IDs and extracts all the Entries fully contained.
func (l *Logbook) RetrieveAndExtractEntries(txids []string) ([]*Entry, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "RetrievingEntries")
	log.Println(trace.Info("retrieving TXs from the blockchain and extracting entries").Add("len txids", fmt.Sprintf("%d", len(txids))).UTC().Append(tr))
	txs, err := l.RetrieveTXs(txids)
	if err != nil {
		log.Println(trace.Alert("error retrieving DataTXs").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error retrieving DataTXs: %w", err)
	}
	entries, err := l.ExtractEntries(txs)
	if err != nil {
		log.Println(trace.Alert("error extracting Entries").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("errorextracting Entries: %w", err)
	}
	return entries, nil
}

//ExtractEntries rebuild all the Entries fully contained in the given TXs array.
func (l *Logbook) ExtractEntries(txs []*DataTX) ([]*Entry, error) {
	tr := trace.New().Source("logbook.go", "Logbook", "ExtractEntries")
	log.Println(trace.Info("reading entries from TXs").Add("len txs", fmt.Sprintf("%d", len(txs))).UTC().Append(tr))
	crypts, err := l.blockchain.UnpackData(txs)
	if err != nil {
		log.Println(trace.Alert("error unpacking data").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error unpacking data: %w", err)
	}
	parts := make([]*EntryPart, 0, len(txs))
	for _, cry := range crypts {
		enco, err := AESDecrypt(l.cryptoKey, cry)
		if err != nil {
			log.Println(trace.Warning("error decrypting OP_RETURN data").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("error decrypting OP_RETURN data: %w", err)
		}
		part, err := EntryPartFromEncodedData(enco)
		if err != nil {
			log.Println(trace.Warning("error decoding EntryPart").UTC().Error(err).Append(tr))
		}
		parts = append(parts, part)
	}
	entries, err := EntriesFromParts(parts)
	if err != nil {
		log.Println(trace.Warning("error while reassembling Entries").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("error while reassembling Entries: %w", err)
	}
	return entries, nil
}

func (l *Logbook) MaxDataSize() int {
	//9 is header size and must never be changed
	avai := l.blockchain.miner.MaxOpReturn() - 9
	cryptFactor := 0.5
	disp := float64(avai) * cryptFactor
	return int(disp)
}
