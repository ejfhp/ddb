package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/ejfhp/ddb"
)

func flagsetDescribe(cmd string, args []string) []string {
	flagset := flag.NewFlagSet("describe", flag.ContinueOnError)
	flagset.BoolVar(&flagLog, "log", false, "true enables log output")
	flagset.BoolVar(&flagHelp, "help", false, "print help")
	flagset.BoolVar(&flagHelp, "h", false, "print help")
	switch cmd {
	case commandStore:
		flagset.StringVar(&flagFilename, "file", "", "path of file to store onchain")
	case commandRetrieve:
		flagset.StringVar(&flagOutputDir, "outdir", "", "path of folder where to save retrived files")
	}
	flagset.Parse(args)
	if flagHelp {
		printHelp(flagset)
	}
	//fmt.Printf("file: %s\n", flagFilename)
	logOn(flagLog)
	return flagset.Args()
}

func cmdDescribe(args []string) error {
	argsLeft := flagset(commandDescribe, args)

	passphrase, passnum, err := checkPassphrase(argsLeft)
	if err != nil {
		return fmt.Errorf("error checking passphrase: %w", err)
	}
	fmt.Printf("\nSecret configuration:\n")
	fmt.Printf("  passnum:    '%d'\n", passnum)
	fmt.Printf("  passphrase: '%s'\n", passphrase)
	logbook, err := newLogbook(passphrase, passnum)
	if err != nil {
		return fmt.Errorf("error creating Logbook: %w", err)
	}
	fmt.Printf("  pasword:     '%s'\n", logbook.EncodingPassword())
	fmt.Printf("\nBitcoin configuration:\n")
	fmt.Printf("  Bitcoin Key (WIF): '%s'\n", logbook.BitcoinPrivateKey())
	if runtime.GOOS != "windows" {
		fmt.Printf("\n")
		ddb.PrintQRCode(os.Stdout, logbook.BitcoinPrivateKey())
		fmt.Printf("\n")
	}
	fmt.Printf("  Bitcoin Address  : '%s'\n", logbook.BitcoinPublicAddress())
	if runtime.GOOS != "windows" {
		fmt.Printf("\n")
		ddb.PrintQRCode(os.Stdout, logbook.BitcoinPublicAddress())
		fmt.Printf("\n")
	}

	history, err := logbook.ListHistory(logbook.BitcoinPublicAddress())
	if err != nil {
		return fmt.Errorf("error getting address history; %w", err)
	}
	fmt.Printf("Transaction History\n")
	if len(history) == 0 {
		fmt.Printf("this address has no history\n")
	}
	for i, tx := range history {
		fmt.Printf("%d: %s\n", i, tx)
	}
	return nil
}
