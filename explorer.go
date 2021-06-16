package remy

import "math"

type UTXO struct {
	TXPos        uint32
	TXHash       string
	Value        float64
	ScriptPubKey *ScriptPubKey
}

func (u *UTXO) Satoshis() uint64 {
	sat := uint64(math.Round(u.Value * 100000000))
	return sat
}
func (v *Vout) Satoshis() uint64 {
	sat := uint64(math.Round(v.Value * 100000000))
	return sat
}

type Explorer interface {
	GetUTXOs(net string, address string) ([]*UTXO, error)
	GetTX(net string, txHash string) (*TX, error)
	GetTXIDs(net string, address string) (*TX, error)
}
