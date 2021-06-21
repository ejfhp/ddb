package ddb

import (
	"fmt"
	"math"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
	"github.com/libsv/go-bt"
)

func BuildOPReturnHexTX(utxo *UTXO, key string, fee Token, payload []byte) (string, error) {
	t := trace.New().Source("transaction.go", "", "BuildOPReturnHexTX")
	tx, err := buildOPReturnTX(utxo, key, fee, payload)
	if err != nil {
		log.Println(trace.Alert("cannot build OP_RETURN TX").UTC().Error(err).Append(t))
		return "", err
	}
	return tx.ToString(), nil
}
func BuildOPReturnBytesTX(utxo *UTXO, key string, fee Token, payload []byte) ([]byte, error) {
	t := trace.New().Source("transaction.go", "", "BuildOPReturnBytesTX")
	tx, err := buildOPReturnTX(utxo, key, fee, payload)
	if err != nil {
		log.Println(trace.Alert("cannot build OP_RETURN TX").UTC().Error(err).Append(t))
		return nil, err
	}
	return tx.ToBytes(), nil
}
func buildOPReturnTX(utxo *UTXO, key string, fee Token, payload []byte) (*bt.Tx, error) {
	t := trace.New().Source("transaction.go", "", "buildOPReturnTX")
	if utxo == nil {
		return nil, fmt.Errorf("origin UTXO is nil")
	}
	toAddress := utxo.ScriptPubKey.Adresses[0]
	log.Println(trace.Info("preparing OP_RETURN transaction").UTC().Add("address", toAddress).Append(t))
	tx := bt.NewTx()
	input, err := bt.NewInputFromUTXO(utxo.TXHash, utxo.TXPos, uint64(utxo.Value.Satoshi()), utxo.ScriptPubKey.Hex, math.MaxUint32)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXO input").UTC().Add("TxHash", utxo.TXHash).Add("TxPos", fmt.Sprintf("%d", utxo.TXPos)).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get UTXO input: %w", err)
	}
	satInput := utxo.Value
	tx.AddInput(input)
	satOutput := satInput.Sub(fee)
	log.Println(trace.Info("calculating fee").UTC().Add("input", fmt.Sprintf("%0.8f", *satInput)).Add("output", fmt.Sprintf("%0.8f", satOutput)).Add("fee", fmt.Sprintf("%0.8f", fee)).Append(t))
	outputS, err := bt.NewP2PKHOutputFromAddress(toAddress, uint64(satOutput.Satoshi()))
	if err != nil {
		log.Println(trace.Alert("cannot create output").UTC().Add("toAddress", toAddress).Add("output", fmt.Sprintf("%0.8f", satOutput)).Error(err).Append(t))
		return nil, fmt.Errorf("cannot create output, address %s amount %0.8f: %w", toAddress, satOutput, err)
	}
	outputD, err := bt.NewOpReturnOutput(payload)
	if err != nil {
		log.Println(trace.Alert("cannot create OP_RETURN output").UTC().Add("toAddress", toAddress).Add("output", fmt.Sprintf("%0.8f", satOutput)).Error(err).Append(t))
		return nil, fmt.Errorf("cannot create OP_RETURN output: %w", err)
	}
	tx.AddOutput(outputS)
	tx.AddOutput(outputD)
	k, err := DecodeWIF(key)
	if err != nil {
		log.Println(trace.Alert("error decoding key").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("error decoding key: %w", err)
	}
	signer := &bt.InternalSigner{PrivateKey: k, SigHashFlag: 0x40 | 0x01}
	for i := range tx.Inputs {
		err = tx.Sign(uint32(i), signer)
		if err != nil {
			log.Println(trace.Alert("cannot sign transaction").UTC().Add("input", fmt.Sprintf("%d", i)).Error(err).Append(t))
			return nil, fmt.Errorf("cannot sign input %d: %w", i, err)
		}
	}
	return tx, nil
}
