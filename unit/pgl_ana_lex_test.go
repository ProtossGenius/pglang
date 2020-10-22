package unit_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ProtossGenius/SureMoonNet/basis/smn_file"
	"github.com/ProtossGenius/pglang/analysis/lex_pgl"
	"github.com/ProtossGenius/pglang/snreader"
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
	res := make([]snreader.ProductItf, 0, len(rc))

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
	// LexPath lex-analysis unit test file path.
	LexPath = "../datas/unit/lex_pgl"
	// LexExt lex-analysis unit test file's extension name.
	LexExt = ".lex"
	// LexOUnit lex-analysis unit test's output.
	LexOUnit = ".to"
	// LexOStd lex-analysis unite test's std-output(for compare with current output.).
	LexOStd = ".std"
)

func lexWrite(t *testing.T, lexPath, lexExt, ext string, doing func(t *testing.T, src, out string, lexs []*lex_pgl.LexProduct)) {
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}
	_, err := smn_file.DeepTraversalDir(lexPath, func(path string, info os.FileInfo) smn_file.FileDoFuncResult {
		if info.IsDir() || !strings.HasSuffix(info.Name(), lexExt) {
			return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
		}
		datas, err := smn_file.FileReadAll(path)
		check(err)
		pro, err := analysis(string(datas))
		check(err)
		doing(t, path, path+ext, pro)

		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})
	check(err)
}

func strDeal(str string) string {
	str = strings.ReplaceAll(str, "\\", "\\\\")

	return strings.ReplaceAll(str, "\n", "\\n")
}

func writeLexProduct(t *testing.T, src, out string, list []*lex_pgl.LexProduct) {
	check := func(err error) {
		if err != nil {
			t.Fatal(err)
		}
	}

	f, err := smn_file.CreateNewFile(out)
	check(err)

	defer f.Close()

	for _, lp := range list {
		_, err = f.WriteString(fmt.Sprintf("%s %s\n", lex_pgl.PglaNameMap[lp.Type], strDeal(lp.Value)))
		check(err)
	}
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
		dStd, err := smn_file.FileReadAll(path + LexOStd)
		check(err)
		dUnit, err := smn_file.FileReadAll(path + LexOUnit)
		check(err)
		if string(dStd) != string(dUnit) {
			t.Fatalf("Error Result Not Equa.")
		}

		return smn_file.FILE_DO_FUNC_RESULT_DEFAULT
	})
	check(err)
}

func TestAnalysis(t *testing.T) {
	lexWrite(t, LexPath, LexExt, LexOUnit, writeLexProduct)
	doCheck(t)
}
