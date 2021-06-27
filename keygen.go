package ddb

import (
	"fmt"

	"crypto/sha256"

	"golang.org/x/crypto/sha3"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
)

const (
	MIN_PHRASE_LEN = 3
	NUM_CONFS      = 4
	NUM_WORDS      = 3
)

var conf_numbers = []int{2, 3, 5, 7}

var hasher map[int64]Hasher = map[int64]Hasher{
	0: sha3_256_1,
	1: sha3_256_2,
	2: sha3_256_2,
	3: sha3_384_1,
	4: sha256_256_2,
	5: sha3_384_2,
	6: sha256_256_1,
}

type Keygen struct {
	num    int
	phrase string
	confs  []int
	words  [][]byte
}

func NewKeygen(num int, phrase string) (*Keygen, error) {
	if len(phrase) < MIN_PHRASE_LEN {
		return nil, fmt.Errorf("secret phrase should be longer than %d chars", MIN_PHRASE_LEN)
	}
	cns := []int{}
	for i := 0; i < len(conf_numbers); i += 1 {
		cns = append(cns, num%conf_numbers[i])
	}
	wLen := len(phrase) / NUM_WORDS
	ws := [][]byte{}
	for i := 0; i < NUM_WORDS; i += 1 {
		ws = append(ws, []byte(phrase[i*wLen:(i+1)*wLen]))
	}
	ws[NUM_WORDS-1] = append(ws[NUM_WORDS-1], phrase[(NUM_WORDS)*wLen:]...)
	return &Keygen{num: num, phrase: phrase, confs: cns, words: ws}, nil
}

func (k *Keygen) Words() [][]byte {
	return k.words
}

func (k *Keygen) Configs() []int {
	return k.confs
}

func (k *Keygen) Describe() {
	fmt.Printf("NUM: %d\n", k.num)
	fmt.Printf("PHRASE: %s\n", k.phrase)
	fmt.Printf("CONFS, 1:%d 2:%d 3:%d 4:%d\n", k.confs[0], k.confs[1], k.confs[2], k.confs[3])
	fmt.Printf("WORDS, 1:'%s' 2:'%s' 3:'%s'\n", k.words[0], k.words[1], k.words[2])
}

func (k *Keygen) MakeWIF() (string, error) {
	var hash []byte
	for i := 0; i < NUM_CONFS; i += 1 {
		h := hasher[1] //k.confs[i]
		hash = h(k.words, k.num, hash)
	}
	priv, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), hash)
	wif, err := bsvutil.NewWIF(priv, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", fmt.Errorf("cannot generate WIF: %v", err)
	}
	return wif.String(), nil
}

type Hasher func(words [][]byte, repeat int, hash []byte) []byte

func sha3_256_1(words [][]byte, repeat int, hash []byte) []byte {
	var out [32]byte
	in := hash
	for i := 0; i < repeat; i += 1 {
		for j := 0; j < NUM_WORDS; j += 1 {
			in = append(in, words[j]...)
			out = sha3.Sum256(in)
			copy(in, out[:])
		}
	}
	return out[:]
}

func sha3_256_2(words [][]byte, repeat int, hash []byte) []byte {
	var out [32]byte
	in := hash
	for i := 0; i < repeat; i += 1 {
		for j := 0; j < NUM_WORDS; j += 1 {
			in = append(words[j], in...)
			out = sha3.Sum256(in)
			copy(in, out[:])
		}
	}
	return out[:]
}

func sha3_384_1(words [][]byte, repeat int, hash []byte) []byte {
	var out [48]byte
	in := hash
	for i := 0; i < repeat; i += 1 {
		for j := 0; j < NUM_WORDS; j += 1 {
			in = append(in, words[j]...)
			out = sha3.Sum384(in)
			copy(in, out[:])
		}
	}
	return out[:]
}

func sha3_384_2(words [][]byte, repeat int, hash []byte) []byte {
	var out [48]byte
	in := hash
	for i := 0; i < repeat; i += 1 {
		for j := 0; j < NUM_WORDS; j += 1 {
			in = append(words[j], in...)
			out = sha3.Sum384(in)
			copy(in, out[:])
		}
	}
	return out[:]
}

func sha256_256_1(words [][]byte, repeat int, hash []byte) []byte {
	var out [32]byte
	in := hash
	for i := 0; i < repeat; i += 1 {
		for j := 0; j < NUM_WORDS; j += 1 {
			in = append(in, words[j]...)
			out = sha256.Sum256(in)
			copy(in, out[:])
		}
	}
	return out[:]
}

func sha256_256_2(words [][]byte, repeat int, hash []byte) []byte {
	var out [32]byte
	in := hash
	for i := 0; i < repeat; i += 1 {
		for j := 0; j < NUM_WORDS; j += 1 {
			in = append(words[j], in...)
			out = sha256.Sum256(in)
			copy(in, out[:])
		}
	}
	return out[:]
}
