package visitor

import (
	"fmt"
	"github.com/pingcap/tidb/pkg/parser"
	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
	"testing"
)

func TestSelectVisitor_Accept_01(t *testing.T) {
	sqlStr := `
SELECT *, a, b FROM table1 WHERE name = 1 or age = 2;
SELECT * FROM table1 WHERE name = 1;
SELECT
    (select 1 FROM dual) as col1
FROM test1;
SELECT
    col1
FROM test1
WHERE name = (select name from test2);
`
	ps := parser.New()
	stmts, _, err := ps.Parse(sqlStr, "", "")
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, stmt := range stmts {
		vst := NewSelectVisitor()
		stmt.Accept(vst)
	}
}

func TestSelectVisitor_MatchTable(t *testing.T) {
	sqlStr := `
SELECT *, a, b
FROM schema.table1
WHERE name = 1
    and age = 11 
    or name = 2
    and (
        name = 3
        or name = 5
        and start_log_file = 'mysql-bin.000001'
        and (
            name = 9
            and start_log_pos = 128
            and name = 8
        )
    )
	and start_log_file = 'mysql-bin.000001'
    and start_log_pos = 128
;
SELECT * FROM schema11.* WHERE name = 1;
SELECT
    (select 1 FROM dual) as col1
FROM test1;
SELECT
    col1
FROM test1
WHERE name = (select name from test2);
`
	ps := parser.New()
	stmts, _, err := ps.Parse(sqlStr, "", "")
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, stmt := range stmts {
		vst := NewSelectVisitor()
		stmt.Accept(vst)
		if vst.Err != nil {
			fmt.Println("Error:", vst.Err.Error())
		}
		fmt.Println(vst.MTable.Table())
		fmt.Println(vst.MTable.CalcOp)
	}
}

func TestSelectVisitor_In(t *testing.T) {
	sqlStr := `
SELECT *, a, b
FROM schema.table1
WHERE name IN(1,2,3,4.2) and age IN(22,33,44)
`
	ps := parser.New()
	stmts, _, err := ps.Parse(sqlStr, "", "")
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, stmt := range stmts {
		vst := NewSelectVisitor()
		stmt.Accept(vst)
		if vst.Err != nil {
			fmt.Println("Error:", vst.Err.Error())
		}
		fmt.Println(vst.MTable.Table())
		fmt.Println(vst.MTable.CalcOp)
		for _, op := range vst.MTable.CalcOp {
			switch v := op.(type) {
			case *Filter:
				switch data := v.Right.(type) {
				case *InElement:
					fmt.Println(data, GetKeyTypeString(data.KeyType))
				}

			}
		}

	}
}

func TestSelectVisitor_BetweenAnd(t *testing.T) {
	sqlStr := `
SELECT *, a, b
FROM schema.table1
WHERE name IN(1,2,3,4.2)
    AND age not between 10 and 20
`
	ps := parser.New()
	stmts, _, err := ps.Parse(sqlStr, "", "")
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, stmt := range stmts {
		vst := NewSelectVisitor()
		stmt.Accept(vst)
		if vst.Err != nil {
			fmt.Println("Error:", vst.Err.Error())
		}
		fmt.Println(vst.MTable.Table())
		fmt.Println(vst.MTable.CalcOp)
		for _, op := range vst.MTable.CalcOp {
			switch v := op.(type) {
			case *Filter:
				switch data := v.Right.(type) {
				case *InElement:
					fmt.Println(data, GetKeyTypeString(data.KeyType))
				}

			}
		}

	}
}
