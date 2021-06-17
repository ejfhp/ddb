package ddb

import (
	"fmt"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

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
	t := trace.New().Source("keys.go", "", "AddressOf")
	w, err := bsvutil.DecodeWIF(wifkey)
	if err != nil {
		log.Println(trace.Alert("cannot decode WIF").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot decode WIF: %w", err)
	}
	fmt.Printf("compressed: %t\n", w.CompressPubKey)
	add, err := bsvutil.NewAddressPubKey(w.SerializePubKey(), &chaincfg.MainNetParams)

	// add, err := bscript.NewAddressFromPublicKeyHash(wif.SerialisePubKey(), true)

	if err != nil {
		log.Println(trace.Alert("cannot generate address from WIF").UTC().Error(err).Append(t))
		return "", fmt.Errorf("cannot generate address from WIF: %w", err)
	}
	return add.EncodeAddress(), nil

}
