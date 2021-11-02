package create

import (
	"context"
	"fmt"
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/config"
	"github.com/daiguadaidai/mysql-flashback/dao"
	"github.com/daiguadaidai/mysql-flashback/models"
	"github.com/daiguadaidai/mysql-flashback/utils"
	"github.com/daiguadaidai/mysql-flashback/visitor"
	"github.com/siddontang/go-mysql/mysql"
	"github.com/siddontang/go-mysql/replication"
	"strings"
	"time"
)

// 获取开始的位点信息
func GetStartPosition(cc *config.CreateConfig, dbc *config.DBConfig) (*models.Position, error) {
	if cc.HaveStartPosInfo() { // 有设置开始位点信息
		startPos := getPositionByPosInfo(cc.StartLogFile, cc.StartLogPos)
		// 检测开始位点是否在系统保留的binlog范围内
		if err := checkStartPosInRange(startPos); err != nil {
			return nil, err
		}
		return startPos, nil
	}

	if cc.HaveStartTime() { // 有设置开始时间
		ts, err := utils.StrTime2Int(cc.StartTime)
		if err != nil {
			return nil, err
		}
		return getStartFileByTime(uint32(ts), dbc)
	}

	return nil, nil
}

// 通过位点信息
func getPositionByPosInfo(logFile string, logPos uint64) *models.Position {
	return &models.Position{
		File:     logFile,
		Position: logPos,
	}
}

// 检测开始位点是否在系统保留的binlog范围内
func checkStartPosInRange(startPos *models.Position) error {
	// 获取 最老和最新的位点信息
	oldestPos, newestPos, err := dao.NewDefaultDao().GetOldestAndNewestPos()
	if err != nil {
		return err
	}

	if startPos.LessThan(oldestPos) {
		return fmt.Errorf("指定的开始位点 %s:%d 太过久远. 存在最老的binlog为: %s",
			startPos.File, startPos.Position, oldestPos.File)
	}

	if newestPos.LessThan(startPos) {
		return fmt.Errorf("指定的开始位点 %s:%d 还没有生成. 存在最新的binlog为: %s:%d",
			startPos.File, startPos.Position, newestPos.File, newestPos.Position)
	}

	return nil
}

// 通过开始时间获取位点信息
func getStartFileByTime(ts uint32, dbc *config.DBConfig) (*models.Position, error) {
	// 如果指定时间大于但前 返回错误.
	nts := utils.NowTimestamp() // 但前时间戳
	if int64(ts) > nts {
		return nil, fmt.Errorf("指定开始时间大于当前时间. start:%v. now:%v",
			utils.TS2String(int64(ts), utils.TIME_FORMAT), utils.TS2String(nts, utils.TIME_FORMAT))
	}

	// 获取所有的binlog
	bLogs, err := dao.NewDefaultDao().ShowBinaryLogs()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	if len(bLogs) < 1 {
		return nil, fmt.Errorf("show binary logs 没有数. 请检查是否开始起binlog")
	}
	lastBLOG := bLogs[len(bLogs)-1] // 获取最后一个binlog

	var preBlog *models.BinaryLog
	for _, bLog := range bLogs {
		eventTS, err := getSecondEventTimeBySyncer(bLog.LogName, 0, dbc)
		if err != nil {
			seelog.Error(err.Error())
			continue
		}

		// 比较binlog event timestamp 是否大于指定的, 如果大于指定的就
		if eventTS > ts {
			if preBlog == nil {
				return nil, fmt.Errorf("指定的开始事件过于久远, 已经找不到对应的binlog. "+
					"您可以指定开始时间: %v", utils.TS2String(int64(eventTS), utils.TIME_FORMAT))
			}
			return getPositionByPosInfo(preBlog.LogName, 0), nil
		}
		preBlog = bLog
	}

	// 获取了每个binarylog的开始event timestamp都没有大于指定的开始时间
	return getPositionByPosInfo(lastBLOG.LogName, 0), nil
}

// 通过 syncer 获取每个日志的第一个 event 事件
func getSecondEventTimeBySyncer(
	logFile string,
	logPos uint32,
	dbc *config.DBConfig,
) (uint32, error) {
	cfg := dbc.GetSyncerConfig()
	syncer := replication.NewBinlogSyncer(cfg)
	defer syncer.Close()

	streamer, err := syncer.StartSync(mysql.Position{logFile, logPos})
	if err != nil {
		return 0, err
	}
	for { // 遍历event获取第二个可用的时间戳
		ev, err := streamer.GetEvent(context.Background())
		if err != nil {
			return 0, fmt.Errorf("获取日志第一个事件出错 logFile:%v, logPos:%v. %v",
				logFile, logPos, err)
		}

		if ev.Header.Timestamp != 0 {
			return ev.Header.Timestamp, nil
		}
	}

	return 0, fmt.Errorf("没有获取到可用的event时间点")
}

func GetAndGeneraLastEvent() (*models.Position, error) {
	// 连续获取两次 show master status, 如果两次查询没有变化则自己对数据进行删除一个不存在的表,
	// 让数据库添加一个binlog. 为了能正常使用 sync获取最后一个位点
	defaultDao := dao.NewDefaultDao()
	pos1, err := defaultDao.ShowMasterStatus()
	if err != nil {
		return nil, err
	}
	time.Sleep(time.Second)
	pos2, err := defaultDao.ShowMasterStatus()
	if err != nil {
		return nil, err
	}
	if !pos1.Equal(pos2) {
		return pos1, nil
	}

	if err = defaultDao.DropNotExistsTable(); err != nil {
		return nil, fmt.Errorf("删除不存在的表.生成binlog event失败. %v")
	}
	return pos1, nil
}

type RollbackType int8

var (
	RollbackNone         RollbackType = 0
	RollbackAllTable     RollbackType = 10
	RollbackPartialTable RollbackType = 20
)

/* 获取需要回滚的表
Return:
[
	{
        cchema: 数据库名,
        table: 表名
    },
    ......
]
*/
func FindRollbackTables(cc *config.CreateConfig, mTables []*visitor.MatchTable) ([]*models.DBTable, RollbackType, error) {
	rollbackTables := make([]*models.DBTable, 0, 1)

	// 没有指定表, 说明使用所有的表
	if len(cc.RollbackSchemas) == 0 && len(cc.RollbackTables) == 0 && len(mTables) == 0 {
		return rollbackTables, RollbackAllTable, nil
	}

	notAllTableSchema := make(map[string]bool) // 保存不需要回滚所有表的 schema
	for _, table := range cc.RollbackTables {
		items := strings.Split(table, ".")
		switch len(items) {
		case 1: // table. 没有指定schema, 只指定了table
			if len(cc.RollbackSchemas) == 0 { // 该表没有指定库
				return nil, RollbackNone, fmt.Errorf("表:%v. 没有指定库", table)
			}
			// 如果有单独指定schema(--rollback-schema=db1), 代表需要回滚的表为 db1.t_n, 并且这个schema则标记为非所有表都需要回滚的库
			for _, cchema := range cc.RollbackSchemas {
				if _, ok := notAllTableSchema[cchema]; !ok {
					notAllTableSchema[cchema] = true
				}

				t := models.NewDBTable(cchema, table)
				rollbackTables = append(rollbackTables, t)
			}
		case 2: // schema.table 的格式, 代表有指定schema 和 table, 并标记 schema 不能进行回滚库中所有表
			if _, ok := notAllTableSchema[items[0]]; !ok {
				notAllTableSchema[items[0]] = true
			}
			t := models.NewDBTable(items[0], items[1])
			rollbackTables = append(rollbackTables, t)
		default:
			return nil, RollbackNone, fmt.Errorf("不能识别需要rollback的表: %v", table)
		}
	}

	// 从 match-sql 中解析需要回滚的表
	for _, mTable := range mTables {
		if _, ok := notAllTableSchema[mTable.SchemaName]; !ok {
			notAllTableSchema[mTable.SchemaName] = true
		}
		t := models.NewDBTable(mTable.SchemaName, mTable.TableName)
		rollbackTables = append(rollbackTables, t)
	}

	// 要是指定的schema, 不存在于 notAllTableSchema 这个变量中, 说明这个schema中的表都需要回滚
	for _, cchema := range cc.RollbackSchemas {
		if _, ok := notAllTableSchema[cchema]; ok {
			continue
		}
		notAllTableSchema[cchema] = true

		tables, err := dao.NewDefaultDao().FindTablesBySchema(cchema)
		if err != nil {
			return nil, RollbackNone, fmt.Errorf("获取数据库[%s]下面的所有表失败. %v", cchema, err)
		}
		rollbackTables = append(rollbackTables, tables...)
	}

	if len(rollbackTables) == 0 {
		return rollbackTables, RollbackAllTable, nil
	}

	return rollbackTables, RollbackPartialTable, nil
}

// 重新设置开始的和结束位点信息
func resetPosInfo(cc *config.CreateConfig, mTables []*visitor.MatchTable) {
	// 设置最终使用的开始位点信息
	minMTable := getMinMTable(mTables)
	if minMTable != nil {
		resetConfigStartPosInfo(cc, minMTable)
	}
	seelog.Infof("任务最终(开始)信息为: %s", cc.StartInfoString())

	// 设置最终使用的结束位点信息
	maxMTable := getMaxMTable(mTables)
	if maxMTable != nil {
		resetConfigEndPosInfo(cc, maxMTable)
	}
	seelog.Infof("任务最终(结束)信息为: %s", cc.EndInfoString())
}

// 获取开始位点信息最小的 MatchTable
func getMinMTable(mTables []*visitor.MatchTable) *visitor.MatchTable {
	var min *visitor.MatchTable
	var err error
	for _, mTable := range mTables {
		if min == nil {
			min = mTable
			continue
		}

		min, err = mTableCompareGetMin(min, mTable)
		if err != nil {
			seelog.Warnf(err.Error())
		}
	}
	return min
}

// 比较matchtable 开始位点信息, 并获取最小的
func mTableCompareGetMin(first *visitor.MatchTable, second *visitor.MatchTable) (*visitor.MatchTable, error) {
	if first.HaveStartPosInfo() && second.HaveStartPosInfo() { // 两个都有填写位点信息
		if first.StartPosInfoLessThan(second) {
			return first, nil
		} else {
			return second, nil
		}
	} else if first.HaveStartPosInfo() && second.HaveStartTime() {
		return first, nil
	} else if first.HaveStartTime() && second.HaveStartPosInfo() {
		return second, nil
	} else if first.HaveStartTime() && second.HaveStartTime() {
		if ok, err := first.StartTimeLessThan(second); err != nil {
			return nil, err
		} else {
			if ok { // first 开始时间小于 second 开始时间
				return first, nil
			} else {
				return second, nil
			}
		}
	} else if first.HaveStartTime() {
		return first, nil
	} else if second.HaveStartTime() {
		return second, nil
	}

	return nil, fmt.Errorf("match table 没有可以使用的(开始)位点信息. %s <=> %s", first.Table(), second.Table())
}

// 获取结束位点信息最大的 MatchTable
func getMaxMTable(mTables []*visitor.MatchTable) *visitor.MatchTable {
	var max *visitor.MatchTable
	var err error
	for _, mTable := range mTables {
		if max == nil {
			max = mTable
			continue
		}

		max, err = mTableCompareGetMax(max, mTable)
		if err != nil {
			seelog.Warnf(err.Error())
		}
	}
	return max
}

// 比较matchtable 结束位点信息, 并获取最大的
func mTableCompareGetMax(first *visitor.MatchTable, second *visitor.MatchTable) (*visitor.MatchTable, error) {
	if first.HaveEndPosInfo() && second.HaveEndPosInfo() { // 两个都有填写位点信息
		if first.EndPostInfoRatherThan(second) {
			return first, nil
		} else {
			return second, nil
		}
	} else if first.HaveEndPosInfo() && second.HaveStartTime() {
		return first, nil
	} else if first.HaveStartTime() && second.HaveEndPosInfo() {
		return second, nil
	} else if first.HaveEndTime() && second.HaveEndTime() {
		if ok, err := first.EndTimeRatherThan(second); err != nil {
			return nil, err
		} else {
			if ok { // first 结束时间 大于 second结束时间
				return first, nil
			} else {
				return second, nil
			}
		}
	} else if first.HaveStartTime() {
		return first, nil
	} else if second.HaveStartTime() {
		return second, nil
	}

	return nil, fmt.Errorf("match table 没有可以使用的(结束)位点信息. %s <=> %s", first.Table(), second.Table())
}

// 设置任务最终开始位点信息
func resetConfigStartPosInfo(cc *config.CreateConfig, mTable *visitor.MatchTable) {
	if cc.HaveStartPosInfo() && mTable.HaveStartPosInfo() {
		if !cc.StartPosInfoLessThan(mTable.StartLogFile, mTable.StartLogPos) {
			cc.StartLogFile = mTable.StartLogFile
			cc.StartLogPos = mTable.StartLogPos
		}
	} else if cc.HaveStartTime() && mTable.HaveStartPosInfo() {
		cc.StartLogFile = mTable.StartLogFile
		cc.StartLogPos = mTable.StartLogPos
	} else if cc.HaveStartTime() && mTable.HaveStartTime() {
		if ok, err := cc.StartTimeLessThan(mTable.StartRollBackTime); err == nil {
			if !ok {
				cc.StartTime = mTable.StartRollBackTime
			}
		}
	} else if mTable.HaveStartPosInfo() {
		cc.StartLogFile = mTable.StartLogFile
		cc.StartLogPos = mTable.StartLogPos
	} else if mTable.HaveStartTime() {
		cc.StartTime = mTable.StartRollBackTime
	}
}

// 设置任务最终结束位点信息
func resetConfigEndPosInfo(cc *config.CreateConfig, mTable *visitor.MatchTable) {
	if cc.HaveEndPosInfo() && mTable.HaveEndPosInfo() {
		if !cc.EndPostInfoRatherThan(mTable.EndLogFile, mTable.EndLogPos) {
			cc.EndLogFile = mTable.EndLogFile
			cc.EndLogPos = mTable.EndLogPos
		}
	} else if cc.HaveEndTime() && mTable.HaveEndPosInfo() {
		cc.EndLogFile = mTable.EndLogFile
		cc.EndLogPos = mTable.EndLogPos
	} else if cc.HaveEndTime() && mTable.HaveEndTime() {
		if ok, err := cc.EndTimeRatherThan(mTable.EndRollBackTime); err == nil {
			if !ok {
				cc.EndTime = mTable.EndRollBackTime
			}
		}
	} else if mTable.HaveEndPosInfo() {
		cc.EndLogFile = mTable.EndLogFile
		cc.EndLogPos = mTable.EndLogPos
	} else if mTable.HaveEndTime() {
		cc.EndTime = mTable.EndRollBackTime
	}
}

// 重置threadId
func resetThreadId(cc *config.CreateConfig, mTables []*visitor.MatchTable) {
	if cc.ThreadID > 0 { // 有显示使用命令参数设置则使用命令参数指定的
		return
	}

	// 没有显示指定 thread id 则使用 sql 里面的
	for _, mTable := range mTables {
		if mTable.ThreadId > 0 {
			cc.ThreadID = mTable.ThreadId
		}
	}
}
