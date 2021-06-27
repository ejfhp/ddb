package ddb_test

import (
	"fmt"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestNewKeygen(t *testing.T) {
	nums := []int{1567}
	phrases := []string{"tanto va la gatta al lardo che ci lascia lo zampino"}
	for i, n := range nums {
		k, err := ddb.NewKeygen(n, phrases[i])
		if err != nil {
			t.Logf("cannot generate Keygen: %v", err)
			t.Fail()
		}
		if len(k.Words()) != ddb.NUM_WORDS {
			t.Logf("wrong num of words: %d", len(k.Words()))
			t.Fail()

		}
		ph := string(append(append(k.Words()[0], k.Words()[1]...), k.Words()[2]...))
		if ph != phrases[i] {
			t.Logf("words don't fit with phrase: '%s' != '%s'", ph, phrases[i])
			t.Fail()
		}
		k.Describe()

	}
}

func TestMakeWIF(t *testing.T) {
	nums := []int{3567}
	phrases := []string{"tanto va la gatta al lardo che ci lascia lo zampino"}
	for i, n := range nums {
		k, err := ddb.NewKeygen(n, phrases[i])
		if err != nil {
			t.Logf("cannot generate Keygen: %v", err)
			t.Fail()
		}
		wif, err := k.MakeWIF()
		if err != nil {
			t.Logf("cannot generate WIF: %v", err)
			t.Fail()
		}
		fmt.Printf("WIF: %s\n", wif)
	}
}
