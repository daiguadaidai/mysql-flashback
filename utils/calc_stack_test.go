package utils

import (
	"fmt"
	"github.com/daiguadaidai/parser/opcode"
	"testing"
)

func TestCalcStack_Calc(t *testing.T) {
	var datas []interface{} = []interface{}{false, true, opcode.LogicAnd, true, true, false, false, true, opcode.LogicAnd, opcode.LogicAnd, opcode.LogicOr, opcode.LogicAnd, opcode.LogicOr}
	calc := NewCalcStack()
	for _, data := range datas {
		calc.PushOrCalc(data)
	}

	fmt.Println(calc.Result())
}

func TestCalcStack_CalcOneData(t *testing.T) {
	var datas []interface{} = []interface{}{true}
	calc := NewCalcStack()
	if !calc.IsEmpty() {
		fmt.Println(calc.Result())
	} else {
		fmt.Println("NULL")
	}

	for _, data := range datas {
		calc.PushOrCalc(data)
	}

	if !calc.IsEmpty() {
		fmt.Println(calc.Result())
	} else {
		fmt.Println("NULL")
	}
}
