package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/ejfhp/ddb/trh"
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

type command struct {
	name        string
	description string
	params      []string
}

var commands = map[string]command{
	"kshow":      {name: "keystore_show", description: "keystore show", params: []string{"pin"}},
	"ktouncry":   {name: "keystore_tounencrypted", description: "keystore export to unencrypted", params: []string{"pin"}},
	"kfromuncry": {name: "keystore_fromunenecrypted", description: "keystore load from unencrypted", params: []string{"pin"}},
	"kgenfromkp": {name: "keystore_fromkeypass", description: "keystore generate from key and password", params: []string{"pin", "key", "password"}},
	"kgenfromph": {name: "keystore_frompassphrase", description: "keystore generate from phrase", params: []string{"pin", "phrase"}},
	"estimate":   {name: "estimate_file", description: "estimate cost to store file", params: []string{"password", "file"}},
	"txshow":     {name: "tx_showall", description: "transactions show all", params: []string{"pin"}},
	"txshowp":    {name: "tx_showpass", description: "transaction show of password", params: []string{"pin", "password"}},
	"collect":    {name: "collect_all", description: "collect unspent money", params: []string{"pin"}},
	"store":      {name: "storefile_file", description: "store file", params: []string{"password", "file", "pin"}},
	"list":       {name: "listfile_all", description: "list all files stored", params: []string{"pin"}},
	"listp":      {name: "listfile_ofpwd", description: "list files stored with password", params: []string{"pin", "password"}},
}
var flagLog bool

func printMainHelp() {
	fmt.Printf(`
TRH (The Rabbit Hole)

TRH is a tool that let you store and retrieve files from the Bitcoin BSV blockchain.

Read instruction on https://ejfhp.com/projects/trh/

Commands:
`)
	w := tabwriter.NewWriter(os.Stdout, 0, 10, 10, ' ', tabwriter.TabIndent)
	cmdNames := make([]string, 0, len(commands))
	for k := range commands {
		cmdNames = append(cmdNames, k)
	}
	sort.Strings(cmdNames)
	for _, cn := range cmdNames {
		cmd := commands[cn]
		fmt.Fprintf(w, "%s\t<%s>\t%s\t\n", cn, strings.Join(cmd.params, "> <"), cmd.description)
	}
	w.Flush()
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

//go:generate go run buildscript/timebuilt.go
func main() {
	flag.BoolVar(&flagLog, "log", false, "enable log")
	flag.Parse()

	cmds := flag.Args()
	if len(cmds) < 2 {
		printMainHelp()
		os.Exit(0)
	}
	cmd := cmds[0]
	command, ok := commands[cmd]
	if !ok {
		fmt.Printf("Command not found.\n")
		os.Exit(1)
	}
	inputs := flag.Args()[1:]
	if len(inputs) != len(command.params) {
		fmt.Printf("Wrong number of input.\n")
		os.Exit(1)
	}
	ksf := getKeystorePath()
	fmt.Printf("TRH Keystore file is: %s\n", ksf)
	var err error
	th := &trh.TRH{}
	switch command.name {
	case "keystore_show":
		err = th.KeystoreShow(inputs[0], ksf)
	case "keystore_tounencrypted":
		ksfu := ksf + ".plain"
		err = th.KeystoreSaveUnencrypted(inputs[0], ksf, ksfu)
		if err == nil {
			fmt.Printf("WARNING: Keystore file saved unencrypted to: %s\n", ksfu)
		}
		fmt.Printf("")
	case "keystore_fromunenecrypted":
		ksfu := ksf + ".plain"
		err = th.KeystoreRestoreFromUnencrypted(inputs[0], ksfu, ksf) //TODO COMPLETE
		if err == nil {
			fmt.Printf("Keystore restored to: %s\n", ksf)
		}
		fmt.Printf("")
	case "keystore_fromkeypass":
		_, err := th.KeystoreGenFromKey(inputs[0], inputs[1], inputs[2], ksf)
		if err == nil {
			fmt.Printf("Keystore created: %s\n", ksf)
		}
	case "keystore_frompassphrase":
	case "estimate_file":
	case "tx_showall":
	case "tx_showpass":
	case "collect_all":
	case "storefile_file":
	case "listfile_all":
	case "listfile_ofpwd":
	}
	if err == nil {
		fmt.Printf("\n\nCommand terminated succesfully.\n")
	} else {
		fmt.Printf("\n\nCommand terminated with error: %v\n", err)

	}

	os.Exit(0)
}

func getKeystorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error retrieving user dir: %v", err)
		home, err = os.Getwd()
		if err != nil {
			fmt.Printf("Error retrieving working dir: %v", err)
			home = "."
		}
	}
	return filepath.Join(home, "keystore.trh")
}
