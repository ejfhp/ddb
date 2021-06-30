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
	miner      Miner
	explorer   Explorer
}

func NewLogbook(wif string, password [32]byte, miner Miner, explorer Explorer) (*Logbook, error) {
	t := trace.New().Source("logbook.go", "Logbook", "NewLogbook")
	log.Println(trace.Debug("new Logbook").UTC().Append(t))
	address, err := AddressOf(wif)
	if err != nil {
		log.Println(trace.Alert("cannot get address of key").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot get address of key: %w", err)
	}
	return &Logbook{bitcoinWif: wif, bitcoinAdd: address, cryptoKey: password, miner: miner, explorer: explorer}, nil

}

//RecordFile store a file (binary or text) on the blockchain, returns the array of the {TXID, TX_HEX} generated.
func (l *Logbook) RecordFile(name string, data []byte) ([][]string, error) {
	t := trace.New().Source("logbook.go", "Logbook", "RecordFile")
	log.Println(trace.Info("preparing file").Add("file", name).Add("size", fmt.Sprintf("%d", len(data))).UTC().Append(t))
	parts, err := EntriesOfFile(name, data, l.maxOpReturnSize())
	if err != nil {
		log.Println(trace.Alert("error generating entries of file").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("error generating entries of file: %w", err)
	}
	txs := make([][]string, 0, len(parts))
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
		id, hex, err := l.PrepareTX(VER_AES, cryptedp)
		if err != nil {
			log.Println(trace.Alert("error preparing entry part TX").UTC().Error(err).Append(t))
			return nil, fmt.Errorf("error preparing entry part TX: %w", err)
		}
		txs = append(txs, []string{id, hex})
	}
	for i, tx := range txs {
		id, err := l.Submit(tx[1])
		if err != nil {
			log.Println(trace.Alert("error submitting entry part TX").Add("num", fmt.Sprintf("%d", i)).UTC().Error(err).Append(t))
			return nil, fmt.Errorf("error submitting entry part TX num %d: %w", i, err)
		}
		if id != tx[0] {
			log.Println(trace.Alert("miner responded with a different TXID").Add("TXID", tx[0]).Add("miner_TXID", id).UTC().Error(err).Append(t))
		}
		txs[i][0] = id
	}
	return txs, nil
}

//PrepareTX encapsulate data in a Bitcoin transaction, returns the TXID and the hex encoded TX
func (l *Logbook) PrepareTX(version string, data []byte) (string, string, error) {
	t := trace.New().Source("logbook.go", "Logbook", "PrepareTX")
	log.Println(trace.Info("preparing TX").UTC().Append(t))
	u, err := l.GetLastUTXO()
	if err != nil {
		log.Println(trace.Alert("cannot get last UTXO").UTC().Error(err).Append(t))
		return "", "", fmt.Errorf("cannot get last UTXO: %w", err)
	}

	dataFee, err := l.miner.GetDataFee()
	if err != nil {
		log.Println(trace.Alert("cannot get data fee from miner").UTC().Add("miner", l.miner.GetName()).Error(err).Append(t))
		return "", "", fmt.Errorf("cannot get data fee from miner: %W", err)
	}

	payload := AddHeader(APP_NAME, VER_AES, data)
	_, txBytes, err := BuildOPReturnBytesTX(u, l.bitcoinWif, Bitcoin(0), payload)
	if err != nil {
		log.Println(trace.Alert("cannot build 0-fee TX").UTC().Error(err).Append(t))
		return "", "", fmt.Errorf("cannot build 0-fee TX: %W", err)
	}
	fee := dataFee.CalculateFee(txBytes)
	txID, txHex, err := BuildOPReturnHexTX(u, l.bitcoinWif, fee, data)
	if err != nil {
		log.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(t))
		return "", "", fmt.Errorf("cannot build TX: %W", err)
	}
	return txID, txHex, nil

}

func (l *Logbook) Submit(txHex string) (string, error) {
	t := trace.New().Source("logbook.go", "Logbook", "Submit")
	log.Println(trace.Info("submit TX hex").UTC().Append(t))
	txid, err := l.miner.SubmitTX(txHex)
	if err != nil {
		log.Println(trace.Alert("cannot submit TX to miner").UTC().Add("miner", l.miner.GetName()).Error(err).Append(t))
		return "", fmt.Errorf("cannot submit TX to miner: %W", err)
	}
	return txid, nil

}

func (l *Logbook) GetLastUTXO() (*UTXO, error) {
	t := trace.New().Source("logbook.go", "Logbook", "getLastUTXO")
	log.Println(trace.Debug("get last UTXO").UTC().Append(t))
	utxos, err := l.explorer.GetUTXOs(l.bitcoinAdd)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXOs").UTC().Add("address", l.bitcoinAdd).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get UTXOs: %w", err)
	}
	if len(utxos) != 1 {
		log.Println(trace.Alert("found multiple or no UTXO").UTC().Add("address", l.bitcoinAdd).Append(t))
		return nil, fmt.Errorf("found multiple or no UTXO")
	}
	return utxos[0], nil

}

func (l *Logbook) maxOpReturnSize() int {
	avai := l.miner.MaxOpReturn() - HEADER_SIZE
	cryptFactor := 0.5
	disp := float64(avai) * cryptFactor
	return int(disp)
}

func AddHeader(appName string, version string, data []byte) []byte {
	header := []byte(fmt.Sprintf("%s;%s;", appName, version))
	payload := append(data, header...)
	copy(payload[HEADER_SIZE:], payload)
	copy(payload, header)
	return payload
}
