package ddb

import (
	"fmt"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Blockchain struct {
	miner    Miner
	explorer Explorer
	Cache    *TXCache
}

//NewBlockchain builds a new Blockchain. This is the access point to write and read from a blockchain.
func NewBlockchain(miner Miner, explorer Explorer, cache *TXCache) *Blockchain {
	return &Blockchain{miner: miner, explorer: explorer, Cache: cache}
}

//CacheDir returns the cache folder path
func (b *Blockchain) CacheDir() string {
	if b.Cache == nil {
		return ""
	}
	return b.Cache.path
}

func (b *Blockchain) NewDataTX(sourceKey string, destinationAddress string, changeAddress string, inutxo []*UTXO, fee Token, amount Token, data []byte, header string) (*DataTX, error) {
	return NewDataTX(sourceKey, destinationAddress, changeAddress, inutxo, fee, amount, data, header)
}

func (b *Blockchain) EstimateDataTXFee(numUTXO int, data []byte, header string) (Satoshi, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "Submit")
	key, add, utxos := fakeKeyAddUTXO(numUTXO)
	dataTX, err := NewDataTX(key, add, add, utxos, Satoshi(1), Satoshi(1), data, header)
	if err != nil {
		trail.Println(trace.Alert("cannot build fake DataTX").UTC().Append(tr).Error(err))
		return 0, fmt.Errorf("cannot submit TX to miner: %w", err)
	}
	fee, err := b.GetDataFee()
	if err != nil {
		trail.Println(trace.Alert("cannot get Fee").UTC().Append(tr).Error(err))
		return 0, fmt.Errorf("cannot get Fee: %w", err)
	}
	return fee.CalculateFee(len(dataTX.ToBytes())), nil
}

//Submit submits all the transactions to the miner to be included in the blockchain, returns the TX IDs
func (b *Blockchain) Submit(txs []*DataTX) ([]string, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "Submit")
	ids := make([]string, len(txs))
	for i, tx := range txs {
		fee := tx.GetTotalInputSatoshis() - tx.GetTotalOutputSatoshis()
		trail.Println(trace.Info("submiting TX").UTC().Add("id", tx.GetTxID()).Add("fee", fmt.Sprintf("%d", fee)).Append(tr))
		id, err := b.miner.SubmitTX(tx.ToString())
		if err != nil {
			trail.Println(trace.Alert("cannot submit TX to miner").UTC().Add("TX n.", fmt.Sprintf("%d", i)).Add("miner", b.miner.GetName()).Error(err).Append(tr))
			return nil, fmt.Errorf("cannot submit TX to miner: %w", err)
		}
		if id != tx.GetTxID() {
			trail.Println(trace.Alert("for TX miner returned a different TXID").UTC().Add("minerTXID", id).Add("TXID", tx.GetTxID()).Add("miner", b.miner.GetName()).Append(tr))
		}
		if b.Cache != nil {
			err = b.Cache.StoreTX(id, tx.ToBytes())
			if err != nil {
				trail.Println(trace.Alert("error while storing submitted TX in cache").UTC().Add("TXID", tx.GetTxID()).Append(tr))
			}
		}
		ids[i] = id
	}
	return ids, nil
}

func (b *Blockchain) GetDataFee() (*Fee, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "GetDataFee")
	if b.miner != nil {
		trail.Println(trace.Debug("get data fee").UTC().Append(tr))
		return b.miner.GetDataFee()
	}
	return &Fee{}, fmt.Errorf("miner undefined")
}

func (b *Blockchain) MaxDataSize() int {
	//9 is header size and must never be changed
	avai := b.miner.MaxOpReturn() - 9
	cryptFactor := 0.5
	disp := float64(avai) * cryptFactor
	return int(disp)
}

func (b *Blockchain) GetUTXO(address string) ([]*UTXO, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "GetLastUTXO")
	trail.Println(trace.Debug("get last UTXO").UTC().Append(tr))
	utxos, err := b.explorer.GetUTXOs(address)
	if err != nil {
		trail.Println(trace.Alert("cannot get UTXOs").UTC().Add("address", address).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get UTXOs: %w", err)
	}
	if len(utxos) < 1 {
		trail.Println(trace.Alert("found no UTXO").UTC().Add("address", address).Append(tr))
		return nil, fmt.Errorf("found no UTXO")
	}
	return utxos, nil
}

func (b *Blockchain) GetTX(id string, cacheOnly bool) (*DataTX, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "GetTX")
	trail.Println(trace.Debug("get TX").UTC().Add("cacheOnly", fmt.Sprintf("%t", cacheOnly)).Append(tr))
	var dataTX *DataTX
	if b.Cache != nil {
		cacheTx, err := b.Cache.RetrieveTX(id)
		if err != nil {
			if err != ErrNotCached {
				trail.Println(trace.Alert("cannot get TX from cache").UTC().Add("id", id).Error(err).Append(tr))
				return nil, fmt.Errorf("cannot get TX with id %s from cache: %w", id, err)
			}
			trail.Println(trace.Info("TX not in cache").UTC().Add("id", id).Error(err).Append(tr))
		} else {
			dataTX, err = DataTXFromBytes(cacheTx)
			if err != nil {
				trail.Println(trace.Alert("cannot build DataTX from cache bytes").UTC().Add("id", id).Error(err).Append(tr))
				return nil, fmt.Errorf("cannot build DataTX with id %s from cache bytes: %w", id, err)
			}
			return dataTX, nil
		}
	}
	if cacheOnly {
		trail.Println(trace.Alert("TX not in cache").UTC().Add("id", id).Append(tr))
		return nil, fmt.Errorf("TX not in cache")
	}
	hex, err := b.explorer.GetRAWTXHEX(id)
	if err != nil {
		trail.Println(trace.Alert("cannot get TX").UTC().Add("id", id).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get TX: %w", err)
	}
	dataTX, err = DataTXFromHex(string(hex))
	if err != nil {
		trail.Println(trace.Alert("cannot build DataTX").UTC().Add("id", id).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build DataTX: %w", err)
	}
	if b.Cache != nil {
		err = b.Cache.StoreTX(id, dataTX.ToBytes())
		if err != nil {
			trail.Println(trace.Warning("error while storing retrieved TX in cache").UTC().Add("TXID", dataTX.GetTxID()).Append(tr))
		}
	}
	return dataTX, nil
}

func (b *Blockchain) GetTXs(ids []string, cacheOnly bool) ([]*DataTX, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "GetTXs")
	trail.Println(trace.Debug("get TXs").Append(tr).UTC().Add("len(txids)", fmt.Sprintf("%d", len(ids))))
	txs := make([]*DataTX, 0, len(ids))
	for _, id := range ids {
		tx, err := b.GetTX(id, cacheOnly)
		if err != nil {
			trail.Println(trace.Alert("error while gettin TX").UTC().Add("id", id).Error(err).Append(tr))
			return nil, fmt.Errorf("error while getting TX with id:%s: %w", id, err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (b *Blockchain) ListTXIDs(address string, cacheOnly bool) ([]string, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "ListTXID")
	trail.Println(trace.Debug("listing TXIDs").UTC().Add("address", address).Append(tr))
	txids := []string{}
	if b.explorer != nil && !cacheOnly {
		ids, err := b.explorer.GetTXIDs(address)
		if err != nil {
			trail.Println(trace.Alert("error while getting TXIDs from explorer").UTC().Add("address", address).Error(err).Append(tr))
			return nil, fmt.Errorf("error while getting TXIDs from explorer: %w", err)
		}
		if b.Cache != nil {
			b.Cache.StoreTXIDs(address, ids)
		}
		txids = append(txids, ids...)
	} else if b.Cache != nil {
		ids, err := b.Cache.GetTXIDs(address)
		if err != nil && err != ErrNotCached {
			trail.Println(trace.Alert("error while getting TXIDs from cache").UTC().Add("address", address).Error(err).Append(tr))
			return nil, fmt.Errorf("error while getting TXIDs from cache: %w", err)
		}
		txids = append(txids, ids...)
	}
	return txids, nil
}

//ListTXHistoryBackward returns all the TXID of the TX history that ends to txid.
//The search follows the given address.
//List length is limited to limit.
// func (b *Blockchain) ListTXHistoryBackward(txid string, folllowAddress string, limit int) ([]string, error) {
// 	tr := trace.New().Source("blockchain.go", "Blockchain", "ListTXHistoryBackwards")
// 	trail.Println(trace.Debug("get TX").UTC().Append(tr))
// 	if txid == "" {
// 		trail.Println(trace.Alert("TXID cannot be empty").UTC().Add("lastTXID", txid).Add("followAddress", folllowAddress).Append(tr))
// 		return nil, fmt.Errorf("TXID cannot be empty, a starting TXID is mandatory")
// 	}
// 	tx, err := b.GetTX(txid)
// 	if err != nil {
// 		trail.Println(trace.Alert("error getting lastTX").UTC().Add("lastTXID", txid).Add("followAddress", folllowAddress).Error(err).Append(tr))
// 		return nil, fmt.Errorf("error getting lastTX: %w", err)
// 	}
// 	path := []string{txid}
// 	for i, in := range tx.Inputs {
// 		history, err := b.walkBackward(in.PreviousTxID, in.PreviousTxOutIndex, folllowAddress, 1, limit)
// 		if err != nil {
// 			trail.Println(trace.Alert("error going back in history").UTC().Add("lastTXID", txid).Add("followAddress", folllowAddress).Error(err).Append(tr))
// 			return nil, fmt.Errorf("error going back in history input:%d txid:%s", i, in.PreviousTxID)
// 		}
// 		path = append(path, history...)
// 	}
// 	return path, nil
// }

//Data returns data inside OP_RETURN and version of TX
func (b *Blockchain) FillSourceOutput(tx *DataTX) error {
	tr := trace.New().Source("blockchain.go", "Blockchain", "FillSourceOutput")
	trail.Println(trace.Debug("filling source outputs").UTC().Append(tr))
	sourceOutputs := make([]*SourceOutput, 0)
	for _, in := range tx.Inputs {
		prevTX, err := b.GetTX(in.PreviousTxID, false)
		if err != nil {
			return fmt.Errorf("error retrieving previous transaction: %w", err)
		}
		sourceOut := SourceOutput{
			TXPos:           in.PreviousTxOutIndex,
			TXHash:          in.PreviousTxID,
			Value:           Satoshi(prevTX.Outputs[in.PreviousTxOutIndex].Satoshis),
			ScriptPubKeyHex: prevTX.Outputs[in.PreviousTxOutIndex].GetLockingScriptHexString(),
		}
		sourceOutputs = append(sourceOutputs, &sourceOut)
	}
	tx.SourceOutputs = sourceOutputs
	return nil
}

func (b *Blockchain) GetFakeUTXO() []*UTXO {
	return []*UTXO{{TXPos: 0, TXHash: "72124e293287ab0ca20a723edb61b58d6ef89aba05508b92198bd948bfb6da40", Value: 100000000000, ScriptPubKeyHex: "76a914330a97979931a961d1e5f05d3c7ace4217fc7adc88ac"}}
}

//TODO This could be parallelized
// func (b *Blockchain) walkBackward(txid string, prevTXpos uint32, mainAddr string, depth int, maxpathlen int) ([]string, error) {
// 	if txid == "" {
// 		return nil, fmt.Errorf("previous tx cannot be empty, a starting TXID is mandatory")
// 	}
// 	depth++
// 	if depth >= maxpathlen {
// 		//fmt.Printf("max pathlen\n")
// 		return []string{}, nil
// 	}
// 	tx, err := b.GetTX(txid)
// 	if err != nil {
// 		return nil, fmt.Errorf("cannot get last path TX: %w", err)
// 	}
// 	if len(tx.Outputs) <= int(prevTXpos) {
// 		return nil, fmt.Errorf("prev output index out of range: %d with %d outputs", prevTXpos, len(tx.Outputs))
// 	}
// 	output := tx.Outputs[prevTXpos]
// 	if output.LockingScript.IsP2PKH() {
// 		//fmt.Printf("found P2PKH in tx %s\n", txid)
// 		pkhash, err := output.LockingScript.GetPublicKeyHash()
// 		if err != nil {
// 			return nil, fmt.Errorf("cannot get PubKeyHash from output: %w", err)
// 		}
// 		addr, err := bscript.NewAddressFromPublicKeyHash(pkhash, true)
// 		if err != nil {
// 			return nil, fmt.Errorf("cannot get address from PubKeyHash: %w", err)
// 		}
// 		destAddr := addr.AddressString
// 		if destAddr == mainAddr {
// 			path := []string{txid}
// 			for i, in := range tx.Inputs {
// 				history, err := b.walkBackward(in.PreviousTxID, in.PreviousTxOutIndex, mainAddr, depth, maxpathlen)
// 				if err != nil {
// 					return nil, fmt.Errorf("cannot get history from input %d of tx %s", i, in.PreviousTxID)
// 				}
// 				path = append(path, history...)
// 			}
// 			return path, nil

// 		} else {
// 			//fmt.Printf("output destination address is NOT main address: %s\n", destAddr)
// 			return []string{}, nil
// 		}
// 	} else {
// 		//fmt.Printf("output is NOT a P2PK in tx %s\n", txid)
// 		return []string{}, nil
// 	}
// }
