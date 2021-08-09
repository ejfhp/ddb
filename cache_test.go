package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
)

func TestNewTXCache(t *testing.T) {
	usercache, _ := os.UserCacheDir()
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "home dir", args: args{path: filepath.Join(usercache, "trh")}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ddb.NewTXCache(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTXCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.DirPath() != tt.args.path {
				t.Errorf("Path = %v, want %s", got.DirPath(), tt.args.path)
			}
		})
	}
}

func TestTXCache_StoreRetrieveTX(t *testing.T) {
	trail.SetWriter(os.Stdout)
	txid := "existingfaketxid"
	random := sha256.Sum256([]byte(time.Now().Format("Mon Jan 2 15:04:05 -0700 MST 2006")))
	ranid := string(hex.EncodeToString(random[:]))
	tx := []byte("jdjdlkajljflvajafldsa;gf;gjdlijljlkngdskjgfksdjfaksj;kjdsf;kajskjf")
	usercache, _ := os.UserCacheDir()

	cache, err := ddb.NewTXCache(filepath.Join(usercache, "trh"))
	if err != nil {
		t.Logf("failed to create cache: %v", err)
		t.FailNow()
	}

	//STORE
	err = cache.StoreTX(txid, tx)
	if err != nil {
		t.Logf("failed to store tx: %v", err)
		t.FailNow()
	}
	err = cache.StoreTX(ranid, tx)
	if err != nil {
		t.Logf("failed to store tx: %v", err)
		t.Fail()
	}

	//RETRIEVE
	rtx, err := cache.RetrieveTX(txid)
	if err != nil {
		t.Logf("failed to retrieve tx: %v", err)
		t.Fail()
	}
	if !reflect.DeepEqual(rtx, tx) {
		t.Logf("retrieved tx is wrong: %v", err)
		t.Fail()
	}
	rtx, err = cache.RetrieveTX(ranid)
	if err != nil {
		t.Logf("failed to retrieve random tx: %v", err)
		t.Fail()
	}
	if !reflect.DeepEqual(rtx, tx) {
		t.Logf("retrieved random tx is wrong: %v", err)
		t.Fail()
	}
	_, err = cache.RetrieveTX("notexists")
	if err != ddb.ErrNotCached {
		t.Logf("unexpected error for not existent tx: %v", err)
		t.Fail()
	}
}

func TestTXCache_StoreRetrieveTXID(t *testing.T) {
	trail.SetWriter(os.Stdout)
	address := "address"
	txid1 := "txid1"
	txid2 := "txid2"
	usercache, _ := os.UserCacheDir()

	cache, err := ddb.NewTXCache(filepath.Join(usercache, "trh"))
	if err != nil {
		t.Logf("failed to create cache: %v", err)
		t.FailNow()
	}

	//STORE
	err = cache.StoreTXIDs(address, []string{txid1})
	if err != nil {
		t.Logf("failed to store sourceoutput 1: %v", err)
		t.FailNow()
	}
	err = cache.StoreTXIDs(address, []string{txid2, txid1})
	if err != nil {
		t.Logf("failed to store sourceoutput 2: %v", err)
		t.Fail()
	}

	//RETRIEVE
	sos, err := cache.RetrieveTXIDs(address)
	if err != nil {
		t.Logf("failed to retrieve txid: %v", err)
		t.Fail()
	}
	if len(sos) != 2 {
		t.Logf("unexpected number of txids: %d", len(sos))
		t.Fail()
	}

	_, err = cache.RetrieveTXIDs("notexists")
	if err != ddb.ErrNotCached {
		t.Logf("unexpected error for not existent sourceoutput: %v", err)
		t.Fail()
	}
}

func TestTXCache_Clear(t *testing.T) {
	trail.SetWriter(os.Stdout)
	txid := "existingfaketxid"
	random := sha256.Sum256([]byte(time.Now().Format("Mon Jan 2 15:04:05 -0700 MST 2006")))
	ranid := string(hex.EncodeToString(random[:]))
	tx := []byte("jdjdlkajljflvajafldsa;gf;gjdlijljlkngdskjgfksdjfaksj;kjdsf;kajskjf")
	usercache, _ := os.UserCacheDir()
	cachepath := filepath.Join(usercache, "clear_trh")
	cache, err := ddb.NewTXCache(cachepath)
	if err != nil {
		t.Logf("failed to create cache: %v", err)
		t.FailNow()
	}

	//STORE
	err = cache.StoreTX(txid, tx)
	if err != nil {
		t.Logf("failed to store tx: %v", err)
		t.FailNow()
	}
	err = cache.StoreTX(ranid, tx)
	if err != nil {
		t.Logf("failed to store tx: %v", err)
		t.Fail()
	}

	size, err := cache.Size()
	if err != nil {
		t.Logf("failed to get cache size: %v", err)
		t.Fail()
	}
	if size != 2 {
		t.Logf("unexpected cache size: %d", size)
		t.Fail()
	}
	//CLEAR
	err = cache.Clear()
	if err != nil {
		t.Logf("failed to clear cache: %v", err)
		t.Fail()
	}
	_, err = os.Open(cachepath)
	if errors.Is(err, os.ErrNotExist) == false {
		t.Logf("unexpected error: %v", err)
		t.Fail()

	}

}
