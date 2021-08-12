package main

import (
	"flag"
	"fmt"
)

var (
	flagRLog       bool
	flagRHelp      bool
	flagRFilename  string
	flagROutputDir string
)

type Retrieve struct {
	flagset *flag.FlagSet
	env     *Environment
}

func NewRetrieve(environment *Environment) *Retrieve {
	flagset := flag.NewFlagSet("describe", flag.ContinueOnError)
	flagset.BoolVar(&flagRLog, "log", false, "true enables log output")
	flagset.BoolVar(&flagRHelp, "help", false, "print help")
	flagset.BoolVar(&flagRHelp, "h", false, "print help")
	flagset.StringVar(&flagROutputDir, "outdir", "", "path of folder where to save retrived files")
	cmd := Retrieve{flagset: flagset, env: environment}
	return &cmd

}

func flagsetRetrieve(cmd string, args []string) []string {
	flagset.Parse(args)
	//fmt.Printf("file: %s\n", flagFilename)
	return flagset.Args()
}

func (cr *Retrieve) Cmd(args []string) error {
	err := cr.flagset.Parse(args)
	if flagHelp {
		printHelp(cr.flagset)
	}
	logOn(flagLog)
	argsLeft := cr.flagset.Args()

	if flagOutputDir == "" {
		fmt.Printf("Output dir not set, using local flolder.\n")
	}
	passphrase, passnum, err := checkPassphrase(argsLeft)
	if err != nil {
		return fmt.Errorf("error checking passphrase: %w", err)
	}
	logbook, err := newLogbook(passphrase, passnum)
	if err != nil {
		return fmt.Errorf("error creating Logbook: %w", err)
	}

	n, err := logbook.DowloadAll(flagOutputDir)
	if err != nil {
		fmt.Errorf("error while retrieving files from address '%s' to floder '%s': %w", logbook.BitcoinPublicAddress(), flagOutputDir, err)
	}
	fmt.Printf("%d files has been retrived from '%s' to '%s'\n", n, logbook.BitcoinPublicAddress(), flagOutputDir)
	return nil
}
