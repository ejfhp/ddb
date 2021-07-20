package ddb_test

import (
	"fmt"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestNewKeygen2(t *testing.T) {
	nums := []int{1567}
	phrases := []string{"tanto va la gatta al lardo che ci lascia lo zampino"}
	for i, n := range nums {
		k, err := ddb.NewKeygen2(n, phrases[i])
		if err != nil {
			t.Logf("cannot generate Keygen2: %v", err)
			t.Fail()
		}
		k.Describe()

	}
}

func TestKeygen2WIF(t *testing.T) {
	nums := []int{3567}
	phrases := []string{"tanto va la gatta al lardo che ci lascia lo zampino"}
	for i, n := range nums {
		k, err := ddb.NewKeygen2(n, phrases[i])
		if err != nil {
			t.Logf("cannot generate Keygen2: %v", err)
			t.Fail()
		}
		wif, err := k.WIF()
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

func TestKeygen2ManyWIF(t *testing.T) {
	template := "this is the phrase number %d, let's hope"
	for i := 0; i < 200000; i += 1791 {
		ph := fmt.Sprintf(template, i)
		k, err := ddb.NewKeygen2(i, ph)
		if err != nil {
			t.Logf("cannot generate Keygen2 %s: %v", ph, err)
			t.Fail()
		}
		wif, err := k.WIF()
		if err != nil {
			t.Logf("cannot generate WIF %s: %v", ph, err)
			t.FailNow()
		}
		address, err := ddb.AddressOf(wif)
		if err != nil {
			t.Logf("cannot get address %s %s: %v", wif, ph, err)
			t.FailNow()
		}
		t.Logf("Phrase: %s  WIF: %s  Address: %s \n", ph, wif, address)
	}
}

func TestKeygen2Keys(t *testing.T) {
	nums := []int{3567, 0, 12, 100, 1001, 1}
	phrases := []string{
		"tanto va la gatta al lardo che ci lascia lo zampino",
		"abc",
		"cippirimerlo",
		"un due tre stella",
		"giro giro tondo gira tutto il mondo",
		"this is 1 test to show how to use Maestrale",
	}
	wifs := []string{
		"Kwuo7yZQsmnAEkypjQcWSyLxGDQgd3M4pDiJc7e3TcP63A877iEo",
		"KyBcTsTPAtSanCsEUcTWrsCS7uC7sdygmGYnDpy4VdhQPtJJTvo8",
		"L1vFyhAVr2VquVQXnbExNZP1e58W9oKaNiUdNTTWCuTNLAQ2VYLy",
		"L52iTUo9NLz9NjdfetCXCgufQpFJ5wE9M2dtXMREsDxKeYLd7JRf",
		"KxemArUPA55L3RR8Tbnry5yWoPaobKq4JxBFJnVcuCkDtiMq6yed",
		"KwVye3MjpGWM6o11swJJbaErapfVn3WP3JCiXZjEHWRAqsDTgDss",
	}
	for i, n := range nums {
		k, err := ddb.NewKeygen2(n, phrases[i])
		if err != nil {
			t.Logf("cannot generate Keygen: %v", err)
			t.Fail()
		}
		wif, err := k.WIF()
		if err != nil {
			t.Logf("cannot generate WIF: %v", err)
			t.Fail()
		}
		_, err = ddb.DecodeWIF(wif)
		if err != nil {
			t.Logf("cannot get address of WIF: %v", err)
			t.Fail()
		}
		if wif != wifs[i] {
			t.Logf("Unexpected WIF: %s != %s\n", wif, wifs[i])
			t.Fail()
		}
	}
}
