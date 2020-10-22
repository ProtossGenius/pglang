package classifygo

import (
	"fmt"

	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
	"github.com/ProtossGenius/pglang/snreader"
)

//BlockPair block pair.
type BlockPair struct {
	Start *lex_pgl.LexProduct
	End   *lex_pgl.LexProduct
}

func read(input snreader.InputItf) *lex_pgl.LexProduct {
	return input.(*lex_pgl.LexProduct)
}

func onErr(reader snreader.StateNodeReader, lex *lex_pgl.LexProduct, reason string) error {
	return fmt.Errorf("Error in [%s], input is [%v] reason is: %s", reader.Name(), lex, reason)
}

//NewBlockReader read block.
func NewBlockReader(start, end *lex_pgl.LexProduct, canIgnore, readEnd bool, key string, finishDo func(stateNode *snreader.StateNode)) *BlockReader {
	if end == nil {
		return nil
	}

	return &BlockReader{MBlockPair: &BlockPair{start, end}, canIgnore: canIgnore, readEndFlag: readEnd, key: key, finishDo: finishDo, name: "BlockReader"}
}

// NewEndFlagReader read until find the end flag.
func NewEndFlagReader(endFlag *lex_pgl.LexProduct, key string, readEndFlag bool, finishDo func(stateNode *snreader.StateNode)) *BlockReader {
	if endFlag == nil {
		return nil
	}

	return &BlockReader{MBlockPair: &BlockPair{nil, endFlag}, canIgnore: false, readEndFlag: readEndFlag, key: key, finishDo: finishDo}
}

//BlockReader block reader [{( .. )}].
type BlockReader struct {
	MBlockPair  *BlockPair
	canIgnore   bool
	key         string
	finishDo    func(stateNode *snreader.StateNode)
	readEndFlag bool
	first       bool
	index       int
	codes       GoCodes
	name        string
}

// SetName .
func (b *BlockReader) SetName(name string) *BlockReader {
	b.name = name
	return b
}

//Name reader's name.
func (b *BlockReader) Name() string {
	return b.name
}

//PreRead only see if should stop read.
func (b *BlockReader) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)
	if lex.Equal(b.MBlockPair.End) {
		if b.index == 1 && b.readEndFlag == false {
			stateNode.Datas[b.key] = b.codes

			if b.finishDo != nil {
				b.finishDo(stateNode)
			}

			return true, nil
		}

		return false, nil
	}

	if !b.first {
		return false, nil
	}

	if b.MBlockPair.Start == nil {
		b.index = 1
		return false, nil
	}

	if !lex.Equal(b.MBlockPair.Start) {
		if b.canIgnore {
			return true, nil
		}

		return true, onErr(b, lex, "except start-input ["+b.MBlockPair.Start.Value+"]")
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (b *BlockReader) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
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
		stateNode.Datas[b.key] = b.codes
		if b.finishDo != nil {
			b.finishDo(stateNode)
		}

		return true, nil
	}

	return false, nil
}

//End when end read.
func (b *BlockReader) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	if b.first && b.canIgnore {
		return true, nil
	}

	return true, onErr(b, nil, "unexcept EOF")
}

//GetProduct return result.
func (b *BlockReader) GetProduct() snreader.ProductItf {
	return nil
}

//Clean let the Reader like new.  it will be call before first Read.
func (b *BlockReader) Clean() {
	b.first = true
	b.index = 0
	b.codes = GoCodes{}
}
