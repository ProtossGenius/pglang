package blocks

import "github.com/ProtossGenius/pglang/analysis/pgl_ana_lex"

//SM for import package quickly.
func SM() {}

//BlockAnaInput intput data (from LexProduct).
type BlockAnaInput struct {
	LexUnit *pgl_ana_lex.LexProduct
}

//BlockType code-block's type
type BlockType int

const (
	//BlockTypeNone not a code block
	BlockTypeNone BlockType = iota
	//BlockTypeParentheses ().
	BlockTypeParentheses
	//BlockTypeSquareBrackets [].
	BlockTypeSquareBrackets
	//BlockTypeCurlyBraces {}.
	BlockTypeCurlyBraces
)

//BlockAnaProduct result.
type BlockAnaProduct struct {
	blockType BlockType
	BlockList []*BlockAnaProduct // if not a
}

//ProductType type.
func (b *BlockAnaProduct) ProductType() int {
	return int(b.blockType)
}

//BlockReader reader.
type BlockReader struct {
}
