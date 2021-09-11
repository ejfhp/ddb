package ddb_test

import (
	"fmt"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestKeygen2_New(t *testing.T) {
	nums := []int{3567, 15}
	phrases := []string{
		"tanto va la gatta al lardo che ci lascia lo zampino",
		"ciao",
	}
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
		t.Logf("password: '%s'\n", string(pass[:]))
		if len(pass) != 32 {
			t.Logf("wrong password: '%s'", string(pass[:]))
			t.Fail()
		}
		t.Logf("WIF: %s\n", wif)
	}
}

func BenchmarkKeygen2_ManyWIF(b *testing.B) {
	template := "this is the phrase number %d, let's hope"
	for i := 0; i < b.N; i += 1791 {
		ph := fmt.Sprintf(template, i)
		k, err := ddb.NewKeygen2(i, ph)
		if err != nil {
			b.Logf("cannot generate Keygen2 %s: %v", ph, err)
			b.Fail()
		}
		wif, err := k.WIF()
		if err != nil {
			b.Logf("cannot generate WIF %s: %v", ph, err)
			b.FailNow()
		}
		address, err := ddb.AddressOf(wif)
		if err != nil {
			b.Logf("cannot get address %s %s: %v", wif, ph, err)
			b.FailNow()
		}
		b.Logf("Phrase: %s  WIF: %s  Address: %s \n", ph, wif, address)
	}
}

func TestKeygen2_Keys(t *testing.T) {
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
