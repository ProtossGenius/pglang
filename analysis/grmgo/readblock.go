package grmgo

import (
	"fmt"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
)

//BlockPair block pair.
type BlockPair struct {
	Start *lex_pgl.LexProduct
	End   *lex_pgl.LexProduct
}

func read(input smn_analysis.InputItf) *lex_pgl.LexProduct {
	return input.(*lex_pgl.LexProduct)
}

func onErr(reader smn_analysis.StateNodeReader, lex *lex_pgl.LexProduct, reason string) error {
	return fmt.Errorf("Error in [%s], input is [%v] reason is: %s", reader.Name(), lex, reason)
}

//BlockReader block reader [{( .. )}].
type BlockReader struct {
	MBlockPair *BlockPair
	first      bool
	CanIgnore  bool
	index      int
	codes      GoCodes
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

	if !lex.Equal(b.MBlockPair.Start) {
		if b.CanIgnore {
			return true, nil
		}

		return true, onErr(b, lex, "")
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (b *BlockReader) Read(stateNode *smn_analysis.StateNode, input smn_analysis.InputItf) (isEnd bool, err error) {
	b.first = false
	lex := read(input)
	b.codes = append(b.codes, lex)

	if lex.Equal(b.MBlockPair.Start) {
		b.index++
	}

	if lex.Equal(b.MBlockPair.End) {
		b.index--
	}

	if b.index == 0 {
		stateNode.Datas["Block"] = b.codes
		return true, nil
	}

	return false, nil
}

//End when end read.
func (b *BlockReader) End(stateNode *smn_analysis.StateNode) (isEnd bool, err error) {
	if b.first && b.CanIgnore {
		return true, nil
	}

	return true, onErr(b, nil, "unexcept EOF")
}

//GetProduct return result.
func (b *BlockReader) GetProduct() smn_analysis.ProductItf {
	return nil
}

//Clean let the Reader like new.  it will be call before first Read.
func (b *BlockReader) Clean() {
	b.first = true
	b.index = 0
	b.codes = GoCodes{}
}
