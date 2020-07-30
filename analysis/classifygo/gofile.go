package classifygo

import (
	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

//NewAnalysiser new analysiser.
func NewAnalysiser() *smn_analysis.StateMachine {
	goFile := &GoFile{}
	sm := new(smn_analysis.StateMachine).Init()
	dft := smn_analysis.NewDftStateNodeReader(sm)
	dft.Register(&CFGoReadPackage{goFile: goFile})
	return sm
}

//CFGoReadPackage read pkg.
type CFGoReadPackage struct {
	goFile *GoFile
	first  bool
}

//Name reader's name.
func (c *CFGoReadPackage) Name() string {
	return "CFGoReadPackage"
}

//PreRead only see if should stop read.
func (c *CFGoReadPackage) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if lex_pgl.IsSpace(lex) || lex_pgl.IsComment(lex) {
		return false, nil
	}

	if c.first {
		if !lex.Equal(ConstPackage) {
			return true, onErr(c, lex, "first ident want package.")
		}

		if c.goFile.Package != "" {
			return true, onErr(c, lex, "not first package.")
		}
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (c *CFGoReadPackage) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if lex_pgl.IsSpace(lex) || lex_pgl.IsComment(lex) {
		return false, nil
	}

	if c.first {
		c.first = false
		return false, nil
	}

	if !lex_pgl.IsIdent(lex) {
		return true, onErr(c, lex, "package name want a ident")
	}

	c.goFile.Package = lex.Value

	return true, nil
}

//End when end read.
func (c *CFGoReadPackage) End(stateNode *smn_analysis.StateNode) (isEnd bool, err error) {
	if c.first {
		return true, onErr(c, nil, "unexcpt EOF")
	}

	stateNode.SendProduct(c.goFile)

	return true, nil
}

//GetProduct return result.
func (c *CFGoReadPackage) GetProduct() smn_analysis.ProductItf {
	return nil
}

//Clean let the Reader like new.  it will be call before first Read.
func (c *CFGoReadPackage) Clean() {
	c.first = true
}
