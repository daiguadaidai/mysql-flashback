package config

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/utils"
	"strings"
	"time"
)

type OfflineConfig struct {
	BinlogFiles          []string
	ThreadID             uint32
	Now                  time.Time
	EnableRollbackUpdate bool
	EnableRollbackInsert bool
	EnableRollbackDelete bool
	SaveDir              string
	SchemaFile           string   // 建表语句
	MatchSqls            []string // 使用sql语句来匹配需要查询的时间段或者
}

func NewOffileConfig() *OfflineConfig {
	return &OfflineConfig{
		Now: time.Now(),
	}
}

// 设置最终的保存文件
func (this *OfflineConfig) GetSaveDir() string {
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

func (this *OfflineConfig) Check() error {
	if err := this.checkCondition(); err != nil {
		return err
	}

	if err := utils.CheckAndCreatePath(this.GetSaveDir(), "回滚文件存放路径"); err != nil {
		return err
	}

	return nil
}

func (this *OfflineConfig) checkCondition() error {
	if len(this.BinlogFiles) == 0 {
		return fmt.Errorf("请输入离线 binlog 文件名以及路径")
	}

	for _, fileName := range this.BinlogFiles {
		ok, err := utils.PathExists(fileName)
		if err != nil {
			return fmt.Errorf("检测离线 binlog 文件是否存在出错, %v", err)
		}
		if !ok {
			return fmt.Errorf("离线 binlog 文件不存在, %v", fileName)
		}
	}

	if strings.TrimSpace(this.SchemaFile) == "" {
		return fmt.Errorf("请指定相关表结构文件")
	}

	ok, err := utils.PathExists(this.SchemaFile)
	if err != nil {
		return fmt.Errorf("检测离线 表结构文件 是否存在出错, %v", err)
	}
	if !ok {
		return fmt.Errorf("表结构文件不存在, %v", this.SchemaFile)
	}

	return nil
}
