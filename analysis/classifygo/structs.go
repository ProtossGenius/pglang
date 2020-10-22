package classifygo

import (
	"strings"

	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

// GoFile anything in a go file.
type GoFile struct {
	Package    string
	Imports    []*GoImport
	Consts     []GoBatchGlobals
	Vars       []GoBatchGlobals
	Funcs      []*GoFunc // global func
	Structs    []*GoStruct
	Interfaces []*GoItf
	TypeFunc   []*GoTypeFunc
	Aliases    []*GoAlias
}

// ProductType result's type. usually should >= 0.
func (g *GoFile) ProductType() int {
	return 1
}

// GoCodes the code witch not analysis.
type GoCodes []*lex_pgl.LexProduct

// Print .
func (gc GoCodes) String() string {
	arr := make([]string, 0, len(gc))
	for _, it := range gc {
		arr = append(arr, it.Value)
	}

	return strings.Join(arr, "")
}

// GoImport .
type GoImport struct {
	Path  string
	Alias string
}

// GoBatchGlobals array of GoGlobals.
type GoBatchGlobals []*GoGlobals

// GoGlobals read until '\n' (if meet char ',' Ignore the line's '\n' ).
type GoGlobals struct {
	Name string
	Code GoCodes
}

// GoFuncDef .
type GoFuncDef struct {
	Name    string
	Params  GoCodes
	Returns GoCodes
}

// GoFunc func <()> XXXX () <()> {}.
type GoFunc struct {
	Scope   GoCodes
	FuncDef *GoFuncDef
	Codes   GoCodes
}

// GoStruct only type XXXX struct {XXXX} .
type GoStruct struct {
	Name  string
	Codes GoCodes
}

// GoItf type XXXX interface { []GoFuncDef }.
type GoItf struct {
	Name  string
	Codes GoCodes
}

// GoAlias like GoConst .
type GoAlias struct {
	Name  string
	Codes GoCodes
}

// GoTypeFunc type XXXX func () ()  .
type GoTypeFunc struct {
	Name    string
	Params  GoCodes
	Returns GoCodes
}
