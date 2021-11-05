package config

import (
	"fmt"
	"strings"
)

const (
	EXECUTE_PARALLER = 1
)

var ec *ExecuteConfig

type ExecuteConfig struct {
	FilePath string
	Paraller int64
}

func SetExecuteConfig(cfg *ExecuteConfig) {
	ec = cfg
}

func (this *ExecuteConfig) Check() error {
	if err := this.checkCondition(); err != nil {
		return err
	}
	return nil
}

func (this *ExecuteConfig) checkCondition() error {
	if this.Paraller < 1 {
		this.Paraller = EXECUTE_PARALLER
	}

	if len(strings.TrimSpace(this.FilePath)) == 0 {
		return fmt.Errorf("请指定需要执行的文件")
	}

	return nil
}
