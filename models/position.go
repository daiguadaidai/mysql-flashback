package models

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type Position struct {
	File              string    `gorm:"column:File"`
	Position          uint32    `gorm:"column:Position"`
	Binlog_Do_DB      string    `gorm:"column:Binlog_Do_DB"`
	Binlog_Ignore_DB  string    `gorm:"column:Binlog_Ignore_DB"`
	Executed_Gtid_Set string    `gorm:"column:Executed_Gtid_Set"`
	TS                time.Time `gorm:"-"`
}

func (this *Position) String() string {
	return fmt.Sprintf("%s:%d", this.File, this.Position)
}

// 比较两个位点是否一样
func (this *Position) Equal(other *Position) bool {
	if this.File != other.File {
		return false
	}
	if this.Position != other.Position {
		return false
	}

	return true
}

// 比较两个位点是否一样
func (this *Position) LessThan(other *Position) bool {
	if this.File < other.File {
		return true
	} else if this.File == other.File {
		if this.Position < other.Position {
			return true
		}
		return false
	} else {
		return false
	}

	return true
}

func (this *Position) GetFileNum() int64 {
	items := strings.Split(this.File, ".")
	numStr := items[len(items)-1]
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return 0
	}
	return num
}

func (this *Position) GetTotalNum() int64 {
	return (this.GetFileNum() * int64(math.MaxUint32)) + int64(this.Position)
}
