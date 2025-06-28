package utils

import (
	"reflect"
)

// Ternary 三元表达式工具函数
// condition: 条件表达式
// trueValue: 条件为真时的返回值
// falseValue: 条件为假时的返回值
// 返回: 根据条件返回对应的值
func Ternary[T any](condition bool, trueValue, falseValue T) T {
	if isNil(condition) {
		panic("condition is nil")
	}
	if condition {
		return trueValue
	}
	return falseValue
}

// isNil 检查值是否为nil
// 支持检查接口、指针、切片、映射、通道等类型
func isNil(value interface{}) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}
