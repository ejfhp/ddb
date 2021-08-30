package ddb

import (
	"encoding/hex"
	"fmt"
	"math"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
	bt "github.com/libsv/go-bt"
	"github.com/libsv/go-bt/bscript"
)

const (
	APP_NAME  = "ddb"  //3 bytes, this must not be changed
	VER_AES   = "0001" //4 bytes
	headerLen = 9
)

type SourceOutput struct {
	TXPos           uint32  `json:"txpos"`
	TXHash          string  `json:"txhash"`
	Value           Satoshi `json:"value"`
	ScriptPubKeyHex string  `json:"scriptpubkeyhex"`
}

func (s *SourceOutput) Equals(so *SourceOutput) bool {
	if s.TXPos != so.TXPos {
		return false
	}
	if s.Value != so.Value {
		return false
	}
	if s.ScriptPubKeyHex != so.ScriptPubKeyHex {
		return false
	}
	if s.TXHash != so.TXHash {
		return false
	}
	return true
}

type DataTX struct {
	SourceOutputs []*SourceOutput
	*bt.Tx
}

//NewDataTX builds a DataTX with the given params. The values in the arrays must be correlated. Generated TX UTXO is in position 0.
func NewDataTX(sourceKey string, destinationAddress string, inutxo []*UTXO, fee Token, data []byte, header string) (*DataTX, error) {
	t := trace.New().Source("transaction.go", "DataTX", "NewDataTX")
	trail.Println(trace.Info("preparing OP_RETURN transaction").UTC().Add("destinationAddress", destinationAddress).Add("num UTXO", fmt.Sprintf("%d", len(inutxo))).Append(t))
	payload, err := addDataHeader(header, data)
	if err != nil {
		trail.Println(trace.Alert("cannot add header").UTC().Add("header", header).Error(err).Append(t))
		return nil, fmt.Errorf("cannot add header: %w", err)
	}
	tx := bt.NewTx()
	satInput := Satoshi(0)
	sourceOutputs := make([]*SourceOutput, 0, len(inutxo))
	for i, utx := range inutxo {
		sourceOutput := SourceOutput{TXPos: utx.TXPos, TXHash: utx.TXHash, Value: utx.Value.Satoshi(), ScriptPubKeyHex: utx.ScriptPubKeyHex}
		sourceOutputs = append(sourceOutputs, &sourceOutput)
		input, err := bt.NewInputFromUTXO(utx.TXHash, utx.TXPos, uint64(utx.Value.Satoshi()), utx.ScriptPubKeyHex, math.MaxUint32)
		if err != nil {
			trail.Println(trace.Alert("cannot get UTXO input").UTC().Add("i", fmt.Sprintf("%d", i)).Add("TXID", utx.TXHash).Add("inpos", fmt.Sprintf("%d", utx.TXPos)).Error(err).Append(t))
			return nil, fmt.Errorf("cannot get UTXO input: %w", err)
		}
		satInput = satInput.Add(utx.Value)
		tx.AddInput(input)
	}
	satOutput := satInput.Sub(fee)
	trail.Println(trace.Info("calculating output values").UTC().Add("input", fmt.Sprintf("%0.8f", satInput.Bitcoin())).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Add("fee", fmt.Sprintf("%0.8f", fee.Bitcoin())).Append(t))
	outputS, err := bt.NewP2PKHOutputFromAddress(destinationAddress, uint64(satOutput.Satoshi()))
	if err != nil {
		trail.Println(trace.Alert("cannot create output").UTC().Add("destinationAddress", destinationAddress).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Error(err).Append(t))
		return nil, fmt.Errorf("cannot create output, destinationAddress %s amount %0.8f: %w", destinationAddress, satOutput.Bitcoin(), err)
	}
	outputD, err := bt.NewOpReturnOutput(payload)
	if err != nil {
		trail.Println(trace.Alert("cannot create OP_RETURN output").UTC().Add("destinationAddress", destinationAddress).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Error(err).Append(t))
		return nil, fmt.Errorf("cannot create OP_RETURN output: %w", err)
	}
	tx.AddOutput(outputS)
	tx.AddOutput(outputD)
	k, err := DecodeWIF(sourceKey)
	if err != nil {
		trail.Println(trace.Alert("error decoding sourcKey").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("error decoding key: %w", err)
	}
	signer := &bt.InternalSigner{PrivateKey: k, SigHashFlag: 0x40 | 0x01}
	for i := range tx.Inputs {
		err = tx.Sign(uint32(i), signer)
		if err != nil {
			trail.Println(trace.Alert("cannot sign transaction").UTC().Add("input", fmt.Sprintf("%d", i)).Error(err).Append(t))
			return nil, fmt.Errorf("cannot sign input %d: %w", i, err)
		}
	}
	dtx := DataTX{SourceOutputs: sourceOutputs, Tx: tx}
	return &dtx, nil
}

func DataTXFromHex(h string) (*DataTX, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "DataTXFromHex")
	b := make([]byte, hex.DecodedLen(len(h)))
	_, err := hex.Decode(b, []byte(h))
	if err != nil {
		trail.Println(trace.Alert("cannot decode hex").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot decode hex: %w", err)
	}
	tx, err := bt.NewTxFromBytes(b)
	if err != nil {
		trail.Println(trace.Alert("cannot build Transaction from HEX").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build Transaction from HEX: %w", err)
	}
	dtx := DataTX{Tx: tx}
	return &dtx, nil
}

func DataTXFromBytes(b []byte) (*DataTX, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "DataTXFromBytes")
	tx, err := bt.NewTxFromBytes(b)
	if err != nil {
		trail.Println(trace.Alert("cannot build Transaction from bytes").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build Transaction from bytes: %w", err)
	}
	dtx := DataTX{Tx: tx}
	return &dtx, nil
}

//OpReturn returns OP_RETURN data
func (t *DataTX) OpReturn() ([]byte, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "Data")
	trail.Println(trace.Info("reading OP_RETURN from DataTX").UTC().Append(tr))
	var data []byte
	for _, o := range t.Outputs {
		if o.LockingScript.IsData() {
			// fmt.Println(o.ToBytes())
			ops, err := bscript.DecodeStringParts(o.GetLockingScriptHexString())
			if err != nil {
				trail.Println(trace.Alert("error decoding output parts").UTC().Error(err).Append(tr))
				return nil, fmt.Errorf("error decoding output parts: %w", err)
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
				data = v
			}
		}
	}
	return data, nil
}

//OpReturn returns data and header of the OP_RETURN data
func (t *DataTX) Data() ([]byte, string, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "Data")
	trail.Println(trace.Info("reading encrypted data from DataTX").UTC().Append(tr))
	opret, err := t.OpReturn()
	if err != nil {
		trail.Println(trace.Alert("error extracting OP_RETURN").UTC().Error(err).Append(tr))
		return nil, "", fmt.Errorf("error extracting OP_RETURN: %w", err)
	}
	header, data, err := stripDataHeader(opret)
	if err != nil {
		trail.Println(trace.Alert("error while stripping header from data").UTC().Error(err).Append(tr))
	}
	return data, header, nil
}

//Data returns data inside OP_RETURN and version of TX
func (t *DataTX) Fee() (Token, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "Fee")
	if t.SourceOutputs == nil || len(t.SourceOutputs) == 0 {
		trail.Println(trace.Alert("transaction has no source utxo").UTC().Append(tr))
		return Satoshi(0), fmt.Errorf("transaction has no source utxo")

	}
	totInput := Satoshi(0)
	for _, in := range t.SourceOutputs {
		trail.Println(trace.Info("input").UTC().Add("value", fmt.Sprintf("%d", in.Value)).Append(tr))
		totInput = totInput.Add(in.Value)
	}
	totOutput := uint64(0)
	for _, out := range t.Outputs {
		trail.Println(trace.Info("output").UTC().Add("value", fmt.Sprintf("%d", out.Satoshis)).Append(tr))
		totOutput += out.Satoshis
	}
	// fmt.Printf("tot input :%d\n", totInput)
	// fmt.Printf("tot output :%d\n", totOutput)
	fee := totInput.Sub(Satoshi(totOutput))
	return fee, nil
}

func BuildDataHeader(version string) (string, error) {
	//header size: len(APP_NAME) + len(";") + len(VER_x) + len(";")
	// ex.  "ddb;0001;"
	if len(version) != 4 {
		return "", fmt.Errorf("version len must be 4")
	}
	header := fmt.Sprintf("%s;%s;", APP_NAME, version)
	if len(header) != 9 {
		return "", fmt.Errorf("header len must be 9")
	}
	return header, nil
}

func stripDataHeader(data []byte) (string, []byte, error) {
	if len(data) < 9 {
		return "", nil, fmt.Errorf("data is shorter than header")
	}
	header := data[:8]
	return string(header), data[9:], nil
}

func addDataHeader(header string, data []byte) ([]byte, error) {
	if len(header) != 9 {
		return nil, fmt.Errorf("header len must be 9")

	}
	payload := make([]byte, len(data)+9)
	copy(payload, header)
	copy(payload[9:], data)
	return payload, nil
}
