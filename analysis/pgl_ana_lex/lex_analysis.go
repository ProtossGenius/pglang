package pgl_ana_lex

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
)

const (
	PGLA_PRODUCT_ = iota
	PGLA_PRODUCT_IDENT
	PGLA_PRODUCT_SPACE
	PGLA_PRODUCT_SYMBOL
	PGLA_PRODUCT_COMMENT
)
const (
	ErrTypeNotMatch = "ErrTypeNotMatch: AnalysisReader %s, input [%s], reason %s "
)

type PglaInput struct {
	Char rune
}

type LexProduct struct {
	Type  int    `json:"type"`
	Value string `json:"value"`
}

func (this *LexProduct) ProductType() int {
	return this.Type
}

func (p *PglaInput) Copy() smn_analysis.InputItf {
	return &PglaInput{Char: p.Char}
}

func ToLexProduct(input smn_analysis.ProductItf) *LexProduct {
	product := &LexProduct{Type: input.ProductType()}
	switch product.Type {
	case -1:
		product.Value = "end"
		return product
	case -2:
		product.Value = input.(*smn_analysis.ProductDftNode).Reason
		return product
	}
	return input.(*LexProduct)
}

func read(input smn_analysis.InputItf) rune {
	return input.(*PglaInput).Char
}

type IdentifierReader struct {
	first  bool
	result *LexProduct
}

func NewLexAnalysiser() *smn_analysis.StateMachine {
	sm := new(smn_analysis.StateMachine).Init()
	dft := smn_analysis.NewDftStateNodeReader(sm)
	dft.Register(&IdentifierReader{})
	dft.Register(&SpaceReader{})
	dft.Register(&SymbolReader{})
	dft.Register(&CommentReader{})
	return sm
}
func (this *IdentifierReader) onErr(inputs, reason string) error {
	return fmt.Errorf(ErrTypeNotMatch, "IdentifierReader", inputs, reason)
}
func (this *IdentifierReader) Name() string {
	return "IdentifierReader"
}

//only see if should stop read.
func (this *IdentifierReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	char := read(input)
	charStr := string([]rune{char})
	if this.first {
		if unicode.IsDigit(char) {
			return true, this.onErr(charStr, "First char can't be Number")
		}
		if !unicode.IsLetter(char) && char != '_' {
			return true, this.onErr(charStr, "Not Identifier")
		}
	} else {
		if !unicode.IsDigit(char) && !unicode.IsLetter(char) && char != '_' {
			return true, nil
		}
	}
	return false, nil
}

//real read. even isEnd == true the input be readed.
func (this *IdentifierReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	this.first = false
	char := read(input)
	charStr := string([]rune{char})
	this.result.Value += charStr
	return false, nil
}

//return result
func (this *IdentifierReader) GetProduct() smn_analysis.ProductItf {
	return this.result
}

//let the Reader like new.  it will be call before first Read
func (this *IdentifierReader) Clean() {
	this.first = true
	this.result = &LexProduct{Type: PGLA_PRODUCT_IDENT}
}

type SpaceReader struct {
	Result *LexProduct
}

func (p *SpaceReader) Name() string {
	return "SpaceReader"
}

//only see if should stop read.
func (p *SpaceReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	char := read(input)
	charStr := string([]rune{char})
	if unicode.IsSpace(char) {
		return false, nil
	}
	return true, fmt.Errorf(ErrTypeNotMatch, "SpaceReader", charStr, "Not Space char")
}

//real read. even isEnd == true the input be readed.
func (p *SpaceReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	char := read(input)
	charStr := string([]rune{char})
	p.Result = &LexProduct{Type: PGLA_PRODUCT_SPACE, Value: charStr}
	return true, nil
}

//return result
func (p *SpaceReader) GetProduct() smn_analysis.ProductItf {
	return p.Result
}

//let the Reader like new.  it will be call before first Read
func (p *SpaceReader) Clean() {
	p.Result = nil
}

type SymbolReader struct {
	Result string
}

//reader's name
func (s *SymbolReader) Name() string {
	return "SymbolReader"
}

func (this *SymbolReader) onErr(inputs, reason string) error {
	return fmt.Errorf(ErrTypeNotMatch, "SymbolReader", inputs, reason)
}

//only see if should stop read.
func (s *SymbolReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	char := read(input)
	charStr := string([]rune{char})
	first := s.Result == ""
	nres := s.Result + charStr
	if !SymbolCharSet[char] {
		if first || !SymbolList[s.Result] {
			return true, s.onErr(nres, "not in SymbolCharSet")
		} else {
			return true, nil
		}
	} else {
		if SymbolUnuse[nres] {
			return true, s.onErr(nres, "not in SymbolList")
		}
		if !SymbolCanContinue[nres] && !SymbolList[nres] {
			return true, nil
		}
	}
	return false, nil
}

//real read. even isEnd == true the input be readed.
func (s *SymbolReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	char := read(input)
	charStr := string([]rune{char})
	s.Result += charStr
	return false, nil
}

//return result
func (s *SymbolReader) GetProduct() smn_analysis.ProductItf {
	return &LexProduct{Type: PGLA_PRODUCT_SYMBOL, Value: s.Result}
}

//let the Reader like new.  it will be call before first Read
func (s *SymbolReader) Clean() {
	s.Result = ""
}

type CommentReader struct {
	Result      string
	mutiLineCmt bool
}

func (this *CommentReader) onErr(inputs, reason string) error {
	return fmt.Errorf(ErrTypeNotMatch, "CommentReader", inputs, reason)
}

//reader's name
func (c *CommentReader) Name() string {
	return "CommentReader"
}

//only see if should stop read.
func (c *CommentReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	char := read(input)
	charStr := string([]rune{char})
	nres := c.Result + charStr
	if c.Result == "" {
		if char != '/' {
			return true, c.onErr(charStr, "not comment")
		}
	} else if len(nres) == 2 {
		if nres == "/*" {
			c.mutiLineCmt = true
		}
		if nres != "//" && nres != "/*" {
			return true, c.onErr(charStr, "not comment")
		}
	}
	if char == '\n' && strings.HasPrefix(nres, "//") {
		return true, nil
	}
	return false, nil
}

//real read. even isEnd == true the input be readed.
func (c *CommentReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	char := read(input)
	charStr := string([]rune{char})
	c.Result += charStr
	if c.mutiLineCmt && strings.HasSuffix(c.Result, "*/") { // muti line comment. start with "/*"
		return true, nil
	}
	return false, nil
}

//return result
func (c *CommentReader) GetProduct() smn_analysis.ProductItf {
	return &LexProduct{Type: PGLA_PRODUCT_COMMENT, Value: c.Result}
}

//let the Reader like new.  it will be call before first Read
func (c *CommentReader) Clean() {
	c.Result = ""
	c.mutiLineCmt = false
}
