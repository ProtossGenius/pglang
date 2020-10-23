package classifygo

import (
	"fmt"

	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
	"github.com/ProtossGenius/pglang/snreader"
)

//LoopReaderStatus .
type LoopReaderStatus int

const (
	//LRStatusStart after clean.
	LRStatusStart LoopReaderStatus = iota
	//LRStatusLooperReady looper is cleaned.
	LRStatusLooperReady
	//LRStatusLooperGoing looper is reading.
	LRStatusLooperGoing
	// LRStatusLooperEnd looper end in preread.
	LRStatusLooperEnd
)

/*NewStateNodeLoopReader use looper in a scope (between start&end).
.
*/
func NewStateNodeLoopReader(looper snreader.StateNodeReader, start, end *lex_pgl.LexProduct, readEnd bool, finishDo func(node *snreader.StateNode)) *StateNodeLoopReader {
	return &StateNodeLoopReader{looper: looper, start: start, end: end, readEnd: readEnd, finishDo: finishDo, name: "StateNodeLoopReader"}
}

//StateNodeLoopReader start loop-body end.
type StateNodeLoopReader struct {
	looper   snreader.StateNodeReader
	start    *lex_pgl.LexProduct
	end      *lex_pgl.LexProduct
	readEnd  bool // if need eat end(or free it to next Reader.).
	finishDo func(node *snreader.StateNode)
	name     string
	status   LoopReaderStatus
}

// SetName .
func (lr *StateNodeLoopReader) SetName(name string) *StateNodeLoopReader {
	lr.name = name
	return lr
}

//Name reader's name.
func (lr *StateNodeLoopReader) Name() string {
	return lr.name
}

//Clean let the Reader like new.  it will be call before first Read.
func (lr *StateNodeLoopReader) Clean() {
	lr.looper.Clean()
	lr.status = LRStatusStart
}
func (lr *StateNodeLoopReader) whenFinish(node *snreader.StateNode) (bool, error) {
	if lr.finishDo != nil {
		lr.finishDo(node)
	}
	return true, nil
}

//PreRead only see if should stop read.
func (lr *StateNodeLoopReader) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)
	fmt.Println("......StateNodeLoopReader PreRead", lex_pgl.PglaNameMap[lex.Type], lex.Value)
	if lr.end == nil {
		return true, onErr(lr, lex, "StateNodeLoopReader's end should't be nil")
	}
	if lr.status == LRStatusStart {
		lr.status = LRStatusLooperReady
		if lr.start == nil || lex.Equal(lr.start) {
			return false, nil
		}

		return true, onErr(lr, lex, "first input except "+lex.Value)
	}

	if lr.status == LRStatusLooperReady {
		if lex.Equal(lr.end) {
			if lr.readEnd {
				return false, nil
			}

			return lr.whenFinish(stateNode)
		}
	}

	lr.status = LRStatusLooperGoing

	lend, lerr := lr.looper.PreRead(stateNode, input)
	if lerr != nil {
		return true, onErr(lr, lex, lerr.Error())
	}
	if lend {
		lr.looper.Clean()
		lr.status = LRStatusLooperEnd
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (lr *StateNodeLoopReader) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)
	fmt.Println("......StateNodeLoopReader Read", lex_pgl.PglaNameMap[lex.Type], lex.Value)
	if lr.status == LRStatusStart {
		return false, nil
	}

	if lr.status == LRStatusLooperEnd {
		lr.status = LRStatusLooperReady

		return false, nil
	}

	if lr.status == LRStatusLooperReady {
		if lex.Equal(lr.end) {
			return lr.whenFinish(stateNode)
		}
	}

	lr.status = LRStatusLooperEnd

	lend, lerr := lr.looper.Read(stateNode, input)
	if lerr != nil {
		return true, onErr(lr, lex, lerr.Error())
	}

	if lend {
		lr.looper.Clean()
		lr.status = LRStatusLooperReady
	}

	return false, nil
}

//End when end read.
func (lr *StateNodeLoopReader) End(stateNode *snreader.StateNode) (isEnd bool, err error) {
	return true, onErr(lr, nil, ErrUnexceptEOF)
}

//GetProduct return result.
func (lr *StateNodeLoopReader) GetProduct() snreader.ProductItf {
	return nil
}
