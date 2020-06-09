package grm_go

import (
	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

const (
	grmTypeStructDef = iota + GrammarTypeStruct*100
	grmTypeStructFieldDef
)

func isSymbol(input *lex_pgl.LexProduct) bool {
	return input.Type == lex_pgl.PGLA_PRODUCT_SYMBOL
}

//GrmGoStructFieldDef struct's field. care of go struct has tags.
type GrmGoStructFieldDef struct {
	FieldName string `json:"field_name"`
	FieldType string `json:"field_type"`
	FieldTags string `json:"field_tags"`
}

//ProductType .
func (g *GrmGoStructFieldDef) ProductType() int {
	return grmTypeStructFieldDef
}

//GrmGoStructDef go struct.
type GrmGoStructDef struct {
	Name string `json:"name"`
	//Type type Int int.
	Type   string                 `json:"type"`
	Fields []*GrmGoStructFieldDef `json:"fields"`
}

//ProductType result's type. usally should >= 0.
func (g *GrmGoStructDef) ProductType() int {
	return grmTypeStructDef
}

//GrmGoTypeStruct for GrammarTypeStruct.
type GrmGoTypeStruct struct {
	List []*GrmGoStructDef `json:"list"`
}

//ProductType result's type. usally should >= 0.
func (g *GrmGoTypeStruct) ProductType() int {
	return GrammarTypeStruct
}

//TypeStructReader inlucde type Struct struct{} and type(Struct1{} Struct2{}) .
type TypeStructReader struct {
	first  bool
	Result *GrmGoTypeStruct
}

//Name reader's name.
func (t *TypeStructReader) Name() string {
	return "TypeStructReader"
}

//PreRead only see if should stop read.
func (t *TypeStructReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if t.first {
		if !isIdent(lex) || lex.Value != "type" {
			return true, onErr(t, lex, "not start with [type].")
		}
		t.first = false
		return false, nil
	}
	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (t *TypeStructReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	panic("not implemented") // TODO: Implement
}

//GetProduct return result.
func (t *TypeStructReader) GetProduct() smn_analysis.ProductItf {
	return t.Result
}

//Clean let the Reader like new.  it will be call before first Read.
func (t *TypeStructReader) Clean() {
	t.Result = &GrmGoTypeStruct{}
	t.first = true
}

type grmTypeStructDefReader struct {
	Result    *GrmGoStructDef
	readName  bool
	checkBody bool //check is the body a type(type Int int).
}

//Name reader's name.
func (g *grmTypeStructDefReader) Name() string {
	return "grmTypeStructDefReader"
}

//PreRead only see if should stop read.
func (g *grmTypeStructDefReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if isSpace(lex) {
		return false, nil
	}
	if g.readName {
		if !isIdent(lex) {
			return true, onErr(g, lex, "Wait Struct Name.")
		}

		g.readName = false
		return false, nil
	}

	if g.checkBody {
		if !isIdent(lex) && (isSymbol(lex) && lex.Value != "{") {
			return true, onErr(g, lex, "Wait a type or struct body.")
		}

		g.checkBody = false
		return false, nil
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (g *grmTypeStructDefReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	panic("not implemented") // TODO: Implement
}

//GetProduct return result.
func (g *grmTypeStructDefReader) GetProduct() smn_analysis.ProductItf {
	return g.Result
}

//Clean let the Reader like new.  it will be call before first Read.
func (g *grmTypeStructDefReader) Clean() {
	g.Result = &GrmGoStructDef{}
	g.readName = true
	g.checkBody = true
}

type grmTypeStructFieldReader struct {
	Result *GrmGoStructFieldDef
	count  int
}

//reader's name
func (g *grmTypeStructFieldReader) Name() string {
	return "grmTypeStructFieldReader"
}

//only see if should stop read.
func (g *grmTypeStructFieldReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	g.count++
	if g.count == 1 {

	}
}

//real read. even isEnd == true the input be readed.
func (g *grmTypeStructFieldReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	panic("not implemented") // TODO: Implement
}

//return result
func (g *grmTypeStructFieldReader) GetProduct() smn_analysis.ProductItf {
	return g.Result
}

//let the Reader like new.  it will be call before first Read
func (g *grmTypeStructFieldReader) Clean() {
	g.Result = &GrmGoStructFieldDef{}
	g.count = 0
}
