package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"math/rand"
	"strings"
	"time"
)

const (
	TIME_FORMAT           = "2006-01-02 15:04:05"
	TIME_FORMAT_FILE_NAME = "2006_01_02_15_04_05"
)

func NewTime(timeStr string) (time.Time, error) {
	loc, _ := time.LoadLocation("Local") //重要：获取时区
	return time.ParseInLocation(TIME_FORMAT, timeStr, loc)
}

func StrTime2Int(tsStr string) (int64, error) {
	t, err := NewTime(tsStr)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

func TS2String(ts int64, format string) string {
	tm := time.Unix(ts, 0)
	return tm.Format(format)
}

func NowTimestamp() int64 {
	return time.Now().Unix()
}

const (
	UUID_LEN      = 30
	UUID_TIME_LEN = 24
)

// 获取唯一自增ID
func GetUUID() string {
	t := time.Now()
	uuid := t.Format("20060102150405123456")
	currUUIDLen := len(uuid)
	for i := 0; i < UUID_TIME_LEN-currUUIDLen; i++ {
		uuid += "0"
	}
	randLen := 6
	if currUUIDLen > UUID_TIME_LEN {
		randLen = UUID_LEN - currUUIDLen
	}
	return fmt.Sprintf("%s%s", uuid, RandString(randLen))
}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandString(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 用掩码实现随机字符串
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, r.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// 字符串重复
func StrRepeat(d string, cnt int, sep string) string {
	dSlice := make([]string, cnt)
	for i := 0; i < cnt; i++ {
		dSlice[i] = d
	}

	return strings.Join(dSlice, sep)
}

// 获取sql表达表达式, 通过字段
func SqlExprPlaceholderByColumns(names []string, symbol string, holder string, sep string) string {
	exprs := make([]string, len(names))
	for i, name := range names {
		exprs[i] = fmt.Sprintf("`%s` %s %s", name, symbol, holder)
	}
	return strings.Join(exprs, sep)
}

// 将row转化为相关类型interface
func ConverSQLType(row []interface{}) ([]interface{}, error) {
	rs := make([]interface{}, len(row))
	for i, field := range row {
		if field == nil {
			rs[i] = "NULL"
			continue
		}
		switch uintData := field.(type) {
		case []uint8:
			strData := string(uintData)
			sqlValue, err := GetSqlStrValue(strData, "'")
			if err != nil {
				return nil, fmt.Errorf("字段数据转化成sql字符串出错. %v", err.Error())
			}
			rs[i] = sqlValue
		case string:
			sqlValue, err := GetSqlStrValue(uintData, "'")
			if err != nil {
				return nil, fmt.Errorf("字段数据转化成sql字符串出错. %v", err.Error())
			}
			rs[i] = sqlValue
		default:
			rs[i] = field
		}
	}

	return rs, nil
}

// 将row转化为相关类型interface
func ReplaceSqlPlaceHolder(sqlStr string, row []interface{}, crc32 uint32, timeStr string) string {
	offset := 2
	rs := make([]interface{}, len(row)+offset)
	rs[0] = crc32
	rs[1] = timeStr
	for i, _ := range row {
		rs[i+offset] = "%v"
	}

	return fmt.Sprintf(sqlStr, rs...)
}

func GetCrc32ByStr(data string) int64 {
	return int64(crc32.ChecksumIEEE([]byte(data)))
}

// 一共有分几片计算出 所在分片号
func GetCrc32ByInterfaceSlice(cols []interface{}) uint32 {
	if len(cols) == 0 {
		return 0
	}

	rawBytes := make([]byte, 0, 50)
	for _, col := range cols {
		raw, err := GetBytes(col)
		if err != nil {
			raw = []byte{0}
		}
		rawBytes = append(rawBytes, raw...)
	}

	// 进行crc32
	return crc32.ChecksumIEEE(rawBytes)
}

func GetBytes(data interface{}) ([]byte, error) {
	switch val := data.(type) {
	case int8, int16, int32, int64, uint8, uint16, uint32, uint64:
		return KindIntToByte(data)
	case int:
		return Int64ToByte(int64(val))
	case uint:
		return Uint64ToByte(uint64(val))
	case string:
		return []byte(val), nil
	case []byte:
		return data.([]byte), nil
	default:
	}

	return nil, fmt.Errorf("未知数据类型%T, 转为[]byte. 数据为: %v", data, data)
}

func Int64ToByte(num int64) ([]byte, error) {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.LittleEndian, num); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func Uint64ToByte(num uint64) ([]byte, error) {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.LittleEndian, num); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func KindIntToByte(num interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.LittleEndian, num); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func GetSQLStmtHearderComment(stmt *string) string {
	var commentBegin bool
	var commentEnd bool
	var meetBeginRod bool
	var meetEndAsterisk bool
	var startContentPos int
	var endContentPos int

	for i, item := range *stmt {
		if !commentBegin { // 注释没开始
			if meetBeginRod { // 开始的反斜杆之后必须是 星号 '*' ascii: 42
				if item != 42 {
					return ""
				}
				commentBegin = true     // 设置注释已经开始
				startContentPos = i + 1 // 设置注释内容开始的位点
			} else { // 上一个字符没有碰到反斜杆 '/' ascii: 47
				switch item {
				case 9, 10, 13, 32: // 空白符
					continue
				case 47: // 反斜杆
					meetBeginRod = true //
				default:
					return ""
				}
			}
		} else { // 注释开始, 获取注释内容结束位点
			if meetEndAsterisk { // 碰到星号 '*' ascii: 42 需要检测是否注释结束
				if item == 47 { // 碰到了  */ 注释结束
					endContentPos = i - 1 // 获取注释内容结束位点
					commentEnd = true
					break
				} else if item == 42 { // 还是星号进行下一次字符判断
					continue
				}

				// 星号后面接的不是 '/'
				meetEndAsterisk = false
			} else { // 没有遇到星号
				if item == 42 {
					meetEndAsterisk = true
				}
			}
		}
	}

	if commentEnd {
		return (*stmt)[startContentPos:endContentPos]
	}

	return ""
}

/*
data: 源数据
warpStr: 最后元数据需要使用什么包括
如: data: aabb, wrapStr: '
最后: 'aabb'
*/
func GetSqlStrValue(data string, wrapStr string) (string, error) {
	oriStrRunes := []rune(data)

	var sb strings.Builder
	// 添加开头单引号
	_, err := fmt.Fprint(&sb, wrapStr)
	if err != nil {
		return "", err
	}
	for _, oriStrRune := range oriStrRunes {
		var s string

		switch oriStrRune {
		case 34: // " 双引号
			s = "\\\""
		case 39: // ' 单引号
			s = "\\'"
		case 92: // \ 反斜杠
			s = "\\\\"
		default:
			s = string(oriStrRune)
		}

		_, err := fmt.Fprint(&sb, s)
		if err != nil {
			return "", err
		}
	}

	// 添加结尾单引号
	_, err = fmt.Fprint(&sb, wrapStr)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}
