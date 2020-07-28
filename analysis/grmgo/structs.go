package grmgo

import "github.com/ProtossGenius/pglang/analysis/lex_pgl"

//GoFile anything in a go file.
type GoFile struct {
	Package    string
	Imports    []string
	Consts     []GoConst
	Vars       []GoVar
	Funcs      []GoFunc //global func
	Structs    []GoStruct
	Interfaces []GoItf
	Aliases    []GoAlias
	TypeFunc   []GoTypeFunc
}

//GoCodes the code witch not analysis.
type GoCodes []*lex_pgl.LexProduct

//GoConst read until '\n' (if meet char ',' Ignore the line's '\n' ).
type GoConst struct {
	Name string
	Code GoCodes
}

//GoVar like GoConst.
type GoVar struct {
	Name string
	Code GoCodes
}

//GoFuncDef .
type GoFuncDef struct {
	Name    string
	Params  GoCodes
	Returns GoCodes
}

//GoFunc func <()> XXXX () <()> {}.
type GoFunc struct {
	Scope   GoCodes
	FuncDef *GoFuncDef
	Codes   GoCodes
}

//GoStruct only type XXXX struct {XXXX} .
type GoStruct struct {
	Name  string
	Codes GoCodes
}

//GoItf type XXXX interface { []GoFuncDef }.
type GoItf struct {
	Name     string
	FuncDefs []*GoFuncDef
}

//GoAlias like GoConst .
type GoAlias struct {
	Name  string
	Codes GoCodes
}

//GoTypeFunc type XXXX func () ()  .
type GoTypeFunc struct {
	Name    string
	Params  GoCodes
	Returns GoCodes
}
