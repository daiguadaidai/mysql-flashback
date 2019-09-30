package dao

import (
	"fmt"
	"github.com/daiguadaidai/mysql-flashback/gdbc"
	"github.com/daiguadaidai/mysql-flashback/models"
	"github.com/jinzhu/gorm"
)

type DefaultDao struct {
	DB *gorm.DB
}

func NewDefaultDao() *DefaultDao {
	return &DefaultDao{
		DB: gdbc.GetOrmInstance().DB,
	}
}

func (this *DefaultDao) ShowBinaryLogs() ([]*models.BinaryLog, error) {
	sql := `SHOW BINARY LOGS;`
	var bLogs []*models.BinaryLog
	if err := this.DB.Raw(sql).Find(&bLogs).Error; err != nil {
		return nil, err
	}
	return bLogs, nil
}

func (this *DefaultDao) ShowMasterStatus() (*models.Position, error) {
	sql := `SHOW MASTER STATUS`
	pos := new(models.Position)
	if err := this.DB.Raw(sql).Scan(pos).Error; err != nil {
		return nil, err
	}
	return pos, nil
}

// 删除一个不存在的表
func (this *DefaultDao) DropNotExistsTable() error {
	sql := "DROP TABLE IF EXISTS `__gmod__`.`__gmod__`"
	return this.DB.Raw(sql).Error
}

// 获取表通过schema
func (this *DefaultDao) FindTablesBySchema(sName string) ([]*models.DBTable, error) {
	sql := `
    SELECT TABLE_SCHEMA,
        TABLE_NAME
    FROM information_schema.TABLES
    WHERE TABLE_TYPE = 'BASE TABLE'
        AND TABLE_SCHEMA = ?
`
	var tables []*models.DBTable
	if err := this.DB.Raw(sql, sName).Find(&tables).Error; err != nil {
		return nil, err
	}

	return tables, nil
}

// 获取表中所有的字段
func (this *DefaultDao) FindTableColumnNames(sName string, tName string) ([]string, error) {
	sql := `
    SELECT COLUMN_NAME
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = ?
        AND TABLE_NAME = ?
    ORDER BY ORDINAL_POSITION ASC
`
	var cNames []string

	if err := this.DB.Raw(sql, sName, tName).Pluck("COLUMN_NAME", &cNames).
		Error; err != nil {
		return nil, err
	}

	return cNames, nil
}

// 获取主键字段名
func (this *DefaultDao) FindTablePKColumnNames(sName string, tName string) ([]string, error) {
	sql := `
    SELECT S.COLUMN_NAME
    FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS AS TC
    LEFT JOIN INFORMATION_SCHEMA.STATISTICS AS S
        ON TC.TABLE_SCHEMA = S.INDEX_SCHEMA
        AND TC.TABLE_NAME = S.TABLE_NAME
        AND TC.CONSTRAINT_NAME = S.INDEX_NAME 
    WHERE TC.TABLE_SCHEMA = ?
        AND TC.TABLE_NAME = ?
        AND TC.CONSTRAINT_TYPE = 'PRIMARY KEY'
        ORDER BY SEQ_IN_INDEX ASC
`

	var cNames []string

	if err := this.DB.Raw(sql, sName, tName).Pluck("COLUMN_NAME", &cNames).
		Error; err != nil {
		return nil, err
	}

	return cNames, nil
}

// 获取唯一键字段
func (this *DefaultDao) FindTableUKColumnNames(sName string, tName string) ([]string, string, error) {
	ukName, err := this.GetUKName(sName, tName)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return make([]string, 0), "", nil
		}
		return nil, "", err
	}
	if ukName == "" {
		return make([]string, 0), "", nil
	}

	sql := `
    SELECT COLUMN_NAME
    FROM INFORMATION_SCHEMA.STATISTICS
    WHERE TABLE_SCHEMA = ?
        AND TABLE_NAME = ?
        AND INDEX_NAME = ?
    ORDER BY SEQ_IN_INDEX ASC
`

	var cNames []string
	if err := this.DB.Raw(sql, sName, tName, ukName).Pluck("COLUMN_NAME", &cNames).
		Error; err != nil {
		return nil, "", err
	}
	return cNames, ukName, nil
}

// 获取第一个唯一键
func (this *DefaultDao) GetUKName(sName string, tName string) (string, error) {
	sql := `
    SELECT CONSTRAINT_NAME
    FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
    WHERE TABLE_SCHEMA = ?
        AND TABLE_NAME = ?
        AND CONSTRAINT_TYPE = 'UNIQUE'
    LIMIT 0, 1;
`

	var ukName string
	if err := this.DB.Raw(sql, sName, tName).Row().Scan(&ukName); err != nil {
		return "", err
	}

	return ukName, nil
}

// 执行dml
func (this *DefaultDao) ExecDML(sql string) error {
	return this.DB.Exec(sql).Error
}

// 获取最老和最新的日志位点
func (this *DefaultDao) GetOldestAndNewestPos() (*models.Position, *models.Position, error) {
	logs, err := this.ShowBinaryLogs()
	if err != nil {
		return nil, nil, err
	}

	if len(logs) == 0 {
		return nil, nil, fmt.Errorf("没有binlog")
	}

	startPos := &models.Position{
		File:     logs[0].LogName,
		Position: 4,
	}
	endPos := &models.Position{
		File:     logs[len(logs)-1].LogName,
		Position: logs[len(logs)-1].FileSize,
	}

	return startPos, endPos, nil
}
