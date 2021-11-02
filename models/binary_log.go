package models

type BinaryLog struct {
	LogName  string `gorm:"column:Log_name"`
	FileSize uint64 `gorm:"column:File_size"`
}

// 比较两个BinaryLog记录是否一样
func (this *BinaryLog) Equal(other *BinaryLog) bool {
	if this.LogName != other.LogName {
		return false
	}
	if this.FileSize != other.FileSize {
		return false
	}

	return true
}
