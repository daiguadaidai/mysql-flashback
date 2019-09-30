package utils

import (
	"fmt"
	"testing"
)

func TestNowTimestamp(t *testing.T) {
	NowTimestamp()
}

func Test_NewTime(t *testing.T) {
	timeStr := "2019-04-01 01:00:00"
	ts, err := NewTime(timeStr)
	if err != nil {
		t.Fatal(err.Error())
	}

	fmt.Println(ts.String())
}

func Test_StrTime2Int(t *testing.T) {
	timeStr := "2019-04-01 01:00:00"
	ts, err := StrTime2Int(timeStr)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(ts)
}
