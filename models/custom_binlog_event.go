package models

import "github.com/go-mysql-org/go-mysql/replication"

type CustomBinlogEvent struct {
	Event    *replication.BinlogEvent
	ThreadId uint32
}
