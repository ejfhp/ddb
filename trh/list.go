package trh

import (
	"fmt"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/ddb/keys"
)

//ListAll returns an array of all the entries found
func (t *TRH) ListAll(keystore *keys.Keystore) ([]*ddb.MetaEntry, error) {
	var mEntries []*ddb.MetaEntry
	mEntries, err := t.btrunk.ListEntries(false)
	if err != nil {
		return nil, fmt.Errorf("error while listing MetaEntry for password: %w", err)
	}
	return mEntries, nil
}
