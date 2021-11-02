package config

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/utils"
	"time"
)

const (
	ENABLE_ROLLBACK_UPDATE = true
	ENABLE_ROLLBACK_INSERT = true
	ENABLE_ROLLBACK_DELETE = true
	SAVE_DIR               = "rollback_sqls"
)

var sc *CreateConfig

type CreateConfig struct {
	StartLogFile         string
	StartLogPos          uint64
	EndLogFile           string
	EndLogPos            uint64
	StartTime            string
	EndTime              string
	RollbackSchemas      []string
	RollbackTables       []string
	ThreadID             uint32
	Now                  time.Time
	EnableRollbackUpdate bool
	EnableRollbackInsert bool
	EnableRollbackDelete bool
	SaveDir              string
	MatchSql             string // 使用sql语句来匹配需要查询的时间段或者
}

func NewStartConfig() *CreateConfig {
	return &CreateConfig{
		Now: time.Now(),
	}
}

func SetStartConfig(cfg *CreateConfig) {
	sc = cfg
}

// 是否有开始位点信息
func (this *CreateConfig) HaveStartPosInfo() bool {
	if this.StartLogFile == "" {
		return false
	}
	return true
}

// 是否所有结束位点信息
func (this *CreateConfig) HaveEndPosInfo() bool {
	if this.EndLogFile == "" {
		return false
	}
	return true
}

// 是否有开始事件
func (this *CreateConfig) HaveStartTime() bool {
	if this.StartTime == "" {
		return false
	}
	return true
}

// 是否有结束时间
func (this *CreateConfig) HaveEndTime() bool {
	if this.EndTime == "" {
		return false
	}
	return true
}

// 设置最终的保存文件
func (this *CreateConfig) GetSaveDir() string {
	if len(this.SaveDir) == 0 {
		cmdDir, err := utils.CMDDir()
		if err != nil {
			saveDir := fmt.Sprintf("./%s", SAVE_DIR)
			seelog.Errorf("获取命令所在路径失败, 使用默认路径: %s. %v",
				saveDir, err.Error())
			return saveDir
		}
		return fmt.Sprintf("%s/%s", cmdDir, SAVE_DIR)
	}

	return this.SaveDir
}

func (this *CreateConfig) Check() error {
	if err := this.checkCondition(); err != nil {
		return err
	}

	if err := utils.CheckAndCreatePath(this.GetSaveDir(), "回滚文件存放路径"); err != nil {
		return err
	}

	return nil
}

func (this *CreateConfig) checkCondition() error {
	if this.StartLogFile != "" && this.StartLogPos >= 0 &&
		this.EndLogFile != "" && this.EndLogPos >= 0 {
		if !this.StartPosInfoLessThan(this.EndLogFile, this.EndLogPos) {
			return fmt.Errorf("指定的结束位点 大于 开始位点")
		}
		return nil
	} else if this.StartLogFile != "" && this.StartLogPos >= 0 &&
		this.EndTime != "" {

		ts, err := utils.StrTime2Int(this.EndTime)
		if err != nil {
			return fmt.Errorf("指定的结束事件有问题")
		}
		if ts > (this.Now.Unix()) {
			return fmt.Errorf("指定的时间还没有到来")
		}
		return nil
	} else if this.StartTime != "" && this.EndLogFile != "" && this.EndLogPos >= 0 {
		return nil
	} else if this.StartTime != "" && this.EndTime != "" {
		ts, err := utils.StrTime2Int(this.EndTime)
		if err != nil {
			return fmt.Errorf("指定的结束事件有问题")
		}
		if ts > (this.Now.Unix()) {
			return fmt.Errorf("指定的时间还没有到来")
		}
		return nil
	}

	return fmt.Errorf("指定的开始位点和结束位点无效")
}

// 开始位点小于其他位点
func (this *CreateConfig) StartPosInfoLessThan(otherStartFile string, otherStartPos uint64) bool {
	if this.StartLogFile < otherStartFile {
		return true
	} else if this.StartLogFile == otherStartFile {
		if this.StartLogPos < otherStartPos {
			return true
		}
	}
	return false
}

// 结束位点大于其他位点
func (this *CreateConfig) EndPostInfoRatherThan(otherEndFile string, otherEndPos uint64) bool {
	if this.EndLogFile > otherEndFile {
		return true
	} else if this.EndLogFile == otherEndFile {
		if this.EndLogPos > otherEndPos {
			return true
		}
	}
	return false
}

// 开始时间小于其他位点
func (this *CreateConfig) StartTimeLessThan(otherStartTime string) (bool, error) {
	ts1, err1 := utils.StrTime2Int(this.StartTime)
	ts2, err2 := utils.StrTime2Int(otherStartTime)
	if err1 == nil && err2 == nil {
		return ts1 < ts2, nil
	} else if err1 == nil && err2 != nil {
		return true, nil
	} else if err1 != nil && err2 == nil {
		return false, nil
	}

	return false, fmt.Errorf("启动配置中的(开始时间)比较出错.. %s. %s", err1.Error(), err2.Error())
}

// 结束时间大于其他位点
func (this *CreateConfig) EndTimeRatherThan(otherEndTime string) (bool, error) {
	ts1, err1 := utils.StrTime2Int(this.EndTime)
	ts2, err2 := utils.StrTime2Int(otherEndTime)
	if err1 == nil && err2 == nil {
		return ts1 > ts2, nil
	} else if err1 == nil && err2 != nil {
		return true, nil
	} else if err1 != nil && err2 == nil {
		return false, nil
	}

	return false, fmt.Errorf("启动配置中的(结束时间)比较出错. %s. %s", err1.Error(), err2.Error())
}

func (this *CreateConfig) StartInfoString() string {
	return fmt.Sprintf("开始位点: %s:%d. 开始时间: %s", this.StartLogFile, this.StartLogPos, this.StartTime)
}

func (this *CreateConfig) EndInfoString() string {
	return fmt.Sprintf("结束位点: %s:%d. 结束时间: %s", this.EndLogFile, this.EndLogPos, this.EndTime)
}
