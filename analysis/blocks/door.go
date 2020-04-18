package blocks

import "github.com/ProtossGenius/pglang/analysis/pgl_ana_lex"

func SM() {} // for more easy to add import.

type BlockAnaInput struct {
	LexUnit *pgl_ana_lex.LexProduct
}

type BlockType int

const (
	BlockType_None           BlockType = iota
	BlockType_Parentheses              // ()
	BlockType_SquareBrackets           // []
	BlockType_CurlyBraces              // {}
)

type BlockAnaProduct struct {
	blockType BlockType
	BlockList []*pgl_ana_lex.LexProduct // if not a
}

func (b *BlockAnaProduct) ProductType() int {
	return int(b.blockType)
}

type NoneBlockReader struct {
}
