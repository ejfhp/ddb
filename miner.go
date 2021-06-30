package ddb

import "time"

type MapiPayload struct {
	ApiVersion                string    `json:"apiVersion"`
	Timestamp                 time.Time `json:"timestamp"`
	ExpiryTime                time.Time `json:"expiryTime"`
	TXID                      string    `json:"txid"`
	ReturnResult              string    `json:"returnResult"`
	ResultDescription         string    `json:"resultDescription"`
	MinerID                   string    `json:"minerId"`
	CurrentHighestBlockHash   string    `json:"currentHighestBlockHash"`
	CurrentHighestBlockHeight int       `json:"currentHighestBlockHeight"`
	Fees                      Fees      `json:"fees"`
}

type MapiSubmitTX struct {
	Rawtx              string `json:"rawtx"`
	CallBackUrl        string `json:"callBackUrl"`
	CallBackToken      string `json:"callBackToken"`
	MerkleProof        bool   `json:"merkleProof"`
	MerkleFormat       string `json:"merkleFormat"`
	DsCheck            bool   `json:"dsCheck"`
	CallBackEncryption string `json:"callBackEncryption"`
}

//Miner is an interface that describes miner interctions
type Miner interface {
	GetName() string
	MaxOpReturn() int
	GetFees() (Fees, error)
	GetDataFee() (*Fee, error)
	GetStandardFee() (*Fee, error)
	//SubmitTX submit the given raw tx to Taal MAPI and if succeed return TXID
	SubmitTX(rawTX string) (string, error)
}
