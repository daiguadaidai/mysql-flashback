package utils

import (
	"github.com/daiguadaidai/parser/opcode"
)

type BoolNode struct {
	Data bool
	Next *BoolNode
}

func NewBoolNode(d bool) *BoolNode {
	return &BoolNode{
		Data: d,
	}
}

type CalcStack struct {
	top     *BoolNode
	result  bool
	isFirst bool
}

func NewCalcStack() *CalcStack {
	return &CalcStack{
		isFirst: true,
	}
}

func (this *CalcStack) PushOrCalc(data interface{}) {
	switch v := data.(type) {
	case opcode.Op:
		this.Calc(v)
	case bool:
		this.Push(v)
	}
}

func (this *CalcStack) Calc(op opcode.Op) {
	data, ok := this.Pop()
	if !ok {
		return
	}
	switch op {
	case opcode.LogicAnd:
		this.result = data && this.result
	case opcode.LogicOr:
		this.result = data || this.result
	}
}

func (this *CalcStack) IsEmpty() bool {
	return this.isFirst && this.top == nil
}

func (this *CalcStack) Result() bool {
	return this.result
}

func (this *CalcStack) Push(data bool) {
	if this.isFirst {
		this.result = data
		this.isFirst = false
		return
	}
	newNode := NewBoolNode(this.result)
	this.result = data
	if this.top == nil { // 空队列
		this.top = newNode
		return
	}

	// 非空队列
	tmpNode := this.top
	this.top = newNode
	newNode.Next = tmpNode
}

func (this *CalcStack) Pop() (bool, bool) {
	// 空队列
	if this.top == nil {
		return false, false
	}

	// 非空队列
	data := this.top.Data
	this.top = this.top.Next

	return data, true
}
