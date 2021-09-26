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
		printHelp(flagset)
		return nil
	}
	opt := areFlagConsistent(flagset, options)
	// keystore, err := loadKeyStore()
	// if err != nil {
	// 	trail.Println(trace.Alert("error while loading keystore").Append(tr).UTC().Error(err))
	// 	return fmt.Errorf("error while loading keystore: %w", err)
	// }
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
		fee, err := btrunk.EstimateFeeOfBranchedEntry(ddb.PasswordFromString(flagPassword), ent, defaultHeader)
		if err != nil {
			trail.Println(trace.Alert("failed to estimate fee for file").Append(tr).UTC().Error(err))
			return fmt.Errorf("failed to estimate fee for file: %w", err)
		}
		fmt.Printf("Estimated Fee: %d satoshi\n", fee)
	default:
		return fmt.Errorf("flag combination invalid")
	}
	return nil
}
