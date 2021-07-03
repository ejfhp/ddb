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

func NewBlockchain(miner Miner, explorer Explorer) *Blockchain {
	return &Blockchain{miner: miner, explorer: explorer}
}

//CastEncryptedEntriesPart writes the raw encrypted entry data on the blockchain, returns the TXID and the hex encoded TX
func (b *Blockchain) CastEncryptedEntriesPart(version string, ownerKey string, encryptedPartsData [][]byte) (string, string, error) {
	// id, hex, err := l.PrepareTX(VER_AES, cryptedp) //TODO da qui in poi il payload deve essere passato alla blockchain
	// if err != nil {
	// 	log.Println(trace.Alert("error preparing entry part TX").UTC().Error(err).Append(t))
	// 	return nil, fmt.Errorf("error preparing entry part TX: %w", err)
	// }
	// txs = append(txs, []string{id, hex})
	// for i, tx := range txs {
	// 	id, err := l.Submit(tx[1])
	// 	if err != nil {
	// 		log.Println(trace.Alert("error submitting entry part TX").Add("num", fmt.Sprintf("%d", i)).UTC().Error(err).Append(t))
	// 		return nil, fmt.Errorf("error submitting entry part TX num %d: %w", i, err)
	// 	}
	// 	if id != tx[0] {
	// 		log.Println(trace.Alert("miner responded with a different TXID").Add("TXID", tx[0]).Add("miner_TXID", id).UTC().Error(err).Append(t))
	// 	}
	// 	txs[i][0] = id
	// }
	t := trace.New().Source("blockchain.go", "Blockchain", "CastEncryptedEntriesPart")
	log.Println(trace.Info("preparing TXs").UTC().Append(t))
	address, err := AddressOf(ownerKey)
	if err != nil {
		log.Println(trace.Alert("cannot get owner address").UTC().Error(err).Append(t))
		return "", "", fmt.Errorf("cannot get owner address: %w", err)
	}
	u, err := b.GetLastUTXO(address)
	if err != nil {
		log.Println(trace.Alert("cannot get last UTXO").UTC().Error(err).Append(t))
		return "", "", fmt.Errorf("cannot get last UTXO: %w", err)
	}
	dataFee, err := b.miner.GetDataFee()
	if err != nil {
		log.Println(trace.Alert("cannot get data fee from miner").UTC().Add("miner", b.miner.GetName()).Error(err).Append(t))
		return "", "", fmt.Errorf("cannot get data fee from miner: %W", err)
	}

	for i, ep := range encryptedPartsData {
		payload := addHeader(APP_NAME, VER_AES, ep)
		_, txBytes, err := BuildOPReturnBytesTX(u, ownerKey, Bitcoin(0), payload)

	}
	if err != nil {
		log.Println(trace.Alert("cannot build 0-fee TX").UTC().Error(err).Append(t))
		return "", "", fmt.Errorf("cannot build 0-fee TX: %W", err)
	}
	fee := dataFee.CalculateFee(txBytes)
	txID, txHex, err := BuildOpReturnHexTX(u, l.bitcoinWif, fee, data)
	if err != nil {
		log.Println(trace.Alert("cannot build TX").UTC().Error(err).Append(t))
		return "", "", fmt.Errorf("cannot build TX: %W", err)
	}
	return txID, txHex, nil

}

func (b *Blockchain) Submit(txsHex []string) ([]string, error) {
	t := trace.New().Source("blockchain.go", "Blockchain", "Submit")
	log.Println(trace.Info("submit TX hex").UTC().Append(t))
	ids := make([]string, len(txsHex))
	for i, tx := range txsHex {
		id, err := b.miner.SubmitTX(tx)
		if err != nil {
			log.Println(trace.Alert("cannot submit TX to miner").UTC().Add("i", fmt.Sprintf("%d", i)).Add("miner", b.miner.GetName()).Error(err).Append(t))
			return nil, fmt.Errorf("cannot submit TX to miner: %W", err)
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

func addHeader(appName string, version string, data []byte) []byte {
	header := []byte(fmt.Sprintf("%s;%s;", appName, version))
	payload := append(data, header...)
	copy(payload[HEADER_SIZE:], payload)
	copy(payload, header)
	return payload
}
