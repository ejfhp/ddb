package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	exitNoPassphrase = iota
	exitNoPassnum
	exitKeygenError
	exitLogbookError
	exitFileError
	exitStoreError
)

var commands = map[string]string{
	"keystore": "keystore",
	// "describe":    "describe",
	// "store":       "store",
	// "retrieveall": "retrieveall",
	// "estimate":    "estimate",
}

func printMainHelp() {
	fmt.Printf(`
TRH (The Rabbit Hole)

TRH is a tool that let you store and retrieve files from the Bitcoin BSV blockchain.

Read instruction on https://ejfhp.com/projects/trh/

Commands:
- keystore: manage keystore
- describe: to show address, keys and transaction IDs
- estimate: to estimate the miner fee before to store a file
- store: to write files on the blockchain
- retrieveall: to download all the files from the blockchain
`)
	for _, command := range commands {
		fmt.Printf("\n")
		fmt.Printf("Options for command '%s': \n", command)
		flagsetK, optionsK := newFlagset(command)
		flagsetK.PrintDefaults()
		fmt.Printf("Accepted combinations:\n")
		for _, c := range optionsK {
			fmt.Printf("     %s\n", c)
		}
	}
	fmt.Printf(`
Examples:

./trh store -help
./trh describe -log -key <key>
./trh keystore generate -phrase "Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF"
./trh estimate -file bitcoin.pdf -log + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
./trh store -file bitcoin.pdf -log + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
./trh retrieveAll -outdir /Users/diego/Desktop/ + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
./trh retrieveall -address 16dEpFZ8nEvSv9MJ9MQqZ7ihk6mypQdrZ -password "Bitcoin: A Peer-to-Peer Electron"
`)
}

func printHelp(flagset *flag.FlagSet) {
	if flagset != nil {
		flagset.SetOutput(os.Stdout)
		flagset.PrintDefaults()
	}
	os.Exit(0)
}

// func cmdDescribe() {
// 	flagset := newFlagset(commandRetrieveAll)
// 	env, err := prepareEnvironment(os.Args, flagset)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(1)
// 	}
// 	if env.help {
// 		printHelp(flagset)
// 		os.Exit(0)
// 	}
// 	cache, err := prepareCache(env)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(2)
// 	}
// 	diary, err := prepareDiary(env, cache)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(3)
// 	}
// 	describe := NewDescribe(env, diary)
// 	err = describe.Describe(os.Stdout)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(4)
// 	}
// }

// func cmdStore() {
// 	flagset := newFlagset(commandStore)
// 	env, err := prepareEnvironment(os.Args, flagset)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(1)
// 	}
// 	if env.help {
// 		printHelp(flagset)
// 		os.Exit(0)
// 	}
// 	if flagFile == "" {
// 		fmt.Printf("WARNING: file not specified\n")
// 		os.Exit(0)
// 	}
// 	cache, err := prepareCache(env)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(2)
// 	}
// 	if cache != nil {
// 		fmt.Printf("INFO: cache folder is: %s.\n", cache.DirPath())
// 	}
// 	diary, err := prepareDiary(env, cache)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(3)
// 	}
// 	store := NewStore(env, diary)
// 	txs, err := store.Store(flagFile)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(4)
// 	}
// 	if len(txs) > 0 {
// 		fmt.Printf("INFO: the file has been stored in %d transactions with the followind ID:\n", len(txs))
// 		for i, tx := range txs {
// 			fmt.Printf("%d: %s\n", i, tx)
// 		}
// 	}
// }

// func cmdEstimate() {
// 	flagset := newFlagset(commandStore)
// 	env, err := prepareEnvironment(os.Args, flagset)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(1)
// 	}
// 	if env.help {
// 		printHelp(flagset)
// 		os.Exit(0)
// 	}
// 	if flagFile == "" {
// 		fmt.Printf("WARNING: file not specified\n")
// 		os.Exit(0)
// 	}
// 	diary, err := prepareDiary(env, nil)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(3)
// 	}
// 	store := NewStore(env, diary)
// 	fee, err := store.Estimate(flagFile)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(4)
// 	}
// 	fmt.Printf("INFO: estimated fee to store the file: %d satoshi\n", fee.Satoshi())
// }

// func cmdRetrieveAll() {
// 	flagset := newFlagset(commandRetrieveAll)
// 	env, err := prepareEnvironment(os.Args, flagset)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(1)
// 	}
// 	if env.help {
// 		printHelp(flagset)
// 		os.Exit(0)
// 	}
// 	if !env.passwordSet {
// 		fmt.Printf("WARNING: password is not set.\n")
// 	}
// 	cache, err := prepareCache(env)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(2)
// 	}
// 	if cache != nil {
// 		fmt.Printf("INFO: cache folder is: %s.\n", cache.DirPath())
// 	}
// 	diary, err := prepareDiary(env, cache)
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(3)
// 	}
// 	retrieve := NewRetrieve(env, diary)
// 	n, err := retrieve.RetrieveAll()
// 	if err != nil {
// 		fmt.Printf("ERROR: %v.\n", err)
// 		os.Exit(4)
// 	}
// }

func main() {
	//fmt.Printf("args: %v\n", os.Args)
	if len(os.Args) < 2 {
		printMainHelp()
		os.Exit(0)
	}
	command := strings.ToLower(os.Args[1])
	var err error
	switch command {
	case commands["keystore"]:
		err = cmdKeystore(os.Args)
	// case commandDescribe:
	// 	cmdDescribe()
	// case commandStore:
	// 	cmdStore()
	// case commandEstimate:
	// 	cmdEstimate()
	// case commandRetrieveAll:
	// 	cmdRetrieveAll()
	default:
		printMainHelp()
	}
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
