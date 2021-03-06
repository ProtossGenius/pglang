//the file product by build.go  ProtossGenius whose email is guyvejianglou@outlook.com
//you should never change this file.
package lex_pgl

var SymbolList = map[string]bool{"+": true, "-": true, "*": true, "/": true, "%": true, "!": true, "^": true, "|": true, "||": true, "&": true, "&&": true, "?": true, "~": true, "(": true, ")": true, "=": true, "[": true, "]": true, "{": true, "}": true, "==": true, "+=": true, "-=": true, "*=": true, "/=": true, "!=": true, "|=": true, "&=": true, "^=": true, "%=": true, "~=": true, ",": true, ".": true, ";": true, "$": true, "@": true, "#": true, "\\": true, ":": true, ">": true, "<": true, ">=": true, "<=": true, ">>": true, "<<": true}

var SymbolCharSet = map[rune]bool{'!': true, '#': true, '$': true, '%': true, '&': true, '(': true, ')': true, '*': true, '+': true, ',': true, '-': true, '.': true, '/': true, ':': true, ';': true, '<': true, '=': true, '>': true, '?': true, '@': true, '[': true, '\\': true, ']': true, '^': true, '{': true, '|': true, '}': true, '~': true}

var SymbolCanContinue = map[string]bool{"|": true, "&": true, "=": true, "+": true, "-": true, "*": true, "/": true, "!": true, "^": true, "%": true, "~": true, ">": true, "<": true}

//some maybe define in another type, but not as symbol. like comment's "//" and "/*"
var SymbolUnuse = map[string]bool{"//": true, "/*": true}

//number charSet
var NumberCharSet = map[rune]bool{'1': true, '2': true, '3': true, '4': true, '5': true, '6': true, '7': true, '8': true, '9': true, '0': true, '.': true, 'x': true, 'X': true, 'a': true, 'A': true, 'b': true, 'B': true, 'c': true, 'C': true, 'd': true, 'D': true, 'e': true, 'E': true, 'f': true, 'F': true, 'l': true, 'L': true, 'u': true, 'U': true}

type PglaProduct int

const (
	PGLA_PRODUCT_ PglaProduct = iota
	PGLA_PRODUCT_IDENT
	PGLA_PRODUCT_SPACE
	PGLA_PRODUCT_SYMBOL
	PGLA_PRODUCT_COMMENT
	PGLA_PRODUCT_INTERGER
	PGLA_PRODUCT_DECIMAL
	PGLA_PRODUCT_STRING
	PGLA_PRODUCT_HAN
)

var PglaNameMap = map[PglaProduct]string{
	-1:                    "EMD",
	PGLA_PRODUCT_IDENT:    "PGLA_PRODUCT_IDENT",
	PGLA_PRODUCT_SPACE:    "PGLA_PRODUCT_SPACE",
	PGLA_PRODUCT_SYMBOL:   "PGLA_PRODUCT_SYMBOL",
	PGLA_PRODUCT_COMMENT:  "PGLA_PRODUCT_COMMENT",
	PGLA_PRODUCT_INTERGER: "PGLA_PRODUCT_INTERGER",
	PGLA_PRODUCT_DECIMAL:  "PGLA_PRODUCT_DECIMAL",
	PGLA_PRODUCT_STRING:   "PGLA_PRODUCT_STRING",
	PGLA_PRODUCT_HAN:      "PGLA_PRODUCT_HAN",
}

//IsIdent chack if lex is Ident.
func IsIdent(lex *LexProduct) bool {
	return lex.ProductType() == int(PGLA_PRODUCT_IDENT)
}

//IsSpace chack if lex is Space.
func IsSpace(lex *LexProduct) bool {
	return lex.ProductType() == int(PGLA_PRODUCT_SPACE)
}

//IsSymbol chack if lex is Symbol.
func IsSymbol(lex *LexProduct) bool {
	return lex.ProductType() == int(PGLA_PRODUCT_SYMBOL)
}

//IsComment chack if lex is Comment.
func IsComment(lex *LexProduct) bool {
	return lex.ProductType() == int(PGLA_PRODUCT_COMMENT)
}

//IsInterger chack if lex is Interger.
func IsInterger(lex *LexProduct) bool {
	return lex.ProductType() == int(PGLA_PRODUCT_INTERGER)
}

//IsDecimal chack if lex is Decimal.
func IsDecimal(lex *LexProduct) bool {
	return lex.ProductType() == int(PGLA_PRODUCT_DECIMAL)
}

//IsString chack if lex is String.
func IsString(lex *LexProduct) bool {
	return lex.ProductType() == int(PGLA_PRODUCT_STRING)
}

//IsHan chack if lex is Han.
func IsHan(lex *LexProduct) bool {
	return lex.ProductType() == int(PGLA_PRODUCT_HAN)
}
