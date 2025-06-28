package utils

import (
	"testing"
)

func TestTernary(t *testing.T) {
	// 测试基本的三元表达式
	result := Ternary(true, "yes", "no")
	if result != "yes" {
		t.Errorf("Ternary(true, \"yes\", \"no\") = %s, want \"yes\"", result)
	}

	result = Ternary(false, "yes", "no")
	if result != "no" {
		t.Errorf("Ternary(false, \"yes\", \"no\") = %s, want \"no\"", result)
	}

	// 测试整数类型
	intResult := Ternary(true, 10, 20)
	if intResult != 10 {
		t.Errorf("Ternary(true, 10, 20) = %d, want 10", intResult)
	}

	intResult = Ternary(false, 10, 20)
	if intResult != 20 {
		t.Errorf("Ternary(false, 10, 20) = %d, want 20", intResult)
	}
}

func TestIsNil(t *testing.T) {
	// 测试isNil函数
	var nilSlice []int
	var nonNilSlice = []int{1, 2, 3}
	var nilMap map[string]int
	var nonNilMap = map[string]int{"key": 1}
	var nilChan chan int
	var nonNilChan = make(chan int)

	// 测试nil值
	if !isNil(nil) {
		t.Error("isNil(nil) = false, want true")
	}

	if !isNil(nilSlice) {
		t.Error("isNil(nilSlice) = false, want true")
	}

	if !isNil(nilMap) {
		t.Error("isNil(nilMap) = false, want true")
	}

	if !isNil(nilChan) {
		t.Error("isNil(nilChan) = false, want true")
	}

	// 测试非nil值
	if isNil(nonNilSlice) {
		t.Error("isNil(nonNilSlice) = true, want false")
	}

	if isNil(nonNilMap) {
		t.Error("isNil(nonNilMap) = true, want false")
	}

	if isNil(nonNilChan) {
		t.Error("isNil(nonNilChan) = true, want false")
	}

	// 测试基本类型
	if isNil(42) {
		t.Error("isNil(42) = true, want false")
	}

	if isNil("hello") {
		t.Error("isNil(\"hello\") = true, want false")
	}

	if isNil(true) {
		t.Error("isNil(true) = true, want false")
	}
}
