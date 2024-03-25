package config

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/daiguadaidai/mysql-flashback/utils"
	"github.com/daiguadaidai/peep"
	"github.com/go-mysql-org/go-mysql/replication"
	"strings"
	"sync"
)

const (
	DB_HOST                = "127.0.0.1"
	DB_PORT                = 3306
	DB_USERNAME            = "root"
	DB_PASSWORD            = "root"
	DB_SCHEMA              = ""
	DB_AUTO_COMMIT         = true
	DB_MAX_OPEN_CONNS      = 8
	DB_MAX_IDEL_CONNS      = 8
	DB_CHARSET             = "utf8mb4"
	DB_TIMEOUT             = 10
	DB_PASSWORD_IS_DECRYPT = true
	SQL_LOG_BIN            = true
)

var dbConfig *DBConfig

type DBConfig struct {
	Username          string
	Password          string
	Database          string
	CharSet           string
	Host              string
	Timeout           int
	Port              int
	MaxOpenConns      int
	MaxIdelConns      int
	AllowOldPasswords int
	AutoCommit        bool
	PasswordIsDecrypt bool
	SqlLogBin         bool
}

func (this *DBConfig) GetDataSource() string {
	var dataSource string

	if this.SqlLogBin {
		dataSource = fmt.Sprintf(
			"%v:%v@tcp(%v:%v)/%v?charset=%v&allowOldPasswords=%v&timeout=%vs&autocommit=%v&parseTime=True&loc=Local",
			this.Username,
			this.GetPassword(),
			this.Host,
			this.Port,
			this.Database,
			this.CharSet,
			this.AllowOldPasswords,
			this.Timeout,
			this.AutoCommit,
		)
	} else {
		dataSource = fmt.Sprintf(
			"%v:%v@tcp(%v:%v)/%v?charset=%v&allowOldPasswords=%v&timeout=%vs&autocommit=%v&parseTime=True&loc=Local&sql_log_bin=%v",
			this.Username,
			this.GetPassword(),
			this.Host,
			this.Port,
			this.Database,
			this.CharSet,
			this.AllowOldPasswords,
			this.Timeout,
			this.AutoCommit,
			this.SqlLogBin,
		)
	}

	return dataSource
}

func (this *DBConfig) Check() error {
	if strings.TrimSpace(this.Database) == "" {
		return fmt.Errorf("数据库不能为空")
	}

	return nil
}

// 设置 DBConfig
func SetDBConfig(dbc *DBConfig) {
	dbConfig = dbc
}

func GetDBConfig() *DBConfig {
	return dbConfig
}

func (this *DBConfig) GetSyncerConfig() replication.BinlogSyncerConfig {
	return replication.BinlogSyncerConfig{
		ServerID: utils.RandRangeUint32(100000000, 200000000),
		Flavor:   "mysql",
		Host:     this.Host,
		Port:     uint16(this.Port),
		User:     this.Username,
		Password: this.GetPassword(),
	}
}

// 获取密码, 有判断是否需要进行解密
func (this *DBConfig) GetPassword() string {
	if this.PasswordIsDecrypt {
		pwd, err := peep.Decrypt(this.Password)
		if err != nil {
			seelog.Warnf("密码解密出错, 将使用未解析串作为密码. %s", err.Error())
			return this.Password
		}
		return pwd
	}
	return this.Password
}

var configMap sync.Map
