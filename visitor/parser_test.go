package visitor

import (
	"fmt"
	"github.com/pingcap/tidb/pkg/parser"
	"testing"
)

func TestParseSelect_01(t *testing.T) {
	querys := `
select emp_no, birth_date from employees.emp
`

	ps := parser.New()
	_, _, err := ps.Parse(querys, "", "")
	if err != nil {
		t.Fatalf("sql语法解析错误: %s", err.Error())
	}
}

func Test_GetMatchTables(t *testing.T) {
	querys := `
select emp_no, birth_date from employees.emp
`

	mTables, err := GetMatchTables(querys)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, mTable := range mTables {
		fmt.Println(mTable.Table())
	}
}
