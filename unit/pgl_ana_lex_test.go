package main

import (
	"testing"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/pglang/analysis/pgl_ana_lex"
)

func analysis(str string) ([]smn_analysis.ProductItf, error) {
	sm := pgl_ana_lex.NewLexAnalysiser()
	for _, char := range str {
		err := sm.Read(&pgl_ana_lex.PglaInput{Char: char})
		if err != nil {
			return nil, err
		}
	}
	sm.End()
	rc := sm.GetResultChan()
	res := make([]smn_analysis.ProductItf, 0, len(rc))
	for len(rc) != 0 {
		res = append(res, <-rc)
	}
	return res, nil
}

func TestAnalysis(t *testing.T) {
	res, err := analysis("hello")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("hello world!!!")
	for _, p := range res {
		t.Log(p.ProductType(), p)
	}
	t.Error("---------------------- end --------------")
	//TODO : input && output should read from file.
}
