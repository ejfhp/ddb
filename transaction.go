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

const (
	APP_NAME = "ddb"  //3 bytes, this must not be changed
	VER_AES  = "0001" //4 bytes
)

type SourceOutput struct {
	TXPos           uint32
	TXHash          string
	Value           Satoshi
	ScriptPubKeyHex string
}

type DataTX struct {
	SourceOutputs []*SourceOutput
	*bt.Tx
}

func NewDataTX(utxo []*SourceOutput, tx *bt.Tx) *DataTX {
	return &DataTX{SourceOutputs: utxo, Tx: tx}
}

//BuildDataTX builds a DataTX with the given params. The values in the arrays must be correlated. Generated TX UTXO is in position 0.
func BuildDataTX(address string, inutxo []*UTXO, key string, fee Token, data []byte, version string) (*DataTX, error) {
	t := trace.New().Source("transaction.go", "DataTX", "BuildDataTX")
	log.Println(trace.Info("preparing OP_RETURN transaction").UTC().Add("address", address).Add("num UTXO", fmt.Sprintf("%d", len(inutxo))).Append(t))
	payload, err := addDataHeader(version, data)
	if err != nil {
		log.Println(trace.Alert("cannot add header").UTC().Add("version", version).Error(err).Append(t))
		return nil, fmt.Errorf("cannot add header: %w", err)
	}
	tx := bt.NewTx()
	satInput := Satoshi(0)
	sourceOutputs := make([]*SourceOutput, len(inutxo))
	for i, utx := range inutxo {
		sourceOutput := SourceOutput{TXPos: utx.TXPos, TXHash: utx.TXHash, Value: utx.Value.Satoshi(), ScriptPubKeyHex: utx.ScriptPubKeyHex}
		sourceOutputs = append(sourceOutputs, &sourceOutput)
		input, err := bt.NewInputFromUTXO(utx.TXHash, utx.TXPos, uint64(utx.Value.Satoshi()), utx.ScriptPubKeyHex, math.MaxUint32)
		if err != nil {
			log.Println(trace.Alert("cannot get UTXO input").UTC().Add("i", fmt.Sprintf("%d", i)).Add("TXID", utx.TXHash).Add("inpos", fmt.Sprintf("%d", utx.TXPos)).Error(err).Append(t))
			return nil, fmt.Errorf("cannot get UTXO input: %w", err)
		}
		satInput = satInput.Add(utx.Value)
		tx.AddInput(input)
	}
	satOutput := satInput.Sub(fee)
	log.Println(trace.Info("calculating output values").UTC().Add("input", fmt.Sprintf("%0.8f", satInput.Bitcoin())).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Add("fee", fmt.Sprintf("%0.8f", fee.Bitcoin())).Append(t))
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
	dtx := NewDataTX(sourceOutputs, tx)
	return dtx, nil
}

func DataTXFromHex(h string) (*DataTX, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "DataTXFromHex")
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
	dtx := DataTX{Tx: tx}
	return &dtx, nil
}

func DataTXFromBytes(b []byte) (*DataTX, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "DataTXFromBytes")
	tx, err := bt.NewTxFromBytes(b)
	if err != nil {
		log.Println(trace.Alert("cannot build Transaction from bytes").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build Transaction from bytes: %w", err)
	}
	dtx := DataTX{Tx: tx}
	return &dtx, nil
}

//Data returns data inside OP_RETURN and version of TX
func (t *DataTX) Data() ([]byte, string, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "Data")
	log.Println(trace.Info("reading OP_RETURN from DataTX").UTC().Append(tr))
	var data []byte
	version := ""
	for _, o := range t.Outputs {
		if o.LockingScript.IsData() {
			// fmt.Println(o.ToBytes())
			ops, err := bscript.DecodeStringParts(o.GetLockingScriptHexString())
			if err != nil {
				log.Println(trace.Alert("cannot decode output parts").UTC().Error(err).Append(tr))
				return nil, version, fmt.Errorf("cannot decode output parts: %w", err)
			}
			for _, v := range ops {
				if v[0] == bscript.OpFALSE {
					//fmt.Printf("%d OP_FALSE %d  %v %d\n", i, v[0], v, len(v))
					continue
				}
				if v[0] == bscript.OpRETURN {
					//fmt.Printf("%d OP_RETURN %d  %v %d\n", i, v[0], v, len(v))
					continue
				}
				//fmt.Printf("%d DATA %v  %v %d\n", i, v[0], string(v), len(v))
				version, data, err = stripDataHeader(v)
				if err != nil {
					log.Println(trace.Alert("cannot decode header").UTC().Error(err).Append(tr))
					continue
					// return nil, version, fmt.Errorf("cannot decode header: %w", err)
				}
			}
		}
	}
	return data, version, nil
}

//Data returns data inside OP_RETURN and version of TX
func (t *DataTX) Fees() (Token, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "Fee")
	if t.SourceOutputs == nil || len(t.SourceOutputs) == 0 {
		log.Println(trace.Alert("transaction has no source utxo").UTC().Append(tr))
		return Satoshi(0), fmt.Errorf("transaction has no source utxo")

	}
	totInput := Satoshi(0)
	for _, in := range t.SourceOutputs {
		log.Println(trace.Info("input").UTC().Add("value", fmt.Sprintf("%d", in.Value)).Append(tr))
		totInput = totInput.Add(in.Value)
	}
	totOutput := uint64(0)
	for _, out := range t.Outputs {
		log.Println(trace.Info("output").UTC().Add("value", fmt.Sprintf("%d", out.Satoshis)).Append(tr))
		totOutput += out.Satoshis
	}
	fmt.Printf("tot input :%d\n", totInput)
	fmt.Printf("tot output :%d\n", totOutput)
	fee := totInput.Sub(Satoshi(totOutput))
	return fee, nil
}

func addDataHeader(version string, data []byte) ([]byte, error) {
	//header size: len(APP_NAME) + len(";") + len(VER_x) + len(";")
	// ex.  "ddb;0001;"
	if len(version) != 4 {
		return nil, fmt.Errorf("version len must be 4")
	}
	header := []byte(fmt.Sprintf("%s;%s;", APP_NAME, version))
	if len(header) != 9 {
		return nil, fmt.Errorf("header len must be 9")

	}
	payload := make([]byte, len(data)+9)
	copy(payload[9:], data)
	copy(payload, header)
	return payload, nil
}

func stripDataHeader(data []byte) (string, []byte, error) {
	if len(data) < 9 {
		return "", nil, fmt.Errorf("data is shorter than header")
	}
	appname := data[:3]
	version := data[4:8]
	if string(appname) != APP_NAME {
		return "", nil, fmt.Errorf("appname doensn't match: '%s'", appname)
	}
	return string(version), data[9:], nil
}
