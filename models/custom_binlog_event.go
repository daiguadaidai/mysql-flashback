package models

import "github.com/siddontang/go-mysql/replication"

type CustomBinlogEvent struct {
	Event    *replication.BinlogEvent
	ThreadId uint32
}
