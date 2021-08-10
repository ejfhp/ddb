package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
)

const (
	exitNoPassphrase = iota
	exitNoPassnum
	exitKeygenError
	exitLogbookError
	exitFileError
	exitStoreError
	commandDescribe = "describe"
	commandStore    = "store"
	commandRetrieve = "retrieve"
	commandEstimate = "estimate"
)

var (
	flagLog       bool
	flagHelp      bool
	flagFilename  string
	flagOutputDir string
)

func checkPassphrase(args []string) (string, int, error) {
	startidx := -1
	for i, t := range args {
		if t == "+" {
			startidx = i + 1
		}
	}
	if startidx < 0 || startidx >= len(args) {
		return "", 0, fmt.Errorf("passphrase is missing")
	}
	passphrase := strings.Join(args[startidx:], " ")
	passnum := 0
	reg, err := regexp.Compile("[^0-9 ]+")
	if err != nil {
		return "", 0, fmt.Errorf("error compiling regexp: %w", err)
	}
	phnum := reg.ReplaceAllString(passphrase, "")
	for _, n := range strings.Split(phnum, " ") {
		num, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			continue
		}
		if num < 0 {
			num = num * -1
		}
		passnum = int(num)
	}
	if passnum == 0 {
		return "", 0, fmt.Errorf("passphrase must contain a number")
	}
	return passphrase, passnum, nil
}

func keyGen(passphrase string, passnum int) (string, [32]byte, error) {
	keygen, err := ddb.NewKeygen2(passnum, passphrase)
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("error while building Keygen: %w", err)
	}
	wif, err := keygen.WIF()
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("error while generating bitcoin key: %w", err)
	}
	password := keygen.Password()
	return wif, password, nil
}

func newLogbook(passphrase string, passnum int) (*ddb.DDB, error) {
	wif, password, err := keyGen(passphrase, passnum)
	if err != nil {
		return nil, fmt.Errorf("error while generating the Bitcoin private key: %w", err)
	}
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc)
	logbook, err := ddb.NewLogbook(wif, password, blockchain)
	if err != nil {
		return nil, fmt.Errorf("error while creating a new Logbook: %w", err)
	}
	return logbook, nil
}

func logOn(on bool) {
	if on {
		trail.SetWriter(os.Stderr)
	}
}

func printMainHelp() {
	fmt.Printf(`
TRH (The Rabbit Hole)

TRH is a tool that let you store and retrieve files from the Bitcoin BSV blockchain.

To start you have to load the address that will be used to store your data onchain.
The address is generated starting from the passphrase given in input.
The passphrase has always to be put at the end of the command, after a plus (+) and can be 
written with or without double quotes ("). Pay attention to write the passphrase exactly 
every time that's the key to access the files onchain. The passphrase must contains a number.
The time it takes the generation of the address is proportional to the value of the number, I
suggest to not go over 9999999999.

To see the address to load and the corrisponding private key do:

>trh describe + <passphrase with a number 9999>

When the address has enough funds, you can store a file onchain. If the address has not enough 
funds the store will fail but the money will be spent anyway. 

To have a raw estimation of the necessary amount of BSV to cover fees do: 

>trh estimate -file <file path> + <passphrase with a number 9999>


To store a file do:

>trh store -file <file path> + <passphrase with a number 9999>

If all is fine, the transactions id of the generated transaction will be shown.

To retrieve the file from the blockchain do:

>trh retrieve -outdir <output folder> + <passphrase with a number 9999>


Options:
-log   true enables log output
-help  print help

Examples:

./trh describe -log + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
./trh estimate -file bitcoin.pdf -log + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
./trh store -file bitcoin.pdf -log + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
./trh retrieve -outdir /Users/diego/Desktop/ + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
`)
}

func printHelp(flagset *flag.FlagSet) {
	if flagset != nil {
		flagset.SetOutput(os.Stdout)
		flagset.PrintDefaults()
	}
	fmt.Printf("Main command: describe, store, retrieve.\n")
	os.Exit(0)
}

func flagset(cmd string, args []string) []string {
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

// func cmdEstimate(args []string) error {
// 	argsLeft := flagset(commandStore, args)

// 	passphrase, passnum, err := checkPassphrase(argsLeft)
// 	if err != nil {
// 		return fmt.Errorf("error checking passphrase: %w", err)
// 	}
// 	logbook, err := newLogbook(passphrase, passnum)
// 	if err != nil {
// 		return fmt.Errorf("error creating Logbook: %w", err)
// 	}
// 	entry, err := ddb.NewEntryFromFile(filepath.Base(flagFilename), flagFilename)
// 	if err != nil {
// 		return fmt.Errorf("error opening file '%s': %v", flagFilename, err)
// 	}
// 	txs, err := logbook.ProcessEntry(entry)
// 	if err != nil {
// 		return fmt.Errorf("error while processing file '%s': %w", flagFilename, err)
// 	}
// 	fmt.Printf("The file has been stored in transactions with the followind IDs\n")
// 	for i, tx := range txids {
// 		fmt.Printf("%d: %s\n", i, tx)
// 	}
// 	return nil
// }

func cmdStore(args []string) error {
	argsLeft := flagset(commandStore, args)

	passphrase, passnum, err := checkPassphrase(argsLeft)
	if err != nil {
		return fmt.Errorf("error checking passphrase: %w", err)
	}
	logbook, err := newLogbook(passphrase, passnum)
	if err != nil {
		return fmt.Errorf("error creating Logbook: %w", err)
	}
	entry, err := ddb.NewEntryFromFile(filepath.Base(flagFilename), flagFilename)
	if err != nil {
		return fmt.Errorf("error opening file '%s': %v", flagFilename, err)
	}
	txids, err := logbook.CastEntry(entry)
	if err != nil {
		return fmt.Errorf("error while storing file '%s' onchain connected to address '%s': %w", flagFilename, logbook.BitcoinPublicAddress(), err)
	}
	fmt.Printf("The file has been stored in transactions with the followind IDs\n")
	for i, tx := range txids {
		fmt.Printf("%d: %s\n", i, tx)
	}
	return nil
}

func cmdRetrieve(args []string) error {
	argsLeft := flagset(commandRetrieve, args)

	if flagOutputDir == "" {
		fmt.Printf("Output dir not set, using local flolder.\n")
	}
	passphrase, passnum, err := checkPassphrase(argsLeft)
	if err != nil {
		return fmt.Errorf("error checking passphrase: %w", err)
	}
	logbook, err := newLogbook(passphrase, passnum)
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

func main() {
	//fmt.Printf("args: %v\n", os.Args)
	if len(os.Args) < 2 {
		printMainHelp()
		os.Exit(0)
	}
	command := strings.ToLower(os.Args[1])
	fmt.Printf("Command is: %s\n", command)
	var err error
	switch command {
	case commandDescribe:
		err = cmdDescribe(os.Args[2:])
	case commandStore:
		err = cmdStore(os.Args[2:])
	case commandRetrieve:
		err = cmdRetrieve(os.Args[2:])
	}
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}
