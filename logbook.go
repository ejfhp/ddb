package ddb

import (
	"fmt"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Logbook struct {
	Key      string
	Miner    Miner
	Explorer Explorer
}

//LogText write a text on the blockchain
func (l *Logbook) LogText(text string) (string, error) {
	t := trace.New().Source("logbook.go", "Logbook", "LogText")
	log.Println(trace.Info("log text").UTC().Append(t))
	u, err := l.getLastUTXO()
	if err != nil {
		log.Println(trace.Alert("cannot get last UTXO").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot get last UTXO: %w", err)
	}

	dataFee, err := l.Miner.GetDataFee()
	if err != nil {
		log.Println(trace.Alert("cannot get data fee from miner").UTC().Add("miner", l.Miner.GetName()).Error(err).Append(t))
		return "", fmt.Errorf("cannot get data fee from miner: %W", err)
	}

	txBytes, err := BuildOPReturnBytesTX(u, l.Key, FromBitcoin(0), []byte(text))
	if err != nil {
		log.Println(trace.Alert("cannot build 0-fee TX").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot build 0-fee TX: %W", err)
	}
	fee := dataFee.CalculateFee(txBytes)
	txHex, err := BuildOPReturnHexTX(u, l.Key, fee, []byte(text))
	if err != nil {
		log.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot build TX: %W", err)
	}
	txid, err := l.Miner.SubmitTX(txHex)
	if err != nil {
		log.Println(trace.Alert("cannot submit TX to miner").UTC().Add("miner", l.Miner.GetName()).Error(err).Append(t))
		return "", fmt.Errorf("cannot submit TX to miner: %W", err)
	}
	return txid, nil

}

func (l *Logbook) getLastUTXO() (*UTXO, error) {
	t := trace.New().Source("logbook.go", "Logbook", "getLastUTXO")
	log.Println(trace.Debug("get last UTXO").UTC().Append(t))
	address, err := AddressOf(l.Key)
	if err != nil {
		log.Println(trace.Alert("cannot get address from key").UTC().Add("key", l.Key).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get address from key: %w", err)
	}
	utxos, err := l.Explorer.GetUTXOs(address)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXOs").UTC().Add("address", address).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get UTXOs: %w", err)
	}
	if len(utxos) != 1 {
		log.Println(trace.Alert("found multiple or no UTXO").UTC().Add("address", address).Append(t))
		return nil, fmt.Errorf("found multiple or no UTXO")
	}
	return utxos[0], nil

}
