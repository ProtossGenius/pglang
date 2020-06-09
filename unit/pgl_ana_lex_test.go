package main

import (
	"os"
	"strings"
	"testing"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_analysis"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_data"
	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
	jsoniter "github.com/json-iterator/go"
)

func analysis(str string) ([]*lex_pgl.LexProduct, error) {
	sm := lex_pgl.NewLexAnalysiser()

	for _, char := range str {
		err := sm.Read(&lex_pgl.PglaInput{Char: char})
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

	arr := []*lex_pgl.LexProduct{}

	for _, p := range res {
		pro := lex_pgl.ToLexProduct(p)
		arr = append(arr, pro)
	}

	return arr, nil
}

const (
	//LexPath lex-analysis unit test file path.
	LexPath = "../datas/unit/lex_pgl"
	//LexExt lex-analysis unit test file's extension name.
	LexExt = ".lex"
	//LexOUnit lex-analysis unit test's output.
	LexOUnit = ".to"
	//LexOStd lex-analysis unite test's std-output(for compare with current output.).
	LexOStd = ".std"
)

func lexWrite(t *testing.T, ext string) {
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := smn_file.DeepTraversalDir(LexPath, func(path string, info os.FileInfo) smn_file.FileDoFuncResult {
		if info.IsDir() || !strings.HasSuffix(info.Name(), LexExt) {
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
	_, err := smn_file.DeepTraversalDir(LexPath, func(path string, info os.FileInfo) smn_file.FileDoFuncResult {
		if info.IsDir() || !strings.HasSuffix(info.Name(), LexExt) {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}
		t.Logf("dealing sameple file .....         %s", path)
		stdOut := []lex_pgl.LexProduct{}
		unitOut := []lex_pgl.LexProduct{}
		datas, err := smn_file.FileReadAll(path + LexOStd)
		check(err)
		err = jsoniter.Unmarshal(datas, &stdOut)
		check(err)
		datas, err = smn_file.FileReadAll(path + LexOUnit)
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
			if stdLp.Type != unitLp.Type {
				t.Fatalf("Error Type Not Equa. Index = %d, Std is %v, UnitOutput is %v ", i, stdLp, unitLp)
			}
		}
		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})
	check(err)
}

func TestAnalysis(t *testing.T) {
	lexWrite(t, LexOUnit)
	doCheck(t)
}
