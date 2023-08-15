# MySQL Flashback

史上最变态的MySQL DML 闪回工具

- [MySQL Flashback](#mysql-flashback)
  - [原理](#%E5%8E%9F%E7%90%86)
  - [再次捣腾这功能原因](#%E5%86%8D%E6%AC%A1%E6%8D%A3%E8%85%BE%E8%BF%99%E5%8A%9F%E8%83%BD%E5%8E%9F%E5%9B%A0)
  - [生成二进制](#%E7%94%9F%E6%88%90%E4%BA%8C%E8%BF%9B%E5%88%B6)
  - [支持的功能](#%E6%94%AF%E6%8C%81%E7%9A%84%E5%8A%9F%E8%83%BD)
    - [大家有我也有](#%E5%A4%A7%E5%AE%B6%E6%9C%89%E6%88%91%E4%B9%9F%E6%9C%89)
    - [我的亮点](#%E6%88%91%E7%9A%84%E4%BA%AE%E7%82%B9)
  - [生成回滚语句](#%E7%94%9F%E6%88%90%E5%9B%9E%E6%BB%9A%E8%AF%AD%E5%8F%A5)
    - [溜溜的玩法](#%E6%BA%9C%E6%BA%9C%E7%9A%84%E7%8E%A9%E6%B3%95)
    - [你可以指定多个 SQL](#%E4%BD%A0%E5%8F%AF%E4%BB%A5%E6%8C%87%E5%AE%9A%E5%A4%9A%E4%B8%AA-sql)
    - [老旧玩法](#%E8%80%81%E6%97%A7%E7%8E%A9%E6%B3%95)
  - [执行回滚SQL](#%E6%89%A7%E8%A1%8C%E5%9B%9E%E6%BB%9Asql)
    - [回滚原理和注意事项](#%E5%9B%9E%E6%BB%9A%E5%8E%9F%E7%90%86%E5%92%8C%E6%B3%A8%E6%84%8F%E4%BA%8B%E9%A1%B9)
    - [执行回滚就简单了](#%E6%89%A7%E8%A1%8C%E5%9B%9E%E6%BB%9A%E5%B0%B1%E7%AE%80%E5%8D%95%E4%BA%86)
  - [离线binlog文件生成回滚信息](#%E7%A6%BB%E7%BA%BFbinlog%E6%96%87%E4%BB%B6%E7%94%9F%E6%88%90%E5%9B%9E%E6%BB%9A%E4%BF%A1%E6%81%AF)

## 原理

在 `MySQL Binlog` 已经被玩烂掉的时代, 我想大家也都知道其中的原理. 无非就是解析 `Binlog` 反向生成需要回滚的 `SQL` 语句.

## 再次捣腾这功能原因

说实在的网上也有很多的关于这种工具的实现, 但都不尽人意.

在网上的大多数都是使用 `Python` 开发的. 但是在量大的时候有经常有各种各样的问题. 所以我这边就使用 `Golang` 重新实现一个

> 会点 `Python` 的 `DBA` 想必都使用过 `python-mysql-replication` 这个解析 `MySQL Binlog` 的包.
> `python-mysql-replication` 这个包有一个时间戳的`bug`不知道修复没有. 我之前是自己在项目中修复的, 但是没有提交 `pr`

## 生成二进制

```
go build
```

## 支持的功能

### 大家有我也有

1. 指定 `开始` 和 `结束` 回滚的位点

2. 指定 `开始` 和 `结束` 回滚的时间

3. 开始时间和位点可以混搭使用, 如果都指定了以位点为准

4. 指定多个库表

5. 指定 thread id

### 我的亮点

1. 可以使用指定 `SQL` 语句, 来获取上面需要回滚的参数

2. 可以支持条件过滤, 条件过滤也是使用 `SQL` 的形式体现的

## 生成回滚语句

### 溜溜的玩法

```
./mysql-flashback create \
    --match-sql="SELECT col_1, col_2, col_3 FROM schema.table WHERE col_1 = 1 AND col_2 IN(1, 2, 3) AND col_3 BETWEEN 10 AND 20 AND start_log_file = 'mysql-bin.000001' AND start_log_pos = 4 AND end_log_file = 'mysql-bin.000004' AND end_log_pos = 0 AND start_rollback_time = '2019-06-06 12:00:01' AND end_rollback_time = '2019-06-07 12:00:01' AND thread_id = 0"
```

上面的 `--match-sql` 参数值特别长, 主要是因为参数要在一个字符串里面导致的.

格式化看 `--match-sql` 参数值

```
SELECT col_1, col_2, col_3                           -- 指定只需要的字段, SELECT * 代表所有字段
FROM schema.table                                    -- 执行需要回滚的表. 需要(显示指定)表所在的数据库
WHERE col_1 = 1                                      -- 过滤条件
    AND col_2 IN(1, 2, 3)                            -- 过滤条件 IN 表达式
    AND col_3 BETWEEN 10 AND 20                      -- 过滤条件 BEWTEEN ... AND ... 表达式
    AND start_log_file = 'mysql-bin.000001'          -- 指定需要回滚的范围(开始binlog[文件]), 非过滤条件.
    AND start_log_pos = 4                            -- 指定需要回滚的范围(开始binlog[位点]), 非过滤条件.
    AND end_log_file = 'mysql-bin.000004'            -- 指定需要回滚的范围(结束binlog[文件]), 非过滤条件.
    AND end_log_pos = 0                              -- 指定需要回滚的范围(结束binlog[位点]), 非过滤条件.
    AND start_rollback_time = '2019-06-06 12:00:01'  -- 指定需要回滚的范围(开始时间), 非过滤条件.
    AND end_rollback_time = '2019-06-07 12:00:01'    -- 指定需要回滚的范围(结束时间), 非过滤条件.
    AND thread_id = 0                                -- 指定需要回滚的 Tread ID.
```

将 `SQL` 格式化成好看的格式后, 就明了了 `SQL` 在做什么事:

* 下面这些条件指定了需要回滚的参数, 解析 `Binlog` 的参数.

```
    AND start_log_file = 'mysql-bin.000001'          -- 指定需要回滚的范围(开始binlog[文件]), 非过滤条件.
    AND start_log_pos = 4                            -- 指定需要回滚的范围(开始binlog[位点]), 非过滤条件.
    AND end_log_file = 'mysql-bin.000004'            -- 指定需要回滚的范围(结束binlog[文件]), 非过滤条件.
    AND end_log_pos = 0                              -- 指定需要回滚的范围(结束binlog[位点]), 非过滤条件.
    AND start_rollback_time = '2019-06-06 12:00:01'  -- 指定需要回滚的范围(开始时间), 非过滤条件.
    AND end_rollback_time = '2019-06-07 12:00:01'    -- 指定需要回滚的范围(结束时间), 非过滤条件.
    AND thread_id = 0                                -- 指定需要回滚的 Tread ID.
```

* 下面条件指定了, 对数据进行过滤. 生成回滚语句的时候会对你指定的条件进行匹配

```
WHERE col_1 = 1                                      -- 过滤条件
    AND col_2 IN(1, 2, 3)                            -- 过滤条件 IN 表达式
    AND col_3 BETWEEN 10 AND 20                      -- 过滤条件 BEWTEEN ... AND ... 表达式
```

* 下面是指定你需要回滚的表

```
FROM schema.table                                    -- 执行需要回滚的表. 需要(显示指定)表所在的数据库
```

* 下面是指定你生成回滚语句只需要哪些字段, 主要针对 `UPDATE`, 对于 `INSERT` 是必须是所有的字段值.

```
SELECT col_1, col_2, col_3                           -- 指定只需要的字段, SELECT * 代表所有字段
```

### 你可以指定多个 SQL

```
./mysql-flashback create \
    --match-sql="指定的SQL1" \
    --match-sql="指定的SQL2"
```

### 老旧玩法

传统指定参数值的玩法, 大家可以执行 `./mysql-flashback create --help` 来查看使用示例

```
./mysql-flashback create --help
生成回滚的sql. 如下:
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

Usage:
  mysql-flashback create [flags]

Flags:
      --db-auto-commit            数据库自动提交 (default true)
      --db-charset string         数据库字符集 (default "utf8mb4")
      --db-host string            数据库host (default "127.0.0.1")
      --db-max-idel-conns int     数据库最大空闲连接数 (default 8)
      --db-max-open-conns int     数据库最大连接数 (default 8)
      --db-password string        数据库密码 (default "root")
      --db-password-is-decrypt    数据库密码是否需要解密 (default true)
      --db-port int               数据库port (default 3306)
      --db-schema string          数据库名称
      --db-timeout int            数据库timeout (default 10)
      --db-username string        数据库用户名 (default "root")
      --enable-rollback-delete    是否启用回滚 delete (default true)
      --enable-rollback-insert    是否启用回滚 insert (default true)
      --enable-rollback-update    是否启用回滚 update (default true)
      --end-log-file string       结束日志文件
      --end-log-pos uint32        结束日志文件点位
      --end-time string           结束时间
  -h, --help                      help for create
      --match-sql string          使用简单的 SELECT 语句来匹配需要的字段和记录
      --rollback-schema strings   指定回滚的数据库, 该命令可以指定多个
      --rollback-table strings    需要回滚的表, 该命令可以指定多个
      --save-dir string           相关文件保存的路径
      --start-log-file string     开始日志文件
      --start-log-pos uint32      开始日志文件点位
      --start-time string         开始时间
      --thread-id uint32          需要rollback的thread id
```

## 执行回滚SQL

在执行 `./mysql-flashback create ...` 之后会生成两个 `SQL` 文件

1. 原来执行`SQL`语句的`SQL`文件, 文件中的语句并非当时执行的语句, 而是影响数据量的一个语句. 如: 一个 `UPDATE` 语句影响的 `10` 条数据, 则原`SQL`语句也将是 `10` 条.

2. 回滚的 `SQL` 文件

### 回滚原理和注意事项

1. 由于生成回滚语句是顺序记入到文件中去的, 所以我们回滚的时候需要倒序读取文件 `SQL` 进行回滚.

2. 在看回滚语句的时候你会发现, 每个`SQL`前面都有注释`/* crc32:xxx */`, 这里面的值`xxx`记录的是每条数据主键的一个`crc32`值. 主要是为了并发而记录的.

```
/* crc32:2313941001 */ INSERT INTO `employees`.`emp1`(`emp_no`, `birth_date`, `first_name`, `last_name`, `gender`, `hire_date`) VALUES(10008, "1958-02-19", "Saniya", "Kalloufi", 1, "1994-09-15");
```

### 执行回滚就简单了

使用 `./mysql-flashback execute --help` 就可以看到有一个使用示例

```
./mysql-flashback execute --help
倒序执行指定的sql回滚文件. 如下:
Example:
./mysql-flashback execute \
    --filepath="/tmp/test.sql" \
    --paraller=8 \
    --db-host="127.0.0.1" \
    --db-port=3306 \
    --db-username="root" \
    --db-password="root"

Usage:
  mysql-flashback execute [flags]

Flags:
      --db-auto-commit           数据库自动提交 (default true)
      --db-charset string        数据库字符集 (default "utf8mb4")
      --db-host string           数据库host
      --db-max-idel-conns int    数据库最大空闲连接数 (default 8)
      --db-max-open-conns int    数据库最大连接数 (default 8)
      --db-password string       数据库密码
      --db-password-is-decrypt   数据库密码是否需要解密 (default true)
      --db-port int              数据库port (default -1)
      --db-schema string         数据库名称
      --db-timeout int           数据库timeout (default 10)
      --db-username string       数据库用户名
      --filepath string          指定执行的文件
  -h, --help                     help for execute
      --paraller int             回滚并发数 (default 1)
```

## 离线binlog文件生成回滚信息

之前展示的生成回滚信息是使用模拟`slave`连接到`mysql`进行获取binlog事件, 经常是`mysql`连接已经不存在了, 需要指定离线`binlog`进行生成回滚信息

```
解析离线binlog, 生成回滚SQL. 如下:
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

Usage:
  mysql-flashback offline [flags]

Flags:
      --binlog-file stringArray   有哪些binlog文件
      --enable-rollback-delete    是否启用回滚 delete (default true)
      --enable-rollback-insert    是否启用回滚 insert (default true)
      --enable-rollback-update    是否启用回滚 update (default true)
  -h, --help                      help for offline
      --match-sql stringArray     使用简单的 SELECT 语句来匹配需要的字段和记录
      --save-dir string           相关文件保存的路径
      --schema-file string        表结构文件
      --thread-id uint32          需要rollback的thread id
```

## 离线解析binlog文件生成相关统计信息

生成4总统计信息, 会保存在相关的目录文件中:

1. **offline_stat_output/table_stat.txt:** 表相关的统计信息(通过影响行数排序)

2. **offline_stat_output/thread_stat.txt:** ThreadId相关的统计信息(通过影响行数排序)

3. **offline_stat_output/timestamp_stat.txt:** 时间相关的统计信息, 通过秒为单位进行统计, 记录的时间是事务执行`BEGIN`的时间, 一个时间点有可能有多个事务.

4. **offline_stat_output/xid_stat.txt:** 事务相关的统计信息, 每个事务的相关统计信息, 通过xid代表不同事务

### 使用方法

```
./mysql-flashback offline-stat --help
解析离线binlog, 统计binlog信息. 如下:
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

Usage:
  mysql-flashback offline-stat [flags]

Flags:
      --binlog-file stringArray   有哪些binlog文件
  -h, --help                      help for offline-stat
      --save-dir string           统计信息保存目录 (default "offline_stat_output")
```

### 每个文件示例

#### table_stat.txt

```
表: employees.emp_01 	dml影响行数: 100, insert: 100, update: 0, delete: 0, 表出现次数: 1
表: employees.emp 	dml影响行数: 10, insert: 0, update: 7, delete: 3, 表出现次数: 3
```

#### thread_stat.txt

```
threadId: 464	dml影响行数: 110, insert: 100, update: 7, delete: 3, 表出现次数: 4
```

#### timestamp_stat.txt

```
2023-08-11 14:35:06: dml影响行数: 1, insert: 0, update: 0, delete: 1, 事务数: 1, 开始位点: /Users/hh/Desktop/mysql-bin.000200:395
2023-08-11 14:35:20: dml影响行数: 2, insert: 0, update: 0, delete: 2, 事务数: 1, 开始位点: /Users/hh/Desktop/mysql-bin.000200:749
2023-08-11 14:36:22: dml影响行数: 100, insert: 100, update: 0, delete: 0, 事务数: 1, 开始位点: /Users/hh/Desktop/mysql-bin.000200:1147
2023-08-11 14:37:28: dml影响行数: 7, insert: 0, update: 7, delete: 0, 事务数: 1, 开始位点: /Users/hh/Desktop/mysql-bin.000200:4217
```

#### xid_stat.txt

```
Xid: 7500 	2023-08-11 14:35:06 	 dml影响行数: 1, insert: 0, update: 0, delete: 1, 开始位点: /Users/hh/Desktop/mysql-bin.000200:395
Xid: 7501 	2023-08-11 14:35:20 	 dml影响行数: 2, insert: 0, update: 0, delete: 2, 开始位点: /Users/hh/Desktop/mysql-bin.000200:749
Xid: 7504 	2023-08-11 14:36:22 	 dml影响行数: 100, insert: 100, update: 0, delete: 0, 开始位点: /Users/hh/Desktop/mysql-bin.000200:1147
Xid: 7507 	2023-08-11 14:37:28 	 dml影响行数: 7, insert: 0, update: 7, delete: 0, 开始位点: /Users/hh/Desktop/mysql-bin.000200:4217
```
