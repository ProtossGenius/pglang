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
	//LRStatusShouldEnd in PreRead & looper is ready & looper returns a error.
	LRStatusShouldEnd
)

//NewStateNodeLoopReader .
func NewStateNodeLoopReader(looper snreader.StateNodeReader, start, end *lex_pgl.LexProduct, readEnd bool, finishDo func(node *snreader.StateNode)) snreader.StateNodeReader {
	return &StateNodeLoopReader{looper: looper, start: start, end: end, readEnd: readEnd, finishDo: finishDo}
}

//StateNodeLoopReader start loop-body end.
type StateNodeLoopReader struct {
	looper   snreader.StateNodeReader
	start    *lex_pgl.LexProduct
	end      *lex_pgl.LexProduct
	readEnd  bool // if need eat end(or free it to next Reader.).
	finishDo func(node *snreader.StateNode)

	status LoopReaderStatus
}

//Name reader's name.
func (lr *StateNodeLoopReader) Name() string {
	return "StateNodeLoopReader"
}

//Clean let the Reader like new.  it will be call before first Read.
func (lr *StateNodeLoopReader) Clean() {
	lr.looper.Clean()
	lr.status = LRStatusStart
}
func (lr *StateNodeLoopReader) whenFinish(node *snreader.StateNode) (bool, error) {
	lr.finishDo(node)
	return true, nil
}

//PreRead only see if should stop read.
func (lr *StateNodeLoopReader) PreRead(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	lex := read(input)
	needLoop := true
	for needLoop {
		needLoop = false
		switch lr.status {
		case LRStatusStart:
			if lr.start == nil {
				lr.status = LRStatusLooperReady
				return false, nil
			}

			if !lex.Equal(lr.start) {
				return true, onErr(lr, lex, fmt.Sprintf("Want %v", lr.start))
			}

			return false, nil
		case LRStatusLooperGoing:
			{
				lend, lerr := lr.looper.PreRead(stateNode, input)
				if lerr != nil {
					return true, onErr(lr, lex, lerr.Error())
				}

				if lend {
					lr.status = LRStatusLooperReady
					lr.looper.Clean()
					needLoop = true
				}

				return false, nil
			}
		case LRStatusLooperReady:
			{
				lend, lerr := lr.looper.PreRead(stateNode, input)
				if lerr != nil {
					lr.status = LRStatusShouldEnd
					return false, nil
				}

				if lend {
					return true, onErr(lr, lex, "Unexcept end in looper's PreRead")
				}

				lr.status = LRStatusLooperGoing
				return false, nil
			}
		case LRStatusShouldEnd:
			{
				if !lex.Equal(lr.end) {
					return true, onErr(lr, lex, fmt.Sprintf("Want %v", lr.end))
				}
			}
		}
	}

	return false, nil
}

//Read real read. even isEnd == true the input be readed.
func (lr *StateNodeLoopReader) Read(stateNode *snreader.StateNode, input snreader.InputItf) (isEnd bool, err error) {
	if lr.status == LRStatusStart {
		lr.status = LRStatusLooperReady
		return false, nil
	}

	if lr.status == LRStatusShouldEnd {
		return lr.whenFinish(stateNode)
	}

	lex := read(input)
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
