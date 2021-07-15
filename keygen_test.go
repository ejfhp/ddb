package ddb_test

import (
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
		if len(k.Words()) != ddb.NumWords {
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
		pass := k.Password()
		if pass[0] != 't' || pass[31] != 'c' {
			t.Logf("wrong password: %v", err)
			t.Fail()
		}
		t.Logf("WIF: %s\n", wif)
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
		"KwsrmwtfphXNGyQq9mo6GtTjMjajrzGQYuoCYWA7XMBRtrEDwgtD",
		"KwDiBf89QgGbjEhKnhXJuH7LrciVrZi3qYjgd9M7rFU73Nd2Mcv1",
		"L2oEpyKF89Mn7XJJuyN7hoY9vU4ugaDAeuLf2nHx6CpquBcNwYve",
		"KyizRQyrAv4ErAgT76karoyBME2t497J7GfuAfrC96sSfENRrKdy",
		"L5EXzcBe2EaTZWhjySZmm2A5m4RCN5Qp5wQqyHrbfGtCKfZBVf4y",
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
