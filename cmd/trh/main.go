package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	defaultHeader    = "TRH202101"
	exitNoPassphrase = iota
	exitNoPassnum
	exitKeygenError
	exitLogbookError
	exitFileError
	exitStoreError
)

var commands = map[string]string{
	"keystore": "keystore",
	"show":     "show",
	// "list":     "list",
	"store":   "store",
	"collect": "collect",
	// "retrieveall": "retrieveall",
	"estimate": "estimate",
}

func printMainHelp() {
	fmt.Printf(`
TRH (The Rabbit Hole)

TRH is a tool that let you store and retrieve files from the Bitcoin BSV blockchain.

Read instruction on https://ejfhp.com/projects/trh/

Commands:
`)
	for _, command := range commands {
		fmt.Printf("       %s\n", command)
	}
	fmt.Printf(`


Options available:

`)
	for _, command := range commands {
		fmt.Printf("\n")
		printHelp(command)
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
	case commands["show"]:
		err = cmdShow(os.Args)
	case commands["collect"]:
		err = cmdCollect(os.Args)
	case commands["store"]:
		err = cmdStore(os.Args)
	case commands["estimate"]:
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
