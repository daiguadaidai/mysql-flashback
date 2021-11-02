package schema

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/dao"
	"github.com/daiguadaidai/mysql-flashback/utils"
	"github.com/daiguadaidai/mysql-flashback/visitor"
	"github.com/daiguadaidai/parser/opcode"
	"strings"
)

type PKType int

var (
	PKTypeAllColumns PKType = 10
	PKTypePK         PKType = 20
	PKTypeUK         PKType = 30
)

type Table struct {
	SchemaName     string
	TableName      string
	ColumnNames    []string       // 表所有的字段
	ColumnPosMap   map[string]int // 每个字段对应的slice位置
	UseColumnNames []string       // 最终需要使用的字段
	UseColumnPos   []int          // 字段对应的位点
	PKColumnNames  []string       // 主键的所有字段
	PKType                        // 主键类型 全部列. 主键. 唯一键
	InsertTemplate string         // insert sql 模板
	UpdateTemplate string         // update sql 模板
	DeleteTemplate string         // delete sql 模板
	CalcOp         []interface{}
}

func (this *Table) String() string {
	return fmt.Sprintf("%s.%s", this.SchemaName, this.TableName)
}

func NewTable(sName string, tName string) (*Table, error) {
	t := new(Table)
	t.SchemaName = sName
	t.TableName = tName

	dao := dao.NewDefaultDao()
	// 添加字段
	if err := t.addColumnNames(dao); err != nil {
		return nil, err
	}

	// 添加主键
	if err := t.addPK(dao); err != nil {
		return nil, err
	}

	// 初始化的时候将所有字段名赋值给需要的字段
	t.UseColumnNames = t.ColumnNames

	t.initAllColumnPos() // 初始化所有字段的位点信息

	t.initUseColumnPos() // 初始化使用字段的位点

	t.InitSQLTemplate()

	return t, nil
}

// 添加表的所有字段名
func (this *Table) addColumnNames(dao *dao.DefaultDao) error {
	var err error
	if this.ColumnNames, err = dao.FindTableColumnNames(this.SchemaName, this.TableName); err != nil {
		return err
	}

	if len(this.ColumnNames) == 0 {
		return fmt.Errorf("表:%s 没有获取到字段, 请确认指定表是否不存在", this.String())
	}

	return nil
}

// 添加主键
func (this *Table) addPK(dao *dao.DefaultDao) error {
	// 获取 主键
	pkColumnNames, err := dao.FindTablePKColumnNames(this.SchemaName, this.TableName)
	if err != nil {
		return fmt.Errorf("获取主键字段名出错. %v", err)
	}
	if len(pkColumnNames) > 0 {
		this.PKColumnNames = pkColumnNames
		this.PKType = PKTypePK
		return nil
	}
	seelog.Warnf("表: %s 没有主键", this.String())

	// 获取唯一键做 主键
	ukColumnNames, ukName, err := dao.FindTableUKColumnNames(this.SchemaName, this.TableName)
	if err != nil {
		return fmt.Errorf("获取唯一键做主键失败. %v", err)
	}
	if len(ukColumnNames) > 0 {
		seelog.Warnf("表: %s 设置唯一键 %s 当作主键", this.String(), ukName)
		this.PKColumnNames = ukColumnNames
		this.PKType = PKTypePK
		return nil
	}
	seelog.Warnf("表: %s 没有唯一键", this.String())

	// 所有字段为 主键
	this.PKColumnNames = this.ColumnNames
	this.PKType = PKTypeAllColumns
	seelog.Warnf("表: %s 所有字段作为该表的唯一键", this.String())

	return nil
}

// 初始所有字段的位置
func (this *Table) initAllColumnPos() {
	columnPosMap := make(map[string]int)
	for pos, name := range this.ColumnNames {
		columnPosMap[name] = pos
	}
	this.ColumnPosMap = columnPosMap
}

// 初始化使用字段位点
func (this *Table) initUseColumnPos() {
	useColumnPos := make([]int, len(this.UseColumnNames))
	for i, columnName := range this.UseColumnNames {
		pos, _ := this.ColumnPosMap[columnName]
		useColumnPos[i] = pos
	}
	this.UseColumnPos = useColumnPos
}

// 初始化sql模板
func (this *Table) InitSQLTemplate() {
	this.initInsertTemplate()
	this.initUpdateTemplate()
	this.initDeleteTemplate()
}

// 初始化 insert sql 模板
func (this *Table) initInsertTemplate() {
	template := "/* crc32:%s, %s */ INSERT INTO `%s`.`%s`(`%s`) VALUES(%s);\n"
	this.InsertTemplate = fmt.Sprintf(template, "%d", "%s", this.SchemaName, this.TableName,
		strings.Join(this.ColumnNames, "`, `"),
		utils.StrRepeat("%s", len(this.ColumnNames), ", "))
}

// 初始化 update sql 模板
func (this *Table) initUpdateTemplate() {
	template := "/* crc32:%s, %s */ UPDATE `%s`.`%s` SET %s WHERE %s;\n"
	this.UpdateTemplate = fmt.Sprintf(template, "%d", "%s", this.SchemaName, this.TableName,
		utils.SqlExprPlaceholderByColumns(this.UseColumnNames, "=", "%s", ", "),
		utils.SqlExprPlaceholderByColumns(this.PKColumnNames, "=", "%s", " AND "))
}

// 初始化 delete sql 模板
func (this *Table) initDeleteTemplate() {
	template := "/* crc32:%s, %s */ DELETE FROM `%s`.`%s` WHERE %s;\n"
	this.DeleteTemplate = fmt.Sprintf(template, "%d", "%s", this.SchemaName, this.TableName,
		utils.SqlExprPlaceholderByColumns(this.PKColumnNames, "=", "%s", " AND "))
}

func (this *Table) SetPKValues(row []interface{}, pkValues []interface{}) {
	for i, v := range this.PKColumnNames {
		pkValues[i] = row[this.ColumnPosMap[v]]
	}
}

// 设置 MTableInfo
func (this *Table) SetMTableInfo(mTable *visitor.MatchTable) error {
	// 设置需要的字段
	if !mTable.AllColumn {
		useColumnNames := make([]string, len(mTable.ColumnNames))
		for i, columnName := range mTable.ColumnNames {
			if _, ok := this.ColumnPosMap[columnName]; !ok {
				return fmt.Errorf("指定的字段不存在, 请确认. 库:%s, 表:%s, 字段:%s", this.SchemaName, this.TableName, columnName)
			}
			useColumnNames[i] = columnName
		}

		this.UseColumnNames = useColumnNames
		this.initUseColumnPos() // 初始化使用字段的位点
		this.InitSQLTemplate()
	}

	// 添加过滤条件
	if mTable.CalcOp != nil && len(mTable.CalcOp) > 0 {
		for _, op := range mTable.CalcOp {
			switch v := op.(type) {
			case *visitor.Filter:
				colPos, ok := this.ColumnPosMap[v.Left]
				if !ok {
					return fmt.Errorf("过滤条件未匹配字段: 库:%s, 表:%s, 字段: %s", mTable.SchemaName, mTable.TableName, v.Left)
				}
				v.ColPos = colPos
			}
		}
		this.CalcOp = mTable.CalcOp
	}

	return nil
}

// 获取只用字段
func (this *Table) GetUseRow(row []interface{}) []interface{} {
	useRow := make([]interface{}, len(this.UseColumnNames))
	for i, pos := range this.UseColumnPos {
		useRow[i] = row[pos]
	}
	return useRow
}

// 过滤行
func (this *Table) FilterRow(row []interface{}) bool {
	if this.CalcOp == nil || len(this.CalcOp) == 0 {
		return true
	}

	calc := utils.NewCalcStack()
	for _, op := range this.CalcOp {
		switch v := op.(type) {
		case *visitor.Filter:
			data := v.Compare(row[v.ColPos])
			calc.PushOrCalc(data)
		case opcode.Op:
			calc.PushOrCalc(v)
		}
	}

	if calc.IsEmpty() {
		return true
	}

	return calc.Result()
}

func (this *Table) GetPKValues(row []interface{}) []interface{} {
	pkValues := make([]interface{}, 0, len(this.PKColumnNames))
	for _, pkName := range this.PKColumnNames {
		pos := this.ColumnPosMap[pkName]
		pkValues = append(pkValues, row[pos])
	}

	return pkValues
}

func (this *Table) GetPKCrc32(row []interface{}) uint32 {
	pkValues := this.GetPKValues(row)
	return utils.GetCrc32ByInterfaceSlice(pkValues)
}
