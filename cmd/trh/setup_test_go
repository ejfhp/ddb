package main

import (
	"strconv"
	"strings"
	"testing"
)

func Test_IsThisAnOption(t *testing.T) {
	options := map[string][]string{
		"1": {"uno", "due", "tre", "quattro", "cinque"},
		"2": {"uno", "due", "tre", "quattro"},
		"3": {"due", "tre", "quattro", "cinque"},
		"4": {"uno", "tre", "quattro", "cinque"},
		"5": {"uno", "due", "cinque"},
		"6": {},
	}
	this := map[string][]string{
		"1_":  {"uno"},
		"2_1": {"uno", "due", "tre", "quattro", "cinque"},
		"3_":  {"uno", "due", "tre", "quattro", "cinque", "sei"},
		"4_5": {"uno", "due", "cinque"},
		"5_1": {"uno", "due", "quattro", "cinque", "tre"},
		"6_":  {"uno", "due", "due", "quattro", "cinque", "tre"},
		"7_":  {"nove"},
		"8_6": {},
	}
	for k, th := range this {
		spl := strings.Split(k, "_")
		i, _ := strconv.ParseInt(spl[0], 10, 64)
		exp := spl[1]
		if val, _ := IsThisAnOption(th, options); val != exp {
			t.Logf("Check of this number %d failed: '%s'", i, exp)
			t.Fail()
		}
	}

}
