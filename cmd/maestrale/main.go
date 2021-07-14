package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
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
)

func printHelp() {
	fmt.Printf("MAESTRALE\n")
}

func checkPassphrase(args []string) (string, int) {
	startidx := -1
	for i, t := range args {
		if t == "=" {
			startidx = i + 1
		}
	}
	if startidx < 0 || startidx >= len(args) {
		quit("because passphrase is missing", exitNoPassphrase)
	}
	passphrase := strings.Join(args[startidx:], " ")
	passnum := 0
	for _, n := range strings.Split(passphrase, " ") {
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
		quit("because passphrase must contain a number", exitNoPassnum)
	}
	fmt.Printf("Secret configuration is:\n")
	fmt.Printf("passnum: '%d'\n", passnum)
	fmt.Printf("passphrase: '%s'\n", passphrase)
	return passphrase, passnum
}

func keyGen(passphrase string, passnum int) (string, [32]byte, error) {
	keygen, err := ddb.NewKeygen(passnum, passphrase)
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("Error while building Keygen: %w", err)
	}
	wif, err := keygen.MakeWIF()
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("Error while generating bitcoin key: %w", err)
	}
	password := keygen.Password()
	return wif, password, nil
}

func newLogbook(passphrase string, passnum int) *ddb.Logbook {
	wif, password, err := keyGen(passphrase, passnum)
	if err != nil {
		quit("while generating Bitcoin private key", exitKeygenError)
	}
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc)
	logbook, err := ddb.NewLogbook(wif, password, blockchain)
	if err != nil {
		quit("while creating the Logbook", exitLogbookError)
	}
	fmt.Printf("Bitcoin configuration is:\n")
	fmt.Printf("Bitcoin Key (WIF) is : '%s'\n", logbook.BitcoinPrivateKey())
	fmt.Printf("Bitcoin Address is   : '%s'\n", logbook.BitcoinPublicAddress())
	return logbook
}

func quit(message string, code int) {
	fmt.Printf("An error has occurred %s.\n", message)
	printHelp()
	os.Exit(code)
}

// maestrale describe <passphrase>
func cmdDescribe(args []string) error {
	flagset := flag.NewFlagSet("describe", flag.ContinueOnError)
	log := flagset.Bool("log", false, "true enables log output")
	flagset.Parse(args)
	logOn(*log)

	phrase := flagset.Args()
	fmt.Printf("log: %t\n", *log)
	fmt.Printf("remained args: %v\n", phrase)

	passphrase, passnum := checkPassphrase(phrase)
	logbook := newLogbook(passphrase, passnum)

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

func logOn(on bool) {
	if on == true {
		log.SetWriter(os.Stderr)
	}
}

// maestrale store <file> <passphrase>
//go run main.go store testo.txt # quando arriva, il maestrale soffia almeno 3 giorni
func cmdStore(args []string) error {
	passphrase, passnum := checkPassphrase(args)
	logbook := newLogbook(passphrase, passnum)
	filename := args[0]
	entry, err := ddb.NewEntryFromFile(filepath.Base(filename), filename)
	if err != nil {
		quit(fmt.Sprintf("while opening file '%s'", filename), exitFileError)
	}
	txids, err := logbook.CastEntry(entry)
	if err != nil {
		quit(fmt.Sprintf("while storing file '%s' onchain", filename), exitStoreError)
	}
	fmt.Printf("The file has been stored in transactions with the followind IDs\n")
	for i, tx := range txids {
		fmt.Printf("%d: %s\n", i, tx)
	}
	return nil
}

//go run main.go describe quando arriva, il maestrale soffia almeno 3 giorni

func main() {
	fmt.Printf("args: %v\n", os.Args)
	if len(os.Args) < 2 {
		printHelp()
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
	}
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}
}
