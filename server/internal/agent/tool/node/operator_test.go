package node

import (
	"testing"
)

// TestCreateNodeTool 测试创建节点工具
func TestCreateNodeTool(t *testing.T) {
	tool, err := CreateNodeTool()
	if err != nil {
		t.Fatalf("CreateNodeTool() error = %v", err)
	}
	if tool == nil {
		t.Fatal("CreateNodeTool() returned nil tool")
	}
}

// TestUpdateNodeTool 测试更新节点工具
func TestUpdateNodeTool(t *testing.T) {
	tool, err := UpdateNodeTool()
	if err != nil {
		t.Fatalf("UpdateNodeTool() error = %v", err)
	}
	if tool == nil {
		t.Fatal("UpdateNodeTool() returned nil tool")
	}
}

// TestDeleteNodeTool 测试删除节点工具
func TestDeleteNodeTool(t *testing.T) {
	tool, err := DeleteNodeTool()
	if err != nil {
		t.Fatalf("DeleteNodeTool() error = %v", err)
	}
	if tool == nil {
		t.Fatal("DeleteNodeTool() returned nil tool")
	}
}

// TestGetAllNodeTools 测试获取所有节点工具
func TestGetAllNodeTools(t *testing.T) {
	tools, err := GetAllNodeTools()
	if err != nil {
		t.Fatalf("GetAllNodeTools() error = %v", err)
	}
	if len(tools) != 3 {
		t.Fatalf("GetAllNodeTools() expected 3 tools, got %d", len(tools))
	}
	for i, tool := range tools {
		if tool == nil {
			t.Fatalf("GetAllNodeTools() tool at index %d is nil", i)
		}
	}
}