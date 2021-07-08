package ddb

type UTXO struct {
	TXPos        uint32
	TXHash       string
	Value        *Bitcoin
	ScriptPubKey *ScriptPubKey
}

type TX struct {
	ID       string  `json:"txid"`
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
	Value        *Bitcoin      `json:"value"`
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

type Explorer interface {
	GetUTXOs(address string) ([]*UTXO, error)
	GetTX(txHash string) (*TX, error)
	GetRAWTXHEX(txHash string) ([]byte, error)
	// GetTXIDs(address string) (*TX, error)
}
