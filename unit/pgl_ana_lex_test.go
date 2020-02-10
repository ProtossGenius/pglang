package main

import (
	"os"
	"strings"
	"testing"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_data"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/pglang/analysis/pgl_ana_lex"
	jsoniter "github.com/json-iterator/go"
)

func analysis(str string) ([]*pgl_ana_lex.LexProduct, error) {
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
	arr := []*pgl_ana_lex.LexProduct{}
	for _, p := range res {
		pro := pgl_ana_lex.ToLexProduct(p)
		arr = append(arr, pro)

	}
	return arr, nil
}

const (
	LEX_PATH   = "../datas/unit/pgl_ana_lex"
	LEX_EXT    = ".lex"
	LEX_O_UNIT = ".to"
	LEX_O_STD  = ".std"
)

func lexWrite(t *testing.T, ext string) {
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := smn_file.DeepTraversalDir(LEX_PATH, func(path string, info os.FileInfo) smn_file.FileDoFuncResult {
		if info.IsDir() || !strings.HasSuffix(info.Name(), LEX_EXT) {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}
		datas, err := smn_file.FileReadAll(path)
		check(err)
		pro, err := analysis(string(datas))
		check(err)
		jstr, err := smn_data.ValToJson(pro)
		check(err)
		f, err := smn_file.CreateNewFile(path + ext)
		check(err)
		f.WriteString(jstr)
		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})
	check(err)
}

func doCheck(t *testing.T) {
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := smn_file.DeepTraversalDir(LEX_PATH, func(path string, info os.FileInfo) smn_file.FileDoFuncResult {
		if info.IsDir() || !strings.HasSuffix(info.Name(), LEX_EXT) {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}
		stdOut := []pgl_ana_lex.LexProduct{}
		unitOut := []pgl_ana_lex.LexProduct{}
		datas, err := smn_file.FileReadAll(path + LEX_O_STD)
		check(err)
		err = jsoniter.Unmarshal(datas, &stdOut)
		check(err)
		datas, err = smn_file.FileReadAll(path + LEX_O_UNIT)
		check(err)
		err = jsoniter.Unmarshal(datas, &unitOut)
		check(err)
		lenStd := len(stdOut)
		lenUnit := len(unitOut)
		if lenStd != lenUnit {
			t.Fatal("Error output len not equal, lex file path = ", path)
		}
		for i := 0; i < lenStd; i++ {
			stdLp := stdOut[i]
			unitLp := unitOut[i]
			t.Logf("type: <%s>,  value :[%s]", pgl_ana_lex.PglaNameMap[stdLp.Type], stdLp.Value)
			if stdLp.Type != unitLp.Type {
				t.Fatalf("Error Type Not Equa. Index = %d, Std is %v, UnitOutput is %v ", i, stdLp, unitLp)
			}
		}
		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})
	check(err)
}

func TestAnalysis(t *testing.T) {
	lexWrite(t, LEX_O_UNIT)
	doCheck(t)
	//TODO : input && output should read from file.
}
