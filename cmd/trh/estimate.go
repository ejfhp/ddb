package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdEstimate(args []string) error {
	tr := trace.New().Source("estimate.go", "", "cmdEstimate")
	flagset, options := newFlagset(commands["estimate"])
	fmt.Printf("cmdEstimate flags: %v\n", args[2:])
	err := flagset.Parse(args[2:])
	if err != nil {
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	if flagHelp {
		printHelp("estimate")
		return nil
	}
	opt, ok := areFlagConsistent(flagset, options)
	if !ok {
		return fmt.Errorf("flag combination invalid")
	}
	switch opt {
	case "file":
		woc := ddb.NewWOC()
		taal := ddb.NewTAAL()
		blockchain := ddb.NewBlockchain(taal, woc, nil)
		btrunk := &ddb.BTrunk{BitcoinWIF: ddb.SampleKey, BitcoinAdd: ddb.SampleAddress, Blockchain: blockchain}
		lff := strings.Split(flagLabels, ",")
		labels := []string{}
		for _, l := range lff {
			labels = append(labels, strings.TrimSpace(strings.ToLower(l)))
		}
		ent, err := ddb.NewEntryFromFile(filepath.Base(flagFile), flagFile, labels, flagNotes)
		if err != nil {
			trail.Println(trace.Alert("failed to generate entry from file").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to generate entry from file: %w", err)
		}
		password := passwordtoBytes(flagPassword)
		bWIF, bAdd, err := btrunk.GenerateKeyAndAddress(password)
		if err != nil {
			trail.Println(trace.Alert("failed to generate branch key and address").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to generate branch key and address: %w", err)
		}
		txs, err := btrunk.TXOfBranchedEntry(bWIF, bAdd, passwordtoBytes(""), ent, defaultHeader, ddb.Bitcoin(ddb.FakeTXValue).Satoshi(), true)
		if err != nil {
			trail.Println(trace.Alert("failed to generate txs for entry").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to generate txs for entry: %w", err)
		}
		totFee := ddb.Satoshi(0)
		for i, t := range txs {
			_, _, fee, err := t.TotInOutFee()
			if err != nil {
				trail.Println(trace.Alert("failed to get fee from tx").Append(tr).UTC().Error(err))
				return fmt.Errorf("failed to get fee from tx num %d: %w", i, err)
			}
			totFee = totFee.Add(fee)
		}
		fmt.Printf("Estimated fee: %d satoshi\n", totFee)
		fmt.Printf("Estimated traffic: %d tx\n", len(txs))
	default:
		return fmt.Errorf("flag combination invalid")
	}
	return nil
}
