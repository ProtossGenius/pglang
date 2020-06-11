package grm_go

import (
	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

//PGL for import package quickly.
func PGL() {}

//BlockType code-block's type.
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

//StringStack .
type StringStack struct {
	d    []string
	size int
}

//Push add to top.
func (s *StringStack) Push(str string) {
	if len(s.d) > s.size {
		s.d[s.size] = str
	} else {
		s.d = append(s.d, str)
	}
	s.size++
}

//Top .
func (s *StringStack) Top() string {
	if s.size != 0 {
		return s.d[s.size-1]
	}
	return ""
}

//Pop delete top.
func (s *StringStack) Pop() string {
	val := s.Top()
	if s.size != 0 {
		s.size--
	}
	return val
}

//Clean delete all data.
func (s *StringStack) Clean() {
	s.size = 0
}

//Size .
func (s *StringStack) Size() int {
	return s.size
}

//Empty is size = 0.
func (s *StringStack) Empty() bool {
	return s.size == 0
}

//BlockAnaProduct result.
type BlockAnaProduct struct {
	blockType BlockType
	BlockList []*lex_pgl.LexProduct // if not a
}

//ProductType type.
func (b *BlockAnaProduct) ProductType() int {
	return int(b.blockType)
}

//BlockReader reader.
type BlockReader struct {
	first  bool
	Result *BlockAnaProduct
	stack  StringStack
}

//Name reader's name.
func (b *BlockReader) Name() string {
	return "BlockReader"
}

//PreRead only see if should stop read.
func (b *BlockReader) PreRead(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if !b.first {
		return false, nil
	}
	result := b.Result
	b.first = false
	switch lex.Value {
	case "{":
		result.blockType = BlockTypeCurlyBraces
	case "(":
		result.blockType = BlockTypeParentheses
	case "[":
		result.blockType = BlockTypeSquareBrackets
	default:
		return true, onErr(b, lex, "except ( or [ or {")
	}
	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (b *BlockReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	lex := read(input)
	b.Result.BlockList = append(b.Result.BlockList, lex)

	switch lex.Value {
	case "{", "[", "(":
		b.stack.Push(lex.Value)
	}

	switch b.stack.Top() + lex.Value {
	case "{}", "[]", "()":
		b.stack.Pop()
	}

	if b.stack.Empty() {
		return true, nil
	}

	return false, nil
}

//GetProduct return result.
func (b *BlockReader) GetProduct() smn_analysis.ProductItf {
	return b.Result
}

//Clean let the Reader like new.  it will be call before first Read.
func (b *BlockReader) Clean() {
	b.Result = &BlockAnaProduct{}
	b.first = false
	b.stack.Clean()
}
