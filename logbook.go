package ddb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

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

//ExtractEntries rebuild all the Entries contained in the given TXs array.
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

// Key should be 32 bytes (AES-256).
func AESEncrypt(key [32]byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cannot initialize key: %w", err)
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	nonce := make([]byte, noncesize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("cannot generate nonce: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize encryption: %w", err)
	}

	crypted := make([]byte, 0)
	crypted = append(crypted, nonce...)
	cipherdata := aesgcm.Seal(nil, nonce, data, nil)
	fmt.Printf("Encrypt CRYTPTED: %s\n", hex.EncodeToString(cipherdata))
	crypted = append(crypted, cipherdata...)
	return crypted, nil
}

// Key should be 32 bytes (AES-256).
func AESDecrypt(key [32]byte, encrypted []byte) ([]byte, error) {
	nonce := encrypted[:noncesize]
	cipherdata := encrypted[noncesize:]

	fmt.Printf("Decrypt CRYTPTED: %s\n", hex.EncodeToString(cipherdata))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("cannot initialize key: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize encryption: %w", err)
	}

	plaindata, err := aesgcm.Open(nil, nonce, cipherdata, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt: %w", err)
	}
	return plaindata, err
}
