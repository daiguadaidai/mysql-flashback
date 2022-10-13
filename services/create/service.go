package create

import (
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/config"
	"github.com/daiguadaidai/mysql-flashback/visitor"
	"strings"
	"syscall"
)

func Start(cc *config.CreateConfig, dbc *config.DBConfig) {
	defer seelog.Flush()
	logger, _ := seelog.LoggerFromConfigAsBytes([]byte(config.LogDefautConfig()))
	seelog.ReplaceLogger(logger)

	config.SetStartConfig(cc)
	config.SetDBConfig(dbc)

	// 解析sql并且获取
	var mTables []*visitor.MatchTable
	if len(cc.MatchSqls) > 0 {
		for _, matchSql := range cc.MatchSqls {
			if strings.TrimSpace(matchSql) == "" {
				continue
			}

			tmpMTables, err := visitor.GetMatchTables(matchSql)
			if err != nil {
				seelog.Errorf(err.Error())
				syscall.Exit(1)
			}

			mTables = append(mTables, tmpMTables...)
		}
		// 重置 位点信息 信息
		resetPosInfo(cc, mTables)
		// 重置 thread id
		resetThreadId(cc, mTables)
	}

	// 检测启动配置信息是否可用
	if err := cc.Check(); err != nil {
		seelog.Error(err.Error())
		syscall.Exit(1)
	}

	flashback, err := NewFlashback(cc, dbc, mTables)
	if err != nil {
		seelog.Error(err.Error())
		syscall.Exit(1)
	}

	if err = flashback.Start(); err != nil {
		seelog.Errorf("生成回滚sql失败. %s", err.Error())
		syscall.Exit(1)
	}
	if !flashback.Successful {
		seelog.Error("生成回滚sql失败")
		syscall.Exit(1)
	}
	seelog.Info("生成回滚sql完成")

}
