package main

import (
	"strconv"
	"strings"
	"testing"
)

func Test_IsThisAnOption(t *testing.T) {
	options := [][]string{
		{"uno", "due", "tre", "quattro", "cinque"},
		{"uno", "due", "tre", "quattro"},
		{"due", "tre", "quattro", "cinque"},
		{"uno", "tre", "quattro", "cinque"},
		{"uno", "due", "cinque"},
		{},
	}
	this := map[string][]string{
		"1_false": {"uno"},
		"2_true":  {"uno", "due", "tre", "quattro", "cinque"},
		"3_false": {"uno", "due", "tre", "quattro", "cinque", "sei"},
		"4_true":  {"uno", "due", "cinque"},
		"5_true":  {"uno", "due", "quattro", "cinque", "tre"},
		"6_false": {"uno", "due", "due", "quattro", "cinque", "tre"},
		"7_false": {"nove"},
		"8_true":  {},
	}
	for k, th := range this {
		spl := strings.Split(k, "_")
		i, _ := strconv.ParseInt(spl[0], 10, 64)
		exp, _ := strconv.ParseBool(spl[1])
		if IsThisAnOption(th, options) != exp {
			t.Logf("Check of this number %d failed", i)
			t.Fail()
		}
	}

}
