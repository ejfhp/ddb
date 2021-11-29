package ddb

import (
	"fmt"

	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/satoshi"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Results struct {
	Cost  satoshi.Satoshi
	TXIDs []string
}

type BTrunk struct {
	key        string
	address    string
	password   string
	passBytes  [32]byte
	blockchain *Blockchain
}

func NewBTrunk(key, address, password string, blockchain *Blockchain) *BTrunk {
	btrunk := BTrunk{
		key:        key,
		address:    address,
		password:   password,
		passBytes:  keys.StringToPassword(password),
		blockchain: blockchain,
	}
	return &btrunk
}

//TXOfBranchedEntry generate all the transactions needed to store the given entry. BranchKey (WIF) and branchAddress must be generated through BTrunk.GenerateKeyAndAddress().
func (bt *BTrunk) TXOfBranchedEntry(node *keys.Node, entry *Entry, header string, maxAmountToSpend satoshi.Satoshi, simulate bool) ([]*DataTX, error) {
	fBranch, err := bt.newFBranch(node.Key(), node.Address(), node.Password())
	if err != nil {
		return nil, fmt.Errorf("error while generating new FBranch: %v", err)
	}
	metaEntry := NewMetaEntry(node, entry)
	metaEntryData, err := metaEntry.Encrypt(bt.passBytes)
	if err != nil {
		return nil, fmt.Errorf("error while encrypting metaEntry: %v", err)
	}
	utxo, err := bt.getUTXOs(simulate)
	if err != nil {
		return nil, fmt.Errorf("error while getting UTXOs: %v", err)
	}
	//Fee for initial transaction of moving fund to FBranch and casting metaEntry
	mefee, err := bt.blockchain.EstimateDataTXFee(len(utxo), metaEntryData, header)
	if err != nil {
		return nil, fmt.Errorf("error while estimating metaEntry TX fee: %v", err)
	}

	//Fee to bring back remaining fund to BTrunk address
	finfee, err := bt.blockchain.EstimateStandardTXFee(1)
	if err != nil {
		return nil, fmt.Errorf("error while estimating final TX fee: %v", err)
	}

	allTXs := make([]*DataTX, 0)
	maxAmountToUse, err := maxAmountToSpend.Sub(mefee)
	if err != nil {
		return nil, fmt.Errorf("error while calculating amount to transfer to branched chain: %v", err)
	}
	//First TX with metaEntry
	meTX, err := NewDataTX(bt.key, fBranch.BitcoinAdd, bt.address, utxo, maxAmountToUse, mefee, metaEntryData, header)
	if err != nil {
		return nil, fmt.Errorf("error while making metaEntry DataTX: %v", err)
	}
	allTXs = append(allTXs, meTX)

	//Entry TXs, only the first UTXO has to be considered.
	entryTXs, err := fBranch.ProcessEntry(entry, meTX.UTXOs()[:1], header)
	if err != nil {
		return nil, fmt.Errorf("error while making entry DataTXs: %v", err)
	}
	allTXs = append(allTXs, entryTXs...)

	//Final transaction to move change back to BTrunk wallet
	lastTX := entryTXs[len(entryTXs)-1]
	if err != nil {
		return nil, fmt.Errorf("error getting output from last eintity TX: %v", err)
	}
	finTX, err := NewDataTX(fBranch.BitcoinWIF, bt.address, bt.address, lastTX.UTXOs()[:1], satoshi.EmptyWallet, finfee, nil, header)
	if err != nil {
		return nil, fmt.Errorf("error while making final DataTX: %v", err)
	}
	allTXs = append(allTXs, finTX)
	return allTXs, nil
}

func (bt *BTrunk) newFBranch(wif, address string, password [32]byte) (*FBranch, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "NewFBranch")
	trail.Println(trace.Debug("generating new FBranch").UTC().Append(tr).Add("address", address))
	fb := FBranch{BitcoinWIF: wif, BitcoinAdd: address, Password: password, Blockchain: bt.blockchain}
	return &fb, nil
}

func (bt *BTrunk) getUTXOs(simulate bool) ([]*UTXO, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "getUTXOs")
	trail.Println(trace.Info("getting green bud UTXO of the btrunk").Append(tr).UTC())
	var utxo []*UTXO
	var err error
	if simulate {
		utxo = bt.blockchain.GetFakeUTXO()
	} else {
		utxo, err = bt.blockchain.GetUTXO(bt.address)
		if err != nil {
			return nil, fmt.Errorf("error getting UTXO for address %s: %w", bt.address, err)
		}
	}
	// for _, u := range utxo {
	// 	fmt.Printf("UTXO: %d\n", u.Value.Satoshi())
	// }
	return utxo, nil
}

//ListEntries of the files stored with the given passwords.
func (bt *BTrunk) ListEntries(password map[string][32]byte, cacheOnly bool) (map[string][]*MetaEntry, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "ListEntries")

	trail.Println(trace.Debug("listing transactions for main address").Append(tr).UTC().Add("address", bt.address))
	TXIDs, err := bt.blockchain.ListTXIDs(bt.address, cacheOnly)
	if err != nil {
		return nil, fmt.Errorf("error while listing BTrunk transactions: %v", err)
	}
	meList := map[string][]*MetaEntry{}
	for _, TXID := range TXIDs {
		tx, err := bt.blockchain.GetTX(TXID, cacheOnly)
		if err != nil {
			return nil, fmt.Errorf("error while getting BTrunk transaction: %v", err)
		}
		data, _, err := tx.Data()
		if err != nil {
			trail.Println(trace.Warning("error while getting transaction data").Append(tr).UTC().Error(err))
		}
		for pass, pwd := range password {
			if meList[pass] == nil {
				meList[pass] = []*MetaEntry{}
			}
			me, _ := MetaEntryFromEncrypted(pwd, data)
			if me != nil && me.Timestamp > 0 {
				meList[pass] = append(meList[pass], me)
			}
		}
	}
	return meList, nil
}
