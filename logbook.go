package ddb

import (
	"fmt"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type LogBook struct {
	Key      string
	Miner    Miner
	Explorer Explorer
}

func (l *LogBook) LogText(text string) (string, error) {
	t := trace.New().Source("logbook.go", "", "LogText")
	log.Println(trace.Info("log text").UTC().Append(t))
	address, err := AddressOf(l.Key)
	if err != nil {
		log.Println(trace.Alert("cannot get address from key").UTC().Add("key", l.Key).Error(err).Append(t))
		return "", fmt.Errorf("cannot get address from key: %w", err)
	}
	utxos, err := l.Explorer.GetUTXOs(address)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXOs").UTC().Add("address", address).Error(err).Append(t))
		return "", fmt.Errorf("cannot get UTXOs: %w", err)
	}
	if len(utxos) != 1 {
		log.Println(trace.Alert("found multiple or no UTXO").UTC().Add("address", address).Append(t))
		return "", fmt.Errorf("found multiple or no UTXO")
	}
	u := utxos[0]

	txHex, err := BuildOPReturnTX(u, l.Key, 0, []byte(text))
	if err != nil {
		log.Println(trace.Alert("cannot build transaction").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot get address from key: %w", err)
	}
	fees, err := l.Miner.GetFees()
	if err != nil {
		log.Println(trace.Alert("cannot get fees").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot get fees: %w", err)
	}
	dataFees, err := fees.GetDataFee()
	if err != nil {
		log.Println(trace.Alert("cannot get data fees").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot get fata fees: %w", err)
	}

}
