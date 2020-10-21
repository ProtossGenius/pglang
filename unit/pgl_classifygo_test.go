package unit_test

import (
	"io"
	"os"
	"testing"

	"github.com/ProtossGenius/pglang/analysis/classifygo"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

func goFile2String(t *testing.T, gf *classifygo.GoFile, out io.StringWriter) {
	writeln := func(str string) {
		_, err := out.WriteString(str + "\n")
		if err != nil {
			t.Fatal(err)
		}
	}

	writeln(gf.Package)

	for _, it := range gf.Imports {
		writeln(it.Alias + " " + it.Path)
	}

	for _, arr := range gf.Consts {
		writeln("#consts")

		for _, it := range arr {
			writeln(it.Name)
		}
	}

	for _, arr := range gf.Vars {
		writeln("#vars")

		for _, it := range arr {
			writeln(it.Name)
		}
	}

	for _, it := range gf.Funcs {
		writeln("func:" + it.FuncDef.Name)
	}

	for _, it := range gf.Structs {
		writeln("struct:" + it.Name)
	}

	for _, it := range gf.Interfaces {
		writeln("interface:" + it.Name)
	}

	for _, it := range gf.TypeFunc {
		writeln("type:" + it.Name)
	}

	for _, it := range gf.Aliases {
		writeln("aliase:" + it.Name)
	}
}

func classifygoAnalysis(t *testing.T, path string, list []*lex_pgl.LexProduct) {
	check := func(err error) {
		if err != nil {
			t.Fatal(path, ":", err)
		}
	}

	sm, gof := classifygo.NewAnalysiser()

	for _, lex := range list {
		err := sm.Read(lex)
		check(err)
	}

	goFile2String(t, gof, os.Stdout)
}

func TestClassifygo(t *testing.T) {
	lexWrite(t, LexOUnit, classifygoAnalysis)
}
