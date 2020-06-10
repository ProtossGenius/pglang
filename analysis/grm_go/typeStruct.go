package grm_go

import (
	"fmt"

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

func isString(input *lex_pgl.LexProduct) bool {
	return input.Type == lex_pgl.PGLA_PRODUCT_STRING
}

func isLineBreak(input *lex_pgl.LexProduct) bool {
	return isSpace(input) && input.Value == "\n"
}

func isComment(input *lex_pgl.LexProduct) bool {
	return input.Type == lex_pgl.PGLA_PRODUCT_COMMENT
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
	Result      *GrmGoStructDef
	fieldReader *grmTypeStructFieldReader
	count       int
}

//Name reader's name.
func (g *grmTypeStructDefReader) Name() string {
	return "grmTypeStructDefReader"
}

//PreRead only see if should stop read.
func (g *grmTypeStructDefReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (g *grmTypeStructDefReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if isSpace(lex) && lex.Value != "\n" {
		return false, nil
	}

	if isComment(lex) {
		return false, nil
	}

	const (
		countReadName  = 1
		countReadType  = 2 //for type Int int.
		countReadPoint = 3
		countReadBody  = 4
	)
	result := g.Result
	switch g.count {
	case countReadName:
		if isIdent(lex) {
			result.Name = lex.Value
			g.count = countReadType
			return false, nil
		}

		return true, onErr(g, lex, "want a ident for name.")
	case countReadType:
		if isIdent(lex) {
			if lex.Value == "func" || lex.Value == "interface" {
				return true, onErr(g, lex, "wait a type or type name")
			}

			result.Type = lex.Value
			g.count = countReadBody
			return true, nil
		}

		return true, onErr(g, lex, "want [indent] as type")
	case countReadBody:
		if isSymbol(lex) && lex.Value == "}" {
			if g.fieldReader.IsNew() {
				return true, nil
			}

			if g.fieldReader.Dirty != "" {
				return true, onErr(g, lex, fmt.Sprintf("when read field get dirty data: info[%s] data[%v]", g.fieldReader.Dirty, g.fieldReader.Result))
			}

			result.Fields = append(result.Fields, g.fieldReader.Result)
			return true, nil
		}

		end, err := g.fieldReader.Read(nil, input)

		if err != nil {
			return isEnd, onErr(g, lex, fmt.Sprintf("when read body error : %s", err.Error()))
		}

		if end {
			result.Fields = append(result.Fields, g.fieldReader.Result)
		}
	}

	return false, nil
}

//GetProduct return result.
func (g *grmTypeStructDefReader) GetProduct() smn_analysis.ProductItf {
	return g.Result
}

//Clean let the Reader like new.  it will be call before first Read.
func (g *grmTypeStructDefReader) Clean() {
	g.Result = &GrmGoStructDef{}
	if g.fieldReader == nil {
		g.fieldReader = &grmTypeStructFieldReader{}
	}
	g.fieldReader.Clean()
	g.count = 1
}

type grmTypeStructFieldReader struct {
	Result *GrmGoStructFieldDef
	count  int
	Dirty  string
}

//reader's name
func (g *grmTypeStructFieldReader) Name() string {
	return "grmTypeStructFieldReader"
}

//only see if should stop read.
func (g *grmTypeStructFieldReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	return false, nil
}

//real read. even isEnd == true the input be readed.
func (g *grmTypeStructFieldReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	g.Dirty = ""
	const (
		countReadName = 1
		countReadType = 2
		countReadTags = 3
		countReadLBrk = 4 //line break \n.
	)
	if isComment(lex) || (isSpace(lex) && lex.Value != "\n") {
		return false, nil
	}

	switch g.count {
	case countReadName:
		if isLineBreak(lex) {
			return false, nil
		}

		if isIdent(lex) {
			g.Result.FieldName = lex.Value
			g.count = countReadType
			return false, nil
		}

		return true, onErr(g, lex, "want a ident")
	case countReadType:
		if isLineBreak(lex) {
			g.Result.FieldType = g.Result.FieldName
			g.Result.FieldName = ""
			return true, nil
		}

		if isSymbol(lex) {
			if lex.Value != "." {
				return true, onErr(g, lex, "want a symbol [.]")
			}
			g.Dirty = "Waiting a Type"
			g.Result.FieldType = g.Result.FieldName + "."
			g.Result.FieldName = ""
			g.count = countReadType
			return false, nil
		}

		if isIdent(lex) {
			g.Result.FieldType += lex.Value
			g.count = countReadTags
			return false, nil
		}

		if isString(lex) {
			g.Result.FieldTags = lex.Value
			g.count = countReadLBrk
			return false, nil
		}

		return true, onErr(g, lex, "want [Line Break] or [.] or [ident] or [tags]")
	case countReadTags:
		if isLineBreak(lex) {
			return true, nil
		}

		if isSymbol(lex) && lex.Value == "." {
			g.Dirty = "Waiting a type"
			g.Result.FieldType += "."
			g.count = countReadType
			return false, nil
		}

		if isString(lex) {
			g.Result.FieldTags = lex.Value
			g.count = countReadLBrk
			return false, nil
		}

		return true, onErr(g, lex, "want [Line Break] or [tags]")
	case countReadLBrk:
		if isLineBreak(lex) {
			return true, nil
		}

		return true, onErr(g, lex, "want [Line Break]")
	default:
		return true, onErr(g, lex, fmt.Sprintf("unexcept g.count = %d", g.count))
	}
}

//return result
func (g *grmTypeStructFieldReader) GetProduct() smn_analysis.ProductItf {
	return g.Result
}

//IsNew .
func (g *grmTypeStructFieldReader) IsNew() bool {
	result := g.Result
	return result.FieldName == "" && result.FieldType == "" && result.FieldTags == ""
}

//let the Reader like new.  it will be call before first Read
func (g *grmTypeStructFieldReader) Clean() {
	g.Result = &GrmGoStructFieldDef{}
	g.count = 1
	g.Dirty = ""
}
