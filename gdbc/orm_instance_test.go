package gdbc

import (
	"fmt"
	"sync"
	"testing"
)

// 测试并发获取数据库链接(使用了单例模式)
func TestGetOrmInstance(t *testing.T) {
	wg := new(sync.WaitGroup)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(_wg *sync.WaitGroup) {
			defer _wg.Done()

			ormInstance := GetOrmInstance()
			ormInstance.DB.DB().Ping()

			// 查询数据库
			rows, err := ormInstance.DB.DB().Query("SELECT * FROM task limit 1")
			if err != nil {
				t.Errorf("error: %v", err)
			}
			defer rows.Close()
			fmt.Println(rows)

			/*
			   columns, _ := rows.Columns()
			   scanArgs := make([]interface{}, len(columns))
			   values := make([]interface{}, len(columns))
			   for j := range values {
			       scanArgs[j] = &values[j]
			   }

			   record := make(map[string]string)
			   for rows.Next() {
			       //将行数据保存到record字典
			       err = rows.Scan(scanArgs...)
			       for i, col := range values {
			           if col != nil {
			               record[columns[i]] = string(col.([]byte))
			           }
			       }
			   }

			   fmt.Println(record)
			*/
		}(wg)
	}

	wg.Wait()
}
