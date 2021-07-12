package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ejfhp/ddb"
	log "github.com/ejfhp/trail"
)

const (
	EXIT_NO_PASSPHRASE = 1
	EXIT_NO_PASSNUM    = 2
	DESCRIBE           = "describe"
	STORE              = "store"
)

func printHelp() {
	fmt.Printf("MAESTRALE\n")
}

func checkPassphrase(args []string) (string, int) {
	if len(args) < 1 {
		fmt.Printf("Missing passphrase\n")
		printHelp()
		os.Exit(EXIT_NO_PASSPHRASE)
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
		os.Exit(EXIT_NO_PASSNUM)
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

func describe(args []string) error {
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
	for i, tx := range history {
		fmt.Printf("%d: %s\n", i, tx)
	}
	return nil
}

func store(args []string) error {
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
	case DESCRIBE:
		err = describe(args[2:])
	case STORE:
		err = store(args[2:])
	}
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}
}
