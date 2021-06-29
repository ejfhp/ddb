package ddb

import (
	"fmt"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Logbook struct {
	key      string
	address  string
	miner    Miner
	explorer Explorer
}

func NewLogbook(key string, miner Miner, explorer Explorer) (*Logbook, error) {
	t := trace.New().Source("logbook.go", "Logbook", "NewLogbook")
	log.Println(trace.Debug("new Logbook").UTC().Append(t))
	address, err := AddressOf(key)
	if err != nil {
		log.Println(trace.Alert("cannot get address of key").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot get address of key: %w", err)
	}
	return &Logbook{key: key, address: address, miner: miner, explorer: explorer}, nil

}

//RecordFile store a file (binary or text) on the blockchain, returns the array of TXs generated.
func (l *Logbook) RecordFile(name string, data []byte) ([][]byte, error) {

}

func (l *Logbook) putOnChain(data []byte) (string, error) {
	t := trace.New().Source("logbook.go", "Logbook", "LogText")
	log.Println(trace.Info("log text").UTC().Append(t))
	u, err := l.getLastUTXO()
	if err != nil {
		log.Println(trace.Alert("cannot get last UTXO").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot get last UTXO: %w", err)
	}

	dataFee, err := l.miner.GetDataFee()
	if err != nil {
		log.Println(trace.Alert("cannot get data fee from miner").UTC().Add("miner", l.miner.GetName()).Error(err).Append(t))
		return "", fmt.Errorf("cannot get data fee from miner: %W", err)
	}

	txBytes, err := BuildOPReturnBytesTX(u, l.key, Bitcoin(0), data)
	if err != nil {
		log.Println(trace.Alert("cannot build 0-fee TX").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot build 0-fee TX: %W", err)
	}
	fee := dataFee.CalculateFee(txBytes)
	txHex, err := BuildOPReturnHexTX(u, l.key, fee, data)
	fmt.Printf("TX Hex: %s\n", txHex)
	if err != nil {
		log.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot build TX: %W", err)
	}
	txid, err := l.miner.SubmitTX(txHex)
	if err != nil {
		log.Println(trace.Alert("cannot submit TX to miner").UTC().Add("miner", l.miner.GetName()).Error(err).Append(t))
		return "", fmt.Errorf("cannot submit TX to miner: %W", err)
	}
	return txid, nil

}

func (l *Logbook) getLastUTXO() (*UTXO, error) {
	t := trace.New().Source("logbook.go", "Logbook", "getLastUTXO")
	log.Println(trace.Debug("get last UTXO").UTC().Append(t))
	utxos, err := l.explorer.GetUTXOs(l.address)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXOs").UTC().Add("address", l.address).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get UTXOs: %w", err)
	}
	if len(utxos) != 1 {
		log.Println(trace.Alert("found multiple or no UTXO").UTC().Add("address", l.address).Append(t))
		return nil, fmt.Errorf("found multiple or no UTXO")
	}
	return utxos[0], nil

}
