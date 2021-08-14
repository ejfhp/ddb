package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Retrieve struct {
	diary     *ddb.Diary
	outfolder string
}

func NewRetrieve(args []string, flagset *flag.FlagSet) (*Retrieve, error) {
	tr := trace.New().Source("retrieve.go", "Retrieve", "NewRetrieve")
	var outputDir string
	var address string
	var key string
	var password string

	passphrase, err := extractPassphrase(os.Args)
	if err != nil {
		fmt.Printf("passphrase not found\n")
	}
	if passphrase == "" && flagBitcoinAddress == "" && flagBitcoinKey == "" {
		trail.Println(trace.Alert("one of (passphrase, key, address) must be set").Append(tr).UTC())
		return nil, fmt.Errorf("one of (passphrase, key, address) must be set")
	}

	if flagPassword != "" && (flagBitcoinAddress != "" || flagBitcoinKey != "") {
		password = flagPassword
		if flagBitcoinAddress != "" {
			address = flagBitcoinAddress
		} else {
			address, err = ddb.AddressOf(flagBitcoinKey)
			if err != nil {
				return nil, fmt.Errorf("bitcoin key is invalid")
			}
		}
	}
	if passphrase != "" {
		key, password, err := processPassphrase(passphrase)
	}
	err := flagset.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("error while parsing args: %w", err)
	}
	if flagHelp {
		printHelp(flagset)
	}
	logOn(flagLog)
	outputDir := flagOutputDir
	if outputDir == "" {
		fmt.Printf("Output dir not set, using local flolder.\n")
		outputDir, _ = filepath.Abs(filepath.Dir(args[0]))
	}

	if flagBitcoinAddress == "" {

	}
	diarty := ddb.NewDiaryRO()
	retrieve := Retrieve{diary: diary, outfolder: outputDir}
	return &retrieve, nil

}

func (cr *Retrieve) Cmd(args []string) error {
	passphrase, num, err := checkPassphrase(argsLeft)
	if err != nil {
		return fmt.Errorf("error checking passphrase: %w", err)
	}
	logbook := newDiary(passphrase, num)
	if err != nil {
		return fmt.Errorf("error creating Logbook: %w", err)
	}

	n, err := logbook.DowloadAll(flagOutputDir)
	if err != nil {
		fmt.Errorf("error while retrieving files from address '%s' to floder '%s': %w", logbook.BitcoinPublicAddress(), flagOutputDir, err)
	}
	fmt.Printf("%d files has been retrived from '%s' to '%s'\n", n, logbook.BitcoinPublicAddress(), flagOutputDir)
	return nil
}
