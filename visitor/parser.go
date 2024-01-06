package visitor

import (
	"fmt"
	"github.com/pingcap/tidb/pkg/parser"
	_ "github.com/pingcap/tidb/pkg/types/parser_driver"
)

func GetMatchTables(querys string) ([]*MatchTable, error) {
	ps := parser.New()
	stmts, _, err := ps.Parse(querys, "", "")
	if err != nil {
		return nil, fmt.Errorf("sql语法解析错误: %s", err.Error())
	}

	mTables := make([]*MatchTable, len(stmts))
	for i, stmt := range stmts {
		vst := NewSelectVisitor()
		stmt.Accept(vst)
		if vst.Err != nil {
			return nil, fmt.Errorf("%s. 语句: %s", vst.Err.Error(), stmt.Text())
		}
		mTables[i] = vst.MTable
	}

	return mTables, nil
}
