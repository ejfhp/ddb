package miner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

//TODO Implemtnet send multi tx and package miners api
type TAAL struct {
	BaseURL string
	fees    Fees
}

func NewTAAL() *TAAL {
	return &TAAL{BaseURL: "https://mapi.taal.com/mapi"}
}

func (l *TAAL) GetName() string {
	return "TAAL"
}
func (l *TAAL) MaxOpReturn() int {
	return 100000
	// return 1000
}

func (l *TAAL) GetFees() (Fees, error) {
	t := trace.New().Source("taal.go", "TAAL", "GetFee")
	if l.fees != nil {
		return l.fees, nil
	}
	url := fmt.Sprintf("%s/feeQuote", l.BaseURL)
	trail.Println(trace.Debug("get fee").UTC().Add("url", url).Append(t))
	resp, err := http.Get(url)
	if err != nil {
		trail.Println(trace.Alert("error while getting fee").UTC().Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while getting fee: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		trail.Println(trace.Alert("error while reading response").UTC().Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while reading response: %w", err)
	}
	trail.Println(trace.Info("miner response").UTC().Add("response", string(body)).Error(err).Append(t))

	mapiResponse := make(map[string]interface{})
	err = json.Unmarshal(body, &mapiResponse)
	if err != nil {
		trail.Println(trace.Alert("error while unmarshalling mapi response").UTC().Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while unmarshalling mapi response: %w", err)
	}
	payload := mapiResponse["payload"].(string)
	mapiPayload := SingleTXResponse{}
	err = json.Unmarshal([]byte(payload), &mapiPayload)
	if err != nil {
		trail.Println(trace.Alert("error while unmarshalling mapi payload").UTC().Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while unmarshalling mapi payload: %w", err)
	}
	l.fees = mapiPayload.Fees
	return mapiPayload.Fees, nil
}

func (l *TAAL) GetDataFee() (*Fee, error) {
	t := trace.New().Source("taal.go", "TAAL", "GetDataFee")
	trail.Println(trace.Debug("get data fee").UTC().Append(t))
	fees, err := l.GetFees()
	if err != nil {
		trail.Println(trace.Alert("cannot get fees").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot get fees: %w", err)
	}
	dataFee, err := fees.GetDataFee()
	if err != nil {
		trail.Println(trace.Alert("cannot get data fee").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot get data fee: %w", err)
	}
	return dataFee, nil
}

func (l *TAAL) GetStandardFee() (*Fee, error) {
	t := trace.New().Source("taal.go", "TAAL", "GetStandardFee")
	trail.Println(trace.Debug("get data fee").UTC().Append(t))
	fees, err := l.GetFees()
	if err != nil {
		trail.Println(trace.Alert("cannot get fees").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot get fees: %w", err)
	}
	stdFee, err := fees.GetStandardFee()
	if err != nil {
		trail.Println(trace.Alert("cannot get standard fee").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot get standard fee: %w", err)
	}
	return stdFee, nil
}

//SubmitTX submit the given raw tx to Tall MAPI and if succeed return TXID
func (l *TAAL) SubmitTX(rawTX string) (string, error) {
	t := trace.New().Source("taal.go", "TAAL", "SubmitTX")
	url := fmt.Sprintf("%s/tx", l.BaseURL)
	trail.Println(trace.Debug("submit tx").UTC().Add("url", url).Append(t))
	//fmt.Printf("\n\n %s \n\n", rawTX)
	mapiSubmitTX := TX{
		Rawtx:       rawTX,
		MerkleProof: false,
		DsCheck:     false,
	}
	payload, err := json.Marshal(mapiSubmitTX)
	if err != nil {
		trail.Println(trace.Alert("error while marshalling MapiSubmitTX").UTC().Add("url", url).Error(err).Append(t))
		return "", fmt.Errorf("error while marshalling MapiSubmitTX: %w", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		trail.Println(trace.Alert("error while posting TX").UTC().Add("url", url).Error(err).Append(t))
		return "", fmt.Errorf("error while posting TX: %w", err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	trail.Println(trace.Info("miner response").UTC().Add("url", url).Add("response", string(body)).Append(t))
	if resp.StatusCode != 200 {
		trail.Println(trace.Alert("miner replied with bad status").UTC().Add("url", url).Add("status", resp.Status).Append(t))
		return "", fmt.Errorf("miner replied with bad status: %s", resp.Status)

	}
	mapiResponse := make(map[string]interface{})
	err = json.Unmarshal(body, &mapiResponse)
	if err != nil {
		trail.Println(trace.Alert("error while unmarshalling mapi response").UTC().Add("url", url).Error(err).Append(t))
		return "", fmt.Errorf("error while unmarshalling mapi response: %w", err)
	}
	responsePayload := mapiResponse["payload"].(string)
	mapiPayload := SingleTXResponse{}
	err = json.Unmarshal([]byte(responsePayload), &mapiPayload)
	if err != nil {
		trail.Println(trace.Alert("error while unmarshalling mapi payload").UTC().Add("url", url).Error(err).Append(t))
		return "", fmt.Errorf("error while unmarshalling mapi payload: %w", err)
	}
	if mapiPayload.ReturnResult != "success" {
		trail.Println(trace.Alert("mapi call unsuccesful").UTC().Add("return", mapiPayload.ReturnResult).Add("returnDecription", mapiPayload.ResultDescription).Append(t))
		return "", fmt.Errorf("mapi call unsuccesfull: %s, %s", mapiPayload.ReturnResult, mapiPayload.ResultDescription)

	}
	txid := mapiPayload.TXID
	return txid, nil
}

func (l *TAAL) SubmitMultiTX(rawTXs []string) ([]string, error) {
	t := trace.New().Source("taal.go", "TAAL", "SubmitMultiTX")
	url := fmt.Sprintf("%s/txs", l.BaseURL)
	trail.Println(trace.Debug("submit multi tx").UTC().Add("url", url).Append(t))
	//fmt.Printf("\n\n %s \n\n", rawTX)
	mapiSubmitMultiTX := make([]TX, len(rawTXs))
	for i, rawTX := range rawTXs {
		mapiSubmitMultiTX[i] = TX{
			Rawtx:       rawTX,
			MerkleProof: false,
			DsCheck:     false,
		}
	}
	payload, err := json.Marshal(mapiSubmitMultiTX)
	if err != nil {
		trail.Println(trace.Alert("error while marshalling MapiSubmitTX").UTC().Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while marshalling []TX: %w", err)
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		trail.Println(trace.Alert("error while posting MultiTX").UTC().Add("url", url).Error(err).Append(t))
		return nil, fmt.Errorf("error while posting MultiTX: %w", err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	trail.Println(trace.Info("miner response").UTC().Add("url", url).Add("response", string(body)).Append(t))
	if resp.StatusCode != 200 {
		trail.Println(trace.Alert("miner replied with bad status").UTC().Add("url", url).Add("status", resp.Status).Append(t))
		return "", fmt.Errorf("miner replied with bad status: %s", resp.Status)

	}
	mapiResponse := make(map[string]interface{})
	err = json.Unmarshal(body, &mapiResponse)
	if err != nil {
		trail.Println(trace.Alert("error while unmarshalling mapi response").UTC().Add("url", url).Error(err).Append(t))
		return "", fmt.Errorf("error while unmarshalling mapi response: %w", err)
	}
	responsePayload := mapiResponse["payload"].(string)
	mapiPayload := SingleTXResponse{}
	err = json.Unmarshal([]byte(responsePayload), &mapiPayload)
	if err != nil {
		trail.Println(trace.Alert("error while unmarshalling mapi payload").UTC().Add("url", url).Error(err).Append(t))
		return "", fmt.Errorf("error while unmarshalling mapi payload: %w", err)
	}
	if mapiPayload.ReturnResult != "success" {
		trail.Println(trace.Alert("mapi call unsuccesful").UTC().Add("return", mapiPayload.ReturnResult).Add("returnDecription", mapiPayload.ResultDescription).Append(t))
		return "", fmt.Errorf("mapi call unsuccesfull: %s, %s", mapiPayload.ReturnResult, mapiPayload.ResultDescription)

	}
	txid := mapiPayload.TXID
	return txid, nil
	return nil, nil
}
