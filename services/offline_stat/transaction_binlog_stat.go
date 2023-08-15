package offline_stat

type TransactionBinlogStat struct {
	DmlStat
	BinlogFile
	Xid       uint64 `json:"xid"`
	Timestamp uint32 `json:"timestamp"`
}

func NewTransactionBinlogStat(timestamp uint32, startLogFile string, startLogPos uint32) *TransactionBinlogStat {
	return &TransactionBinlogStat{
		Timestamp: timestamp,
		BinlogFile: BinlogFile{
			StartLogFile: startLogFile,
			StartLogPos:  startLogPos,
		},
	}
}
