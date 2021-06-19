package ddb

import (
	"fmt"
	"math"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
	"github.com/libsv/go-bt"
)

type Bitcoin struct {
	value uint64
}

func FromSatoshis(satoshi uint64) Bitcoin {
	return Bitcoin{value: satoshi}
}
func FromBitcoin(bitcoin float64) Bitcoin {
	sat := uint64(math.Round(bitcoin * 100000000))
	return Bitcoin{value: sat}
}

func (b Bitcoin) Satoshis() uint64 {
	// sat := uint64(math.Round(float64(b) * 100000000))
	return b.value
}

func (b Bitcoin) Value() float64 {
	return float64(b.value) / 100000000
}

func (b Bitcoin) Sub(s Bitcoin) Bitcoin {
	res := b.value - s.value
	return Bitcoin{value: res}
}
func (b Bitcoin) Add(s Bitcoin) Bitcoin {
	res := b.value + s.value
	return Bitcoin{value: res}
}

func BuildOPReturnHexTX(utxo *UTXO, key string, fee Bitcoin, payload []byte) (string, error) {
	t := trace.New().Source("transaction.go", "", "BuildOPReturnHexTX")
	tx, err := buildOPReturnTX(utxo, key, fee, payload)
	if err != nil {
		log.Println(trace.Alert("cannot build OP_RETURN TX").UTC().Error(err).Append(t))
		return "", err
	}
	return tx.ToString(), nil
}
func BuildOPReturnBytesTX(utxo *UTXO, key string, fee Bitcoin, payload []byte) ([]byte, error) {
	t := trace.New().Source("transaction.go", "", "BuildOPReturnBytesTX")
	tx, err := buildOPReturnTX(utxo, key, fee, payload)
	if err != nil {
		log.Println(trace.Alert("cannot build OP_RETURN TX").UTC().Error(err).Append(t))
		return nil, err
	}
	return tx.ToBytes(), nil
}
func buildOPReturnTX(utxo *UTXO, key string, fee Bitcoin, payload []byte) (*bt.Tx, error) {
	t := trace.New().Source("transaction.go", "", "buildOPReturnTX")
	if utxo == nil {
		return nil, fmt.Errorf("origin UTXO is nil")
	}
	toAddress := utxo.ScriptPubKey.Adresses[0]
	log.Println(trace.Info("preparing OP_RETURN transaction").UTC().Add("address", toAddress).Append(t))
	tx := bt.NewTx()
	input, err := bt.NewInputFromUTXO(utxo.TXHash, utxo.TXPos, utxo.Value.Satoshis(), utxo.ScriptPubKey.Hex, math.MaxUint32)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXO input").UTC().Add("TxHash", utxo.TXHash).Add("TxPos", fmt.Sprintf("%d", utxo.TXPos)).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get UTXO input: %w", err)
	}
	satInput := utxo.Value
	tx.AddInput(input)
	satOutput := satInput.Sub(fee)
	log.Println(trace.Info("calculating fee").UTC().Add("input", fmt.Sprintf("%0.8f", satInput.Value())).Add("output", fmt.Sprintf("%0.8f", satOutput.Value())).Add("fee", fmt.Sprintf("%0.8f", fee.Value())).Append(t))
	outputS, err := bt.NewP2PKHOutputFromAddress(toAddress, satOutput.Satoshis())
	if err != nil {
		log.Println(trace.Alert("cannot create output").UTC().Add("toAddress", toAddress).Add("output", fmt.Sprintf("%0.8f", satOutput.Value())).Error(err).Append(t))
		return nil, fmt.Errorf("cannot create output, address %s amount %0.8f: %w", toAddress, satOutput.Value(), err)
	}
	outputD, err := bt.NewOpReturnOutput(payload)
	if err != nil {
		log.Println(trace.Alert("cannot create OP_RETURN output").UTC().Add("toAddress", toAddress).Add("output", fmt.Sprintf("%0.8f", satOutput.Value())).Error(err).Append(t))
		return nil, fmt.Errorf("cannot create OP_RETURN output: %w", err)
	}
	tx.AddOutput(outputS)
	tx.AddOutput(outputD)
	return tx, nil
}
