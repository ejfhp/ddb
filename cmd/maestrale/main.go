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
	commandRetrieve = "retrieve"
)

var (
	flagLog       bool
	flagHelp      bool
	flagFilename  string
	flagOutputDir string
)

func checkPassphrase(args []string) (string, int) {
	startidx := -1
	for i, t := range args {
		if t == "+" {
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
	//fmt.Printf("Secret configuration is:\n")
	//fmt.Printf("passnum: '%d'\n", passnum)
	//fmt.Printf("passphrase: '%s'\n", passphrase)
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
	//fmt.Printf("Bitcoin configuration is:\n")
	//fmt.Printf("Bitcoin Key (WIF) is : '%s'\n", logbook.BitcoinPrivateKey())
	//fmt.Printf("Bitcoin Address is   : '%s'\n", logbook.BitcoinPublicAddress())
	return logbook
}

func logOn(on bool) {
	if on == true {
		log.SetWriter(os.Stderr)
	}
}

func printHelp(flagset *flag.FlagSet) {
	//fmt.Printf("MAESTRALE\n")
	if flagset != nil {
		flagset.SetOutput(os.Stdout)
		flagset.PrintDefaults()
	}
	//fmt.Printf("Main command: describe, store, retrieve.\n")
	os.Exit(0)
}

func quit(message string, code int) {
	//fmt.Printf("An error has occurred %s.\n", message)
	os.Exit(code)
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
		flagset.StringVar(&flagFilename, "outdir", "", "path of folder where to save retrived files")
	}
	flagset.Parse(args)
	if flagHelp == true {
		printHelp(flagset)
	}
	//fmt.Printf("file: %s\n", flagFilename)
	logOn(flagLog)
	return flagset.Args()
}

// maestrale describe <passphrase>
func cmdDescribe(args []string) error {
	argsLeft := flagset(commandDescribe, args)

	passphrase, passnum := checkPassphrase(argsLeft)
	logbook := newLogbook(passphrase, passnum)

	history, err := logbook.ListHistory(logbook.BitcoinPublicAddress())
	if err != nil {
		return fmt.Errorf("error getting address history; %w", err)
	}
	//fmt.Printf("Transaction History\n")
	if len(history) == 0 {
		//fmt.Printf("this address has no history\n")
	}
	for i, tx := range history {
		//fmt.Printf("%d: %s\n", i, tx)
	}
	return nil
}

// maestrale store -file <file> = <passphrase>
//go run main.go store -file testo.txt + quando arriva, il maestrale soffia almeno 3 giorni
func cmdStore(args []string) error {
	argsLeft := flagset(commandStore, args)

	passphrase, passnum := checkPassphrase(argsLeft)
	logbook := newLogbook(passphrase, passnum)

	entry, err := ddb.NewEntryFromFile(filepath.Base(flagFilename), flagFilename)
	if err != nil {
		quit(fmt.Sprintf("while opening file '%s'", flagFilename), exitFileError)
	}
	txids, err := logbook.CastEntry(entry)
	if err != nil {
		quit(fmt.Sprintf("while storing file '%s' onchain connected to address '%s'", flagFilename, logbook.BitcoinPublicAddress()), exitStoreError)
	}
	//fmt.Printf("The file has been stored in transactions with the followind IDs\n")
	for i, tx := range txids {
		//fmt.Printf("%d: %s\n", i, tx)
	}
	return nil
}

func cmdRetrieve(args []string) error {
	argsLeft := flagset(commandRetrieve, args)

	passphrase, passnum := checkPassphrase(argsLeft)
	logbook := newLogbook(passphrase, passnum)

	n, err := logbook.DowloadAll(flagOutputDir)
	if err != nil {
		quit(fmt.Sprintf("while retrieving files from address '%s' to floder '%s'", logbook.BitcoinPublicAddress(), flagOutputDir), exitFileError)
	}
	//fmt.Printf("%d files has been retrived from '%s' to '%s'\n", n, logbook.BitcoinPublicAddress(), flagOutputDir)
	return nil
}

//go run main.go describe quando arriva, il maestrale soffia almeno 3 giorni
func main() {
	//fmt.Printf("args: %v\n", os.Args)
	if len(os.Args) < 2 {
		printHelp(nil)
	}
	command := strings.ToLower(os.Args[1])
	//fmt.Printf("Command is: %s\n", command)
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
		//fmt.Printf("ERROR: %v", err)
	}
}
