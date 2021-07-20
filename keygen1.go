package ddb

import (
	"fmt"
	"math"

	"crypto/sha256"

	"golang.org/x/crypto/sha3"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

const (
	MinPhraseLen = 3
	NumConfs     = 4
	NumWords     = 3
)

var conf_numbers = []int{2, 3, 5, 7}

var hashers map[int64]hasher = map[int64]hasher{
	0: sha3_256_1,
	1: sha3_256_2,
	2: sha3_256_2,
	3: sha3_384_1,
	4: sha256_256_2,
	5: sha3_384_2,
	6: sha256_256_1,
}

type Keygen1 struct {
	num    int
	phrase string
	confs  []int
	words  [][]byte
}

func NewKeygen1(number int, phrase string) (*Keygen1, error) {
	tr := trace.New().Source("keygen1.go", "Keygen1", "NewKeygen1")
	log.Println(trace.Debug("new keygen1").UTC().Append(tr))
	if len(phrase) < MinPhraseLen {
		return nil, fmt.Errorf("secret phrase should be longer than %d chars", MinPhraseLen)
	}
	cns := []int{}
	for i := 0; i < len(conf_numbers); i++ {
		cns = append(cns, number%conf_numbers[i])
	}
	wLen := len(phrase) / NumWords
	ws := [][]byte{}
	for i := 0; i < NumWords; i++ {
		ws = append(ws, []byte(phrase[i*wLen:(i+1)*wLen]))
	}
	ws[NumWords-1] = append(ws[NumWords-1], phrase[(NumWords)*wLen:]...)
	num := math.Abs(float64(number)) + 3
	n := int(math.Ceil((math.Log(float64(num)) * 100)))
	log.Println(trace.Debug("number of iteraction").UTC().Add("n", fmt.Sprintf("%d", n)).Append(tr))
	return &Keygen1{num: n, phrase: phrase, confs: cns, words: ws}, nil
}

func (k *Keygen1) Describe() {
	fmt.Printf("NUM: %d\n", k.num)
	fmt.Printf("PHRASE: %s\n", k.phrase)
	fmt.Printf("CONFS, 1:%d 2:%d 3:%d 4:%d\n", k.confs[0], k.confs[1], k.confs[2], k.confs[3])
	fmt.Printf("WORDS, 1:'%s' 2:'%s' 3:'%s'\n", k.words[0], k.words[1], k.words[2])
}

func (k *Keygen1) WIF() (string, error) {
	var hash []byte
	for i := 0; i < NumConfs; i++ {
		h := hashers[1] //k.confs[i]
		hash = h(k.words, k.num, hash)
	}
	priv, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), hash)
	wif, err := bsvutil.NewWIF(priv, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", fmt.Errorf("cannot generate WIF: %v", err)
	}
	return wif.String(), nil
}

func (k *Keygen1) Password() [32]byte {
	var password [32]byte
	copy(password[:], []byte(k.phrase)[:32])
	return password
}

type hasher func(words [][]byte, repeat int, hash []byte) []byte

func sha3_256_1(words [][]byte, repeat int, hash []byte) []byte {
	var out [32]byte
	in := hash
	for i := 0; i < repeat; i++ {
		for j := 0; j < NumWords; j++ {
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
	for i := 0; i < repeat; i++ {
		for j := 0; j < NumWords; j++ {
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
	for i := 0; i < repeat; i++ {
		for j := 0; j < NumWords; j++ {
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
	for i := 0; i < repeat; i++ {
		for j := 0; j < NumWords; j++ {
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
	for i := 0; i < repeat; i++ {
		for j := 0; j < NumWords; j++ {
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
	for i := 0; i < repeat; i++ {
		for j := 0; j < NumWords; j++ {
			in = append(words[j], in...)
			out = sha256.Sum256(in)
			copy(in, out[:])
		}
	}
	return out[:]
}
