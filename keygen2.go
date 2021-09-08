package ddb

import (
	"fmt"

	"golang.org/x/crypto/sha3"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
)

type Keygen2 struct {
	num    int
	phrase string
}

func NewKeygen2(number int, phrase string) (*Keygen2, error) {
	if len(phrase) < MinPhraseLen {
		return nil, fmt.Errorf("secret phrase should be longer than %d chars", MinPhraseLen)
	}
	return &Keygen2{num: number, phrase: phrase}, nil
}

func (k *Keygen2) Describe() {
	fmt.Printf("NUM: %d\n", k.num)
	fmt.Printf("PHRASE: %s\n", k.phrase)
}

func (k *Keygen2) WIF() (string, error) {
	var hash = sha3_256([]byte(k.phrase), k.num)
	priv, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), hash)
	wif, err := bsvutil.NewWIF(priv, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", fmt.Errorf("cannot generate WIF: %v", err)
	}
	return wif.String(), nil
}

func (k *Keygen2) Password() [32]byte {
	var password [32]byte
	copy(password[:], []byte(k.phrase))
	return password
}

func sha3_256(word []byte, repeat int) []byte {
	start := sha3.Sum256(word)
	var out [32]byte
	copy(out[:], start[:32])
	for i := 0; i < repeat; i++ {
		out = sha3.Sum256(out[:])
	}
	return out[:]
}
