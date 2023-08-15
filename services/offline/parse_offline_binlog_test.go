package offline

import (
	"github.com/siddontang/go-mysql/replication"
	"os"
	"testing"
)

func Test_ParseOfflineBinlog_01(t *testing.T) {
	binlogFileName := "/Users/hh/Desktop/mysql-bin.000200"

	// 创建一个 BinlogParser 对象
	parser := replication.NewBinlogParser()

	err := parser.ParseFile(binlogFileName, 0, func(event *replication.BinlogEvent) error {
		event.Dump(os.Stdout)

		return nil
	})

	if err != nil {
		t.Fatal(err.Error())
	}
}
