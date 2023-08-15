package offline_stat

import "fmt"

type DmlStat struct {
	InsertCount int `json:"insert_count"`
	UpdateCount int `json:"update_count"`
	DeleteCount int `json:"delete_count"`
}

type BinlogFile struct {
	StartLogFile string `json:"start_log_file"`
	StartLogPos  uint32 `json:"start_log_pos"`
}

func (this *BinlogFile) FilePos() string {
	return fmt.Sprintf("%v:%v", this.StartLogFile, this.StartLogPos)
}

func (this *DmlStat) DmlCount() int {
	return this.InsertCount + this.UpdateCount + this.DeleteCount
}

type TableBinlogStat struct {
	DmlStat
	SchemaName  string `json:"schema_name"`
	TableName   string `json:"table_name"`
	AppearCount int64  `json:"appear_count"`
}
