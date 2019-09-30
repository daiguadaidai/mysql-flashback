package visitor

import (
	"fmt"
	"github.com/daiguadaidai/tidb/types"
	driver "github.com/daiguadaidai/tidb/types/parser_driver"
)

func GetValueExprValue(node *driver.ValueExpr) (interface{}, error) {
	value := node.GetValue()
	switch data := value.(type) {
	case *types.MyDecimal:
		v, err := data.ToFloat64()
		if err != nil {
			return nil, fmt.Errorf("类型:*types.MyDecimal. 转化为Float64出错. %s", err.Error())
		}
		return v, nil
	}
	return value, nil
}
