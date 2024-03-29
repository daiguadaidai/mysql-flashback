// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"github.com/daiguadaidai/mysql-flashback/services/offline"
	"github.com/daiguadaidai/mysql-flashback/services/offline_stat"
	"os"

	"github.com/daiguadaidai/mysql-flashback/config"
	"github.com/daiguadaidai/mysql-flashback/services/create"
	"github.com/daiguadaidai/mysql-flashback/services/execute"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mysql-flashback",
	Short: "MySQL flashback 工具",
}

// cerateCmd 是 rootCmd 的一个子命令
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "生成回滚SQL",
	Long: `生成回滚的sql. 如下:
Example:
指定 开始位点 和 结束位点
./mysql-flashback create \
    --start-log-file="mysql-bin.000090" \
    --start-log-pos=0 \
    --end-log-file="mysql-bin.000092" \
    --end-log-pos=424 \
    --thread-id=15 \
    --rollback-table="schema1.table1" \
    --rollback-table="schema1.table2" \
    --rollback-table="schema2.table1" \
    --save-dir="" \
    --db-host="127.0.0.1" \
    --db-port=3306 \
    --db-username="root" \
    --db-password="root" \
    --match-sql="select * from schema1.table1 where name = 'aa'"

指定 开始位点 和 结束时间
./mysql-flashback create \
    --start-log-file="mysql-bin.000090" \
    --start-log-pos=0 \
    --end-time="2018-12-17 15:36:58" \
    --thread-id=15 \
    --rollback-table="schema1.table1" \
    --rollback-table="schema1.table2" \
    --rollback-table="schema2.table1" \
    --save-dir="" \
    --db-host="127.0.0.1" \
    --db-port=3306 \
    --db-username="root" \
    --db-password="root" \
    --match-sql="select name, age from schema1.table1 where name = 'aa'"

指定 开始时间 和 结束时间
./mysql-flashback create \
    --start-time="2018-12-14 15:00:00" \
    --end-time="2018-12-17 15:36:58" \
    --thread-id=15 \
    --rollback-schema="schema1" \
    --rollback-table="table1" \
    --rollback-table="schema1.table2" \
    --rollback-table="schema2.table1" \
    --save-dir="" \
    --db-host="127.0.0.1" \
    --db-port=3306 \
    --db-username="root" \
    --db-password="root" \
    --match-sql="select name, age from schema1.table1 where name = 'aa' and age = 2"
`,
	Run: func(cmd *cobra.Command, args []string) {
		create.Start(cc, cdbc)
	},
}

// cerateCmd 是 rootCmd 的一个子命令
var offlineCmd = &cobra.Command{
	Use:   "offline",
	Short: "解析离线binlog, 生成回滚SQL",
	Long: `解析离线binlog, 生成回滚SQL. 如下:
Example:
./mysql-flashback offline \
    --enable-rollback-insert=true \
    --enable-rollback-update=true \
    --enable-rollback-delete=true \
    --thread-id=15 \
    --save-dir="" \
    --schema-file="" \
    --match-sql="select * from schema1.table1 where name = 'aa'" \
    --match-sql="select * from schema2.table1 where name = 'aa'" \
    --binlog-file="mysql-bin.0000001" \
    --binlog-file="mysql-bin.0000002"
`,
	Run: func(cmd *cobra.Command, args []string) {
		offline.Start(offlineCfg)
	},
}

// executeCmd 是 rootCmd 的一个子命令
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "执行sql回滚文件",
	Long: `倒序执行指定的sql回滚文件. 如下:
Example:
./mysql-flashback execute \
    --filepath="/tmp/test.sql" \
    --paraller=8 \
    --sql-log-bin=true \
    --db-host="127.0.0.1" \
    --db-port=3306 \
    --db-username="root" \
    --db-password="root"
`,
	Run: func(cmd *cobra.Command, args []string) {
		execute.Start(ec, edbc)
	},
}

// cerateCmd 是 rootCmd 的一个子命令
var offlineStatCmd = &cobra.Command{
	Use:   "offline-stat",
	Short: "解析离线binlog, 统计binlog信息",
	Long: `解析离线binlog, 统计binlog信息. 如下:
执行成功后会在当前目录生成 4 个文件
offline_stat_output/table_stat.txt # 保存表统计信息
offline_stat_output/thread_stat.txt # 保存线程统计信息
offline_stat_output/timestamp_stat.txt # 保存时间统计信息(记录的是每个事务执行 BEGIN 的时间)
offline_stat_output/xid_stat.txt # 保存 xid 统计信息

Example:
./mysql-flashback offline-stat \
    --save-dir="offline_stat_output" \
    --binlog-file="mysql-bin.0000001" \
    --binlog-file="mysql-bin.0000002"
`,
	Run: func(cmd *cobra.Command, args []string) {
		offline_stat.Start(offlineStatCfg)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	addCreateCMD()
	addExecuteCMD()
	addOfflineCMD()
	addOfflineStatCMD()
}

var cc *config.CreateConfig
var cdbc *config.DBConfig

// 添加创建回滚SQL子命令
func addCreateCMD() {
	rootCmd.AddCommand(createCmd)
	cc = config.NewStartConfig()
	createCmd.PersistentFlags().StringVar(&cc.StartLogFile, "start-log-file", "", "开始日志文件")
	createCmd.PersistentFlags().Uint64Var(&cc.StartLogPos, "start-log-pos", 0, "开始日志文件点位")
	createCmd.PersistentFlags().StringVar(&cc.EndLogFile, "end-log-file", "", "结束日志文件")
	createCmd.PersistentFlags().Uint64Var(&cc.EndLogPos, "end-log-pos", 0, "结束日志文件点位")
	createCmd.PersistentFlags().StringVar(&cc.StartTime, "start-time", "", "开始时间")
	createCmd.PersistentFlags().StringVar(&cc.EndTime, "end-time", "", "结束时间")
	createCmd.PersistentFlags().StringArrayVar(&cc.RollbackSchemas, "rollback-schema", []string{}, "指定回滚的数据库, 该命令可以指定多个")
	createCmd.PersistentFlags().StringArrayVar(&cc.RollbackTables, "rollback-table", []string{}, "需要回滚的表, 该命令可以指定多个")
	createCmd.PersistentFlags().Uint32Var(&cc.ThreadID, "thread-id", 0, "需要rollback的thread id")
	createCmd.PersistentFlags().BoolVar(&cc.EnableRollbackInsert, "enable-rollback-insert", config.ENABLE_ROLLBACK_INSERT, "是否启用回滚 insert")
	createCmd.PersistentFlags().BoolVar(&cc.EnableRollbackUpdate, "enable-rollback-update", config.ENABLE_ROLLBACK_UPDATE, "是否启用回滚 update")
	createCmd.PersistentFlags().BoolVar(&cc.EnableRollbackDelete, "enable-rollback-delete", config.ENABLE_ROLLBACK_DELETE, "是否启用回滚 delete")
	createCmd.PersistentFlags().StringVar(&cc.SaveDir, "save-dir", "", "相关文件保存的路径")
	createCmd.PersistentFlags().StringArrayVar(&cc.MatchSqls, "match-sql", []string{}, "使用简单的 SELECT 语句来匹配需要的字段和记录")

	cdbc = new(config.DBConfig)
	// 链接的数据库配置
	createCmd.PersistentFlags().StringVar(&cdbc.Host, "db-host", config.DB_HOST, "数据库host")
	createCmd.PersistentFlags().IntVar(&cdbc.Port, "db-port", config.DB_PORT, "数据库port")
	createCmd.PersistentFlags().StringVar(&cdbc.Username, "db-username", config.DB_USERNAME, "数据库用户名")
	createCmd.PersistentFlags().StringVar(&cdbc.Password, "db-password", config.DB_PASSWORD, "数据库密码")
	createCmd.PersistentFlags().StringVar(&cdbc.Database, "db-schema", config.DB_SCHEMA, "数据库名称")
	createCmd.PersistentFlags().StringVar(&cdbc.CharSet, "db-charset", config.DB_CHARSET, "数据库字符集")
	createCmd.PersistentFlags().IntVar(&cdbc.Timeout, "db-timeout", config.DB_TIMEOUT, "数据库timeout")
	createCmd.PersistentFlags().IntVar(&cdbc.MaxIdelConns, "db-max-idel-conns", config.DB_MAX_IDEL_CONNS, "数据库最大空闲连接数")
	createCmd.PersistentFlags().IntVar(&cdbc.MaxOpenConns, "db-max-open-conns", config.DB_MAX_OPEN_CONNS, "数据库最大连接数")
	createCmd.PersistentFlags().BoolVar(&cdbc.AutoCommit, "db-auto-commit", config.DB_AUTO_COMMIT, "数据库自动提交")
	createCmd.PersistentFlags().BoolVar(&cdbc.PasswordIsDecrypt, "db-password-is-decrypt", config.DB_PASSWORD_IS_DECRYPT, "数据库密码是否需要解密")
	createCmd.PersistentFlags().BoolVar(&cdbc.SqlLogBin, "sql-log-bin", config.SQL_LOG_BIN, "执行sql是否记录binlog")
}

var offlineCfg *config.OfflineConfig

// 添加离线创建回滚SQL子命令
func addOfflineCMD() {
	rootCmd.AddCommand(offlineCmd)

	offlineCfg = config.NewOffileConfig()
	offlineCmd.PersistentFlags().Uint32Var(&offlineCfg.ThreadID, "thread-id", 0, "需要rollback的thread id")
	offlineCmd.PersistentFlags().BoolVar(&offlineCfg.EnableRollbackInsert, "enable-rollback-insert", config.ENABLE_ROLLBACK_INSERT, "是否启用回滚 insert")
	offlineCmd.PersistentFlags().BoolVar(&offlineCfg.EnableRollbackUpdate, "enable-rollback-update", config.ENABLE_ROLLBACK_UPDATE, "是否启用回滚 update")
	offlineCmd.PersistentFlags().BoolVar(&offlineCfg.EnableRollbackDelete, "enable-rollback-delete", config.ENABLE_ROLLBACK_DELETE, "是否启用回滚 delete")
	offlineCmd.PersistentFlags().StringVar(&offlineCfg.SaveDir, "save-dir", "", "相关文件保存的路径")
	offlineCmd.PersistentFlags().StringVar(&offlineCfg.SchemaFile, "schema-file", "", "表结构文件")
	offlineCmd.PersistentFlags().StringArrayVar(&offlineCfg.MatchSqls, "match-sql", []string{}, "使用简单的 SELECT 语句来匹配需要的字段和记录")
	offlineCmd.PersistentFlags().StringArrayVar(&offlineCfg.BinlogFiles, "binlog-file", []string{}, "有哪些binlog文件")
}

// 添加创建回滚SQL子命令
var ec *config.ExecuteConfig
var edbc *config.DBConfig

func addExecuteCMD() {
	rootCmd.AddCommand(executeCmd)

	ec = new(config.ExecuteConfig)
	executeCmd.PersistentFlags().StringVar(&ec.FilePath, "filepath", "", "指定执行的文件")
	executeCmd.PersistentFlags().Int64Var(&ec.Paraller, "paraller", config.EXECUTE_PARALLER, "回滚并发数")

	edbc = new(config.DBConfig)
	// 链接的数据库配置
	executeCmd.PersistentFlags().StringVar(&edbc.Host, "db-host", "", "数据库host")
	executeCmd.PersistentFlags().IntVar(&edbc.Port, "db-port", -1, "数据库port")
	executeCmd.PersistentFlags().StringVar(&edbc.Username, "db-username", "", "数据库用户名")
	executeCmd.PersistentFlags().StringVar(&edbc.Password, "db-password", "", "数据库密码")
	executeCmd.PersistentFlags().StringVar(&edbc.Database, "db-schema", "", "数据库名称")
	executeCmd.PersistentFlags().StringVar(&edbc.CharSet, "db-charset", config.DB_CHARSET, "数据库字符集")
	executeCmd.PersistentFlags().IntVar(&edbc.Timeout, "db-timeout", config.DB_TIMEOUT, "数据库timeout")
	executeCmd.PersistentFlags().IntVar(&edbc.MaxIdelConns, "db-max-idel-conns", config.DB_MAX_IDEL_CONNS, "数据库最大空闲连接数")
	executeCmd.PersistentFlags().IntVar(&edbc.MaxOpenConns, "db-max-open-conns", config.DB_MAX_OPEN_CONNS, "数据库最大连接数")
	executeCmd.PersistentFlags().BoolVar(&edbc.AutoCommit, "db-auto-commit", config.DB_AUTO_COMMIT, "数据库自动提交")
	executeCmd.PersistentFlags().BoolVar(&edbc.PasswordIsDecrypt, "db-password-is-decrypt", config.DB_PASSWORD_IS_DECRYPT, "数据库密码是否需要解密")
	executeCmd.PersistentFlags().BoolVar(&cdbc.SqlLogBin, "sql-log-bin", config.SQL_LOG_BIN, "执行sql是否记录binlog")
}

var offlineStatCfg *config.OfflineStatConfig

// 添加离线创建回滚SQL子命令
func addOfflineStatCMD() {
	rootCmd.AddCommand(offlineStatCmd)

	offlineStatCfg = new(config.OfflineStatConfig)
	offlineStatCmd.PersistentFlags().StringArrayVar(&offlineStatCfg.BinlogFiles, "binlog-file", []string{}, "有哪些binlog文件")
	offlineStatCmd.PersistentFlags().StringVar(&offlineStatCfg.SaveDir, "save-dir", config.DefaultOfflineStatSaveDir, "统计信息保存目录")
}
