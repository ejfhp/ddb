package ddb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

// curl https://api.whatsonchain.com/v1/bsv/main/chain/info
// curl https://api.whatsonchain.com/v1/bsv/main/address/1NRoySJ9Lvby6DuE2UQYnyT67AASwNZxGb/unspent
// curl https://api.whatsonchain.com/v1/bsv/main/address/12LwgC8RQ6ScX5mLbNYL6twZba6SpkoLh2/unspent
// curl https://api.whatsonchain.com/v1/bsv/main/tx/hash/73908464acc24e75af7d0046c2ee01a305e235e18b4e941ee75b9dd371b0169b

type wocu struct {
	Height int32  `json:"height"`
	TXPos  uint32 `json:"tx_pos"`
	TXHash string `json:"tx_hash"`
	Value  uint64 `json:"value"`
}

type WOC struct {
	BaseURL string
}

func NewWOC() *WOC {
	w := WOC{BaseURL: "https://api.whatsonchain.com/v1/bsv/main"}
	return &w
}

func (w *WOC) GetUTXOs(address string) ([]*UTXO, error) {
	t := trace.New().Source("whatsonchain.go", "WOC", "GetUTXOs")
	url := fmt.Sprintf("%s/address/%s/unspent", w.BaseURL, address)
	resp, err := http.Get(url)
	if err != nil {
		trail.Println(trace.Alert("error while getting unspent").UTC().Add("address", address).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while getting unspent: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		trail.Println(trace.Alert("error while reading response").UTC().Add("address", address).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	unspent := []*wocu{}
	err = json.Unmarshal(body, &unspent)
	if err != nil {
		trail.Println(trace.Alert("error while unmarshalling").UTC().Add("address", address).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while unmarshalling: %w", err)
	}

	outs := make([]*UTXO, 0)
	for _, u := range unspent {
		tx, err := w.GetTX(u.TXHash)
		if err != nil {
			trail.Println(trace.Alert("cannot get TX").UTC().Add("address", address).Add("TxHash", u.TXHash).Error(err).Append(t))
			return nil, fmt.Errorf("cannot get TX: %w", err)
		}
		for _, to := range tx.Out {
			if to.N == u.TXPos {
				u := UTXO{
					TXHash:          u.TXHash,
					TXPos:           to.N,
					Value:           to.Value,
					ScriptPubKeyHex: to.ScriptPubKey.Hex,
				}
				outs = append(outs, &u)
			}
		}
	}
	return outs, nil
}

func (w *WOC) GetTX(txHash string) (*TX, error) {
	t := trace.New().Source("whatsonchain.go", "WOC", "GetTX")
	url := fmt.Sprintf("%s/tx/hash/%s", w.BaseURL, txHash)
	trail.Println(trace.Debug("get tx").UTC().Add("hash", txHash).Add("url", url).Append(t))
	resp, err := http.Get(url)
	if err != nil {
		trail.Println(trace.Alert("error while getting TX").UTC().Add("txHash", txHash).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while getting TX: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		trail.Println(trace.Alert("error while reading response").UTC().Add("txHash", txHash).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	tx := TX{}
	err = json.Unmarshal(body, &tx)
	if err != nil {
		trail.Println(trace.Alert("error while unmarshalling").UTC().Add("txHash", txHash).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while unmarshalling: %w", err)
	}
	return &tx, nil
}

func (w *WOC) GetRAWTXHEX(txHash string) ([]byte, error) {
	t := trace.New().Source("whatsonchain.go", "WOC", "GetTX")
	template := "%s/tx/%s/hex"
	url := fmt.Sprintf(template, w.BaseURL, txHash)
	trail.Println(trace.Debug("get tx").UTC().Add("hash", txHash).Add("url", url).Append(t))
	resp, err := http.Get(url)
	if err != nil {
		trail.Println(trace.Alert("error while getting TX").UTC().Add("txHash", txHash).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while getting TX: %w", err)
	}
	defer resp.Body.Close()
	hex, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		trail.Println(trace.Alert("error while reading response").UTC().Add("txHash", txHash).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	return hex, nil
}

func (w *WOC) GetTXIDs(address string) ([]string, error) {
	t := trace.New().Source("whatsonchain.go", "WOC", "GetTXIDs")
	url := fmt.Sprintf("%s/address/%s/history", w.BaseURL, address)
	resp, err := http.Get(url)
	if err != nil {
		trail.Println(trace.Alert("error while getting unspent").UTC().Add("address", address).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while getting unspent: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		trail.Println(trace.Alert("error while reading response").UTC().Add("address", address).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	txs := []*wocu{}
	err = json.Unmarshal(body, &txs)
	if err != nil {
		trail.Println(trace.Alert("error while unmarshalling").UTC().Add("address", address).Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while unmarshalling: %w", err)
	}
	txids := make([]string, len(txs))
	for i, t := range txs {
		txids[i] = t.TXHash
	}
	return txids, nil
}
