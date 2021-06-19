package ddb

import (
	"fmt"
	"math"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type FeeUnit struct {
	Satoshis int `json:"satoshis"` // Fee in satoshis of the amount of Bytes
	Bytes    int `json:"bytes"`    // Number of bytes that the Fee covers
}

// Fee displays the MiningFee as well as the RelayFee for a specific
// FeeType, for example 'standard' or 'data'
// see https://github.com/bitcoin-sv-specs/brfc-misc/tree/master/feespec
type Fee struct {
	FeeType   string  `json:"feeType"` // standard || data
	MiningFee FeeUnit `json:"miningFee"`
	RelayFee  FeeUnit `json:"relayFee"` // Fee for retaining Tx in secondary mempool
}

//Fees is the returned array of Fee from the miner
type Fees []*Fee

func (f Fees) GetStandardFee() (*Fee, error) {
	for _, t := range f {
		if t.FeeType == "standard" {
			return t, nil
		}
	}
	return nil, fmt.Errorf("standard fee not found")
}

func (f Fees) GetDataFee() (*Fee, error) {
	for _, t := range f {
		if t.FeeType == "data" {
			return t, nil
		}
	}
	return nil, fmt.Errorf("data fee not found")
}

//CalculateFee return the amount of satoshi to set as fee for the given TX
func (f *Fee) CalculateFee(tx []byte) Bitcoin {
	t := trace.New().Source("fees.go", "Fee", "CalculateFee")
	size := len(tx)
	log.Println(trace.Info("TX size").UTC().Add("bytes len", fmt.Sprintf("%d", size)).Append(t))
	miningFee := (float64(size) / float64(f.MiningFee.Bytes)) * float64(f.MiningFee.Satoshis)
	// relayFee := (float64(size) / float64(standardFee.RelayFee.Bytes)) * float64(standardFee.RelayFee.Satoshis)
	relayFee := 0.0
	totalFee := uint64(math.Ceil(miningFee + relayFee))
	log.Println(trace.Info("calculating fee").UTC().Add("size", fmt.Sprintf("%d", size)).Add("miningFee", fmt.Sprintf("%.9f", miningFee)).Add("relayFee", fmt.Sprintf("%.9f", relayFee)).Add("totalFee", fmt.Sprintf("%d", totalFee)).Append(t))
	return FromSatoshis(totalFee)
}
