package ddb

import (
	"fmt"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const (
	HEADER_SIZE = 9      // len(APP_NAME) + len(";") + len(VER_x) + len(";")
	APP_NAME    = "ddb"  //this must not be changed
	VER_AES     = "0001" //max 3 bytes
)

type Logbook struct {
	bitcoinWif string
	bitcoinAdd string
	cryptoKey  [32]byte
	blockchain *Blockchain
}

func NewLogbook(wif string, password [32]byte, blockchain *Blockchain) (*Logbook, error) {
	t := trace.New().Source("logbook.go", "Logbook", "NewLogbook")
	log.Println(trace.Debug("new Logbook").UTC().Append(t))
	address, err := AddressOf(wif)
	if err != nil {
		log.Println(trace.Alert("cannot get address of key").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot get address of key: %w", err)
	}
	return &Logbook{bitcoinWif: wif, bitcoinAdd: address, cryptoKey: password, blockchain: blockchain}, nil

}

//RecordFile store a file (binary or text) on the blockchain, returns the array of the {TXID, TX_HEX} generated.
func (l *Logbook) RecordEntry(entry *Entry) ([]DataTX, error) {
	t := trace.New().Source("logbook.go", "Logbook", "RecordFile")
	log.Println(trace.Info("preparing file").Add("file", entry.Name).Add("size", fmt.Sprintf("%d", len(entry.Data))).UTC().Append(t))
	parts, err := entry.Parts(l.maxDataSize())
	if err != nil {
		log.Println(trace.Alert("error generating entries of file").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("error generating entries of file: %w", err)
	}
	encryptedEntryParts := make([][]byte, 0, len(parts))
	for _, p := range parts {
		encodedp, err := p.Encode()
		if err != nil {
			log.Println(trace.Alert("error encoding entry part").UTC().Error(err).Append(t))
			return nil, fmt.Errorf("error encoding entry part: %w", err)
		}
		cryptedp, err := AESEncrypt(l.cryptoKey, encodedp)
		if err != nil {
			log.Println(trace.Alert("error encrypting entry part").UTC().Error(err).Append(t))
			return nil, fmt.Errorf("error encrypting entry part: %w", err)
		}
		encryptedEntryParts = append(encryptedEntryParts, cryptedp)
	}
	return nil, nil
}

func (l *Logbook) ReadEntries(txs []DataTX) ([]Entry, error) {
	return nil, nil
}

func (l *Logbook) maxDataSize() int {
	avai := l.blockchain.miner.MaxOpReturn() - HEADER_SIZE
	cryptFactor := 0.5
	disp := float64(avai) * cryptFactor
	return int(disp)
}
