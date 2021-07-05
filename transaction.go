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

type DataTX struct{ bt.Tx }

func TransactionFromHex(h string) (*DataTX, error) {
	tr := trace.New().Source("transaction.go", "Transaction", "TransactionFromHex")
	b := make([]byte, hex.DecodedLen(len(h)))
	_, err := hex.Decode(b, []byte(h))
	if err != nil {
		log.Println(trace.Alert("cannot decode hex").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot decode hex: %w", err)
	}
	tx, err := bt.NewTxFromBytes(b)
	if err != nil {
		log.Println(trace.Alert("cannot build Transaction from HEX").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build Transaction from HEX: %w", err)
	}
	dtx := DataTX{*tx}
	return &dtx, nil
}

func TransactionFromBytes(b []byte) (*DataTX, error) {
	tr := trace.New().Source("transaction.go", "Transaction", "TransactionFromBytes")
	tx, err := bt.NewTxFromBytes(b)
	if err != nil {
		log.Println(trace.Alert("cannot build Transaction from bytes").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build Transaction from bytes: %w", err)
	}
	dtx := DataTX{*tx}
	return &dtx, nil
}

func (t *DataTX) OpReturn() ([]byte, error) {
	tr := trace.New().Source("transaction.go", "Transaction", "OpReturn")
	var opRet []byte
	for _, o := range t.Outputs {
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

// //BuildOPReturnHexTX returns the TXID and the hex encoded TX
// func BuildOpReturnHexTX(utxo *UTXO, key string, fee Token, payload []byte) (string, string, error) {
// 	t := trace.New().Source("transaction.go", "", "BuildOpReturnHexTX")
// 	tx, err := buildOPReturnTX(utxo, key, fee, payload)
// 	if err != nil {
// 		log.Println(trace.Alert("cannot build OP_RETURN TX").UTC().Error(err).Append(t))
// 		return "", "", fmt.Errorf("cannot build OP_RETURN TX: %w", err)
// 	}
// 	return tx.GetTxID(), tx.ToString(), nil
// }

// //BuildOPReturnHexTX returns the TXID and the []byte of the TX
// func BuildOPReturnBytesTX(utxo *UTXO, key string, fee Token, payload []byte) (string, []byte, error) {
// 	t := trace.New().Source("transaction.go", "", "BuildOPReturnBytesTX")
// 	tx, err := buildOPReturnTX(utxo, key, fee, payload)
// 	if err != nil {
// 		log.Println(trace.Alert("cannot build OP_RETURN TX").UTC().Error(err).Append(t))
// 		return "", nil, err
// 	}
// 	return tx.GetTxID(), tx.ToBytes(), nil
// }

//BuildOPReturnHexTX returns the TXID and the []byte of the TX
func BuildOPReturnTX(address string, inTXID string, in Satoshi, inpos uint32, inScriptHex string, key string, fee Token, payload []byte) (*DataTX, error) {
	t := trace.New().Source("transaction.go", "", "buildOPReturnTX")
	log.Println(trace.Info("preparing OP_RETURN transaction").UTC().Add("address", address).Append(t))
	tx := bt.NewTx()
	input, err := bt.NewInputFromUTXO(inTXID, inpos, uint64(in), inScriptHex, math.MaxUint32)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXO input").UTC().Add("TXID", inTXID).Add("inpos", fmt.Sprintf("%d", inpos)).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get UTXO input: %w", err)
	}
	satInput := in
	tx.AddInput(input)
	satOutput := satInput.Sub(fee)
	log.Println(trace.Info("calculating fee").UTC().Add("input", fmt.Sprintf("%0.8f", in.Bitcoin())).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Add("fee", fmt.Sprintf("%0.8f", fee)).Append(t))
	outputS, err := bt.NewP2PKHOutputFromAddress(address, uint64(satOutput.Satoshi()))
	if err != nil {
		log.Println(trace.Alert("cannot create output").UTC().Add("address", address).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Error(err).Append(t))
		return nil, fmt.Errorf("cannot create output, address %s amount %0.8f: %w", address, satOutput.Bitcoin(), err)
	}
	outputD, err := bt.NewOpReturnOutput(payload)
	if err != nil {
		log.Println(trace.Alert("cannot create OP_RETURN output").UTC().Add("address", address).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Error(err).Append(t))
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
	dtx := DataTX{*tx}
	return &dtx, nil
}
