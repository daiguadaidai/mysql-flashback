package execute

import (
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/config"
	"syscall"
)

func Start(ec *config.ExecuteConfig, dbc *config.DBConfig) {
	defer seelog.Flush()
	logger, _ := seelog.LoggerFromConfigAsBytes([]byte(config.LogDefautConfig()))
	seelog.ReplaceLogger(logger)

	seelog.Infof("回滚开始")

	if err := checkConfig(ec, dbc); err != nil {
		seelog.Error(err.Error())
		syscall.Exit(1)
	}

	config.SetExecuteConfig(ec)
	config.SetDBConfig(dbc)

	executor := NewExecutor(ec, dbc)
	if err := executor.Start(); err != nil {
		seelog.Error(err.Error())
		syscall.Exit(1)
	}
	if !executor.EmitSuccess || !executor.ExecSuccess {
		seelog.Errorf("回滚未执行成功. 执行了 %d 条", executor.ExecCount)
		syscall.Exit(1)
	}

	seelog.Infof("回滚执行成功. 执行行数: %d", executor.ExecCount)
}

func checkConfig(ec *config.ExecuteConfig, dbc *config.DBConfig) error {
	// 检测执行子命令配置文件, 设置执行的类型
	if err := ec.Check(); err != nil {
		return err
	}

	return nil
}
