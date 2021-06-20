package ddb

import (
	"encoding/json"
	"fmt"
	"math"

	log "github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
	"github.com/libsv/go-bt"
)

type Token interface{
	Bitcoin() Bitcoin
	Satoshi() Satoshi
}

type Bitcoin struct {
	value float64
}

func NewBitcoin(bitcoin float64) *Bitcoin {
	return &Bitcoin{value: bitcoin}
}

func (b Bitcoin) Satoshi() Satoshi {
	return Satoshi(uint64(math.Round(float64(b) * 100000000)))
}

func (b *Bitcoin) Value() float64 {
	return float64(b.satoshis) / 100000000
}

func (b *Bitcoin) Sub(s *Bitcoin) *Bitcoin {
	return FromSatoshis(b.satoshis - s.satoshis)
}

func (b *Bitcoin) Add(s *Bitcoin) *Bitcoin {
	return FromSatoshis(b.satoshis + s.satoshis)
}

func (b *Bitcoin) UnmarshalJSON(bytes []byte) error {
	var val interface{}
	if err := json.Unmarshal(bytes, &val); err != nil {
		return err
	}
	bit, okb := 
	sat, oks := val.(uint64)
	fmt.Printf("okb: %t   oks: %t\n", okb, oks)
	if oks == true {
		fmt.Printf("value in satoshi: %d\n", sat)
		b.SetSatoshis(sat)
	} else if okb == true {
		fmt.Printf("value in bitcoin: %f\n", bit)
		b.SetValue(bit)
	} else {
		return fmt.Errorf("bitcoin value unparsable")
	}
	return nil
}

func (b *Bitcoin) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Value())
}











type Satoshi uint64

func (s Satoshi) Bitcoin() Bitcoin {
	return Bitcoin(float64(b.satoshis) / 100000000)
}
func (s Satoshi) Satoshi() Satoshi {
	return s
}

func (b *Bitcoin) Sub(s *Bitcoin) *Bitcoin {
	return FromSatoshis(b.satoshis - s.satoshis)
}

func (b *Bitcoin) Add(s *Bitcoin) *Bitcoin {
	return FromSatoshis(b.satoshis + s.satoshis)
}

func (b *Bitcoin) UnmarshalJSON(bytes []byte) error {
	var val interface{}
	if err := json.Unmarshal(bytes, &val); err != nil {
		return err
	}
	bit, okb := 
	sat, oks := val.(uint64)
	fmt.Printf("okb: %t   oks: %t\n", okb, oks)
	if oks == true {
		fmt.Printf("value in satoshi: %d\n", sat)
		b.SetSatoshis(sat)
	} else if okb == true {
		fmt.Printf("value in bitcoin: %f\n", bit)
		b.SetValue(bit)
	} else {
		return fmt.Errorf("bitcoin value unparsable")
	}
	return nil
}

func (b *Bitcoin) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.Value())
}
