package offline_stat

import (
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/config"
	"syscall"
)

func Start(offlineStatCfg *config.OfflineStatConfig) {
	defer seelog.Flush()
	logger, _ := seelog.LoggerFromConfigAsBytes([]byte(config.LogDefautConfig()))
	seelog.ReplaceLogger(logger)

	// 检测启动配置信息是否可用
	if err := offlineStatCfg.Check(); err != nil {
		seelog.Error(err.Error())
		syscall.Exit(1)
	}

	offlineStat := NewOfflineStat(offlineStatCfg)
	if err := offlineStat.Start(); err != nil {
		seelog.Errorf("统计出错. %v", err)
	} else {
		seelog.Info("统计完成")
	}
}
