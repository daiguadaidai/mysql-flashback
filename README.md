# MySQL Flashback

史上最变态的MySQL DML 闪回工具

@[toc]

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
