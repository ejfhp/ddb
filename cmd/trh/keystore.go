package main

import (
	"fmt"

	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

func cmdKeystore(args []string) error {
	tr := trace.New().Source("setup.go", "", "cmdKeystore")
	flagset := newFlagset(commandRetrieveAll)
	err := flagset.Parse(args[2:])
	if err != nil {
		trail.Println(trace.Alert("error while parsing args").Append(tr).UTC().Error(err))
		return fmt.Errorf("error while parsing args: %w", err)
	}
	if flagHelp {
		printHelp(flagset)
		return nil
	}
	return nil
}
