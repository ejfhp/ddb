package ddb_test

import (
	"fmt"
	"testing"

	"github.com/ejfhp/ddb"
)

func TestNewKeygen1(t *testing.T) {
	nums := []int{1567}
	phrases := []string{"tanto va la gatta al lardo che ci lascia lo zampino"}
	for i, n := range nums {
		k, err := ddb.NewKeygen1(n, phrases[i])
		if err != nil {
			t.Logf("cannot generate Keygen: %v", err)
			t.Fail()
		}
		k.Describe()

	}
}

func TestKeygen1WIF(t *testing.T) {
	nums := []int{3567}
	phrases := []string{"tanto va la gatta al lardo che ci lascia lo zampino"}
	for i, n := range nums {
		k, err := ddb.NewKeygen1(n, phrases[i])
		if err != nil {
			t.Logf("cannot generate Keygen1: %v", err)
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

func BenchmarkKeygen1ManyWIF(b *testing.B) {
	template := "this is the phrase number %d, let's hope"
	for i := 0; i < b.N; i++ {
		ph := fmt.Sprintf(template, i)
		k, err := ddb.NewKeygen1(i, ph)
		if err != nil {
			b.Logf("cannot generate Keygen1 %s: %v", ph, err)
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

func TestKeygen1Keys(t *testing.T) {
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
		"L48cWSssxbFnRuuJCVes9NEYP1W987kfpSgWG2RKSaZtcs6iCHpT",
		"KzxKMJoJ13Ug2E8mBb9npbqavs9hbX3rZ3XPq3jBNUriQNk5rMUc",
		"Kyypokm7KphGVa6QdqpWM4bkdTocQmD6f2waMBREcVq9UJKHow3o",
		"KzLEePeTR2utHtBLfoPRjf7hJeDzBodnfApN8WFb4gaEkneCP7KP",
		"L5n7n4ntJD3YyqUtcaekyqHZiv5nB71yhZE5SRzwWwtQocqEgwiv",
		"L4JnikU8C8z8nJgipUEAbwQfqCRW19FhpXs8cWnw25mYjjVu32jC",
	}
	for i, n := range nums {
		k, err := ddb.NewKeygen1(n, phrases[i])
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
