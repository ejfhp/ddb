package keys

import (
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/sha3"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
)

type Keygen2 struct {
	initialized bool
	num         int
	phrase      string
}

func (k *Keygen2) Init(number int, phrase string) error {
	if len(phrase) < MinPhraseLen {
		return fmt.Errorf("secret phrase should be longer than %d chars", MinPhraseLen)
	}
	k.num = number
	k.phrase = phrase
	k.initialized = true
	return nil
}

func (k *Keygen2) Describe() {
	if !k.initialized {
		fmt.Printf("NOT INITIALIZED\n")
	}
	fmt.Printf("Keygen ver. 2\n")
	fmt.Printf("NUM: %d\n", k.num)
	fmt.Printf("PHRASE: %s\n", k.phrase)
}

func (k *Keygen2) WIF() (string, error) {
	if !k.initialized {
		return "", fmt.Errorf("keygen not initialized")
	}
	var hash = sha3_256_2([]byte(k.phrase), k.num)
	priv, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), hash)
	wif, err := bsvutil.NewWIF(priv, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", fmt.Errorf("cannot generate WIF: %v", err)
	}
	return wif.String(), nil
}

func (k *Keygen2) Password() ([32]byte, error) {
	if !k.initialized {
		return [32]byte{}, fmt.Errorf("keygen not initialized")
	}
	var hash = sha3.Sum256([]byte(k.phrase))
	text := hex.EncodeToString(hash[:])
	var password [32]byte
	copy(password[:], []byte(text))
	return password, nil
}

func sha3_256_2(word []byte, repeat int) []byte {
	start := sha3.Sum256(word)
	var out [32]byte
	copy(out[:], start[:])
	for i := 0; i < repeat; i++ {
		out = sha3.Sum256(out[:])
	}
	return out[:]
}
