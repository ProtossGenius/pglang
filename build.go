package main

import (
	"fmt"
	"strings"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func SymbolVarCfg() {
	fmt.Println("[start]read symbol list from file and write to code")
	defer fmt.Println("[end]read symbol list from file and write to code")

	datas, err := smn_file.FileReadAll("./datas/analysis/pgl_ana_lex/symbol.cfg")
	check(err)
	fo, err := smn_file.CreateNewFile("./analysis/pgl_ana_lex/cfg_vars.go")
	defer fo.Close()
	check(err)
	write := func(str string) {
		_, err = fo.WriteString(str)
		check(err)
	}
	writef := func(format string, a ...interface{}) {
		write(fmt.Sprintf(format, a...))
	}
	check(err)
	charMap := map[rune]bool{}
	write(`package pgl_ana_lex

var SymbolList = map[string]bool{`)
	smbList := strings.Split(string(datas), "\n")
	for i := range smbList {
		smbList[i] = strings.TrimSpace(smbList[i])
		smbList[i] = strings.Replace(smbList[i], "\\", "\\\\", -1)
		line := smbList[i]
		if line == "" {
			continue
		}
		for _, char := range line {
			charMap[char] = true
		}
		writef("\"%s\":true,", line)
	}

	write(`}

var SymbolCharSet = map[rune]bool{`)
	for char := range charMap {
		if char == '\\' {
			writef(`'\\':true,`)
		} else {
			writef("'%c':true,", char)
		}

	}
	ccMap := map[string]bool{"": true}
	write(`}

var SymbolCanContinue = map[string]bool{`)

	for _, c1 := range smbList {
		for _, c2 := range smbList {
			if ccMap[c2] || c1 == c2 {
				continue
			}
			if strings.HasPrefix(c1, c2) {
				writef("\"%s\":true, ", c2)
				ccMap[c2] = true
			}
		}
	}
	write(`}

//some maybe define in another type, but not as symbol. like comment's "//" and "/*"
var SymbolUnuse = map[string]bool{"//":true, "/*":true}
`)
}

func main() {
	fmt.Println("$$$$$$$$$$$$$$$$$$$ start build project $$$$$$$$$$$$$$$$$$$$$")
	//read symbol list from file and write to code
	SymbolVarCfg()
	fmt.Println("SUCCESS")
}
