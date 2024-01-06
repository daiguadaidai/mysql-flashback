package sql_parser

import (
	"fmt"
	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/model"
	"strings"
)

func ParseCreateStmts(query string) ([]*ast.CreateTableStmt, error) {
	ps := parser.New()
	stmts, _, err := ps.Parse(query, "", "")
	if err != nil {
		return nil, fmt.Errorf("解析建表语句 SQL 出错. %v", err)
	}

	useDatabase := ""
	stmtNodes := make([]*ast.CreateTableStmt, 0, len(stmts))
	for _, stmt := range stmts {
		switch stmtNode := stmt.(type) {
		case *ast.CreateTableStmt:
			if strings.TrimSpace(stmtNode.Table.Schema.String()) == "" {
				stmtNode.Table = &ast.TableName{
					Schema: model.NewCIStr(useDatabase),
					Name:   model.NewCIStr(stmtNode.Table.Name.String()),
				}
			}

			stmtNodes = append(stmtNodes, stmtNode)
		case *ast.UseStmt:
			useDatabase = stmtNode.DBName
		}
	}

	return stmtNodes, nil
}

func GetCreateTableColumnNames(createTableStmt *ast.CreateTableStmt) []string {
	columnNames := make([]string, 0, len(createTableStmt.Cols))

	for _, col := range createTableStmt.Cols {
		columnNames = append(columnNames, col.Name.Name.String())
	}

	return columnNames
}

func GetCreateTablePKColumnNames(createTableStmt *ast.CreateTableStmt) []string {
	columnNames := make([]string, 0, len(createTableStmt.Cols))

	for _, constraint := range createTableStmt.Constraints {
		switch constraint.Tp {
		case ast.ConstraintPrimaryKey:
			for _, column := range constraint.Keys {
				columnNames = append(columnNames, column.Column.Name.String())
			}
		}
	}

	return columnNames
}

func GetCreateTableFirstUKColumnNames(createTableStmt *ast.CreateTableStmt) (string, []string) {
	columnNames := make([]string, 0, len(createTableStmt.Cols))
	var indexName string

	for _, constraint := range createTableStmt.Constraints {
		indexName = constraint.Name
		switch constraint.Tp {
		case ast.ConstraintUniq, ast.ConstraintUniqIndex, ast.ConstraintUniqKey:
			for _, column := range constraint.Keys {
				columnNames = append(columnNames, column.Column.Name.String())
			}
		}

		if len(columnNames) != 0 {
			return indexName, columnNames
		}
	}

	return indexName, columnNames
}
