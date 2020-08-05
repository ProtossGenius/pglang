package classifygo

import "github.com/ProtossGenius/pglang/analysis/lex_pgl"

func constIdent(val string) *lex_pgl.LexProduct {
	return &lex_pgl.LexProduct{Type: lex_pgl.PGLA_PRODUCT_IDENT, Value: val}
}

var ( //Ident
	ConstPackage   = constIdent("package")
	ConstImport    = constIdent("import")
	ConstVar       = constIdent("var")
	ConstConst     = constIdent("const")
	ConstFuncs     = constIdent("func")
	ConstType      = constIdent("type")
	ConstStruct    = constIdent("struct")
	ConstInterface = constIdent("interface")
)

func constSymbol(val string) *lex_pgl.LexProduct {
	return &lex_pgl.LexProduct{Type: lex_pgl.PGLA_PRODUCT_SYMBOL, Value: val}
}

var ( //symbol
	ConstLeftParentheses  = constSymbol("(")
	ConstRightParentheses = constSymbol(")")
	ConstSemicolon        = constSymbol(";")
	ConstLeftCurlyBraces  = constSymbol("{")
	ConstRightCurlyBraces = constSymbol("}")
)

func constSpace(val string) *lex_pgl.LexProduct {
	return &lex_pgl.LexProduct{Type: lex_pgl.PGLA_PRODUCT_SPACE, Value: val}
}

var ( //space
	ConstBreakLine = constSpace("\n")
)

const ( //Err
	//ErrUnexceptEOF .
	ErrUnexceptEOF = "Unexcept EOF"
)
