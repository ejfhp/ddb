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

//ListTXHistoryBackward returns all the TXID of the TX history that ends to lastTXID.
//The search follows the address of the first (and there should be only one) P2PKH output of last TX.
//List length is limited to limit if limit > 0.
func (b *Blockchain) ListTXHistoryBackward(lastTXID string, limit int) ([]string, error) {
	tr := trace.New().Source("blockchain.go", "Blockchain", "ListTXHistoryBackwards")
	log.Println(trace.Debug("get TX").UTC().Append(tr))
	tx, err := b.GetTX(lastTXID)
	if err != nil {
		log.Println(trace.Alert("cannot get last TX").UTC().Add("lastTXID", lastTXID).Error(err).Append(tr))
		return nil, fmt.Errorf("cannot get last TX: %w", err)
	}
	firstAddr := ""
	for _, o := range tx.Outputs {
		if o.LockingScript.IsP2PKH() {
			pkhash, err := o.LockingScript.GetPublicKeyHash()
			if err != nil {
				log.Println(trace.Alert("cannot get PubKeyHash from output").UTC().Error(err).Append(tr))
				return nil, fmt.Errorf("cannot get PubKeyHash from output: %w", err)
			}
			addr, err := bscript.NewAddressFromPublicKeyHash(pkhash, true)
			if err != nil {
				log.Println(trace.Alert("cannot get address from PubKeyHash").UTC().Add("pubKeyHash", string(pkhash)).Error(err).Append(tr))
				return nil, fmt.Errorf("cannot get address from PubKeyHash: %w", err)
			}
			firstAddr = addr.AddressString
		}
	}
	//Scan and search the inputs
	return nil, nil

}
