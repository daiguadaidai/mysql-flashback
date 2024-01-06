package visitor

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/utils"
	"github.com/pingcap/tidb/pkg/parser/opcode"
	"strings"
)

type MatchTable struct {
	SchemaName        string
	TableName         string
	ColumnNames       []string
	AllColumn         bool
	StartLogFile      string
	StartLogPos       uint64
	EndLogFile        string
	EndLogPos         uint64
	StartRollBackTime string
	EndRollBackTime   string
	ThreadId          uint32
	CalcOp            []interface{}
}

func NewMatchTable() *MatchTable {
	return &MatchTable{
		ColumnNames: make([]string, 0),
		CalcOp:      make([]interface{}, 0),
	}
}

func (this *MatchTable) Table() string {
	return fmt.Sprintf("%s.%s", this.SchemaName, this.TableName)
}

// 是否有开始位点信息
func (this *MatchTable) HaveStartPosInfo() bool {
	if strings.TrimSpace(this.StartLogFile) == "" {
		return false
	}
	return true
}

// 是否所有结束位点信息
func (this *MatchTable) HaveEndPosInfo() bool {
	if strings.TrimSpace(this.EndLogFile) == "" {
		return false
	}
	return true
}

// 是否有开始事件
func (this *MatchTable) HaveStartTime() bool {
	if strings.TrimSpace(this.StartRollBackTime) == "" {
		return false
	}
	return true
}

// 是否有结束时间
func (this *MatchTable) HaveEndTime() bool {
	if strings.TrimSpace(this.EndRollBackTime) == "" {
		return false
	}
	return true
}

// 开始位点小于其他位点
func (this *MatchTable) StartPosInfoLessThan(other *MatchTable) bool {
	if this.StartLogFile < other.StartLogFile {
		return true
	} else if this.StartLogFile == other.StartLogFile {
		if this.StartLogPos < other.StartLogPos {
			return true
		}
	}
	return false
}

// 结束位点大于其他位点
func (this *MatchTable) EndPostInfoRatherThan(other *MatchTable) bool {
	if this.EndLogFile > other.EndLogFile {
		return true
	} else if this.EndLogFile == other.EndLogFile {
		if this.EndLogPos > other.EndLogPos {
			return true
		}
	}
	return false
}

// 开始时间小于其他位点
func (this *MatchTable) StartTimeLessThan(other *MatchTable) (bool, error) {
	ts1, err1 := utils.StrTime2Int(this.StartRollBackTime)
	ts2, err2 := utils.StrTime2Int(other.StartRollBackTime)
	if err1 == nil && err2 == nil {
		return ts1 < ts2, nil
	} else if err1 == nil && err2 != nil {
		return true, nil
	} else if err1 != nil && err2 == nil {
		return false, nil
	}

	return false, fmt.Errorf("MatchTable StartTimeLessThan 比较出错. %s. %s", err1.Error(), err2.Error())
}

// 结束时间大于其他位点
func (this *MatchTable) EndTimeRatherThan(other *MatchTable) (bool, error) {
	ts1, err1 := utils.StrTime2Int(this.EndRollBackTime)
	ts2, err2 := utils.StrTime2Int(other.EndRollBackTime)
	if err1 == nil && err2 == nil {
		return ts1 > ts2, nil
	} else if err1 == nil && err2 != nil {
		return true, nil
	} else if err1 != nil && err2 == nil {
		return false, nil
	}

	return false, fmt.Errorf("MatchTable EndTimeRatherThan 比较出错. %s. %s", err1.Error(), err2.Error())
}

type Filter struct {
	Left   string
	Op     opcode.Op
	Right  interface{}
	ColPos int // 字段所在的位置
}

func NewFilter(col string, op opcode.Op, val interface{}) *Filter {
	return &Filter{
		Left:  col,
		Op:    op,
		Right: val,
	}
}

func (this *Filter) String() string {
	return this.Left
}

func (this *Filter) Compare(other interface{}) bool {
	switch this.Op {
	case opcode.EQ: // ==
		return utils.Equal(other, this.Right)
	case opcode.NE: // <>
		return utils.NotEqual(other, this.Right)
	case opcode.LT: // <
		return utils.Less(other, this.Right)
	case opcode.LE: // <=
		return utils.LessEqual(other, this.Right)
	case opcode.GT: // >
		return utils.Rather(other, this.Right)
	case opcode.GE: // >=
		return utils.RatherEqual(other, this.Right)
	case opcode.IsNull: // is null
		return utils.IsNull(other)
	case opcode.In:
		return this.compareIn(other)
	case opcode.NullEQ: // 没有 bewteen 只能使用这个来代替
		return this.compareBetween(other)
	}
	return false
}

// 比较 IN 表达式
func (this *Filter) compareIn(other interface{}) bool {
	inElement, ok := this.Right.(*InElement)
	if !ok {
		seelog.Warnf("进行 IN 比较, 但是无法转化InElement类型进行比较")
		return false
	}
	if inElement.Not {
		return !inElement.Matched(other)
	}
	return inElement.Matched(other)
}

// 比较 between ... and ... 表达式
func (this *Filter) compareBetween(other interface{}) bool {
	bewteenElement, ok := this.Right.(*BetweenAndElement)
	if !ok {
		seelog.Warnf("进行 BEWTEEN ... AND ... 比较, 但是无法转化 BetweenAndElement 类型")
	}
	if bewteenElement.Not {
		return !bewteenElement.Matched(other)
	}
	return bewteenElement.Matched(other)
}

const (
	IN_KEY_TYPE_NONE = iota
	IN_KEY_TYPE_INT64
	IN_KEY_TYPE_UINT64
	IN_KEY_TYPE_FLOAT64
	IN_KEY_TYPE_STR
)

var keyTypeMap map[int]string = map[int]string{
	IN_KEY_TYPE_NONE:    "None",
	IN_KEY_TYPE_INT64:   "Int64",
	IN_KEY_TYPE_UINT64:  "Uint64",
	IN_KEY_TYPE_FLOAT64: "Float64",
	IN_KEY_TYPE_STR:     "String",
}

type InElement struct {
	KeyType int
	Not     bool
	Data    map[interface{}]struct{}
}

func NewInElement(keyType int, not bool) *InElement {
	return &InElement{
		KeyType: keyType,
		Not:     not,
		Data:    make(map[interface{}]struct{}),
	}
}

func (this *InElement) Matched(other interface{}) bool {
	switch this.KeyType {
	case IN_KEY_TYPE_INT64:
		key, err := utils.InterfaceToInt64(other)
		if err != nil {
			seelog.Warnf("进行 IN 比较但是, 将数据转化为 Int64 出错. 需要转化的值:%v. %s", other, err.Error())
		}
		if _, ok := this.Data[key]; !ok {
			return false
		}
		return true
	case IN_KEY_TYPE_UINT64:
		key, err := utils.InterfaceToUint64(other)
		if err != nil {
			seelog.Warnf("进行 IN 比较但是, 将数据转化为 Uint64 出错. 需要转化的值:%v. %s", other, err.Error())
		}
		if _, ok := this.Data[key]; !ok {
			return false
		}
		return true
	case IN_KEY_TYPE_FLOAT64:
		key, err := utils.InterfaceToFloat64(other)
		if err != nil {
			seelog.Warnf("进行 IN 比较但是, 将数据转化为 Float64 出错. 需要转化的值:%v. %s", other, err.Error())
		}
		if _, ok := this.Data[key]; !ok {
			return false
		}
	case IN_KEY_TYPE_STR:
		key := utils.InterfaceToStr(other)
		if _, ok := this.Data[key]; !ok {
			return false
		}
		return true
	}
	return false
}

func GetKeyType(data interface{}) int {
	switch data.(type) {
	case int8, int16, int32, int64, int:
		return IN_KEY_TYPE_INT64
	case uint8, uint16, uint32, uint64, uint:
		return IN_KEY_TYPE_UINT64
	case float32, float64:
		return IN_KEY_TYPE_FLOAT64
	case string:
		return IN_KEY_TYPE_STR
	case []uint8:
		return IN_KEY_TYPE_STR
	}
	return IN_KEY_TYPE_NONE
}

func GetKeyTypeString(key int) string {
	if typeStr, ok := keyTypeMap[key]; ok {
		return typeStr
	}
	return "未识别"
}

type BetweenAndElement struct {
	Not   bool
	Left  interface{}
	Right interface{}
}

func NewBetweenAndElement(not bool, left interface{}, right interface{}) *BetweenAndElement {
	return &BetweenAndElement{
		Not:   not,
		Left:  left,
		Right: right,
	}
}

func (this *BetweenAndElement) Matched(other interface{}) bool {
	if utils.RatherEqual(other, this.Left) && utils.LessEqual(other, this.Right) {
		return true
	}
	return false
}
