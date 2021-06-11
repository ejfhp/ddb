package spocs

type UTXO struct {
	TXPos        uint32
	TXHash       string
	Value        float64
	ScriptPubKey *ScriptPubKey
}

type ChainStorage interface {
	GetUTXOs(net string, address string) ([]*UTXO, error)
	GetTX(net string, txHash string) (*TX, error)
	GetTXIDs(net string, address string) (*TX, error)
}
