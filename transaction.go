package remy

import (
	"fmt"
	"math"

	"github.com/bitcoinsv/bsvd/bsvec"
	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
	"github.com/libsv/go-bk/wif"
	"github.com/libsv/go-bt"
)

type TX struct {
	ID       string  `json:"txidv"`
	Hash     string  `json:"hash"`
	Hex      string  `json:"hex"`
	Version  int     `json:"version"`
	Size     uint32  `json:"size"`
	Locktime uint32  `json:"locktime"`
	In       []*Vin  `json:"vin"`
	Out      []*Vout `json:"vout"`
}

type Vin struct {
	Coinbase  string     `json:"coinbase"`
	TXID      string     `json:"txid"`
	Vout      int        `json:"vout"`
	ScriptSig *ScriptSig `json:"scriptSig>"`
	Sequence  int        `json:"sequence"`
}

type Vout struct {
	Value        float64       `json:"value"`
	N            uint32        `json:"n"`
	ScriptPubKey *ScriptPubKey `json:"scriptPubKey"`
}

type ScriptSig struct {
	Asm string `json:"asm"`
	Hex string `json:"hex"`
}

type ScriptPubKey struct {
	Asm      string   `json:"asm"`
	Hex      string   `json:"hex"`
	ReqSigs  int      `json:"reqSigs"`
	Type     string   `json:"type"`
	Adresses []string `json:"addresses"`
}

func OPReturn(utxo *UTXO, key string, fee uint64, payload []byte) (*bt.Tx, error) {
	t := trace.New().Source("transaction.go", "", "OPReturn")
	tx := bt.NewTx()
	var satInput uint64 = 0
	log.Println(trace.Info("preparing OPReturn transaction").UTC().Append(t))
	input, err := bt.NewInputFromUTXO(utxo.TXHash, utxo.TXPos, utxo.Satoshis(), utxo.ScriptPubKey.Hex, math.MaxUint32)
	if err != nil {
		log.Println(trace.Alert("cannot add UTXO").UTC().Add("TxHash", u.TXHash).Add("TxPos", fmt.Sprintf("%d", u.TXPos)).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get UTXOs: %w", err)
	}
	satInput += u.Satoshis()
	tx.AddInput(input)
	satOutput := satInput - fee
	log.Println(trace.Info("calculating fee").UTC().Add("inputs", fmt.Sprintf("%d", satInput)).Add("outputs", fmt.Sprintf("%d", satOutput)).Add("fee", fmt.Sprintf("%d", fee)).Append(t))
	output, err := bt.NewP2PKHOutputFromAddress(toAddress, satOutput)
	if err != nil {
		log.Println(trace.Alert("cannot create output").UTC().Add("toAddress", toAddress).Add("satoshi", fmt.Sprintf("%d", satOutput)).Error(err).Append(t))
		return nil, fmt.Errorf("cannot create output to %s for %d satoshi: %w", toAddress, satOutput, err)
	}
	tx.AddOutput(output)

}
func UTXOsToAddress(utxos []*UTXO, toAddress string, key string, fee uint64) (*bt.Tx, error) {
	t := trace.New().Source("transaction.go", "", "SendUTXOsToAddress")
	tx := bt.NewTx()
	var satInput uint64 = 0
	log.Println(trace.Info("reading UTXO").UTC().Add("len UTXO", fmt.Sprintf("%d", len(utxos))).Append(t))
	for _, u := range utxos {
		input, err := bt.NewInputFromUTXO(u.TXHash, u.TXPos, u.Satoshis(), u.ScriptPubKey.Hex, math.MaxUint32)
		if err != nil {
			log.Println(trace.Alert("cannot add UTXO").UTC().Add("TxHash", u.TXHash).Add("TxPos", fmt.Sprintf("%d", u.TXPos)).Error(err).Append(t))
			return nil, fmt.Errorf("cannot get UTXOs: %w", err)
		}
		satInput += u.Satoshis()
		tx.AddInput(input)
	}
	satOutput := satInput - fee
	log.Println(trace.Info("calculating fee").UTC().Add("inputs", fmt.Sprintf("%d", satInput)).Add("outputs", fmt.Sprintf("%d", satOutput)).Add("fee", fmt.Sprintf("%d", fee)).Append(t))
	output, err := bt.NewP2PKHOutputFromAddress(toAddress, satOutput)
	if err != nil {
		log.Println(trace.Alert("cannot create output").UTC().Add("toAddress", toAddress).Add("satoshi", fmt.Sprintf("%d", satOutput)).Error(err).Append(t))
		return nil, fmt.Errorf("cannot create output to %s for %d satoshi: %w", toAddress, satOutput, err)
	}
	tx.AddOutput(output)
	w, err := wif.DecodeWIF(key)
	if err != nil {
		log.Println(trace.Alert("error decoding key").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("error decoding key: %w", err)
	}
	signer := &bt.InternalSigner{PrivateKey: (*bsvec.PrivateKey)(w.PrivKey), SigHashFlag: 0x40 | 0x01}
	// signed, err := tx.SignAuto(signer)
	// if err != nil || len(signed) != len(utxos) {
	// 	log.Println(trace.Alert("cannot sign transaction inputs").UTC().Error(err).Append(t))
	// 	return nil, fmt.Errorf("cannot sign transaction inputs: %w", err)
	// }
	for i := range tx.Inputs {
		err = tx.Sign(uint32(i), signer)
		if err != nil {
			log.Println(trace.Alert("cannot sign transaction").UTC().Add("input", fmt.Sprintf("%d", i)).Error(err).Append(t))
			return nil, fmt.Errorf("cannot sign input %d: %w", i, err)
		}
	}
	return tx, nil
}
