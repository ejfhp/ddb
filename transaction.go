package ddb

import (
	"fmt"
	"math"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
	"github.com/libsv/go-bt"
)

func BuildOPReturnTX(utxo *UTXO, key string, fee uint64, payload []byte) (string, error) {
	t := trace.New().Source("transaction.go", "", "BuildOPReturnTX")
	if utxo == nil {
		return "", fmt.Errorf("origin UTXO is nil")
	}
	toAddress := utxo.ScriptPubKey.Adresses[0]
	log.Println(trace.Info("preparing OP_RETURN transaction").UTC().Add("address", toAddress).Append(t))
	tx := bt.NewTx()
	input, err := bt.NewInputFromUTXO(utxo.TXHash, utxo.TXPos, utxo.Satoshis(), utxo.ScriptPubKey.Hex, math.MaxUint32)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXO input").UTC().Add("TxHash", utxo.TXHash).Add("TxPos", fmt.Sprintf("%d", utxo.TXPos)).Error(err).Append(t))
		return "", fmt.Errorf("cannot get UTXO input: %w", err)
	}
	satInput := utxo.Satoshis()
	tx.AddInput(input)
	satOutput := satInput - fee
	log.Println(trace.Info("calculating fee").UTC().Add("input", fmt.Sprintf("%d", satInput)).Add("output", fmt.Sprintf("%d", satOutput)).Add("fee", fmt.Sprintf("%d", fee)).Append(t))
	outputS, err := bt.NewP2PKHOutputFromAddress(toAddress, satOutput)
	if err != nil {
		log.Println(trace.Alert("cannot create output").UTC().Add("toAddress", toAddress).Add("output", fmt.Sprintf("%d", satOutput)).Error(err).Append(t))
		return "", fmt.Errorf("cannot create output, address %s amount %d: %w", toAddress, satOutput, err)
	}
	outputD, err := bt.NewOpReturnOutput(payload)
	if err != nil {
		log.Println(trace.Alert("cannot create OP_RETURN output").UTC().Add("toAddress", toAddress).Add("output", fmt.Sprintf("%d", satOutput)).Error(err).Append(t))
		return "", fmt.Errorf("cannot create OP_RETURN output: %w", err)
	}
	tx.AddOutput(outputS)
	tx.AddOutput(outputD)
	return tx.ToString(), nil
}
