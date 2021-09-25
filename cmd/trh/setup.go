package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

var (
	flagLog            bool
	flagHelp           bool
	flagAction         string
	flagFile           string
	flagOutputDir      string
	flagCacheDir       string
	flagDisableCache   bool
	flagOnlyCache      bool
	flagBitcoinAddress string
	flagBitcoinKey     string
	flagPassword       string
	flagPIN            string
	flagKeygenID       int64
	flagPhrase         string
	flagNotToCheck     = []string{"log", "help", "h", "keygen"}
)

func newFlagset(command string) (*flag.FlagSet, map[string][]string) {
	flagset := flag.NewFlagSet(command, flag.ContinueOnError)
	flagset.BoolVar(&flagLog, "log", false, "true enables log output")
	flagset.BoolVar(&flagHelp, "help", false, "prints help")
	flagset.BoolVar(&flagHelp, "h", false, "prints help")
	//KEYSTORE
	if command == commands["keystore"] {
		flagset.StringVar(&flagAction, "action", "show", "what to do (show, generate)")
		flagset.Int64Var(&flagKeygenID, "keygen", 2, "keygen to be used for key and password generation")
		flagset.StringVar(&flagBitcoinKey, "key", "", "bitcoin key")
		flagset.StringVar(&flagPhrase, "phrase", "", "passphrase to generate key and password, if key is not set")
		flagset.StringVar(&flagPassword, "password", "", "encryption password, required if key is set")
		flagset.StringVar(&flagPIN, "pin", "", "the pin to use to encrypt the keystore")
		options := map[string][]string{
			"empty":        {""},
			"key":          {"key", "password"},
			"phrase":       {"phrase"},
			"actionkey":    {"action", "pin", "key", "password"},
			"actionphrase": {"action", "pin", "phrase"},
		}
		return flagset, options
	}
	//DESCRIBE
	// if command == commandDescribe {
	// 	flagset.StringVar(&flagBitcoinAddress, "address", "", "bitcoin address")
	// 	flagset.StringVar(&flagBitcoinKey, "key", "", "bitcoin key")
	// 	flagset.StringVar(&flagPassword, "password", "", "encryption password")
	// 	flagset.Int64Var(&flagKeygenID, "keygen", 2, "keygen to be used for key and password generation")
	// 	return flagset, nil
	// }
	//RETRIEVE
	// if command == commandRetrieveAll {
	// 	flagset.StringVar(&flagBitcoinAddress, "address", "", "bitcoin address")
	// 	flagset.StringVar(&flagBitcoinKey, "key", "", "bitcoin key")
	// 	flagset.StringVar(&flagPassword, "password", "", "encryption password")
	// 	flagset.Int64Var(&flagKeygenID, "keygen", 2, "keygen to be used for key and password generation")
	// 	flagset.StringVar(&flagOutputDir, "outdir", "", "path of the folder where to save retrived files")
	// 	flagset.BoolVar(&flagDisableCache, "nocache", false, "true disables cache")
	// 	flagset.BoolVar(&flagOnlyCache, "onlycache", false, "true retrieves only from cache")
	// 	flagset.StringVar(&flagCacheDir, "cachedir", "", "path of the folder to be used as cache")
	// 	return flagset, nil
	// }
	//STORE
	// if command == commandStore {
	// 	flagset.StringVar(&flagBitcoinAddress, "address", "", "bitcoin address")
	// 	flagset.StringVar(&flagBitcoinKey, "key", "", "bitcoin key")
	// 	flagset.StringVar(&flagPassword, "password", "", "encryption password")
	// 	flagset.Int64Var(&flagKeygenID, "keygen", 2, "keygen to be used for key and password generation")
	// 	flagset.StringVar(&flagFile, "file", "", "path of file to store")
	// 	flagset.BoolVar(&flagDisableCache, "nocache", false, "true disables cache")
	// 	flagset.StringVar(&flagCacheDir, "cachedir", "", "path of the folder to be used as cache")
	// 	return flagset, nil
	// }
	return flagset, nil
}

func areFlagConsistent(flagset *flag.FlagSet, options map[string][]string) string {
	foundFlag := []string{}
	flagset.Visit(func(f *flag.Flag) {
		if !isInArray(f.Name, flagNotToCheck) {
			foundFlag = append(foundFlag, f.Name)
		}
	})
	// fmt.Printf("areFlagConsistent, foundFlag: %v\n", foundFlag)
	consistent := IsThisAnOption(foundFlag, options)
	return consistent
}

func IsThisAnOption(this []string, options map[string][]string) string {
	found := ""
	for name, opt := range options {
		if sameContent(this, opt) {
			found = name
			break
		}
	}
	return found
}

func isInArray(field string, array []string) bool {
	is := false
	for _, e := range array {
		if field == e {
			is = true
			break
		}
	}
	return is
}

func sameContent(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	same := true
	for _, e1 := range slice1 {
		if !isInArray(e1, slice2) {
			same = false
			break
		}
	}
	return same
}

func passwordtoBytes(password string) [32]byte {
	var pass [32]byte
	copy(pass[:], []byte(password))
	return pass
}

type Environment struct {
	passphrase    string
	log           bool
	help          bool
	workingDir    string
	key           string
	address       string
	password      [32]byte
	passwordSet   bool
	outFolder     string
	cacheFolder   string
	cacheDisabled bool
	cacheOnly     bool
}

func (e *Environment) passwordString() string {
	return string(bytes.Trim(e.password[:], string([]byte{0})))
}

func prepareEnvironment(args []string, flagset *flag.FlagSet) (*Environment, error) {
	tr := trace.New().Source("setup.go", "Environment", "BuildEnvironment")
	err := flagset.Parse(args[2:])
	if err != nil {
		trail.Println(trace.Alert("error while parsing args").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error while parsing args: %w", err)
	}
	env := Environment{}
	env.log = flagLog
	if env.log {
		trail.SetWriter(os.Stderr)
	}
	env.help = flagHelp

	env.workingDir, err = filepath.Abs(filepath.Dir(args[0]))
	if err != nil {
		trail.Println(trace.Alert("error gettign current working dir").Append(tr).UTC().Error(err))
		return nil, fmt.Errorf("error gettign current working dir")
	}

	keygenID := flagKeygenID
	if keygenID < 1 || keygenID > 2 {
		trail.Println(trace.Warning("keygenID out of range using default").Append(tr).UTC())
		keygenID = 2
	}
	trail.Println(trace.Info("keygenID defined").Append(tr).UTC().Add("keygenID", fmt.Sprintf("%d", keygenID)))

	passphrase, err := extractPassphrase(args)
	if err != nil {
		trail.Println(trace.Warning("passphrase not found").Append(tr).UTC().Error(err))
	}
	trail.Println(trace.Info("passphrase extracted").Append(tr).UTC().Add("passphrase", passphrase))
	if passphrase != "" {
		env.passphrase = passphrase
		env.key, env.password, err = processPassphrase(passphrase, int(keygenID))
		if err != nil {
			trail.Println(trace.Warning("error processing passphrase").Append(tr).UTC().Add("passphrase", passphrase).Error(err))
			return nil, fmt.Errorf("error procesing passphrase")
		}
		env.passwordSet = true
	}

	if flagBitcoinKey != "" {
		trail.Println(trace.Info("using key from command line").Append(tr).UTC().Add("key", flagBitcoinKey))
		env.key = flagBitcoinKey
	}
	if flagBitcoinAddress != "" && env.key == "" {
		trail.Println(trace.Info("using address from command line").Append(tr).UTC().Add("address", flagBitcoinAddress))
		env.address = flagBitcoinAddress
	}
	if flagPassword != "" {
		trail.Println(trace.Info("using password from command line").Append(tr).UTC().Add("flagPassword", flagPassword))
		for i := 0; i < len(env.password); i++ {
			env.password[i] = 0
		}
		copy(env.password[:], []byte(flagPassword))
		env.passwordSet = true
	}

	if flagOutputDir != "" {
		env.outFolder = flagOutputDir
	}
	if flagDisableCache && flagOnlyCache {
		trail.Println(trace.Warning("onlycache and nocache both enabled").Append(tr).UTC())
		return nil, fmt.Errorf("onlycache and nocache both enabled")
	}
	env.cacheDisabled = flagDisableCache
	env.cacheOnly = flagOnlyCache
	env.cacheFolder = flagCacheDir
	trail.Println(trace.Info("environment prepared").Append(tr).UTC().Add("key", env.key).Add("address", env.address).Add("password", env.passwordString()))
	return &env, nil
}

func prepareCache(env *Environment) (*ddb.TXCache, error) {
	tr := trace.New().Source("setup.go", "", "prepareCache")
	if env.cacheDisabled {
		trail.Println(trace.Info("cache disabled").Append(tr).UTC())
		return nil, nil
	}
	usercache, _ := os.UserCacheDir()
	cacheDir := filepath.Join(usercache, "trh")
	if env.cacheFolder != "" {
		cacheDir = flagCacheDir
	}
	cache, err := ddb.NewTXCache(cacheDir)
	if err != nil {
		trail.Println(trace.Info("error while building cache").Append(tr).UTC())
		return nil, fmt.Errorf("error while parsing args: %w", err)
	}
	return cache, nil
}

// func prepareDiary(env *Environment, cache *ddb.TXCache) (*ddb.FBranch, error) {
// 	tr := trace.New().Source("setup.go", "", "prepareCache")
// 	if env.key == "" && env.address == "" {
// 		trail.Println(trace.Alert("bitcoin key and address are both empty").Append(tr).UTC())
// 		return nil, fmt.Errorf("cannot prepare diary, bitcoin key and address are both empty")
// 	}
// 	var woc ddb.Explorer
// 	if !env.cacheOnly {
// 		woc = ddb.NewWOC()
// 	}
// 	taal := ddb.NewTAAL()
// 	blockchain := ddb.NewBlockchain(taal, woc, cache)
// 	var err error
// 	var diary *ddb.FBranch
// 	if env.key != "" {
// 		trail.Println(trace.Info("building Diary").Append(tr).UTC())
// 		diary, err = ddb.NewFBranch(env.key, env.password, blockchain)
// 		if err != nil {
// 			return nil, fmt.Errorf("error while creating a new Diary: %w", err)
// 		}
// 	} else {
// 		trail.Println(trace.Info("building read only Diary").Append(tr).UTC())
// 		diary, err = ddb.NewDiaryRO(env.address, env.password, blockchain)
// 		if err != nil {
// 			return nil, fmt.Errorf("error while creating a new read only Diary: %w", err)
// 		}
// 	}
// 	return diary, nil
// }
