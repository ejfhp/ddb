package ddb

import (
	"encoding/hex"
	"fmt"
	"math"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
	bt "github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bscript"
)

type Transaction struct {
	Data []byte
}

func TransactionFromHex(h string) (*Transaction, error) {
	tr := trace.New().Source("transaction.go", "Transaction", "TransactionFromHex")
	b := make([]byte, hex.DecodedLen(len(h)))
	_, err := hex.Decode(b, []byte(h))
	if err != nil {
		log.Println(trace.Alert("cannot build Transaction from HEX").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build Transaction from HEX: %w", err)
	}
	return &Transaction{Data: b}, nil
}

func (t *Transaction) ID() string {
	tr := trace.New().Source("transaction.go", "Transaction", "ID")
	tx, err := bt.NewTxFromBytes(t.Data)
	if err != nil {
		log.Println(trace.Alert("cannot build bt.TX from bytes").UTC().Error(err).Append(tr))
	}
	return tx.GetTxID()
}

func (t *Transaction) HexData() string {
	return hex.EncodeToString(t.Data)
}

func (t *Transaction) OpReturn() ([]byte, error) {
	tr := trace.New().Source("transaction.go", "Transaction", "OpReturn")
	tx, err := bt.NewTxFromBytes(t.Data)
	if err != nil {
		log.Println(trace.Alert("cannot build bt.TX from bytes").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build bt.TX from bytes: %w", err)
	}
	var opRet []byte
	for _, o := range tx.Outputs {
		if o.LockingScript.IsData() {
			// fmt.Println(o.ToBytes())
			ops, err := bscript.DecodeStringParts(o.GetLockingScriptHexString())
			if err != nil {
				log.Println(trace.Alert("cannot decode output parts").UTC().Error(err).Append(tr))
				return nil, fmt.Errorf("cannot decode output parts: %w", err)
			}
			for i, v := range ops {
				if v[0] == bscript.OpFALSE {
					fmt.Printf("%d OP_FALSE %d  %v %d\n", i, v[0], v, len(v))
					continue
				}
				if v[0] == bscript.OpRETURN {
					fmt.Printf("%d OP_RETURN %d  %v %d\n", i, v[0], v, len(v))
					continue
				}
				// fmt.Printf("%d DATA %v  %v %d\n", i, v[0], string(v), len(v))
				opRet = v
			}
		}
	}
	return opRet, nil
}

//BuildOPReturnHexTX returns the TXID and the hex encoded TX
func BuildOpReturnHexTX(utxo *UTXO, key string, fee Token, payload []byte) (string, string, error) {
	t := trace.New().Source("transaction.go", "", "BuildOpReturnHexTX")
	tx, err := buildOPReturnTX(utxo, key, fee, payload)
	if err != nil {
		log.Println(trace.Alert("cannot build OP_RETURN TX").UTC().Error(err).Append(t))
		return "", "", fmt.Errorf("cannot build OP_RETURN TX: %w", err)
	}
	return tx.GetTxID(), tx.ToString(), nil
}

//BuildOPReturnHexTX returns the TXID and the []byte of the TX
func BuildOPReturnBytesTX(utxo *UTXO, key string, fee Token, payload []byte) (string, []byte, error) {
	t := trace.New().Source("transaction.go", "", "BuildOPReturnBytesTX")
	tx, err := buildOPReturnTX(utxo, key, fee, payload)
	if err != nil {
		log.Println(trace.Alert("cannot build OP_RETURN TX").UTC().Error(err).Append(t))
		return "", nil, err
	}
	return tx.GetTxID(), tx.ToBytes(), nil
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
