package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ejfhp/ddb/keys"
	"github.com/ejfhp/ddb/trh"
	"github.com/ejfhp/trail"
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
	"estimate":   {name: "estimate_file", description: "estimate cost to store file", params: []string{"file", "comma separated labels", "notes"}},
	"utxos":      {name: "utxo_show", description: "show all utxos", params: []string{"pin"}},
	"txshow":     {name: "tx_showall", description: "show all transactions", params: []string{"pin"}},
	"txshowp":    {name: "tx_showpass", description: "show transactions of password", params: []string{"pin", "password"}},
	"collect":    {name: "collect_all", description: "collect unspent money", params: []string{"pin"}},
	"store":      {name: "storefile_file", description: "store file", params: []string{"pin", "file", "comma separated labels", "notes", "max spend (satoshi)"}},
	"storewithp": {name: "storefile_filepassword", description: "store file", params: []string{"pin", "password", "file", "comma separated labels", "notes", "max spend (satoshi)"}},
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

   trh kshow 1346
   trh ktouncry 1346
   trh kfromuncry 1346
   trh kgenfromkp 1346 Kxn6wiqVGzGjMq7JA8m9fxRdukwzzjGgYkXir5eyRwvvrRs7GZKZ therabbithole 
   trh kgenfromph 1346 "Lunedi 8 Novembre 2021"
   trh estimate keystore.trh "keystore,bitcoin,trh" "very important"
   trh txshow 1346
   trh txshowp 1346 main
   trh utxos 1346
   trh collect 1346
   trh store 1346 bitcoin.pdf "bitcoin,pdf" "test import" 200000 
   trh listp 1346
   trh listp 1346 main
`)
	fmt.Printf("\nBuilt time: %s\n", buildTimestamp)
}

//go:generate go run buildscript/timebuilt.go
func main() {
	flag.BoolVar(&flagLog, "log", false, "enable log")
	flag.Parse()
	if flagLog {
		trail.SetWriter(os.Stderr)
	}
	cmds := flag.Args()
	if len(cmds) < 2 {
		printMainHelp()
		os.Exit(0)
	}
	cmd := cmds[0]
	command, ok := commands[cmd]
	if !ok {
		fmt.Printf("Command not found.\n")
		fmt.Printf("Run TRH without params for help.\n")
		os.Exit(1)
	}
	inputs := flag.Args()[1:]
	if len(inputs) != len(command.params) {
		fmt.Printf("Wrong number of input.\n")
		os.Exit(1)
	}
	ksf := getKeystorePath()
	fmt.Printf("TRH Keystore file is: %s\n", ksf)
	fmt.Printf("%s: %s\n", command.name, command.description)
	fmt.Printf("\n")
	var mainerr error
	th := &trh.TRH{}
	switch command.name {
	case "keystore_show":
		mainerr = th.KeystoreShow(inputs[0], ksf)
	case "keystore_tounencrypted":
		ksfu := ksf + ".plain"
		mainerr = th.KeystoreSaveUnencrypted(inputs[0], ksf, ksfu)
		if mainerr == nil {
			fmt.Printf("WARNING: Keystore file saved unencrypted to: %s\n", ksfu)
		}
		fmt.Printf("")
	case "keystore_fromunenecrypted":
		ksfu := ksf + ".plain"
		mainerr = th.KeystoreRestoreFromUnencrypted(inputs[0], ksfu, ksf)
		if mainerr == nil {
			fmt.Printf("Keystore restored to: %s\n", ksf)
		}
		fmt.Printf("")
	case "keystore_fromkeypass":
		_, mainerr = th.KeystoreGenFromKey(inputs[0], inputs[1], inputs[2], ksf)
		if mainerr == nil {
			fmt.Printf("Keystore created: %s\n", ksf)
		}
	case "keystore_frompassphrase":
		_, mainerr = th.KeystoreGenFromPhrase(inputs[0], inputs[1], 3, ksf)
		if mainerr == nil {
			fmt.Printf("Keystore created: %s\n", ksf)
		}
	case "estimate_file":
		filePar := inputs[0]
		labelPar := inputs[1]
		notePar := inputs[2]
		labels := strings.Split(labelPar, ",")
		lbls := make([]string, len(labels))
		for i, l := range labels {
			lbls[i] = strings.TrimSpace(l)
		}
		txs, cost, err := th.Estimate(filePar, lbls, notePar)
		if err == nil {
			fmt.Printf("Estimated cost: %d satoshi\n", cost)
			fmt.Printf("Estimated num of txs: %d\n", len(txs))
		}
		mainerr = err
	case "utxo_show":
		ks, err := keys.LoadKeystore(ksf, inputs[0])
		if err != nil {
			mainerr = err
			break
		}
		utxos, err := th.ListUTXOs(ks)
		if err == nil {
			tot := uint64(0)
			fmt.Printf("UTOXs tied to this keystore:\n")
			for id, amount := range utxos {
				tot = tot + amount
				fmt.Printf(" TXID: '%s' amount: '%d' satoshi\n", id, amount)
			}
			fmt.Printf("Total: %d\n", tot)
		}
		mainerr = err
	case "tx_showall":
		ks, err := keys.LoadKeystore(ksf, inputs[0])
		if err != nil {
			mainerr = err
			break
		}
		alltxs, err := th.ListAllTX(ks)
		if err == nil {
			fmt.Printf("IDs of transactions tied to this keystore:\n")
			for pwd, txs := range alltxs {
				fmt.Printf(" Password: '%s' address: '%s'\n", pwd, ks.Address(pwd))
				fmt.Printf("Number of transactions found: %d\n", len(txs))
				for _, id := range txs {
					fmt.Printf("  %s\n", id)
				}
			}
		}
		mainerr = err
	case "tx_showpass":
		ks, err := keys.LoadKeystore(ksf, inputs[0])
		if err != nil {
			mainerr = err
			break
		}
		txs, err := th.ListSinglePasswordTX(ks, inputs[1])
		fmt.Printf("Number of transactions found: %d\n", len(txs))
		if err == nil {
			fmt.Printf("IDs of transactions tied to this keystore:\n")
			fmt.Printf(" Password: '%s' address: '%s'\n", inputs[1], ks.Address(inputs[1]))
			for _, id := range txs {
				fmt.Printf("  %s\n", id)
			}
		}
		mainerr = err
	case "collect_all":
		ks, err := keys.LoadKeystore(ksf, inputs[0])
		if err != nil {
			mainerr = err
			break
		}
		txResults, err := th.Collect(ks)
		if err == nil {
			if len(txResults) == 0 {
				fmt.Printf("No UTXO found and so no transaction has been submitted.\n")
				break
			}
			fmt.Printf("IDs of transactions to collect UTXO of the keystore:\n")
			for _, id := range txResults {
				fmt.Printf("  %s\n", id)
			}
		}
		mainerr = err
	case "storefile_file":
		pinPar := inputs[0]
		filePar := inputs[1]
		labelPar := inputs[2]
		notePar := inputs[3]
		spendPar := inputs[4]
		ks, err := keys.LoadKeystore(ksf, pinPar)
		if err != nil {
			mainerr = err
			break
		}
		labels := strings.Split(labelPar, ",")
		lbls := make([]string, len(labels))
		for i, l := range labels {
			lbls[i] = strings.TrimSpace(l)
		}
		maxSpend, err := strconv.ParseUint(spendPar, 10, 64)
		if err != nil {
			mainerr = err
			break
		}
		node, _ := ks.Node(keys.NodeMainTrunk)
		_, cost, err := th.Estimate(filePar, lbls, notePar)
		if err != nil {
			mainerr = err
			break
		}
		if maxSpend < cost {
			fmt.Printf("Amount to spend (%d) is not enough, estimation is: %d\n", maxSpend, cost)
			break
		}
		txs, err := th.Store(ks, node, filePar, lbls, notePar, defaultHeader, maxSpend)
		if err == nil {
			fmt.Printf("IDs of transactions that store the file\n")
			for num, txid := range txs {
				fmt.Printf("%d: %s\n", num, txid)
			}
		}
		mainerr = err
	case "storefile_filepassword":
		pinPar := inputs[0]
		pwdPar := inputs[1]
		filePar := inputs[2]
		labelPar := inputs[3]
		notePar := inputs[4]
		spendPar := inputs[5]
		ks, err := keys.LoadKeystore(ksf, pinPar)
		if err != nil {
			mainerr = err
			break
		}
		labels := strings.Split(labelPar, ",")
		lbls := make([]string, len(labels))
		for i, l := range labels {
			lbls[i] = strings.TrimSpace(l)
		}
		maxSpend, err := strconv.ParseUint(spendPar, 10, 64)
		if err != nil {
			mainerr = err
			break
		}
		node, err := ks.NodeFromPassword(pwdPar)
		if err != nil {
			mainerr = err
			break
		}
		needsUpdate := ks.StoreNode(node)
		if needsUpdate {
			err = ks.Update()
			if err != nil {
				mainerr = err
				break
			}
		}
		_, cost, err := th.Estimate(filePar, lbls, notePar)
		if err != nil {
			mainerr = err
			break
		}
		if maxSpend < cost {
			fmt.Printf("Amount to spend (%d) is not enough, estimation is: %d\n", maxSpend, cost)
			break
		}
		txs, err := th.Store(ks, node, filePar, lbls, notePar, defaultHeader, maxSpend)
		if err == nil {
			fmt.Printf("IDs of transactions that store the file\n")
			for num, txid := range txs {
				fmt.Printf("%d: %s\n", num, txid)
			}
		}
		mainerr = err
	case "listfile_all":
		ks, err := keys.LoadKeystore(ksf, inputs[0])
		if err != nil {
			mainerr = err
			break
		}
		allent, err := th.ListAll(ks)
		if err == nil {
			fmt.Printf("Files stored tied to this keystore:\n")
			for p, ents := range allent {
				fmt.Printf("Password %s\n", p)
				for i, ent := range ents {
					fmt.Printf("%d  Name: '%s' hash: '%s'  time: %s\n", i, ent.Name, ent.Hash, time.Unix(ent.Timestamp, 0).Format("2006-01-02 15:04 EST"))
				}
			}
		}
		mainerr = err
	case "listfile_ofpwd":
		ks, err := keys.LoadKeystore(ksf, inputs[0])
		if err != nil {
			mainerr = err
			break
		}
		allent, err := th.ListSinglePassword(ks, inputs[1])
		if err == nil {
			fmt.Printf("Files stored tied to this keystore:\n")
			for p, ents := range allent {
				fmt.Printf("Password %s\n", p)
				for i, ent := range ents {
					fmt.Printf("%d  Name: '%s' hash: '%s'  time: %s\n", i, ent.Name, ent.Hash, time.Unix(ent.Timestamp, 0).Format("2006-01-02 15:04 EST"))
				}
			}
		}
		mainerr = err
	}
	if mainerr == nil {
		fmt.Printf("\n\nCommand terminated succesfully.\n")
		os.Exit(0)
	} else {
		fmt.Printf("\n\nCommand terminated with error: %v\n", mainerr)
		os.Exit(1)
	}
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
