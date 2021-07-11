package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ejfhp/ddb"
)

const (
	EXIT_NO_PASSPHRASE = 1
	EXIT_NO_PASSNUM    = 2
	DESCRIBE           = "describe"
)

func printHelp() {
	fmt.Printf("MAESTRALE\n")
}

func describe(args []string) error {
	if len(args) < 1 {
		fmt.Printf("Missing passphrase\n")
		printHelp()
		os.Exit(EXIT_NO_PASSPHRASE)
	}
	passphrase := strings.Join(args, " ")
	fmt.Printf("passphrase is : '%s'\n", passphrase)
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
	fmt.Printf("passnum is : '%d'\n", passnum)
	keygen, err := ddb.NewKeygen(passnum, passphrase)
	if err != nil {
		return fmt.Errorf("Error while building Keygen: %w", err)
	}
	keygen.Describe()
	wif, err := keygen.MakeWIF()
	if err != nil {
		return fmt.Errorf("Error while generating bitcoin key: %w", err)
	}
	address, err := ddb.AddressOf(wif)
	if err != nil {
		return fmt.Errorf("Error while generating bitcoin address: %w", err)
	}
	fmt.Printf("Bitcoin Key (WIF) is : '%s'\n", wif)
	fmt.Printf("Bitcoin Address is   : '%s'\n", address)
	return nil

}

func main() {
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
	}
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}
}
