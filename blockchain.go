package ddb

import (
	"fmt"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Blockchain struct {
	miner    Miner
	explorer Explorer
}

//NewBlockchain builds a new Blockchain. This is the access point to write and read from a blockchain.
func NewBlockchain(miner Miner, explorer Explorer) *Blockchain {
	return &Blockchain{miner: miner, explorer: explorer}
}

//PackEncryptedEntriesPart writes the raw encrypted entry data on a chained serie of TX, returns the TXID and the hex encoded TX
func (b *Blockchain) PackEncryptedEntriesPart(version string, ownerKey string, encryptedPartsData [][]byte) ([]*DataTX, error) {
	t := trace.New().Source("blockchain.go", "Blockchain", "PackEncryptedEntriesPart")
	log.Println(trace.Info("preparing TXs").UTC().Append(t))
	address, err := AddressOf(ownerKey)
	if err != nil {
		log.Println(trace.Alert("cannot get owner address").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot get owner address: %w", err)
	}
	u, err := b.GetLastUTXO(address)
	if err != nil {
		log.Println(trace.Alert("cannot get last UTXO").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot get last UTXO: %w", err)
	}
	inTXID := u.TXHash
	inSat := u.Value.Satoshi()
	inPos := u.TXPos
	inScr := u.ScriptPubKey.Hex

	dataFee, err := b.miner.GetDataFee()
	if err != nil {
		log.Println(trace.Alert("cannot get data fee from miner").UTC().Add("miner", b.miner.GetName()).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get data fee from miner: %W", err)
	}
	dataTXs := make([]*DataTX, len(encryptedPartsData))
	for i, ep := range encryptedPartsData {
		payload := b.AddHeader(APP_NAME, VER_AES, ep)
		dataTx, err := BuildOPReturnTX(address, inTXID, inSat, inPos, inScr, ownerKey, Bitcoin(0), payload)
		if err != nil {
			log.Println(trace.Alert("cannot build 0-fee TX").UTC().Error(err).Append(t))
			return nil, fmt.Errorf("cannot build 0-fee TX: %W", err)
		}
		fee := dataFee.CalculateFee(dataTx.ToBytes())
		dataTx, err = BuildOPReturnTX(address, inTXID, inSat, inPos, inScr, ownerKey, fee, payload)
		if err != nil {
			log.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(t))
			return nil, fmt.Errorf("cannot build TX: %W", err)
		}
		inTXID = dataTx.GetTxID()
		inSat = Satoshi(dataTx.Outputs[0].Satoshis)
		inPos = 1
		inScr = dataTx.Outputs[0].GetLockingScriptHexString()
		dataTXs[i] = dataTx
	}
	return dataTXs, nil
}

//UnpackEncriptedEntriesPart extract the OP_RETURN data from the given transaxtions byte arrays
func (b *Blockchain) UnpackEncriptedEntriesPart(txs [][]byte) ([][]byte, error) {
	t := trace.New().Source("blockchain.go", "Blockchain", "UnpackEncryptedEntriesPart")
	log.Println(trace.Info("opening TXs").UTC().Append(t))
	cryptedData := make([][]byte, 0, len(txs))
	for i, tx := range txs {
		dataTX, err := TransactionFromBytes(tx)
		if err != nil {
			log.Println(trace.Alert("error while building DataTX from bytes").UTC().Error(err).Append(t))
			return nil, fmt.Errorf("error while building DataTX from bytes: %W", err)
		}
		opr, err := dataTX.OpReturn()
		if err != nil {
			log.Println(trace.Alert("error while getting OpReturn data from DataTX").UTC().Error(err).Append(t))
			return nil, fmt.Errorf("error while getting OpReturn data from DataTX: %W", err)
		}
		cryptedData = append(cryptedData, opr)
	}
	return cryptedData, nil
}

func (b *Blockchain) Submit(txs []*DataTX) ([]string, error) {
	t := trace.New().Source("blockchain.go", "Blockchain", "Submit")
	log.Println(trace.Info("submit TX hex").UTC().Append(t))
	ids := make([]string, len(txs))
	for i, tx := range txs {
		id, err := b.miner.SubmitTX(tx.ToString())
		if err != nil {
			log.Println(trace.Alert("cannot submit TX to miner").UTC().Add("i", fmt.Sprintf("%d", i)).Add("miner", b.miner.GetName()).Error(err).Append(t))
			return nil, fmt.Errorf("cannot submit TX to miner: %W", err)
		}
		if id != tx.GetTxID() {
			log.Println(trace.Alert("miner returned a different TXID").UTC().Add("minerTXID", id).Add("TXID", tx.GetTxID()).Add("miner", b.miner.GetName()).Append(t))
		}
		ids[i] = id
	}
	return ids, nil

}

func (b *Blockchain) GetLastUTXO(address string) (*UTXO, error) {
	t := trace.New().Source("blockchain.go", "Blockchain", "getLastUTXO")
	log.Println(trace.Debug("get last UTXO").UTC().Append(t))
	utxos, err := b.explorer.GetUTXOs(address)
	if err != nil {
		log.Println(trace.Alert("cannot get UTXOs").UTC().Add("address", address).Error(err).Append(t))
		return nil, fmt.Errorf("cannot get UTXOs: %w", err)
	}
	if len(utxos) != 1 {
		log.Println(trace.Alert("found multiple or no UTXO").UTC().Add("address", address).Append(t))
		return nil, fmt.Errorf("found multiple or no UTXO")
	}
	return utxos[0], nil

}

func (b *Blockchain) AddHeader(appName string, version string, data []byte) []byte {
	header := []byte(fmt.Sprintf("%s;%s;", appName, version))
	payload := append(data, header...)
	copy(payload[HEADER_SIZE:], payload)
	copy(payload, header)
	return payload
}
