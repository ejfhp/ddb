package ddb

import (
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/sha3"
)

const (
	MIN_NUM        = 1000
	MIN_PHRASE_LEN = 15
	NUM_CONF       = 4
	NUM_WORDS      = 3
)

var hasher map[int64]Hasher = map[int64]Hasher{
	1: sha3_256_1,
	2: sha3_256_2,
	3: sha3_384,
}

type Keygen struct {
	num    int
	phrase string
	confs  []int
	words  [][]byte
}

func NewKeygen(num int, phrase string) (*Keygen, error) {
	if num < MIN_NUM {
		return nil, fmt.Errorf("Secret number should be greater than %d", MIN_NUM)
	}
	if len(phrase) < MIN_PHRASE_LEN {
		return nil, fmt.Errorf("Sectret phrase should be longer than %d chars", MIN_PHRASE_LEN)
	}
	conf1 := num % 2
	conf2 := num % 3
	conf3 := num % 5
	conf4 := num % 7

	wordLen := len(phrase) / NUM_WORDS
	word1 := []byte(phrase[:wordLen])
	word2 := []byte(phrase[wordLen : wordLen*2])
	word3 := []byte(phrase[wordLen*2:])
	return &Keygen{num: num, phrase: phrase, confs: []int{conf1, conf2, conf3, conf4}, words: [][]byte{word1, word2, word3}}, nil
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

func (k *Keygen) MakeWIF() string {
	var hash []byte
	for i := 0; i < NUM_CONF; i += 1 {
		h := hasher[1] //k.confs[i]
		hash = h(k.words, k.num, hash)
	}
	return hex.EncodeToString(hash)
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
			fmt.Println(hex.EncodeToString(out[:]))
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
			fmt.Println(hex.EncodeToString(out[:]))
		}
	}
	return out[:]
}

func sha3_384(words [][]byte, repeat int, hash []byte) []byte {
	var out [48]byte
	in := hash
	for i := 0; i < repeat; i += 1 {
		for j := 0; j < NUM_WORDS; j += 1 {
			in = append(in, words[j]...)
			out = sha3.Sum384(in)
			copy(in, out[:])
			fmt.Println(hex.EncodeToString(out[:]))
		}
	}
	return out[:]
}
