package classifygo

import (
	"fmt"
	"strings"

	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
	"github.com/ProtossGenius/pglang/snreader"
)

/*
* finished pakcage, import, globals(const, var), funcs type.
 */

type statistics struct {
	path string
	line int
	col  int
}

// Position print current position.
func (s *statistics) Position() string {
	return fmt.Sprintf("%s|%d|%d|", s.path, s.line+1, s.col)
}

func strLen(str string) int {
	return len([]rune(str))
}

// Read read input and calc position.
func (s *statistics) Read(input snreader.InputItf) {
	lex := read(input)
	if strings.Contains(lex.Value, "\n") {
		spls := strings.Split(lex.Value, "\n")
		s.col = strLen(spls[len(spls)-1])
		s.line += len(spls) - 1
	} else {
		s.col += strLen(lex.Value)
	}
}

func (s *statistics) Clean() {
	s.path = ""
	s.line = 0
	s.col = 0
}

//NewAnalysiser new analysiser.
func NewAnalysiser(path string) (*snreader.StateMachine, *GoFile) {
	goFile := &GoFile{}
	sm := new(snreader.StateMachine).Init()
	sm.Statistic = &statistics{path: path}
	dft := snreader.NewDftStateNodeReader(sm)
	dft.Register(NewCFReadIgnore(func(lex *lex_pgl.LexProduct) bool {
		return lex_pgl.IsSpace(lex) || lex_pgl.IsComment(lex)
	}, IgnoreTypeWithError))
	dft.Register(&CFGoReadPackage{goFile: goFile})
	dft.Register(&CFGoReadImports{goFile: goFile})
	dft.Register(&CFGoReadGlobals{goFile: goFile})
	dft.Register(NewCFGoReadFuncs(goFile))
	dft.Register(NewCFgoReadTypes(goFile))
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

// IgnoreType .
type IgnoreType int

const (
	// IgnoreTypeWithError when finish return error, and only read one lex.
	IgnoreTypeWithError IgnoreType = iota
	// IgnoreTypeNoError when finish not with error.
	IgnoreTypeNoError
)

// NewCFReadIgnore create ignore reader.
func NewCFReadIgnore(ignoreWhat func(lex *lex_pgl.LexProduct) bool, ignoreType IgnoreType) snreader.StateNodeReader {
	return &CFGoReadIgnore{ignoreWhat: ignoreWhat, ignoreType: ignoreType}
}

//CFGoReadIgnore ignore the spase between types.
type CFGoReadIgnore struct {
	ignoreWhat func(lex *lex_pgl.LexProduct) bool
	ignoreType IgnoreType
}

//Name reader's name.
func (rign *CFGoReadIgnore) Name() string {
	return "CFGoReadIgnore"
}

//Clean let the Reader like new.  it will be call before first Read.
func (rign *CFGoReadIgnore) Clean() {
}

//PreRead only see if should stop read.
func (rign *CFGoReadIgnore) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)
	ignore := false
	if rign.ignoreWhat != nil && rign.ignoreWhat(lex) {
		ignore = true
	}

	if ignore {
		return false, nil
	}
	// if cant ignore.
	if rign.ignoreType == IgnoreTypeWithError {
		return true, onErr(rign, lex, "cant be ignore.")
	}

	return true, nil
}

//Read real read. even isEnd == true the input be readed.
func (rign *CFGoReadIgnore) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	if rign.ignoreType == IgnoreTypeWithError {
		return true, nil
	}

	return false, nil
}

//End when end read.
func (rign *CFGoReadIgnore) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	return true, nil
}

//GetProduct return result.
func (rign *CFGoReadIgnore) GetProduct() snreader.ProductItf {
	return nil
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

//Read real read. even isEnd == true the input be readed.
func (rg *CFGoReadGlobals) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)

	if ignore(lex) {
		return false, nil
	}

	if rg.first {
		rg.first, rg.varType = false, lex.Value
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

//End when end read.
func (rg *CFGoReadGlobals) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	return true, onErr(rg, nil, ErrUnexceptEOF)
}

//GetProduct return result.
func (rg *CFGoReadGlobals) GetProduct() snreader.ProductItf {

	return nil
}

//CFGoReadLexUnit read a ident and save it to stateNode's datas.
type CFGoReadLexUnit struct {
	preCheck func(*lex_pgl.LexProduct) bool
	readDo   func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error
	name     string
}

// SetName .
func (ridt *CFGoReadLexUnit) SetName(name string) *CFGoReadLexUnit {
	ridt.name = name

	return ridt
}

//Name reader's name.
func (ridt *CFGoReadLexUnit) Name() string {
	return ridt.name
}

//Clean let the Reader like new.  it will be call before first Read.
func (ridt *CFGoReadLexUnit) Clean() {
}

//PreRead only see if should stop read.
func (ridt *CFGoReadLexUnit) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if nil != ridt.preCheck {
		if ridt.preCheck(lex) {
			return false, nil
		}

		return true, onErr(ridt, lex, "preCheck fail")
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (ridt *CFGoReadLexUnit) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)
	return true, ridt.readDo(ridt, stateNode, lex)
}

//End when end read.
func (ridt *CFGoReadLexUnit) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	return true, onErr(ridt, nil, ErrUnexceptEOF)
}

//GetProduct return result.
func (ridt *CFGoReadLexUnit) GetProduct() snreader.ProductItf {
	return nil
}

// NewIdentReader .
func NewIdentReader(preCheck func(lex *lex_pgl.LexProduct) bool,
	readDo func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error) *CFGoReadLexUnit {
	res := &CFGoReadLexUnit{readDo: readDo}
	if preCheck != nil {
		res.preCheck = func(lex *lex_pgl.LexProduct) bool {
			if !lex_pgl.IsIdent(lex) {
				return false
			}

			return preCheck(lex)
		}
	} else {
		res.preCheck = lex_pgl.IsIdent
	}

	return res.SetName("IdentReader")

}

// NewIdentSaver read a ident and save to SN.Datas[key].
func NewIdentSaver(key string) *CFGoReadLexUnit {
	return NewIdentReader(nil, func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error {
		stateNode.Datas[key] = lex.Value
		return nil
	}).SetName("IdentSaver")
}

// NewLexChecker read a ident and check is it equ chk.
func NewLexChecker(chk *lex_pgl.LexProduct) *CFGoReadLexUnit {
	return &CFGoReadLexUnit{readDo: func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error {
		if !lex.Equal(chk) {
			return onErr(reader, lex, "want a <ident>"+chk.Value)
		}

		return nil
	}, name: "LexChecker"}
}

// NewLexPreChecker pre check, not eat the lex.
func NewLexPreChecker(chk *lex_pgl.LexProduct) *CFGoReadLexUnit {
	return &CFGoReadLexUnit{preCheck: chk.Equal}
}

// NewLexExcluder if in excs return error.
func NewLexExcluder(excs ...*lex_pgl.LexProduct) *CFGoReadLexUnit {
	return &CFGoReadLexUnit{
		readDo: func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error {
			for _, exc := range excs {
				if lex.Equal(exc) {
					return onErr(reader, lex, "dont want this Lex")
				}
			}
			return nil
		}}
}

//NewCFGoReadFuncDef .
func NewCFGoReadFuncDef(end *lex_pgl.LexProduct, finishDo func(node *snreader.StateNode)) snreader.StateNodeReader {

	return snreader.NewStateNodeListReader(
		//read funcName.
		NewIdentSaver("funcName"),
		NewCFReadIgnore(ignore, IgnoreTypeNoError),
		//read params.
		NewBlockReader(ConstLeftParentheses, ConstRightParentheses, false, true, "params", nil).SetName("FuncDefParamsReader"),
		//read returns.
		NewEndFlagReader(end, "returns", false, finishDo).SetName("FuncDefReturnsReader"),
	)
}

//NewCFGoReadFuncs read funcs.
func NewCFGoReadFuncs(goFile *GoFile) snreader.StateNodeReader {
	return snreader.NewStateNodeListReader(
		//read func.
		NewLexChecker(ConstFuncs),
		//space & comment
		NewCFReadIgnore(ignore, IgnoreTypeNoError),
		//read scope
		NewBlockReader(ConstLeftParentheses, ConstRightParentheses, true, true, "scope", nil).SetName("FuncScopeReader"),
		//space & comment
		NewCFReadIgnore(ignore, IgnoreTypeNoError),
		//read func def
		NewCFGoReadFuncDef(ConstLeftCurlyBraces, func(node *snreader.StateNode) {
			datas := node.Datas

			fDef := &GoFuncDef{
				Name:    datas["funcName"].(string),
				Params:  datas["params"].(GoCodes),
				Returns: datas["returns"].(GoCodes),
			}
			datas["funcDef"] = fDef
		}),
		//read codes
		NewBlockReader(ConstLeftCurlyBraces, ConstRightCurlyBraces, false, true, "code", func(stateNode *snreader.StateNode) {
			datas := stateNode.Datas
			var scope GoCodes
			if val, exist := datas["scope"]; exist {
				scope = val.(GoCodes)
			}
			goFile.Funcs = append(goFile.Funcs, &GoFunc{
				Scope:   scope,
				FuncDef: datas["funcDef"].(*GoFuncDef),
				Codes:   datas["code"].(GoCodes),
			})
		}).SetName("FuncCodeReader"),
	).SetName("CFGoFuncReader")
}

// NewCFgoReadTypes read type XX .
func NewCFgoReadTypes(goFile *GoFile) snreader.StateNodeReader {
	return snreader.NewStateNodeListReader(
		//read type.
		NewLexChecker(ConstType),
		// space & coment
		&CFGoReadIgnore{ignore, IgnoreTypeNoError},
		snreader.NewStateNodeSelectReader(
			// type XXXX struct
			snreader.NewStateNodeListReader(
				//read name
				NewIdentSaver("typeName"),
				&CFGoReadIgnore{ignore, IgnoreTypeNoError},
				NewLexChecker(ConstStruct),
				// space & comment.
				&CFGoReadIgnore{ignore, IgnoreTypeNoError},
				NewBlockReader(ConstLeftCurlyBraces, ConstRightCurlyBraces, false, true, "code", func(stateNode *snreader.StateNode) {
					datas := stateNode.Datas
					goFile.Structs = append(goFile.Structs, &GoStruct{
						Name:  datas["typeName"].(string),
						Codes: datas["code"].(GoCodes),
					})
				}),
			),
			// type XXXX interface
			snreader.NewStateNodeListReader(
				//read name
				NewIdentSaver("typeName"),
				&CFGoReadIgnore{ignore, IgnoreTypeNoError},
				NewLexChecker(ConstInterface),
				&CFGoReadIgnore{ignore, IgnoreTypeNoError},
				NewBlockReader(ConstLeftCurlyBraces, ConstRightCurlyBraces, false, true, "code", func(stateNode *snreader.StateNode) {
					datas := stateNode.Datas
					goFile.Interfaces = append(goFile.Interfaces, &GoItf{Name: datas["typeName"].(string), Codes: datas["code"].(GoCodes)})
				}),
			),

			// type XXXX func
			snreader.NewStateNodeListReader(
				NewCFGoReadFuncDef(ConstBreakLine, func(stateNode *snreader.StateNode) {
					datas := stateNode.Datas
					goFile.TypeFunc = append(goFile.TypeFunc, &GoTypeFunc{Name: datas["funcName"].(string), Params: datas["params"].(GoCodes), Returns: datas["returns"].(GoCodes)})
				}),
			),
			//another
			snreader.NewStateNodeListReader(
				//read name
				NewIdentSaver("typeName"),
				&CFGoReadIgnore{ignore, IgnoreTypeNoError},
				NewLexExcluder(ConstStruct, ConstInterface),
				NewBlockReader(nil, ConstBreakLine, false, true, "code", func(stateNode *snreader.StateNode) {
					datas := stateNode.Datas
					goFile.Aliases = append(goFile.Aliases, &GoAlias{Name: datas["typeName"].(string), Codes: datas["code"].(GoCodes)})
				}),
			),
		),
	).SetName("CFGoTypeReader")
}
