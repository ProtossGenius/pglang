package grm_go

import (
	"fmt"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

const (
	//GrammarPackage package ...
	GrammarPackage = iota
	//GrammarList list when deal type(...) etc..
	GrammarList
	//GrammarTypeStruct struct.
	GrammarTypeStruct
	//GrammarTypeFunc type a func ..
	GrammarTypeFunc
	//GrammarFunc func.
	GrammarFunc
	//GrammarVar var.
	GrammarVar
	//GrammarConst const.
	GrammarConst
)

const (
	//ErrTypeNotMatch .
	ErrTypeNotMatch = "ErrTypeNotMatch: AnalysisReader %s, inputType [%s], inputValue[%s], reason %s "
)

//GrmGoPackage grammar product go package.
type GrmGoPackage struct {
	Name string
}

//ProductType .
func (g *GrmGoPackage) ProductType() int {
	return GrammarPackage
}

//PackageReader read package.
type PackageReader struct {
	first  bool
	result *GrmGoPackage
}

func read(input smn_analysis.InputItf) *lex_pgl.LexProduct {
	return input.Copy().(*lex_pgl.LexProduct)
}

//Name reader's name.
func (p *PackageReader) Name() string {
	return "PackageReader"
}

func (p *PackageReader) onErr(input *lex_pgl.LexProduct, reason string) error {
	return fmt.Errorf(ErrTypeNotMatch, p.Name(), lex_pgl.PglaNameMap[input.Type], input.Value, reason)
}

//PreRead only see if should stop read.
func (p *PackageReader) PreRead(stateNode *smn_analysis.StateNode,
	input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if p.first && lex.Value != "package" {
		return false, p.onErr(lex, "first input should be [package]")
	}

	if !p.first && lex.Type != lex_pgl.PGLA_PRODUCT_IDENT {
		return false, p.onErr(lex, "second input should be indent")
	}

	return true, nil
}

//Read real read. even isEnd == true the input be readed.
func (p *PackageReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if p.first && lex.Value == "package" {
		p.first = false
		return true, nil
	}

	p.result.Name = lex.Value

	return true, nil
}

//GetProduct return result.
func (p *PackageReader) GetProduct() smn_analysis.ProductItf {
	return p.result
}

//Clean let the Reader like new.  it will be call before first Read.
func (p *PackageReader) Clean() {
	p.result = &GrmGoPackage{}
}
