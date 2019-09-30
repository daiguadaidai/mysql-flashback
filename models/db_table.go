package models

import "fmt"

type DBTable struct {
	TableSchema string `gorm:"column:TABLE_SCHEMA"`
	TableName   string `gorm:"column:TABLE_NAME"`
}

func (this *DBTable) String() string {
	return fmt.Sprintf("%s.%s", this.TableSchema, this.TableName)
}

func NewDBTable(schema string, table string) *DBTable {
	return &DBTable{
		TableSchema: schema,
		TableName:   table,
	}
}
