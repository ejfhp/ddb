package ddb

//Miner is an interface that describes miner interctions
type Miner interface {
	GetName() string
	GetFees() (Fees, error)
	GetDataFee() (*Fee, error)
	GetStandardFee() (*Fee, error)
	//SubmitTX submit the given raw tx to Tall MAPI and if succeed return TXID
	SubmitTX(rawTX string) (string, error)
}
