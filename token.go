package ddb

import (
	"encoding/json"
	"math"
)

const EmptyWallet = Satoshi(math.MaxUint64)

type Token interface {
	Bitcoin() Bitcoin
	Satoshi() Satoshi
}

type Bitcoin float64

func (b Bitcoin) Satoshi() Satoshi {
	return Satoshi(uint64(math.Round(float64(b) * 100000000)))
}

func (b Bitcoin) Bitcoin() Bitcoin {
	return b
}

func (b Bitcoin) Sub(t Token) Token {
	return Satoshi(b.Satoshi() - t.Satoshi())
}

func (b Bitcoin) Add(t Token) Token {
	return Satoshi(b.Satoshi() + t.Satoshi())
}

func (b *Bitcoin) UnmarshalJSON(bytes []byte) error {
	var val float64
	if err := json.Unmarshal(bytes, &val); err != nil {
		return err
	}
	bit := Bitcoin(val)
	*b = bit
	return nil
}

func (b *Bitcoin) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(*b))
}

type Satoshi uint64

func (s Satoshi) Bitcoin() Bitcoin {
	return Bitcoin(float64(s) / 100000000)
}
func (s Satoshi) Satoshi() Satoshi {
	return s
}

func (s Satoshi) Sub(t Token) Satoshi {
	return Satoshi(s - t.Satoshi())
}

func (s Satoshi) Add(t Token) Satoshi {
	return Satoshi(s + t.Satoshi())
}

func (s *Satoshi) UnmarshalJSON(bytes []byte) error {
	var val uint64
	if err := json.Unmarshal(bytes, &val); err != nil {
		return err
	}
	sat := Satoshi(val)
	*s = sat
	return nil
}

func (s *Satoshi) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint64(*s))
}
