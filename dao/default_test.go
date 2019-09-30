package dao

import (
	"fmt"
	"github.com/daiguadaidai/mysql-flashback/config"
	"testing"
)

func initDBConfig() {
	dbConfig := &config.DBConfig{
		Host:         "10.10.10.21",
		Port:         3307,
		Username:     "HH",
		Password:     "oracle12",
		Database:     "poow",
		CharSet:      "utf8mb4",
		AutoCommit:   true,
		MaxOpenConns: 100,
		MaxIdelConns: 100,
		Timeout:      10,
	}

	config.SetDBConfig(dbConfig)
}

func TestDefaultDao_QueryBinaryLogs(t *testing.T) {
	initDBConfig()

	logs, err := NewDefaultDao().ShowBinaryLogs()
	if err != nil {
		t.Fatal(err.Error())
	}
	for _, log := range logs {
		fmt.Println(log)
	}
}

func TestDefaultDao_ShowMasterStatus(t *testing.T) {
	initDBConfig()

	pos, err := NewDefaultDao().ShowMasterStatus()
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(pos)
}
