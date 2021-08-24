package main

import (
	"fmt"
	"io"
	"runtime"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

type Describe struct {
	diary *ddb.Diary
	env   *Environment
}

func NewDescribe(env *Environment, diary *ddb.Diary) *Describe {
	describe := Describe{diary: diary, env: env}
	return &describe

}

func (d *Describe) Describe(writer io.Writer) error {
	tr := trace.New().Source("describe.go", "Describe", "CmdDescribe")
	trail.Println(trace.Info("printing environment configuration").Append(tr).UTC())

	fmt.Fprintf(writer, "Current configuration:\n")

	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "Secret:\n")
	fmt.Fprintf(writer, "WARNING: Save the passphrase in a safe place.\n")
	if d.env.passphrase != "" {
		fmt.Fprintf(writer, "passphrase: -->%s<--\n", d.env.passphrase)
	}
	if d.env.passwordSet {
		fmt.Fprintf(writer, "pasword: -->%s<--\n", d.env.passwordString())
	}

	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "Bitcoin:\n")
	if d.diary.BitcoinPrivateKey() != "" {
		fmt.Fprintf(writer, "key (WIF): -->%s<--\n", d.diary.BitcoinPrivateKey())
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			fmt.Fprintf(writer, "Bitcoin key QRCode\n")
			ddb.PrintQRCode(writer, d.diary.BitcoinPrivateKey())
			fmt.Fprintf(writer, "\n")
		}
	}
	if d.diary.BitcoinPublicAddress() != "" {
		fmt.Fprintf(writer, "address: -->%s<--\n", d.diary.BitcoinPublicAddress())
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			fmt.Fprintf(writer, "Bitcoin address QRCode\n")
			ddb.PrintQRCode(writer, d.diary.BitcoinPublicAddress())
			fmt.Fprintf(writer, "\n")
		}
	}

	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "Transaction:\n")
	if d.diary.BitcoinPublicAddress() != "" {
		history, err := d.diary.ListHistory(d.diary.BitcoinPublicAddress())
		if err != nil {
			trail.Println(trace.Alert("error getting transaction history").Append(tr).UTC().Add("address", d.diary.BitcoinPublicAddress()))
			return fmt.Errorf("error getting transaction history for address '%s': %w", d.diary.BitcoinPublicAddress(), err)
		}
		if len(history) == 0 {
			fmt.Printf("this address has no history\n")
		}
		for i, tx := range history {
			fmt.Fprintf(writer, "%d: %s\n", i, tx)
		}
	}

	fmt.Fprintf(writer, "\n")
	fmt.Fprintf(writer, "Cache:\n")
	if !d.env.cacheDisabled {
		fmt.Fprintf(writer, "cache folder: '%s'\n", d.diary.Blockchain.CacheDir())
	}
	return nil
}
