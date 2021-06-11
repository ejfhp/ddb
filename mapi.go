package spocs

import (
	"math"
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

func (u *UTXO) Satoshis() uint64 {
	sat := uint64(math.Round(u.Value * 100000000))
	return sat
}
func (v *Vout) Satoshis() uint64 {
	sat := uint64(math.Round(v.Value * 100000000))
	return sat
}
