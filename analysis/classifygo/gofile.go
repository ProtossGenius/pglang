package classifygo

import (
	"fmt"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

//NewAnalysiser new analysiser.
func NewAnalysiser() (*smn_analysis.StateMachine, *GoFile) {
	goFile := &GoFile{}
	sm := new(smn_analysis.StateMachine).Init()
	dft := smn_analysis.NewDftStateNodeReader(sm)
	dft.Register(&CFGoReadPackage{goFile: goFile})
	dft.Register(&CFGoReadImports{goFile: goFile})
	return sm, goFile
}

//CFGoReadPackage read pkg.
type CFGoReadPackage struct {
	goFile *GoFile
	first  bool
}

//Name reader's name.
func (rp *CFGoReadPackage) Name() string {
	return "CFGoReadPackage"
}

func ignore(lex *lex_pgl.LexProduct) bool {
	return (lex_pgl.IsSpace(lex) && lex.Value != "\n") || lex_pgl.IsComment(lex)
}

//PreRead only see if should stop read.
func (rp *CFGoReadPackage) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if ignore(lex) {
		return false, nil
	}

	if rp.first {
		if !lex.Equal(ConstPackage) {
			return true, onErr(rp, lex, "first ident want package.")
		}

		if rp.goFile.Package != "" {
			return true, onErr(rp, lex, "not first package.")
		}
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (rp *CFGoReadPackage) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if ignore(lex) {
		return false, nil
	}

	if rp.first {
		rp.first = false
		return false, nil
	}

	if !lex_pgl.IsIdent(lex) {
		return true, onErr(rp, lex, "package name want a ident")
	}

	rp.goFile.Package = lex.Value

	return true, nil
}

//End when end read.
func (rp *CFGoReadPackage) End(stateNode *smn_analysis.StateNode) (isEnd bool, err error) {
	if rp.first {
		return true, onErr(rp, nil, "unexcpt EOF")
	}

	return true, nil
}

//GetProduct return result.
func (rp *CFGoReadPackage) GetProduct() smn_analysis.ProductItf {
	return nil
}

//Clean let the Reader like new.  it will be call before first Read.
func (rp *CFGoReadPackage) Clean() {
	rp.first = true
}

const (
	mutiStatusUnknown = 0 // default
	mutiStatusSingle  = 1 // singel. import ""
	mutiStatusMuti    = 2 //muti. import("" "")
)

//CFGoReadImports read imports.
type CFGoReadImports struct {
	first      bool
	mutiStatus int // import("" "" )
	curPath    string
	curAlias   string
	goFile     *GoFile
}

//Name reader's name.
func (ri *CFGoReadImports) Name() string {
	return "CFGoReadImports"
}

//PreRead only see if should stop read.
func (ri *CFGoReadImports) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)

	if ignore(lex) {
		return false, nil
	}

	if ri.first && !lex.Equal(ConstImport) {
		return true, onErr(ri, lex, "not a import")
	}
	return false, nil
}

func (ri *CFGoReadImports) addImport() {
	ri.goFile.Imports = append(ri.goFile.Imports, &GoImport{Path: ri.curPath, Alias: ri.curAlias})
	ri.curPath = ""
	ri.curAlias = ""
}

//Read real read. even isEnd == true the input be readed.
func (ri *CFGoReadImports) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)

	if ignore(lex) {
		return false, nil
	}

	if ri.first { // import
		ri.first = false
		return false, nil
	}

	if ri.mutiStatus == mutiStatusUnknown {
		ri.mutiStatus = mutiStatusSingle

		if lex.Equal(ConstLeftParentheses) {
			ri.mutiStatus = mutiStatusMuti
		}

		return false, nil
	}

	if ri.mutiStatus == mutiStatusMuti && lex.Equal(ConstRightParentheses) {
		if ri.curPath != "" {
			ri.addImport()
		}

		return true, nil
	}

	if lex.Equal(ConstBreakLine) {
		if ri.mutiStatus == mutiStatusSingle {
			if ri.curPath == "" {
				return true, onErr(ri, lex, "expect package path[string]")
			}

			ri.addImport()
			return true, nil
		}
		//mutiStatusMuti
		if ri.curPath != "" {
			ri.addImport()
			return false, nil
		}
		//curPath == ""
		if ri.curAlias != "" {
			return true, onErr(ri, lex, "expect package path[string], found newline")
		}
	}

	if lex.Equal(ConstSemicolon) {
		if ri.curPath == "" {
			return true, onErr(ri, lex, "except package path[string]")
		}

		if ri.mutiStatus == mutiStatusMuti {
			ri.addImport()
		}

		return false, nil
	}

	if lex_pgl.IsIdent(lex) {
		if ri.curAlias != "" {
			return true, onErr(ri, lex, fmt.Sprintf("import Alias[%s] exist.", ri.curAlias))
		}

		ri.curAlias = lex.Value
		return false, nil
	}

	if lex_pgl.IsString(lex) {
		if ri.curPath != "" {
			return true, onErr(ri, lex, fmt.Sprintf("import Path[%s] exist.", ri.curPath))
		}

		ri.curPath = lex.Value
	}

	return false, nil
}

//End when end read.
func (ri *CFGoReadImports) End(stateNode *smn_analysis.StateNode) (isEnd bool, err error) {
	return true, onErr(ri, nil, "Unexcept EOF")
}

//GetProduct return result.
func (ri *CFGoReadImports) GetProduct() smn_analysis.ProductItf {
	return nil
}

//Clean let the Reader like new.  it will be call before first Read.
func (ri *CFGoReadImports) Clean() {
	ri.curPath = ""
	ri.curPath = ""
	ri.first = true
	ri.mutiStatus = mutiStatusUnknown
}
