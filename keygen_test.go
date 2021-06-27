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

func TestConsitencyMakeWIF(t *testing.T) {
	nums := []int{3567, 0, 12, 100, 1001}
	phrases := []string{
		"tanto va la gatta al lardo che ci lascia lo zampino",
		"abc",
		"cippirimerlo",
		"un due tre stella",
		"giro giro tondo gira tutto il mondo",
	}
	wifs := []string{
		"Kxyhq2q28TjHDtQsHw5iqG9tKkzD61T9kbRA6r8PGgtgw4DN1HhB",
		"KwDiBf89QgGbjEhKnhXJuH7LrciVrZi3qYjgd9M7rFU73Nd2Mcv1",
		"KyhADgji8DRRGL9Sozq2Tsr3mMiVqxNuKovM1seoi3jMQJenVmTE",
		"L5P2mJKM9MNtucYkGUT9LALtVzFPcBmmkiYeSkyPjorQ2QSFbBnj",
		"KzWJfnzfNxd26egrsRcALpNpvBBAXnzqkK5eyAUkGcAaPCaWc4nJ",
	}
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
		if wif != wifs[i] {
			t.Logf("Unexpected WIF: %s != %s\n", wif, wifs[i])
			t.Fail()
		}
	}
}
