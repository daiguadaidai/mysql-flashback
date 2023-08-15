package offline_stat

type ThreadBinlogStat struct {
	DmlStat
	ThreadId    uint32 `json:"thread_id"`
	AppearCount int64  `json:"appear_count"`
}
