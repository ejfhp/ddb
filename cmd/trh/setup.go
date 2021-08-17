package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ejfhp/ddb"
	"github.com/ejfhp/trail"
	"github.com/ejfhp/trail/trace"
)

var (
	flagLog            bool
	flagHelp           bool
	flagFilename       string
	flagOutputDir      string
	flagCacheDir       string
	flagDisableCache   bool
	flagBitcoinAddress string
	flagBitcoinKey     string
	flagPassword       string
	flagKeygenID       int64
)

type Environment struct {
	log           bool
	help          bool
	workingDir    string
	key           string
	address       string
	password      [32]byte
	outFolder     string
	cacheFolder   string
	cacheDisabled bool
}

func prepareEnvironment(args []string, flagset *flag.FlagSet) (*Environment, error) {
	tr := trace.New().Source("setup.go", "Environment", "BuildEnvironment")
	err := flagset.Parse(args)
	if err != nil {
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

	passphrase, err := extractPassphrase(os.Args)
	if err != nil {
		trail.Println(trace.Warning("passphrase not found").Append(tr).UTC().Error(err))
	}
	if passphrase != "" {
		env.key, env.password, err = processPassphrase(passphrase, int(keygenID))
		if err != nil {
			trail.Println(trace.Warning("error processing passphrase").Append(tr).UTC().Add("passphrase", passphrase).Error(err))
			return nil, fmt.Errorf("error procesing passphrase")
		}
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
		copy(env.password[:], []byte(flagPassword))
	}

	if flagOutputDir != "" {
		env.outFolder = flagOutputDir
	}
	env.cacheDisabled = flagDisableCache
	env.cacheFolder = flagCacheDir

	if env.key == "" && env.address == "" {
		trail.Println(trace.Alert("bitcoin key and address are both empty").Append(tr).UTC())
		return nil, fmt.Errorf("bitcoin key and address are both empty")
	}
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

func prepareDiary(env *Environment, cache *ddb.TXCache) (*ddb.Diary, error) {
	tr := trace.New().Source("setup.go", "", "prepareCache")
	woc := ddb.NewWOC()
	taal := ddb.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	var err error
	var diary *ddb.Diary
	if env.key != "" {
		trail.Println(trace.Info("building Diary").Append(tr).UTC())
		diary, err = ddb.NewDiary(env.key, env.password, blockchain)
		if err != nil {
			return nil, fmt.Errorf("error while creating a new Diary: %w", err)
		}
	} else {
		trail.Println(trace.Info("building read only Diary").Append(tr).UTC())
		diary, err = ddb.NewDiaryRO(env.address, env.password, blockchain)
		if err != nil {
			return nil, fmt.Errorf("error while creating a new read only Diary: %w", err)
		}
	}
	return diary, nil
}

func newFlagset(command string) *flag.FlagSet {
	flagset := flag.NewFlagSet(command, flag.ContinueOnError)
	flagset.BoolVar(&flagLog, "log", false, "true enables log output")
	flagset.BoolVar(&flagHelp, "help", false, "print help")
	flagset.BoolVar(&flagHelp, "h", false, "print help")
	flagset.StringVar(&flagBitcoinAddress, "address", "", "bitcoin address")
	flagset.StringVar(&flagBitcoinKey, "key", "", "bitcoin key")
	flagset.StringVar(&flagPassword, "password", "", "encryption password")
	flagset.Int64Var(&flagKeygenID, "keygen", 2, "keygen to be used for key and password generation")
	//DESCRIBE
	if command == commandDescribe {
		flagset := flag.NewFlagSet("describe", flag.ContinueOnError)
		flagset.StringVar(&flagBitcoinAddress, "address", "", "bitcoin address")
		return flagset
	}
	//RETRIEVE
	if command == commandRetrieveAll {
		flagset.StringVar(&flagOutputDir, "outdir", "", "path of the folder where to save retrived files")
		flagset.BoolVar(&flagDisableCache, "nocache", false, "true disable cache")
		flagset.StringVar(&flagCacheDir, "cachedir", "", "path of the folder to be used as cache")
		return flagset
	}
	return flagset
}

func extractPassphrase(args []string) (string, error) {
	startidx := -1
	for i, t := range args {
		if t == "+" {
			startidx = i + 1
		}
	}
	if startidx < 0 || startidx >= len(args) {
		return "", fmt.Errorf("passphrase is missing")
	}
	passphrase := strings.Join(args[startidx:], " ")
	return passphrase, nil
}

func processPassphrase(passphrase string, keygenID int) (string, [32]byte, error) {
	var passnum int
	reg, err := regexp.Compile("[^0-9 ]+")
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("error compiling regexp: %w", err)
	}
	phrasenum := reg.ReplaceAllString(passphrase, "")
	for _, n := range strings.Split(phrasenum, " ") {
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
		return "", [32]byte{}, fmt.Errorf("passphrase must contain a number")
	}
	var keygen ddb.Keygen
	if keygenID == 2 {
		keygen, err = ddb.NewKeygen2(passnum, passphrase)
		if err != nil {
			return "", [32]byte{}, fmt.Errorf("error building Keygen2: %w", err)
		}
	}
	wif, err := keygen.WIF()
	if err != nil {
		return "", [32]byte{}, fmt.Errorf("error while generating bitcoin key: %w", err)
	}
	password := keygen.Password()
	return wif, password, nil
}
