package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	defaultHeader    = "TRH202101"
	keystoreCmd      = "keystore"
	txCmd            = "tx"
	storeCmd         = "store"
	listCmd          = "list"
	collectCmd       = "collect"
	estimateCmd      = "estimate"
	exitNoPassphrase = iota
	exitNoPassnum
	exitKeygenError
	exitLogbookError
	exitFileError
	exitStoreError
)

var commands = map[string][]string{
	keystoreCmd: "creates and shows keystore file",
	txCmd:       "lists transaction connected to keystore and password",
	listCmd:     "list",
	storeCmd:    "stores a file on the blockchain",
	collectCmd:  "collects amount left on branched transactions due to errors",
	// "retrieveall": "retrieveall",
	estimateCmd: "estimates the amount to spend to save a file on the blockchain",
}

func printMainHelp() {
	fmt.Printf(`
TRH (The Rabbit Hole)

TRH is a tool that let you store and retrieve files from the Bitcoin BSV blockchain.

Read instruction on https://ejfhp.com/projects/trh/

Commands:
`)
	for cmd, desc := range commands {
		fmt.Printf("       %s: %s\n", cmd, desc)
	}
	fmt.Printf(`


Options available:

`)
	for cmd, _ := range commands {
		fmt.Printf("\n")
		printHelp(cmd)
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
	fmt.Printf("\nBuilt time: %s\n", buildTimestamp)
}

func printHelp(command string) {
	fmt.Printf("Options for command '%s': \n", command)
	flagsetK, optionsK := newFlagset(command)
	flagsetK.PrintDefaults()
	fmt.Printf("Accepted combinations:\n")
	for n, c := range optionsK {
		if n != "ignored" {
			fmt.Printf("     %s\n", c)
		}
	}
}

//go:generate go run buildscript/timebuilt.go
func main() {
	//fmt.Printf("args: %v\n", os.Args)
	if len(os.Args) < 2 {
		printMainHelp()
		os.Exit(0)
	}
	command := strings.ToLower(os.Args[1])
	var err error
	switch command {
	case keystoreCmd:
		err = cmdKeystore(os.Args)
	case txCmd:
		err = cmdTx(os.Args)
	case collectCmd:
		err = cmdCollect(os.Args)
	case storeCmd:
		err = cmdStore(os.Args)
	case listCmd:
		err = cmdList(os.Args)
	case estimateCmd:
		err = cmdEstimate(os.Args)
	// case commandRetrieveAll:
	// 	cmdRetrieveAll()
	default:
		printMainHelp()
	}
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		if err.Error() == "flag combination invalid" {
			printHelp(command)
		}
		os.Exit(1)
	}
	os.Exit(0)
}
