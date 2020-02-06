package pgl_ana_lex

import (
	"fmt"
	"unicode"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
)

const (
	PGLA_PRODUCT_ = iota
	PGLA_PRODUCT_IDENT
	PGLA_PRODUCT_SPACE
)
const (
	ErrTypeNotMatch = "ErrTypeNotMatch: AnalysisReader %s, input [%s], reason %s "
)

type PglaInput struct {
	Char rune
}

func (p *PglaInput) Copy() smn_analysis.InputItf {
	return &PglaInput{Char: p.Char}
}
func read(input smn_analysis.InputItf) rune {
	return input.(*PglaInput).Char
}

type PglaIdent struct {
	Name string
}

func (p *PglaIdent) ProductType() int {
	return PGLA_PRODUCT_IDENT
}

type IdentifierReader struct {
	first  bool
	result *PglaIdent
}

func NewLexAnalysiser() *smn_analysis.StateMachine {
	sm := new(smn_analysis.StateMachine).Init()
	dft := smn_analysis.NewDftStateNodeReader(sm)
	dft.Register(&IdentifierReader{})
	dft.Register(&SpaceReader{})
	return sm
}
func (this *IdentifierReader) onErr(inputs, reason string) error {
	return fmt.Errorf(ErrTypeNotMatch, "IdentifierReader", inputs, reason)
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
	this.result.Name += charStr
	return false, nil
}

//return result
func (this *IdentifierReader) GetProduct() smn_analysis.ProductItf {
	return this.result
}

//let the Reader like new.  it will be call before first Read
func (this *IdentifierReader) Clean() {
	this.first = true
	this.result = &PglaIdent{}
}

type PglaSpace struct {
	Char rune
}

func (p *PglaSpace) ProductType() int {
	return PGLA_PRODUCT_SPACE
}

type SpaceReader struct {
	Result *PglaSpace
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
	p.Result = &PglaSpace{Char: char}
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
