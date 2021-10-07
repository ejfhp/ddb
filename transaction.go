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
	APP_NAME    = "ddb"  //3 bytes, this must not be changed
	VER_AES     = "0001" //4 bytes
	headerLen   = 9
	FakeTXValue = 100
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

//DataTX is a wrapper for a TX and its source output
type DataTX struct {
	SourceOutputs []*SourceOutput
	*bt.Tx
}

//NewDataTX builds a DataTX with the given params. Output order is: 1:destination, 2:opreturn, 3:change.
func NewDataTX(sourceKey string, destinationAddress string, changeAddress string, inutxo []*UTXO, amount Token, fee Token, data []byte, header string) (*DataTX, error) {
	t := trace.New().Source("transaction.go", "DataTX", "NewDataTX")
	var payload []byte = nil
	var err error
	if data != nil {
		payload, err = addDataHeader(header, data)
		if err != nil {
			trail.Println(trace.Alert("cannot add header").UTC().Add("header", header).Error(err).Append(t))
			return nil, fmt.Errorf("cannot add header: %w", err)
		}
	}
	tx, err := NewTX(sourceKey, destinationAddress, changeAddress, inutxo, amount, fee, payload)
	if err != nil {
		trail.Println(trace.Alert("error creating a new transaction").UTC().Append(t).Error(err))
		return nil, fmt.Errorf("error creating a new transaction: %w", err)
	}
	sourceOutputs := []*SourceOutput{}
	for _, utx := range inutxo {
		sourceOutput := SourceOutput{TXPos: utx.TXPos, TXHash: utx.TXHash, Value: utx.Value.Satoshi(), ScriptPubKeyHex: utx.ScriptPubKeyHex}
		sourceOutputs = append(sourceOutputs, &sourceOutput)
	}
	dtx := DataTX{SourceOutputs: sourceOutputs, Tx: tx}
	return &dtx, nil
}

//NewMultiInputTX builds a DataTX transaction collecting the amount from the multiple UTXO given.
func NewMultiInputTX(destinationAddress string, inputs map[string][]*UTXO, fee Token) (*DataTX, error) {
	t := trace.New().Source("transaction.go", "", "NewMultiInputTX")
	tx := bt.NewTx()
	satInput := Satoshi(0)
	sourceOutputs := []*SourceOutput{}
	signers := []*bt.InternalSigner{}
	for k, utxos := range inputs {
		key, err := DecodeWIF(k)
		if err != nil {
			trail.Println(trace.Alert("error decoding WIF").UTC().Error(err).Append(t))
			return nil, fmt.Errorf("error decoding WIF: %w", err)
		}
		signer := &bt.InternalSigner{PrivateKey: key, SigHashFlag: 0x40 | 0x01}
		for i, utx := range utxos {
			input, err := bt.NewInputFromUTXO(utx.TXHash, utx.TXPos, uint64(utx.Value.Satoshi()), utx.ScriptPubKeyHex, math.MaxUint32)
			if err != nil {
				trail.Println(trace.Alert("cannot get UTXO input").UTC().Add("i", fmt.Sprintf("%d", i)).Add("TXID", utx.TXHash).Add("inpos", fmt.Sprintf("%d", utx.TXPos)).Error(err).Append(t))
				return nil, fmt.Errorf("cannot get UTXO input: %w", err)
			}
			satInput = satInput.Add(utx.Value)
			tx.AddInput(input)
			signers = append(signers, signer)
			sourceOutput := SourceOutput{TXPos: utx.TXPos, TXHash: utx.TXHash, Value: utx.Value.Satoshi(), ScriptPubKeyHex: utx.ScriptPubKeyHex}
			sourceOutputs = append(sourceOutputs, &sourceOutput)
		}
	}
	satOutput := satInput.Sub(fee)
	outputDest, err := bt.NewP2PKHOutputFromAddress(destinationAddress, uint64(satOutput.Satoshi()))
	if err != nil {
		trail.Println(trace.Alert("cannot create output").UTC().Add("destinationAddress", destinationAddress).Append(t).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Error(err))
		return nil, fmt.Errorf("cannot create output, destinationAddress %s amount %0.8f: %w", destinationAddress, satOutput.Bitcoin(), err)
	}
	tx.AddOutput(outputDest)
	if len(signers) != len(tx.Inputs) {
		return nil, fmt.Errorf("signers and inputs have different length")
	}
	for i := range tx.Inputs {
		err = tx.Sign(uint32(i), signers[i])
		if err != nil {
			trail.Println(trace.Alert("cannot sign transaction").UTC().Add("input", fmt.Sprintf("%d", i)).Error(err).Append(t))
			return nil, fmt.Errorf("cannot sign input %d: %w", i, err)
		}
	}
	dtx := DataTX{SourceOutputs: sourceOutputs, Tx: tx}
	return &dtx, nil
}

//NewTX builds a bt.TX transaction with the given params. The optional OP_RETURN data is in position 0. To move all the amount connected to the address use put EmptyWallet as amount.
//No nil field allowed.
func NewTX(sourceKey string, destinationAddress string, changeAddress string, inutxo []*UTXO, amount Token, fee Token, opreturn []byte) (*bt.Tx, error) {
	t := trace.New().Source("transaction.go", "", "NewTX")
	tx := bt.NewTx()
	satInput := Satoshi(0)
	for i, utx := range inutxo {
		input, err := bt.NewInputFromUTXO(utx.TXHash, utx.TXPos, uint64(utx.Value.Satoshi()), utx.ScriptPubKeyHex, math.MaxUint32)
		if err != nil {
			trail.Println(trace.Alert("cannot get UTXO input").UTC().Add("i", fmt.Sprintf("%d", i)).Add("TXID", utx.TXHash).Add("inpos", fmt.Sprintf("%d", utx.TXPos)).Error(err).Append(t))
			return nil, fmt.Errorf("cannot get UTXO input: %w", err)
		}
		satInput = satInput.Add(utx.Value)
		tx.AddInput(input)
	}
	if satInput == 0 {
		trail.Println(trace.Alert("input is 0").Append(t).UTC())
		return nil, fmt.Errorf("input is 0")
	}
	satOutput := Satoshi(0)
	satChange := Satoshi(0)
	if amount.Bitcoin() < 0 {
		trail.Println(trace.Alert("requested output amount is negative").Append(t).UTC())
		return nil, fmt.Errorf("requested output amount is negative")

	}
	if satInput < amount.Satoshi().Add(fee) {
		trail.Println(trace.Alert("output is less than 0").Append(t).UTC())
		return nil, fmt.Errorf("output is less than 0")
	}
	if amount.Satoshi() == EmptyWallet {
		trail.Println(trace.Warning("requested output is EmptyWallet").Append(t).UTC())
		satOutput = satInput.Sub(fee)
	} else {
		satOutput = amount.Satoshi()
	}

	outputDest, err := bt.NewP2PKHOutputFromAddress(destinationAddress, uint64(satOutput.Satoshi()))
	if err != nil {
		trail.Println(trace.Alert("cannot create output").UTC().Add("destinationAddress", destinationAddress).Append(t).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Error(err))
		return nil, fmt.Errorf("cannot create output, destinationAddress %s amount %0.8f: %w", destinationAddress, satOutput.Bitcoin(), err)
	}
	tx.AddOutput(outputDest)

	if opreturn != nil {
		outOpRet, err := bt.NewOpReturnOutput(opreturn)
		if err != nil {
			trail.Println(trace.Alert("cannot create OP_RETURN output").UTC().Add("destinationAddress", destinationAddress).Add("output", fmt.Sprintf("%0.8f", satOutput.Bitcoin())).Error(err).Append(t))
			return nil, fmt.Errorf("cannot create OP_RETURN output: %w", err)
		}
		tx.AddOutput(outOpRet)
	}

	satChange = satInput.Sub(satOutput.Satoshi().Add(fee))
	if satChange.Satoshi() > 0 {
		outputChange, err := bt.NewP2PKHOutputFromAddress(changeAddress, uint64(satChange.Satoshi()))
		if err != nil {
			trail.Println(trace.Alert("cannot create output").UTC().Add("changeAddress", changeAddress).Append(t).Add("output", fmt.Sprintf("%0.8f", satChange.Bitcoin())).Error(err))
			return nil, fmt.Errorf("cannot create output, changeAddress %s amount %0.8f: %w", changeAddress, satChange.Bitcoin(), err)
		}
		tx.AddOutput(outputChange)
	}

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
	return tx, nil
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
	// trail.Println(trace.Info("reading OP_RETURN from DataTX").UTC().Append(tr))
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
				//Third place in ops
				return v, nil
			}
		}
	}
	return nil, fmt.Errorf("no OP_RETURN found")
}

//OpReturn returns data and header of the OP_RETURN data
func (t *DataTX) Data() ([]byte, string, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "Data")
	// trail.Println(trace.Info("reading encrypted data from DataTX").UTC().Append(tr))
	opret, err := t.OpReturn()
	if err != nil {
		trail.Println(trace.Alert("error extracting OP_RETURN").UTC().Error(err).Append(tr))
		return nil, "", fmt.Errorf("error extracting OP_RETURN: %w", err)
	}
	header, data, err := stripDataHeader(opret)
	if err != nil {
		trail.Println(trace.Alert("error while stripping header from data").UTC().Error(err).Append(tr))
		return nil, "", fmt.Errorf("error stripping header from data: %w", err)
	}
	return data, header, nil
}

//Fee returns fee of TX
func (t *DataTX) TotInOutFee() (Satoshi, Satoshi, Satoshi, error) {
	tr := trace.New().Source("transaction.go", "DataTX", "Fee")
	if t.SourceOutputs == nil || len(t.SourceOutputs) == 0 {
		trail.Println(trace.Alert("transaction has no source utxo").UTC().Append(tr))
		return Satoshi(0), Satoshi(0), Satoshi(0), fmt.Errorf("transaction has no source utxo")

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
	fee := totInput.Sub(Satoshi(totOutput))
	// fmt.Printf("TX IN: %d  OUT: %d  FEE: %d\n", totInput, totOutput, fee)
	return totInput, Satoshi(totOutput), fee, nil
}

func (t *DataTX) UTXOs() []*UTXO {
	utxos := make([]*UTXO, 0)
	for op, out := range t.Outputs {
		u := UTXO{TXPos: uint32(op), TXHash: t.GetTxID(), Value: Satoshi(out.Satoshis).Bitcoin(), ScriptPubKeyHex: out.GetLockingScriptHexString()}
		utxos = append(utxos, &u)
	}
	return utxos
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
	header := data[:9]
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

func fakeKeyAddUTXO(num int) (string, string, []*UTXO) {
	address := SampleAddress
	key := SampleKey
	// var changeAddress string = "1EpFjTzJoNAFyJKVGATzxhgqXigUWLNWM6"
	// var changeKey string = "L2mk9qzXebT1gfwUuALMJrbqBtrJxGUN5JnVeqQTGRXytqpXsPr8"
	txid := "e6706b900df5a46253b8788f691cbe1506c1e9b76766f1f9d6b3602e1458f055"
	scriptHex := "76a9142f353ff06fe8c4d558b9f58dce952948252e5df788ac"
	utxos := make([]*UTXO, 0, num)
	for i := 0; i < num; i++ {
		utxos = append(utxos, &UTXO{TXHash: txid, TXPos: uint32(i), Value: FakeTXValue, ScriptPubKeyHex: scriptHex})
	}
	return key, address, utxos
}
