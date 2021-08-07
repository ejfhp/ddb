package ddb_test

import (
	"crypto/sha256"
	"encoding/hex"
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
			if got.Path() != tt.args.path {
				t.Errorf("Path = %v, want %s", got.Path(), tt.args.path)
			}
		})
	}
}

func TestTXCache_StoreRetrieve(t *testing.T) {
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
	err = cache.Store(txid, tx)
	if err != nil {
		t.Logf("failed to store tx: %v", err)
		t.FailNow()
	}
	err = cache.Store(ranid, tx)
	if err != nil {
		t.Logf("failed to store tx: %v", err)
		t.Fail()
	}

	//RETRIEVE
	rtx, err := cache.Retrieve(txid)
	if err != nil {
		t.Logf("failed to retrieve tx: %v", err)
		t.Fail()
	}
	if !reflect.DeepEqual(rtx, tx) {
		t.Logf("retrieved tx is wrong: %v", err)
		t.Fail()
	}
	rtx, err = cache.Retrieve(ranid)
	if err != nil {
		t.Logf("failed to retrieve random tx: %v", err)
		t.Fail()
	}
	if !reflect.DeepEqual(rtx, tx) {
		t.Logf("retrieved random tx is wrong: %v", err)
		t.Fail()
	}
	_, err = cache.Retrieve("notexists")
	if err != ddb.ErrTXNotExist {
		t.Logf("unexpected error for not existent tx: %v", err)
		t.Fail()
	}

}
