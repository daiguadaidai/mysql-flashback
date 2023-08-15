package offline_stat

type TimestampBinlogStat struct {
	DmlStat
	BinlogFile
	Timestamp uint32 `json:"timestamp"`
	TxCount   int64  `json:"tx_count"`
}

func NewTimestampBinlogStat(timestamp uint32, startLogFile string, startLogPos uint32) *TimestampBinlogStat {
	return &TimestampBinlogStat{
		Timestamp: timestamp,
		BinlogFile: BinlogFile{
			StartLogFile: startLogFile,
			StartLogPos:  startLogPos,
		},
	}
}
