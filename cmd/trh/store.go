package main

// func flagsetStore(cmd string, args []string) []string {
// 	flagset := flag.NewFlagSet("describe", flag.ContinueOnError)
// 	flagset.BoolVar(&flagLog, "log", false, "true enables log output")
// 	flagset.BoolVar(&flagHelp, "help", false, "print help")
// 	flagset.BoolVar(&flagHelp, "h", false, "print help")
// 	switch cmd {
// 	case commandStore:
// 		flagset.StringVar(&flagFilename, "file", "", "path of file to store onchain")
// 	case commandRetrieve:
// 		flagset.StringVar(&flagOutputDir, "outdir", "", "path of folder where to save retrived files")
// 	}
// 	flagset.Parse(args)
// 	if flagHelp {
// 		printHelp(flagset)
// 	}
// 	//fmt.Printf("file: %s\n", flagFilename)
// 	logOn(flagLog)
// 	return flagset.Args()
// }

// func cmdStore(args []string) error {
// 	argsLeft := flagset(commandStore, args)

// 	passphrase, passnum, err := checkPassphrase(argsLeft)
// 	if err != nil {
// 		return fmt.Errorf("error checking passphrase: %w", err)
// 	}
// 	logbook, err := newLogbook(passphrase, passnum)
// 	if err != nil {
// 		return fmt.Errorf("error creating Logbook: %w", err)
// 	}
// 	entry, err := ddb.NewEntryFromFile(filepath.Base(flagFilename), flagFilename)
// 	if err != nil {
// 		return fmt.Errorf("error opening file '%s': %v", flagFilename, err)
// 	}
// 	txids, err := logbook.CastEntry(entry)
// 	if err != nil {
// 		return fmt.Errorf("error while storing file '%s' onchain connected to address '%s': %w", flagFilename, logbook.BitcoinPublicAddress(), err)
// 	}
// 	fmt.Printf("The file has been stored in transactions with the followind IDs\n")
// 	for i, tx := range txids {
// 		fmt.Printf("%d: %s\n", i, tx)
// 	}
// 	return nil
// }
