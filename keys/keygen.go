package keys

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	KeygenVersion1 = 1
	KeygenVersion2 = 2
	KeygenVersion3 = 3
)

type Keygen interface {
	Init(number int, phrase string) error
	WIF() (string, error)
	Password() ([32]byte, error)
}

func MakeKeygen(version int) (Keygen, error) {
	switch version {
	case KeygenVersion1:
		return &Keygen1{}, nil
	case KeygenVersion2:
		return &Keygen2{}, nil
	case KeygenVersion3:
		return &Keygen3{}, nil
	}
	return nil, fmt.Errorf("Keygen version not found: %d", version)

}

func exctactNum(passphrase string) (int, error) {
	var passnum int
	reg, err := regexp.Compile("[^0-9 ]+")
	if err != nil {
		return 0, fmt.Errorf("error compiling regexp: %w", err)
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
		return 0, fmt.Errorf("passphrase must contain a number not 0")
	}
	return passnum, nil
}

//FromPassphrase generate WIF and password with the given keygen
func FromPassphrase(passphrase string, keygenVersion int) (string, string, error) {
	var passnum int
	reg, err := regexp.Compile("[^0-9 ]+")
	if err != nil {
		return "", "", fmt.Errorf("error compiling regexp: %w", err)
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
		return "", "", fmt.Errorf("passphrase must contain a number")
	}
	keygen, err := MakeKeygen(keygenVersion)
	if err != nil {
		return "", "", fmt.Errorf("error building Keygen: %w", err)
	}
	keygen.Init(passnum, passphrase)
	wif, err := keygen.WIF()
	if err != nil {
		return "", "", fmt.Errorf("error while generating bitcoin key: %w", err)
	}
	password, err := keygen.Password()
	if err != nil {
		return "", "", fmt.Errorf("error while generating password: %w", err)
	}
	return wif, passwordToString(password), nil
}

// func PasswordFromString(pwd string) [32]byte {
// 	var password = [32]byte{}
// 	for i := 0; i < len(password); i++ {
// 		password[i] = '#'
// 	}
// 	copy(password[:], pwd[:])
// 	return password
// }

// func PasswordToString(password [32]byte) string {
// 	return strings.TrimSpace(string(password[:]))
// }
