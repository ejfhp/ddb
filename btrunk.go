package ddb

import (
	"crypto/sha256"
	"fmt"

	"github.com/bitcoinsv/bsvd/bsvec"
	"github.com/bitcoinsv/bsvd/chaincfg"
	"github.com/bitcoinsv/bsvutil"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Results struct {
	Cost  Satoshi
	TXIDs []string
}

type BTrunk struct {
	BitcoinWIF string
	BitcoinAdd string
	Blockchain *Blockchain
}

func (bt *BTrunk) GenerateKeyAndAddress(password [32]byte) (string, string, error) {
	keySeed := []byte{}
	keySeed = append(keySeed, []byte(bt.BitcoinAdd)...)
	keySeed = append(keySeed, password[:]...)
	keySeedHash := sha256.Sum256(keySeed)
	key, _ := bsvec.PrivKeyFromBytes(bsvec.S256(), keySeedHash[:])
	fbwif, err := bsvutil.NewWIF(key, &chaincfg.MainNetParams, true)
	if err != nil {
		return "", "", fmt.Errorf("error while generating key: %v", err)
	}
	fbWIF := fbwif.String()
	fbAdd, err := AddressOf(fbWIF)
	if err != nil {
		return "", "", fmt.Errorf("error while generating address: %v", err)
	}
	return fbWIF, fbAdd, nil

}

func (bt *BTrunk) TXOfBranchedEntry(wif, address string, password [32]byte, entry *Entry, header string, maxAmountToSpend Satoshi, simulate bool) ([]*DataTX, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "SameKeyFBranch")
	fBranch, err := bt.newFBranch(wif, address, password)
	if err != nil {
		trail.Println(trace.Debug("error while generating new FBranch").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while generating new FBranch: %v", err)
	}
	metaEntry := NewMetaEntry(entry)
	metaEntryData, err := metaEntry.Encrypt(password)
	if err != nil {
		trail.Println(trace.Debug("error while encrypting metaEntry").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while encrypting metaEntry: %v", err)
	}
	utxo, err := bt.getUTXOs(simulate)
	if err != nil {
		trail.Println(trace.Debug("error while getting UTXOs").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while getting UTXOs: %v", err)
	}
	//Fee for initial transaction of moving fund to FBranch and casting metaEntry
	mefee, err := bt.Blockchain.EstimateDataTXFee(len(utxo), metaEntryData, header)
	if err != nil {
		trail.Println(trace.Debug("error while estimating metaEntry TX fee").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while estimating metaEntry TX fee: %v", err)
	}
	//Fee for casting Entry
	efee, err := fBranch.EstimateEntryFee(header, entry)
	if err != nil {
		trail.Println(trace.Debug("error while estimating entry TXs fee").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while estimating entry TXs fee: %v", err)
	}
	//Fee to bring back remaining fund to BTrunk address
	finfee, err := bt.Blockchain.EstimateStandardTXFee(1)
	if err != nil {
		trail.Println(trace.Debug("error while estimating final TX fee").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while estimating final TX fee: %v", err)
	}
	totalFee := mefee.Add(efee)
	totalFee = totalFee.Add(finfee)
	fmt.Printf("BTRUNK METAE FEE: %d\n", mefee)
	fmt.Printf("BTRUNK ENTYT FEE: %d\n", efee)
	fmt.Printf("BTRUNK FINAL FEE: %d\n", finfee)
	fmt.Printf("BTRUNK TOTAL FEE: %d\n", totalFee)
	if totalFee > maxAmountToSpend {
		trail.Println(trace.Debug("given amount is less than estimated required fee").Append(tr).UTC().Add("fee/amount", fmt.Sprintf("%d/%d", totalFee, maxAmountToSpend)))
		return nil, fmt.Errorf("given amount (%d) is less than estimated required fee (%d)", maxAmountToSpend, totalFee)
	}
	allTXs := make([]*DataTX, 0)
	//First TX with metaEntry
	meTX, err := NewDataTX(bt.BitcoinWIF, fBranch.BitcoinAdd, bt.BitcoinAdd, utxo, maxAmountToSpend, mefee, metaEntryData, header)
	if err != nil {
		trail.Println(trace.Debug("error while making metaEntry DataTX").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while making metaEntry DataTX: %v", err)
	}
	allTXs = append(allTXs, meTX)
	//Enrtry TXs
	entryTXs, err := fBranch.ProcessEntry(header, entry, simulate)
	if err != nil {
		trail.Println(trace.Debug("error while making entry DataTXs").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while making entry DataTXs: %v", err)
	}
	allTXs = append(allTXs, entryTXs...)
	lastTX := entryTXs[len(entryTXs)-1]
	//Final transaction to move change back to BTrunk wallet
	_, lout, _, err := lastTX.TotInOutFee()
	if err != nil {
		trail.Println(trace.Debug("error getting output from last entity TX").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error getting output from last eintity TX: %v", err)
	}
	finTX, err := NewDataTX(fBranch.BitcoinWIF, bt.BitcoinAdd, bt.BitcoinAdd, lastTX.UTXOs(), lout, finfee, nil, "")
	if err != nil {
		trail.Println(trace.Debug("error while making final DataTX").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while making final DataTX: %v", err)
	}
	allTXs = append(allTXs, finTX)
	return allTXs, nil
}

func (bt *BTrunk) newFBranch(wif, address string, password [32]byte) (*FBranch, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "NewFBranch")
	trail.Println(trace.Debug("generating new FBranch").UTC().Append(tr).Add("address", address))
	fb := FBranch{BitcoinWIF: wif, BitcoinAdd: address, Password: password, Blockchain: bt.Blockchain}
	return &fb, nil
}

func (bt *BTrunk) getUTXOs(simulate bool) ([]*UTXO, error) {
	tr := trace.New().Source("btrunk.go", "BTrunk", "getUTXOs")
	trail.Println(trace.Info("getting green bud UTXO of the btrunk").Append(tr).UTC())
	var utxo []*UTXO
	var err error
	if simulate {
		utxo = bt.Blockchain.GetFakeUTXO()
	} else {
		utxo, err = bt.Blockchain.GetUTXO(bt.BitcoinAdd)
		if err != nil {
			trail.Println(trace.Alert("error getting UTXO").UTC().Add("address", bt.BitcoinAdd).Error(err).Append(tr))
			return nil, fmt.Errorf("error getting UTXO for address %s: %w", bt.BitcoinAdd, err)
		}
	}
	return utxo, nil
}

//DownloadAll saves locally all the files connected to the address. Return the number of entries saved.
// func (bt *BTrunk) DowloadAll(outPath string) (int, error) {
// 	tr := trace.New().Source("btrunk.go", "BTrunk", "DownloadAll")
// 	trail.Println(trace.Info("download all entries locally").Add("outpath", outPath).UTC().Append(tr))
// 	history, err := bt.ListHistory(bt.BitcoinAdd)
// 	if err != nil {
// 		trail.Println(trace.Alert("error getting address history").UTC().Error(err).Append(tr))
// 		return 0, fmt.Errorf("error getting address history: %w", err)
// 	}
// 	entries, err := bt.RetrieveAndExtractEntries(history)
// 	if err != nil {
// 		trail.Println(trace.Alert("error retrieving entries").UTC().Error(err).Append(tr))
// 		return 0, fmt.Errorf("error retrieving entries: %w", err)
// 	}
// 	for _, e := range entries {
// 		ioutil.WriteFile(filepath.Join(outPath, e.Name), e.Data, 0444)
// 	}
// 	return len(entries), nil

// }
