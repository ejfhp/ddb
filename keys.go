package ddb

import (
	"fmt"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Keygen interface {
	WIF() (string, error)
	Password() [32]byte
}

func DecodeWIF(wifkey string) (*bsvec.PrivateKey, error) {
	t := trace.New().Source("keys.go", "", "DecodeWIF")
	wif, err := bsvutil.DecodeWIF(wifkey)
	if err != nil {
		log.Println(trace.Alert("cannot decode WIF").UTC().Error(err).Append(t))
		return nil, fmt.Errorf("cannot decode WIF: %w", err)
	}
	priv := wif.PrivKey
	return priv, nil
}

func AddressOf(wifkey string) (string, error) {
	tr := trace.New().Source("keys.go", "", "AddressOf")
	w, err := bsvutil.DecodeWIF(wifkey)
	if err != nil {
		log.Println(trace.Alert("cannot decode WIF").UTC().Add("wif", wifkey).Error(err).Append(tr))
		return "", fmt.Errorf("cannot decode WIF: %w", err)
	}
	_, err = bsvec.ParsePubKey(w.SerializePubKey(), bsvec.S256())
	if err != nil {
		log.Println(trace.Alert("cannot parse").UTC().Add("wif", wifkey).Error(err).Append(tr))
		return "", err
	}
	// fmt.Printf("pubk: %s\n", string(pubk.SerializeCompressed()))
	// fmt.Printf("pubk ser: %s\n", string(w.SerializePubKey()))
	add, err := bsvutil.NewAddressPubKey(w.SerializePubKey(), &chaincfg.MainNetParams)
	if err != nil {
		log.Println(trace.Alert("cannot generate address from WIF").UTC().Add("wif", wifkey).Error(err).Append(tr))
		return "", fmt.Errorf("cannot generate address from WIF: %w", err)
	}
	return add.EncodeAddress(), nil

}
