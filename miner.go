package ddb

type Miner interface {
	GetFees() (Fees, error)
	SubmitTX(rawTX string) (string, error)
}
