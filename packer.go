package ddb

import (
	"fmt"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

//PackEncryptedEntriesPart writes each []data on a single TX chained with the others, returns the TXIDs and the hex encoded TXs
func PackData(version string, ownerKey string, data [][]byte, utxos []*UTXO, dataFee *Fee) ([]*DataTX, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "PackEncryptedEntriesPart")
	trail.Println(trace.Info("packing bytes in an array of DataTX").UTC().Append(tr))
	address, err := AddressOf(ownerKey)
	if err != nil {
		trail.Println(trace.Alert("cannot get owner address").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get owner address: %w", err)
	}
	dataTXs := make([]*DataTX, len(data))
	for i, ep := range data {
		tempTx, err := BuildDataTX(address, utxos, ownerKey, Bitcoin(0), ep, version)
		if err != nil {
			trail.Println(trace.Alert("cannot build 0 fee DataTX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("cannot build 0 fee DataTX: %w", err)
		}
		fee := dataFee.CalculateFee(tempTx.ToBytes())
		dataTx, err := BuildDataTX(address, utxos, ownerKey, fee, ep, version)
		if err != nil {
			trail.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("cannot build TX: %w", err)
		}
		trail.Println(trace.Info("estimated fee").UTC().Add("fee", fmt.Sprintf("%0.8f", fee.Bitcoin())).Add("txid", dataTx.GetTxID()).Append(tr))
		//UTXO in TX built by BuildDataTX is in position 0
		inPos := 0
		utxos = []*UTXO{{TXPos: 0, TXHash: dataTx.GetTxID(), Value: Satoshi(dataTx.Outputs[inPos].Satoshis).Bitcoin(), ScriptPubKeyHex: dataTx.Outputs[inPos].GetLockingScriptHexString()}}
		dataTXs[i] = dataTx
	}
	return dataTXs, nil
}

//UnpackData extract the OP_RETURN data from the given transaxtions byte arrays
func UnpackData(txs []*DataTX) ([][]byte, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "UnpackEncryptedEntriesPart")
	trail.Println(trace.Info("opening TXs").UTC().Append(tr))
	data := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		opr, ver, err := tx.Data()
		// trail.Println(trace.Info("DataTX version").Add("version", ver).UTC().Error(err).Append(tr))
		if err != nil {
			trail.Println(trace.Alert("error while getting OpReturn data from DataTX").UTC().Add("version", ver).Error(err).Append(tr))
			return nil, fmt.Errorf("error while getting OpReturn data from DataTX ver%s: %w", ver, err)
		}
		data = append(data, opr)
	}
	return data, nil
}
