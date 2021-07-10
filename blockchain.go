package ddb

import (
	"fmt"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
	"github.com/libsv/go-bt/bscript"
)

type Blockchain struct {
	miner    Miner
	explorer Explorer
}

//NewBlockchain builds a new Blockchain. This is the access point to write and read from a blockchain.
func NewBlockchain(miner Miner, explorer Explorer) *Blockchain {
	return &Blockchain{miner: miner, explorer: explorer}
}

//PackEncryptedEntriesPart writes each []data on a single TX chained with the others, returns the TXIDs and the hex encoded TXs
func (b *Blockchain) PackData(version string, ownerKey string, data [][]byte) ([]*DataTX, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "PackEncryptedEntriesPart")
	log.Println(trace.Info("preparing TXs").UTC().Append(tr))
	address, err := AddressOf(ownerKey)
	if err != nil {
		log.Println(trace.Alert("cannot get owner address").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get owner address: %w", err)
	}
	utxos, err := b.GetLastUTXO(address)
	if err != nil {
		log.Println(trace.Alert("cannot get last UTXO").UTC().Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get last UTXO: %w", err)
	}
	dataFee, err := b.miner.GetDataFee()
	if err != nil {
		log.Println(trace.Alert("cannot get data fee from miner").UTC().Add("miner", b.miner.GetName()).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get data fee from miner: %w", err)
	}
	dataTXs := make([]*DataTX, len(data))
	for i, ep := range data {
		tempTx, err := BuildDataTX(address, utxos, ownerKey, Bitcoin(0), ep, version)
		if err != nil {
			log.Println(trace.Alert("cannot build 0-fee TX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("cannot build 0-fee TX: %w", err)
		}
		fee := dataFee.CalculateFee(tempTx.ToBytes())
		dataTx, err := BuildDataTX(address, utxos, ownerKey, fee, ep, version)
		if err != nil {
			log.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(tr))
			return nil, fmt.Errorf("cannot build TX: %w", err)
		}
		log.Println(trace.Info("estimated fee").UTC().Add("fee", fmt.Sprintf("%0.8f", fee.Bitcoin())).Add("txid", dataTx.GetTxID()).Append(tr))
		//UTXO in TX built by BuildDataTX is in position 0
		inPos := 0
		utxos = []*UTXO{{TXPos: 0, TXHash: dataTx.GetTxID(), Value: Satoshi(dataTx.Outputs[inPos].Satoshis).Bitcoin(), ScriptPubKeyHex: dataTx.Outputs[inPos].GetLockingScriptHexString()}}
		dataTXs[i] = dataTx
	}
	return dataTXs, nil
}

//UnpackData extract the OP_RETURN data from the given transaxtions byte arrays
func (b *Blockchain) UnpackData(txs []*DataTX) ([][]byte, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "UnpackEncryptedEntriesPart")
	log.Println(trace.Info("opening TXs").UTC().Append(tr))
	data := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		opr, ver, err := tx.Data()
		// log.Println(trace.Info("DataTX version").Add("version", ver).UTC().Error(err).Append(tr))
		if err != nil {
			log.Println(trace.Alert("error while getting OpReturn data from DataTX").UTC().Add("version", ver).Error(err).Append(tr))
			return nil, fmt.Errorf("error while getting OpReturn data from DataTX ver%s: %w", ver, err)
		}
		data = append(data, opr)
	}
	return data, nil
}

//Submit submits all the transactions to the miner to be included in the blockchain, returns the TX IDs
func (b *Blockchain) Submit(txs []*DataTX) ([]string, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "Submit")
	ids := make([]string, len(txs))
	for i, tx := range txs {
		fee := tx.GetTotalInputSatoshis() - tx.GetTotalOutputSatoshis()
		log.Println(trace.Info("submiting TX").UTC().Add("id", tx.GetTxID()).Add("fee", fmt.Sprintf("%d", fee)).Append(tr))
		id, err := b.miner.SubmitTX(tx.ToString())
		if err != nil {
			log.Println(trace.Alert("cannot submit TX to miner").UTC().Add("TX n.", fmt.Sprintf("%d", i)).Add("miner", b.miner.GetName()).Error(err).Append(tr))
			return nil, fmt.Errorf("cannot submit TX to miner: %w", err)
		}
		if id != tx.GetTxID() {
			log.Println(trace.Alert("for TX miner returned a different TXID").UTC().Add("minerTXID", id).Add("TXID", tx.GetTxID()).Add("miner", b.miner.GetName()).Append(tr))
		}
		ids[i] = id
	}
	return ids, nil

}

func (b *Blockchain) GetLastUTXO(address string) ([]*UTXO, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "GetLastUTXO")
	log.Println(trace.Debug("get last UTXO").UTC().Append(tr))
	utxos, err := b.explorer.GetUTXOs(address)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXOs").UTC().Add("address", address).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get UTXOs: %w", err)
	}
	if len(utxos) < 1 {
		log.Println(trace.Alert("found no UTXO").UTC().Add("address", address).Append(tr))
		return nil, fmt.Errorf("found no UTXO")
	}
	return utxos, nil

}

func (b *Blockchain) GetTX(id string) (*DataTX, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "GetTX")
	log.Println(trace.Debug("get TX").UTC().Append(tr))
	hex, err := b.explorer.GetRAWTXHEX(id)
	if err != nil {
		log.Println(trace.Alert("cannot get TX").UTC().Add("id", id).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get TX: %w", err)
	}
	dataTX, err := DataTXFromHex(string(hex))
	if err != nil {
		log.Println(trace.Alert("cannot build DataTX").UTC().Add("id", id).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot build DataTX: %w", err)
	}
	return dataTX, nil
}

//ListTXHistoryBackward returns all the TXID of the TX history that ends to txid.
//The search follows the given address.
//List length is limited to limit.
func (b *Blockchain) ListTXHistoryBackward(txid string, folllowAddress string, limit int) ([]string, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "ListTXHistoryBackwards")
	log.Println(trace.Debug("get TX").UTC().Append(tr))
	if txid == "" {
		log.Println(trace.Alert("TXID cannot be empty").UTC().Add("lastTXID", txid).Add("followAddress", folllowAddress).Append(tr))
		return nil, fmt.Errorf("TXID cannot be empty, a starting TXID is mandatory")
	}
	tx, err := b.GetTX(txid)
	if err != nil {
		log.Println(trace.Alert("error getting lastTX").UTC().Add("lastTXID", txid).Add("followAddress", folllowAddress).Error(err).Append(tr))
		return nil, fmt.Errorf("error getting lastTX: %w", err)
	}
	path := []string{txid}
	for i, in := range tx.Inputs {
		history, err := b.walkBackward(in.PreviousTxID, in.PreviousTxOutIndex, folllowAddress, 1, limit)
		if err != nil {
			log.Println(trace.Alert("error going back in history").UTC().Add("lastTXID", txid).Add("followAddress", folllowAddress).Error(err).Append(tr))
			return nil, fmt.Errorf("error going back in history input:%d txid:%s", i, in.PreviousTxID)
		}
		path = append(path, history...)
	}
	return path, nil
}

//TODO This could be parallelized
func (b *Blockchain) walkBackward(txid string, prevTXpos uint32, mainAddr string, depth int, maxpathlen int) ([]string, error) {
	if txid == "" {
		return nil, fmt.Errorf("previous tx cannot be empty, a starting TXID is mandatory")
	}
	depth++
	if depth >= maxpathlen {
		fmt.Printf("max pathlen\n")
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
		fmt.Printf("found P2PKH in tx %s\n", txid)
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
			fmt.Printf("output destination address is NOT main address: %s\n", destAddr)
			return []string{}, nil
		}
	} else {
		fmt.Printf("output is NOT a P2PK in tx %s\n", txid)
		return []string{}, nil
	}
}
