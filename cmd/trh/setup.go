package main

import (
	"bytes"
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
	flagFile           string
	flagOutputDir      string
	flagCacheDir       string
	flagDisableCache   bool
	flagOnlyCache      bool
	flagBitcoinAddress string
	flagBitcoinKey     string
	flagPassword       string
	flagKeygenID       int64
)

func newFlagset(command string) *flag.FlagSet {
	flagset := flag.NewFlagSet(command, flag.ContinueOnError)
	flagset.BoolVar(&flagLog, "log", false, "true enables log output")
	flagset.BoolVar(&flagHelp, "help", false, "prints help")
	flagset.BoolVar(&flagHelp, "h", false, "prints help")
	flagset.StringVar(&flagBitcoinAddress, "address", "", "bitcoin address")
	flagset.StringVar(&flagBitcoinKey, "key", "", "bitcoin key")
	flagset.StringVar(&flagPassword, "password", "", "encryption password")
	flagset.Int64Var(&flagKeygenID, "keygen", 2, "keygen to be used for key and password generation")
	//DESCRIBE
	if command == commandDescribe {
		return flagset
	}
	//RETRIEVE
	if command == commandRetrieveAll {
		flagset.StringVar(&flagOutputDir, "outdir", "", "path of the folder where to save retrived files")
		flagset.BoolVar(&flagDisableCache, "nocache", false, "true disables cache")
		flagset.BoolVar(&flagOnlyCache, "onlycache", false, "true retrieves only from cache")
		flagset.StringVar(&flagCacheDir, "cachedir", "", "path of the folder to be used as cache")
		return flagset
	}
	//STORE
	if command == commandStore {
		flagset.StringVar(&flagFile, "file", "", "path of file to store")
		flagset.BoolVar(&flagDisableCache, "nocache", false, "true disables cache")
		flagset.StringVar(&flagCacheDir, "cachedir", "", "path of the folder to be used as cache")
		return flagset
	}
	return flagset
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

func prepareDiary(env *Environment, cache *ddb.TXCache) (*ddb.FBranch, error) {
	tr := trace.New().Source("setup.go", "", "prepareCache")
	if env.key == "" && env.address == "" {
		trail.Println(trace.Alert("bitcoin key and address are both empty").Append(tr).UTC())
		return nil, fmt.Errorf("cannot prepare diary, bitcoin key and address are both empty")
	}
	var woc ddb.Explorer
	if !env.cacheOnly {
		woc = ddb.NewWOC()
	}
	taal := ddb.NewTAAL()
	blockchain := ddb.NewBlockchain(taal, woc, cache)
	var err error
	var diary *ddb.FBranch
	if env.key != "" {
		trail.Println(trace.Info("building Diary").Append(tr).UTC())
		diary, err = ddb.NewFBranch(env.key, env.password, blockchain)
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