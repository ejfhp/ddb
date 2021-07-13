package main

import (
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
	exitFileError
	exitStoreError
	commandDescribe = "describe"
	commandStore    = "store"
)

func printHelp() {
	fmt.Printf("MAESTRALE\n")
}

func checkPassphrase(args []string) (string, int) {
	if len(args) < 1 {
		fmt.Printf("Missing passphrase\n")
		printHelp()
		os.Exit(exitNoPassphrase)
	}
	passphrase := strings.Join(args, " ")
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
		fmt.Printf("Passphrase must contain a number\n")
		printHelp()
		os.Exit(exitNoPassnum)
	}
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

func newLogbook(wif string, password [32]byte) (*ddb.Logbook, error) {
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc)
	logbook, err := ddb.NewLogbook(wif, password, blockchain)
	if err != nil {
		return nil, fmt.Errorf("error creating new Logbook; %w", err)
	}
	return logbook, nil
}

// maestrale describe <passphrase>
func cmdDescribe(args []string) error {
	passphrase, passnum := checkPassphrase(args)
	wif, password, err := keyGen(passphrase, passnum)
	if err != nil {
		return err
	}
	logbook, err := newLogbook(wif, password)
	if err != nil {
		return err
	}
	address := logbook.BitcoinPublicAddress()
	fmt.Printf("passphrase is : '%s'\n", passphrase)
	fmt.Printf("passnum is : '%d'\n", passnum)
	fmt.Printf("Bitcoin Key (WIF) is : '%s'\n", wif)
	fmt.Printf("Bitcoin Address is   : '%s'\n", address)

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

// maestrale store <file> <passphrase>
func cmdStore(args []string) error {
	filename := args[0]
	entry, err := ddb.NewEntryFromFile(filepath.Base(filename), filename)
	if err != nil {
		fmt.Printf("Error while opening file '%s': %v \n", filename, err)
		printHelp()
		os.Exit(exitFileError)
	}
	passphrase, passnum := checkPassphrase(args)
	wif, password, err := keyGen(passphrase, passnum)
	if err != nil {
		return err
	}
	logbook, err := newLogbook(wif, password)
	if err != nil {
		return err
	}
	txids, err := logbook.CastEntry(entry)
	if err != nil {
		fmt.Printf("Error while storing file '%s' on-chain: %v \n", filename, err)
		printHelp()
		os.Exit(exitFileError)
	}
	fmt.Printf("The file has been stored in transactions with the followind IDs\n")
	for i, tx := range txids {
		fmt.Printf("%d: %s\n", i, tx)
	}
	return nil
}

//go run main.go describe quando arriva, il maestrale soffia almeno 3 giorni

func main() {
	log.SetWriter(os.Stdout)
	args := os.Args
	fmt.Printf("args: %v\n", args)
	if len(args) < 2 {
		printHelp()
		os.Exit(0)
	}
	command := strings.ToLower(args[1])
	var err error
	switch command {
	case commandDescribe:
		err = cmdDescribe(args[2:])
	case commandStore:
		err = cmdStore(args[2:])
	}
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}
}
