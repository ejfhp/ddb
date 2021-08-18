package ddb

import (
	"fmt"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
	"github.com/libsv/go-bt/bscript"
)

type Blockchain struct {
	Miner    Miner
	Explorer Explorer
	Cache    *TXCache
}

//NewBlockchain builds a new Blockchain. This is the access point to write and read from a blockchain.
func NewBlockchain(miner Miner, explorer Explorer, cache *TXCache) *Blockchain {
	return &Blockchain{Miner: miner, Explorer: explorer, Cache: cache}
}

//Submit submits all the transactions to the miner to be included in the blockchain, returns the TX IDs
func (b *Blockchain) Submit(txs []*DataTX) ([]string, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "Submit")
	ids := make([]string, len(txs))
	for i, tx := range txs {
		fee := tx.GetTotalInputSatoshis() - tx.GetTotalOutputSatoshis()
		trail.Println(trace.Info("submiting TX").UTC().Add("id", tx.GetTxID()).Add("fee", fmt.Sprintf("%d", fee)).Append(tr))
		id, err := b.Miner.SubmitTX(tx.ToString())
		if err != nil {
			trail.Println(trace.Alert("cannot submit TX to miner").UTC().Add("TX n.", fmt.Sprintf("%d", i)).Add("miner", b.Miner.GetName()).Error(err).Append(tr))
			return nil, fmt.Errorf("cannot submit TX to miner: %w", err)
		}
		if id != tx.GetTxID() {
			trail.Println(trace.Alert("for TX miner returned a different TXID").UTC().Add("minerTXID", id).Add("TXID", tx.GetTxID()).Add("miner", b.Miner.GetName()).Append(tr))
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

func (b *Blockchain) GetUTXO(address string) ([]*UTXO, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "GetLastUTXO")
	trail.Println(trace.Debug("get last UTXO").UTC().Append(tr))
	utxos, err := b.Explorer.GetUTXOs(address)
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

func (b *Blockchain) GetTX(id string) (*DataTX, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "GetTX")
	trail.Println(trace.Debug("get TX").UTC().Append(tr))
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
		}
	}
	if dataTX == nil {
		hex, err := b.Explorer.GetRAWTXHEX(id)
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
	}
	return dataTX, nil
}

func (b *Blockchain) ListTXID(address string, cacheonly bool) ([]string, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "ListTXID")
	trail.Println(trace.Debug("listing TXIDs").UTC().Add("address", address).Add("cacheonly", fmt.Sprintf("%t", cacheonly)).Append(tr))
	txids := make([]string, 0)
	if cacheonly && b.Cache != nil {
		txids, err := b.Cache.RetrieveTXIDs(address)
		if err != ErrNotCached {
			trail.Println(trace.Alert("error while getting TXID from cache").UTC().Add("address", address).Error(err).Append(tr))
			return nil, fmt.Errorf("error while getting TXID from cache: %w", err)
		}
		return txids, nil
	}
	txids, err := b.Explorer.GetTXIDs(address)
	if err != nil {
		trail.Println(trace.Alert("error while getting TXID from explorer").UTC().Add("address", address).Error(err).Append(tr))
		return nil, fmt.Errorf("error while getting TXID from explorer: %w", err)
	}
	if b.Cache != nil {
		b.Cache.StoreTXIDs(address, txids)
	}
	return txids, nil
}

//ListTXHistoryBackward returns all the TXID of the TX history that ends to txid.
//The search follows the given address.
//List length is limited to limit.
func (b *Blockchain) ListTXHistoryBackward(txid string, folllowAddress string, limit int) ([]string, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "ListTXHistoryBackwards")
	trail.Println(trace.Debug("get TX").UTC().Append(tr))
	if txid == "" {
		trail.Println(trace.Alert("TXID cannot be empty").UTC().Add("lastTXID", txid).Add("followAddress", folllowAddress).Append(tr))
		return nil, fmt.Errorf("TXID cannot be empty, a starting TXID is mandatory")
	}
	tx, err := b.GetTX(txid)
	if err != nil {
		trail.Println(trace.Alert("error getting lastTX").UTC().Add("lastTXID", txid).Add("followAddress", folllowAddress).Error(err).Append(tr))
		return nil, fmt.Errorf("error getting lastTX: %w", err)
	}
	path := []string{txid}
	for i, in := range tx.Inputs {
		history, err := b.walkBackward(in.PreviousTxID, in.PreviousTxOutIndex, folllowAddress, 1, limit)
		if err != nil {
			trail.Println(trace.Alert("error going back in history").UTC().Add("lastTXID", txid).Add("followAddress", folllowAddress).Error(err).Append(tr))
			return nil, fmt.Errorf("error going back in history input:%d txid:%s", i, in.PreviousTxID)
		}
		path = append(path, history...)
	}
	return path, nil
}

//Data returns data inside OP_RETURN and version of TX
func (b *Blockchain) FillSourceOutput(tx *DataTX) error {
	tr := trace.New().Source("blockchain.go", "Blockchain", "FillSourceOutput")
	trail.Println(trace.Debug("filling source outputs").UTC().Append(tr))
	sourceOutputs := make([]*SourceOutput, 0)
	for _, in := range tx.Inputs {
		prevTX, err := b.GetTX(in.PreviousTxID)
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

//TODO This could be parallelized
func (b *Blockchain) walkBackward(txid string, prevTXpos uint32, mainAddr string, depth int, maxpathlen int) ([]string, error) {
	if txid == "" {
		return nil, fmt.Errorf("previous tx cannot be empty, a starting TXID is mandatory")
	}
	depth++
	if depth >= maxpathlen {
		//fmt.Printf("max pathlen\n")
		return []string{}, nil
	}
	tx, err := b.GetTX(txid)
	if err != nil {
		return nil, fmt.Errorf("cannot get last path TX: %w", err)
	}
	if len(tx.Outputs) <= int(prevTXpos) {
		return nil, fmt.Errorf("prev output index out of range: %d with %d outputs", prevTXpos, len(tx.Outputs))
	}
	output := tx.Outputs[prevTXpos]
	if output.LockingScript.IsP2PKH() == true {
		//fmt.Printf("found P2PKH in tx %s\n", txid)
		pkhash, err := output.LockingScript.GetPublicKeyHash()
		if err != nil {
			return nil, fmt.Errorf("cannot get PubKeyHash from output: %w", err)
		}
		addr, err := bscript.NewAddressFromPublicKeyHash(pkhash, true)
		if err != nil {
			return nil, fmt.Errorf("cannot get address from PubKeyHash: %w", err)
		}
		destAddr := addr.AddressString
		if destAddr == mainAddr {
			path := []string{txid}
			for i, in := range tx.Inputs {
				history, err := b.walkBackward(in.PreviousTxID, in.PreviousTxOutIndex, mainAddr, depth, maxpathlen)
				if err != nil {
					return nil, fmt.Errorf("cannot get history from input %d of tx %s", i, in.PreviousTxID)
				}
				path = append(path, history...)
			}
			return path, nil

		} else {
			//fmt.Printf("output destination address is NOT main address: %s\n", destAddr)
			return []string{}, nil
		}
	} else {
		//fmt.Printf("output is NOT a P2PK in tx %s\n", txid)
		return []string{}, nil
	}
}
