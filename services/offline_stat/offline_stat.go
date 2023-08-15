package offline_stat

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/config"
	"github.com/daiguadaidai/mysql-flashback/utils"
	"github.com/siddontang/go-mysql/replication"
	"os"
	"sort"
)

const (
	QueryEventBegin = "BEGIN"
)

type OfflineStat struct {
	OfflineStatCfg         *config.OfflineStatConfig
	TotalTableStatMap      map[string]*TableBinlogStat
	TotalThreadStatMap     map[uint32]*ThreadBinlogStat
	TotalTransactionStats  []*TransactionBinlogStat
	TotalTimestampStats    []*TimestampBinlogStat
	CurrentTimestampStat   *TimestampBinlogStat
	CurrentTransactionStat *TransactionBinlogStat
	CurrentTreadId         uint32
	CurrentXid             uint64
	CurrentTimestamp       uint32
	CurrentSchemaName      string
	CurrentTableName       string
	CurrentLogFile         string
	CurrentLogPos          uint32
}

func NewOfflineStat(cfg *config.OfflineStatConfig) *OfflineStat {
	return &OfflineStat{
		OfflineStatCfg:        cfg,
		TotalTableStatMap:     make(map[string]*TableBinlogStat),
		TotalThreadStatMap:    make(map[uint32]*ThreadBinlogStat),
		TotalTransactionStats: make([]*TransactionBinlogStat, 0, 1000),
		TotalTimestampStats:   make([]*TimestampBinlogStat, 0, 1000),
	}
}

func (this *OfflineStat) Start() error {
	for i, binlogFile := range this.OfflineStatCfg.BinlogFiles {
		this.CurrentLogFile = binlogFile
		seelog.Infof("开始解析Binlog: %v/%v, binlog文件: %v", i+i, len(this.OfflineStatCfg.BinlogFiles), binlogFile)

		// 创建一个 BinlogParser 对象
		parser := replication.NewBinlogParser()
		if err := parser.ParseFile(binlogFile, 0, func(event *replication.BinlogEvent) error {
			return this.handleEvent(event)
		}); err != nil {
			return fmt.Errorf("解析binlog出错. 进度: %v/%v, binlog文件: %v. %v", i+1, len(this.OfflineStatCfg.BinlogFiles), binlogFile, err)
		}
	}

	// 将统计信息输出到文件中
	this.statToFile()

	return nil
}

// 处理binlog事件
func (this *OfflineStat) handleEvent(ev *replication.BinlogEvent) error {
	switch e := ev.Event.(type) {
	case *replication.XIDEvent:
		this.handleXIDEvent(e)
	case *replication.QueryEvent:
		this.handleQueryEvent(e, ev)
	case *replication.TableMapEvent:
		this.handleTableMapEvent(e, ev)
	case *replication.RowsEvent:
		this.handleRowEvent(e, ev)
	}

	return nil
}

func (this *OfflineStat) handleXIDEvent(e *replication.XIDEvent) {
	if this.CurrentTransactionStat == nil {
		return
	}

	// 添加事务统计
	this.CurrentTransactionStat.Xid = e.XID
	this.TotalTransactionStats = append(this.TotalTransactionStats, this.CurrentTransactionStat)

	this.CurrentTransactionStat = nil
}

func (this *OfflineStat) handleQueryEvent(e *replication.QueryEvent, ev *replication.BinlogEvent) {
	this.CurrentTreadId = e.SlaveProxyID

	// 遇到 BEGIN
	if QueryEventBegin == string(e.Query) {
		// 添加和初始化时间统计
		if ev.Header.Timestamp != this.CurrentTimestamp {
			this.CurrentTimestampStat = NewTimestampBinlogStat(ev.Header.Timestamp, this.CurrentLogFile, ev.Header.LogPos)
			// 添加时间统计
			if this.CurrentTimestampStat != nil {
				this.TotalTimestampStats = append(this.TotalTimestampStats, this.CurrentTimestampStat)
			}
		}
		this.CurrentTimestampStat.TxCount += 1

		// 初始化事务统计
		this.CurrentTransactionStat = NewTransactionBinlogStat(ev.Header.Timestamp, this.CurrentLogFile, ev.Header.LogPos)

		// 初始化 threadId
		threadStat, ok := this.TotalThreadStatMap[e.SlaveProxyID]
		if !ok {
			threadStat = &ThreadBinlogStat{
				ThreadId: e.SlaveProxyID,
			}
			this.TotalThreadStatMap[e.SlaveProxyID] = threadStat
		}
		threadStat.AppearCount += 1
	}
}

func (this *OfflineStat) handleTableMapEvent(e *replication.TableMapEvent, ev *replication.BinlogEvent) {
	this.CurrentSchemaName = string(e.Schema)
	this.CurrentTableName = string(e.Table)
	table := fmt.Sprintf("%v.%v", this.CurrentSchemaName, this.CurrentTableName)

	tableStat, ok := this.TotalTableStatMap[table]
	if !ok {
		tableStat = &TableBinlogStat{
			SchemaName: this.CurrentSchemaName,
			TableName:  this.CurrentTableName,
		}

		this.TotalTableStatMap[table] = tableStat
	}

	// 统计表出现次数
	tableStat.AppearCount += 1
}

func (this *OfflineStat) handleRowEvent(e *replication.RowsEvent, ev *replication.BinlogEvent) {
	table := fmt.Sprintf("%v.%v", this.CurrentSchemaName, this.CurrentTableName)

	switch ev.Header.EventType {
	case replication.WRITE_ROWS_EVENTv0, replication.WRITE_ROWS_EVENTv1, replication.WRITE_ROWS_EVENTv2:
		// 表统计
		tableStat, ok := this.TotalTableStatMap[table]
		if ok {
			tableStat.InsertCount += len(e.Rows)
		}

		// 时间统计
		if this.CurrentTimestampStat != nil {
			this.CurrentTimestampStat.InsertCount += len(e.Rows)
		}

		// 事务统计
		if this.CurrentTransactionStat != nil {
			this.CurrentTransactionStat.InsertCount += len(e.Rows)
		}

		// Thread 统计
		threadStat, ok := this.TotalThreadStatMap[this.CurrentTreadId]
		if ok {
			threadStat.InsertCount += len(e.Rows)
		}
	case replication.UPDATE_ROWS_EVENTv0, replication.UPDATE_ROWS_EVENTv1, replication.UPDATE_ROWS_EVENTv2:
		// 表统计
		tableStat, ok := this.TotalTableStatMap[table]
		if ok {
			tableStat.UpdateCount += len(e.Rows) / 2
		}

		// 时间统计
		if this.CurrentTimestampStat != nil {
			this.CurrentTimestampStat.UpdateCount += len(e.Rows) / 2
		}

		// 事务统计
		if this.CurrentTransactionStat != nil {
			this.CurrentTransactionStat.UpdateCount += len(e.Rows) / 2
		}

		// Thread 统计
		threadStat, ok := this.TotalThreadStatMap[this.CurrentTreadId]
		if ok {
			threadStat.UpdateCount += len(e.Rows) / 2
		}
	case replication.DELETE_ROWS_EVENTv0, replication.DELETE_ROWS_EVENTv1, replication.DELETE_ROWS_EVENTv2:
		// 表统计
		tableStat, ok := this.TotalTableStatMap[table]
		if ok {
			tableStat.DeleteCount += len(e.Rows)
		}

		// 时间统计
		if this.CurrentTimestampStat != nil {
			this.CurrentTimestampStat.DeleteCount += len(e.Rows)
		}

		// 事务统计
		if this.CurrentTransactionStat != nil {
			this.CurrentTransactionStat.DeleteCount += len(e.Rows)
		}

		// Thread 统计
		threadStat, ok := this.TotalThreadStatMap[this.CurrentTreadId]
		if ok {
			threadStat.DeleteCount += len(e.Rows)
		}
	}
}

// 统计信息到文件中
func (this *OfflineStat) statToFile() {
	// 表统计
	this.tableStatToFile()

	// thread统计
	this.threadStatToFile()

	// 时间统计
	this.TimestampStatToFile()

	// 事务统计
	this.XidStatToFile()
}

// 表统计信息写入到文件中
func (this *OfflineStat) tableStatToFile() {
	stats := make([]*TableBinlogStat, 0, len(this.TotalTableStatMap))
	for _, stat := range this.TotalTableStatMap {
		stats = append(stats, stat)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].DmlCount() > stats[j].DmlCount()
	})

	filename := this.OfflineStatCfg.TableStatFilePath()
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		seelog.Errorf("将(表)统计信息写入文件出错. 打开文件出错. 文件: %v. %v", filename, err)
		return
	}
	defer f.Close()

	for _, stat := range stats {
		if _, err := f.WriteString(fmt.Sprintf("表: %v.%v \tdml影响行数: %v, insert: %v, update: %v, delete: %v, 表出现次数: %v\n",
			stat.SchemaName, stat.TableName, stat.DmlCount(), stat.InsertCount, stat.UpdateCount, stat.DeleteCount, stat.AppearCount)); err != nil {
			seelog.Errorf("写入(表)统计信息出错. 文件: %v. 表: %v.%v, %v", filename, stat.SchemaName, stat.TableName, err)
			return
		}
	}
}

func (this *OfflineStat) threadStatToFile() {
	stats := make([]*ThreadBinlogStat, 0, len(this.TotalTableStatMap))
	for _, stat := range this.TotalThreadStatMap {
		stats = append(stats, stat)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].DmlCount() > stats[j].DmlCount()
	})

	filename := this.OfflineStatCfg.ThreadStatFilePath()
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		seelog.Errorf("将(thread)统计信息写入文件出错. 打开文件出错. 文件: %v. %v", filename, err)
		return
	}
	defer f.Close()

	for _, stat := range stats {
		if _, err := f.WriteString(fmt.Sprintf("threadId: %v\tdml影响行数: %v, insert: %v, update: %v, delete: %v, 表出现次数: %v\n",
			stat.ThreadId, stat.DmlCount(), stat.InsertCount, stat.UpdateCount, stat.DeleteCount, stat.AppearCount)); err != nil {
			seelog.Errorf("写入(thread)统计信息出错. 文件: %v. ThreadId: %v, %v", filename, stat.ThreadId, err)
			return
		}
	}
}

func (this *OfflineStat) TimestampStatToFile() {
	filename := this.OfflineStatCfg.TimestampStatFilePath()
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		seelog.Errorf("将(时间)统计信息写入文件出错. 打开文件出错. 文件: %v. %v", filename, err)
		return
	}
	defer f.Close()

	for _, stat := range this.TotalTimestampStats {
		if _, err := f.WriteString(fmt.Sprintf("%v: dml影响行数: %v, insert: %v, update: %v, delete: %v, 事务数: %v, 开始位点: %v\n",
			utils.TS2String(int64(stat.Timestamp), utils.TIME_FORMAT), stat.DmlCount(), stat.InsertCount, stat.UpdateCount, stat.DeleteCount, stat.TxCount, stat.FilePos())); err != nil {
			seelog.Errorf("写入(时间)统计信息出错. 文件: %v. %v. 开始位点: %v. %v", filename, utils.TS2String(int64(stat.Timestamp), utils.TIME_FORMAT), stat.FilePos(), err)
			return
		}
	}
}

func (this *OfflineStat) XidStatToFile() {
	filename := this.OfflineStatCfg.TransactionStatFilePath()
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		seelog.Errorf("将(xid)统计信息写入文件出错. 打开文件出错. 文件: %v. %v", filename, err)
		return
	}
	defer f.Close()

	for _, stat := range this.TotalTransactionStats {
		if _, err := f.WriteString(fmt.Sprintf("Xid: %v \t%v \t dml影响行数: %v, insert: %v, update: %v, delete: %v, 开始位点: %v\n",
			stat.Xid, utils.TS2String(int64(stat.Timestamp), utils.TIME_FORMAT), stat.DmlCount(), stat.InsertCount, stat.UpdateCount, stat.DeleteCount, stat.FilePos())); err != nil {
			seelog.Errorf("写入(xid)统计信息出错. 文件: %v. Xid: %v, %v. 开始位点: %v. %v", filename, stat.Xid, utils.TS2String(int64(stat.Timestamp), utils.TIME_FORMAT), stat.FilePos(), err)
			return
		}
	}
}
