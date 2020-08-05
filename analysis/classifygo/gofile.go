package classifygo

import (
	"fmt"

	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
	"github.com/ProtossGenius/pglang/snreader"
)

/*
* finished pakcage, import, globals(const, var)
 */

//NewAnalysiser new analysiser.
func NewAnalysiser() (*snreader.StateMachine, *GoFile) {
	goFile := &GoFile{}
	sm := new(snreader.StateMachine).Init()
	dft := snreader.NewDftStateNodeReader(sm)
	dft.Register(&CFGoReadPackage{goFile: goFile})
	dft.Register(&CFGoReadImports{goFile: goFile})
	dft.Register(&CFGoReadGlobals{goFile: goFile})
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
func (rp *CFGoReadPackage) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)

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
func (rp *CFGoReadPackage) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
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
func (rp *CFGoReadPackage) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	if rp.first {
		return true, onErr(rp, nil, "unexcpt EOF")
	}

	return true, nil
}

//GetProduct return result.
func (rp *CFGoReadPackage) GetProduct() snreader.ProductItf {
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
func (ri *CFGoReadImports) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)

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
func (ri *CFGoReadImports) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)

	if ignore(lex) {
		return false, nil
	}

	if ri.first { // import
		ri.first = false
		return false, nil
	}

	if ri.mutiStatus == mutiStatusUnknown {
		if lex.Equal(ConstLeftParentheses) {
			ri.mutiStatus = mutiStatusMuti
			return false, nil
		}

		ri.mutiStatus = mutiStatusSingle
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

		return false, nil
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
func (ri *CFGoReadImports) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	return true, onErr(ri, nil, "Unexcept EOF")
}

//GetProduct return result.
func (ri *CFGoReadImports) GetProduct() snreader.ProductItf {
	return nil
}

//Clean let the Reader like new.  it will be call before first Read.
func (ri *CFGoReadImports) Clean() {
	ri.curPath = ""
	ri.curPath = ""
	ri.first = true
	ri.mutiStatus = mutiStatusUnknown
}

//CFGoReadIgnore ignore the spase between types.
type CFGoReadIgnore struct {
}

//Name reader's name.
func (rign *CFGoReadIgnore) Name() string {
	return "CFGoReadIgnore"
}

//PreRead only see if should stop read.
func (rign *CFGoReadIgnore) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (rign *CFGoReadIgnore) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if lex_pgl.IsSpace(lex) || lex_pgl.IsComment(lex) {
		return true, nil
	}

	return false, onErr(rign, lex, "can't ignore")
}

//End when end read.
func (rign *CFGoReadIgnore) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	return true, nil
}

//GetProduct return result.
func (rign *CFGoReadIgnore) GetProduct() snreader.ProductItf {
	return nil
}

//Clean let the Reader like new.  it will be call before first Read.
func (rign *CFGoReadIgnore) Clean() {
}

//CFGoReadGlobals get consts and vars.
type CFGoReadGlobals struct {
	goFile         *GoFile
	first          bool
	goBatchGlobals GoBatchGlobals
	varType        string
	mutiStatus     int
	gName          string
	gCode          GoCodes
	lleftPNum      int // lonly left parentheses nums;
	lleftCNum      int //lonly left curly braces nums;
}

//Clean let the Reader like new.  it will be call before first Read.
func (rg *CFGoReadGlobals) Clean() {
	rg.first = true
	rg.goBatchGlobals = nil
	rg.mutiStatus = mutiStatusUnknown
	rg.lleftPNum = 0
	rg.varType = ""
	rg.gName = ""
	rg.gCode = nil
	rg.lleftCNum = 0
}

//Name reader's name.
func (rg *CFGoReadGlobals) Name() string {
	return "CFGoReadGlobals"
}

//PreRead only see if should stop read.
func (rg *CFGoReadGlobals) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)

	if rg.first && !lex.Equal(ConstConst) && !lex.Equal(ConstVar) {
		return true, onErr(rg, lex, "not a global value")
	}

	return false, nil
}

func (rg *CFGoReadGlobals) checkFinish() string {
	if rg.varType == "" {
		return "need const or var"
	}

	if rg.mutiStatus == mutiStatusSingle && rg.lleftPNum < 0 {
		return "too much '('"
	}

	if rg.gName == "" {
		return "need var name"
	}

	return ""
}

func (rg *CFGoReadGlobals) addGlobal(lex *lex_pgl.LexProduct) error {
	if cf := rg.checkFinish(); cf != "" {
		return onErr(rg, lex, cf)
	}

	rg.goBatchGlobals = append(rg.goBatchGlobals, &GoGlobals{Name: rg.gName, Code: rg.gCode})
	rg.gName = ""
	rg.gCode = nil
	return nil
}

func (rg *CFGoReadGlobals) countLonlyBrackets(lex *lex_pgl.LexProduct) {
	if lex.Equal(ConstLeftParentheses) {
		rg.lleftPNum++
	}

	if lex.Equal(ConstRightParentheses) {
		rg.lleftPNum--
	}

	if lex.Equal(ConstLeftCurlyBraces) {
		rg.lleftCNum++
	}

	if lex.Equal(ConstRightCurlyBraces) {
		rg.lleftCNum--
	}
}

//Read real read. even isEnd == true the input be readed.
func (rg *CFGoReadGlobals) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)

	if ignore(lex) {
		return false, nil
	}

	if rg.first {
		rg.first, rg.varType = true, lex.Value
		return false, nil
	}

	if rg.mutiStatus == mutiStatusUnknown {
		if lex.Equal(ConstLeftParentheses) {
			rg.mutiStatus = mutiStatusMuti
			return false, nil
		}

		rg.mutiStatus = mutiStatusSingle
	}

	rg.countLonlyBrackets(lex)

	if rg.mutiStatus == mutiStatusMuti && rg.lleftPNum == -1 {
		if rg.gName == "" {
			return true, nil
		}

		return true, rg.addGlobal(lex)
	}

	if lex.Equal(ConstBreakLine) {
		if rg.lleftCNum != 0 {
			return false, nil
		}

		if rg.gName == "" && rg.mutiStatus == mutiStatusMuti {
			return false, nil
		}

		return rg.mutiStatus == mutiStatusSingle, rg.addGlobal(lex)
	}

	if rg.gName == "" {
		if !lex_pgl.IsIdent(lex) {
			return true, onErr(rg, lex, "var name want a ident")
		}

		rg.gName = lex.Value
		return false, nil
	}

	rg.gCode = append(rg.gCode, lex)
	return false, nil
}

//End when end read.
func (rg *CFGoReadGlobals) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	return true, onErr(rg, nil, ErrUnexceptEOF)
}

//GetProduct return result.
func (rg *CFGoReadGlobals) GetProduct() snreader.ProductItf {

	return nil
}

//CFGoReadIdent read a ident and save it to stateNode's datas.
type CFGoReadIdent struct {
	readDo func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error
}

//Name reader's name.
func (ridt *CFGoReadIdent) Name() string {
	return "CFGoReadIdent"
}

//Clean let the Reader like new.  it will be call before first Read.
func (ridt *CFGoReadIdent) Clean() {
}

//PreRead only see if should stop read.
func (ridt *CFGoReadIdent) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)

	if !lex_pgl.IsIdent(lex) {
		return true, onErr(ridt, lex, "want a ident.")
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (ridt *CFGoReadIdent) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)
	return true, ridt.readDo(ridt, stateNode, lex)
}

//End when end read.
func (ridt *CFGoReadIdent) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	return true, onErr(ridt, nil, ErrUnexceptEOF)
}

//GetProduct return result.
func (ridt *CFGoReadIdent) GetProduct() snreader.ProductItf {
	return nil
}

//NewCFGoReadFuncDef .
func NewCFGoReadFuncDef(end *lex_pgl.LexProduct, finishDo func(node *snreader.StateNode)) snreader.StateNodeReader {
	return snreader.NewStateNodeListReader(
		//read funcName.
		&CFGoReadIdent{readDo: func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error {
			stateNode.Datas["funcName"] = lex.Value
			return nil
		}},
		//read params.
		NewBlockReader(ConstLeftParentheses, ConstRightParentheses, false, "params", nil),
		//read returns.A
		//TODO it may end with \n or {
		NewStateNodeLoopReader(nil, nil, end, false, func(node *snreader.StateNode) {
			//save to stateNode's datas.
		}),
	)
}

//NewCFGoReadFuncs read funcs.
func NewCFGoReadFuncs(goFile *GoFile) snreader.StateNodeReader {
	return snreader.NewStateNodeListReader(
		//read func.
		&CFGoReadIdent{readDo: func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error {
			if !lex.Equal(ConstFuncs) {
				return onErr(reader, lex, "want a <ident>func")
			}
			return nil
		}},
		//read scope
		NewBlockReader(ConstLeftParentheses, ConstRightParentheses, true, "scope", nil),
		//read codes
		NewBlockReader(ConstLeftCurlyBraces, ConstRightCurlyBraces, false, "code", func(stateNode *snreader.StateNode) {
			datas := stateNode.Datas
			goFile.Funcs = append(goFile.Funcs, &GoFunc{
				Scope:   datas["scope"].(GoCodes),
				FuncDef: nil,
				Codes:   datas["code"].(GoCodes),
			})
		}),
	)
}
