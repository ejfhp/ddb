package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ejfhp/trail"
)

const (
	exitNoPassphrase = iota
	exitNoPassnum
	exitKeygenError
	exitLogbookError
	exitFileError
	exitStoreError
	commandDescribe    = "describe"
	commandStore       = "store"
	commandRetrieveAll = "retrieveall"
	commandEstimate    = "estimate"
)

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

To retrieve all the files from the blockchain do:

>trh retrieveAll -outdir <output folder> + <passphrase with a number 9999>


Options:
-log   true enables log output
-help  print help

Examples:

./trh describe -log + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
./trh estimate -file bitcoin.pdf -log + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
./trh store -file bitcoin.pdf -log + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
./trh retrieveAll -outdir /Users/diego/Desktop/ + Bitcoin: A Peer-to-Peer Electronic Cash System - 2008 PDF
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

func main() {
	//fmt.Printf("args: %v\n", os.Args)
	if len(os.Args) < 2 {
		printMainHelp()
		os.Exit(0)
	}
	command := strings.ToLower(os.Args[1])
	fmt.Printf("Command is: %s\n", command)
	switch command {
	case commandDescribe:
		// err = cmdDescribe(os.Args)
	case commandStore:
		// err = cmdStore(os.Args)
	case commandRetrieveAll:
		flagset := newFlagset(commandRetrieveAll)
		env, err := prepareEnvironment(os.Args, flagset)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(-1)
		}
		cache, err := prepareCache(env)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(-1)
		}
		if cache != nil {
			fmt.Printf("Using cache folder: %s\n", cache.DirPath())
		}
		diary, err := prepareDiary(env, cache)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(-1)
		}
		retrieve := NewRetrieve(env, diary)
		n, err := retrieve.CmdDownloadAll()
		fmt.Printf("%d files has been retrived from '%s' to '%s'\n", n, diary.BitcoinPublicAddress(), flagOutputDir)
	}
}
