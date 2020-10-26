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
	dft := snreader.NewDftStateNodeReader(sm).SetName("ClassifyGoAnalysiser")
	dft.Register(NewCFReadIgnore(ignoreAll, IgnoreTypeWithError))
	dft.Register(&CFGoReadPackage{goFile: goFile})
	dft.Register(NewCFGoImports(goFile))
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

func ignoreWithoutBreakline(lex *lex_pgl.LexProduct) bool {
	return (lex_pgl.IsSpace(lex) && lex.Value != "\n") || lex_pgl.IsComment(lex)
}

func ignoreAll(lex *lex_pgl.LexProduct) bool {
	return lex_pgl.IsSpace(lex) || lex_pgl.IsComment(lex)
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

	if ignoreWithoutBreakline(lex) {
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

	if ignoreWithoutBreakline(lex) {
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
	preCheck func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) (isEnd bool, err error)
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
		return ridt.preCheck(ridt, stateNode, lex)
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (ridt *CFGoReadLexUnit) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	if ridt.readDo == nil {
		return true, nil
	}

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
func NewIdentReader(preCheck func(lex *lex_pgl.LexProduct) (isEnd bool, err error),
	readDo func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error) *CFGoReadLexUnit {
	res := &CFGoReadLexUnit{readDo: readDo}
	if preCheck != nil {
		res.preCheck = func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) (bool, error) {
			if !lex_pgl.IsIdent(lex) {
				return true, onErr(res, lex, "except a Ident")
			}

			return preCheck(lex)
		}
	} else {
		res.preCheck = func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) (bool, error) {
			if !lex_pgl.IsIdent(lex) {
				return true, onErr(res, lex, "except a Ident")
			}

			return false, nil
		}
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

// NewLexTryRead try read a lex, if not eqa will not read, if eqa will eat it.
func NewLexTryRead(chk *lex_pgl.LexProduct) *CFGoReadLexUnit {
	res := &CFGoReadLexUnit{}
	res.preCheck = func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) (isEnd bool, err error) {
		if chk.Equal(lex) { // if eqa then goto read and eat it.
			return false, nil
		}

		return true, nil
	}

	return res
}

// NewLexPreCheck if input in set then return success, or return error.
func NewLexPreCheck(chk ...*lex_pgl.LexProduct) *CFGoReadLexUnit {
	cMap := make(map[string]*lex_pgl.LexProduct, len(chk))

	for _, it := range chk {
		cMap[it.Value] = it
	}

	return &CFGoReadLexUnit{
		preCheck: func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) (isEnd bool, err error) {
			if val, exist := cMap[lex.Value]; exist {
				if val.Equal(lex) {
					return true, nil
				}
			}

			return true, onErr(reader, lex, "not except Input")
		},
	}
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
		NewCFReadIgnore(ignoreWithoutBreakline, IgnoreTypeNoError),
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
		NewCFReadIgnore(ignoreWithoutBreakline, IgnoreTypeNoError),
		//read scope
		NewBlockReader(ConstLeftParentheses, ConstRightParentheses, true, true, "scope", nil).SetName("FuncScopeReader"),
		//space & comment
		NewCFReadIgnore(ignoreWithoutBreakline, IgnoreTypeNoError),
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
		&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
		snreader.NewStateNodeSelectReader(
			// type XXXX struct
			snreader.NewStateNodeListReader(
				//read name
				NewIdentSaver("typeName"),
				&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
				NewLexChecker(ConstStruct),
				// space & comment.
				&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
				NewBlockReader(ConstLeftCurlyBraces, ConstRightCurlyBraces, false, true, "code", func(stateNode *snreader.StateNode) {
					datas := stateNode.Datas
					goFile.Structs = append(goFile.Structs, &GoStruct{
						Name:  datas["typeName"].(string),
						Codes: datas["code"].(GoCodes),
					})
				}),
			).SetName("CFGoTypeStructReader"),
			// type XXXX interface
			snreader.NewStateNodeListReader(
				//read name
				NewIdentSaver("typeName"),
				&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
				NewLexChecker(ConstInterface),
				&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
				NewBlockReader(ConstLeftCurlyBraces, ConstRightCurlyBraces, false, true, "code", func(stateNode *snreader.StateNode) {
					datas := stateNode.Datas
					goFile.Interfaces = append(goFile.Interfaces, &GoItf{Name: datas["typeName"].(string), Codes: datas["code"].(GoCodes)})
				}),
			).SetName("CFGoTypeInterfaceReader"),

			// type XXXX func
			snreader.NewStateNodeListReader(
				&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
				NewLexChecker(ConstFuncs),
				NewCFGoReadFuncDef(ConstBreakLine, func(stateNode *snreader.StateNode) {
					datas := stateNode.Datas
					goFile.TypeFunc = append(goFile.TypeFunc, &GoTypeFunc{Name: datas["funcName"].(string), Params: datas["params"].(GoCodes), Returns: datas["returns"].(GoCodes)})
				}),
			).SetName("CFGoTypeFuncReader"),
			//another
			snreader.NewStateNodeListReader(
				//read name
				NewIdentSaver("typeName"),
				&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
				NewLexExcluder(ConstStruct, ConstInterface, ConstFuncs),
				NewBlockReader(nil, ConstBreakLine, false, true, "code", func(stateNode *snreader.StateNode) {
					datas := stateNode.Datas
					goFile.Aliases = append(goFile.Aliases, &GoAlias{Name: datas["typeName"].(string), Codes: datas["code"].(GoCodes)})
				}),
			),
		),
	).SetName("CFGoTypeReader")
}

// import alias<ident> path<string>
func newGoOneImport(goFile *GoFile) *snreader.StateNodeListReader {
	return snreader.NewStateNodeListReader(
		// read alias name.
		&CFGoReadLexUnit{
			preCheck: func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) (isEnd bool, err error) {
				// if is ident then read it
				return !lex_pgl.IsIdent(lex), nil
			},
			readDo: func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error {
				stateNode.Datas["alias"] = lex.Value
				return nil
			},
		},
		&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
		&CFGoReadLexUnit{
			preCheck: func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) (isEnd bool, err error) {
				if !lex_pgl.IsString(lex) {
					return true, onErr(reader, lex, "except a string for import path")
				}
				return false, nil
			},
			readDo: func(reader snreader.StateNodeReader, stateNode *snreader.StateNode, lex *lex_pgl.LexProduct) error {
				vAlias := stateNode.Datas["alias"]
				alias := ""
				if vAlias != nil {
					alias = vAlias.(string)
				}
				stateNode.Datas["alias"] = ""
				goFile.Imports = append(goFile.Imports, &GoImport{Path: lex.Value, Alias: alias})
				return nil
			},
		},
		&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
		NewLexTryRead(ConstSemicolon),
		&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
		NewLexTryRead(ConstBreakLine),
	)
}

// NewCFGoImports .
func NewCFGoImports(goFile *GoFile) snreader.StateNodeReader {
	return snreader.NewStateNodeListReader(
		//read import
		NewLexChecker(ConstImport).SetName("ImportReader-ReadImport"),
		//read ignore
		&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
		//select one or muti-import
		snreader.NewStateNodeSelectReader(
			// import xxx "xxx"
			newGoOneImport(goFile).SetName("OneImportReader"),
			// import ( ... )
			NewStateNodeLoopReader(snreader.NewStateNodeListReader(
				//read ignroe (all-type sapce and comment)
				&CFGoReadIgnore{ignoreAll, IgnoreTypeNoError},
				newGoOneImport(goFile),
			).SetName("MutiImportLoopUnit"), ConstLeftParentheses, ConstRightParentheses, true, nil).SetName("MutiImportReader"),
		),
	).SetName("CFGoImportReader")
}

// NewCFGoReadGoType .
func NewCFGoReadGoType() {
}

// NewCFGoReadGlobals .
func NewCFGoReadGlobals(goFile *GoFile) snreader.StateNodeReader {
	return snreader.NewStateNodeListReader(
		// must be const or var
		NewLexPreCheck(ConstVar, ConstType),
		NewIdentSaver("type"),
		&CFGoReadIgnore{ignoreWithoutBreakline, IgnoreTypeNoError},
		snreader.NewStateNodeSelectReader(),
	)
}
