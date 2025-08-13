# 增量计划更新功能

## 概述

增量计划更新功能允许LLM通过指定具体的操作（添加、修改、删除、重排序）来更新执行计划，而不是重新生成整个计划。这种方式更加高效，并且能够保持已完成步骤的状态。

## 功能特性

### 支持的操作类型

1. **添加步骤** (`add`): 在指定位置添加新的执行步骤
2. **修改步骤** (`modify`): 更新现有步骤的属性（不能修改已完成的步骤）
3. **删除步骤** (`remove`): 删除指定的步骤
4. **重排序步骤** (`reorder`): 调整步骤的执行顺序

### 智能状态管理

- **保护已完成步骤**: 已完成的步骤不能被修改或删除
- **选择性清除结果**: 只清除受影响步骤的专家执行结果
- **版本控制**: 每次更新都会增加计划版本号并保存历史记录

## LLM 输出格式

当需要更新计划时，LLM应该返回以下JSON格式：

```json
{
  "plan_metadata": {
    "name": "更新后的计划名称",
    "description": "更新后的计划描述",
    "update_reason": "更新原因说明"
  },
  "operations": [
    {
      "type": "add",
      "position": 2,
      "step_data": {
        "name": "新步骤名称",
        "description": "步骤描述",
        "assigned_specialist": "specialist_name",
        "priority": 1,
        "dependencies": ["step1"],
        "parameters": {"key": "value"}
      },
      "reason": "添加此步骤的原因"
    },
    {
      "type": "modify",
      "step_id": "existing_step_id",
      "step_data": {
        "name": "修改后的步骤名称",
        "description": "修改后的描述"
      },
      "reason": "修改此步骤的原因"
    },
    {
      "type": "remove",
      "step_id": "step_to_remove",
      "reason": "删除此步骤的原因"
    },
    {
      "type": "reorder",
      "step_id": "step_to_move",
      "position": 1,
      "reason": "重排序的原因"
    }
  ]
}
```

## 操作详细说明

### 添加操作 (add)
- `position`: 新步骤插入的位置（从0开始）
- `step_data`: 完整的步骤数据
- 会自动生成新的步骤ID

### 修改操作 (modify)
- `step_id`: 要修改的步骤ID
- `step_data`: 只需包含要修改的字段
- 不能修改已完成的步骤

### 删除操作 (remove)
- `step_id`: 要删除的步骤ID
- 不能删除已完成的步骤
- 会自动清除相关的专家执行结果

### 重排序操作 (reorder)
- `step_id`: 要移动的步骤ID
- `position`: 目标位置（从0开始）
- 不影响步骤的其他属性

## 实现细节

### 核心方法

- `clonePlan()`: 深度复制当前计划
- `applyOperation()`: 应用单个操作到计划
- `determineUpdateType()`: 根据操作确定更新类型
- `selectiveClearSpecialistResults()`: 选择性清除专家结果

### 错误处理

- 尝试修改已完成步骤时返回错误
- 无效的步骤ID或位置会被忽略
- 操作失败不会影响其他操作的执行

## 使用示例

```go
// 在PlanUpdateHandler中使用
handler := &PlanUpdateHandler{}

// LLM返回的更新数据会被自动解析并应用
// 系统会：
// 1. 解析操作列表
// 2. 验证操作的有效性
// 3. 应用所有有效操作
// 4. 更新计划版本和历史
// 5. 选择性清除受影响的专家结果
```

## 测试覆盖

项目包含完整的测试套件，覆盖：
- 所有操作类型的功能测试
- 边界条件和错误处理
- 状态保护机制
- 选择性结果清除

运行测试：
```bash
go test -v -run "TestPlanUpdateHandler|TestSelectiveClearSpecialistResults"
```

## 优势

1. **效率提升**: 避免重新生成整个计划
2. **状态保护**: 保持已完成步骤的执行结果
3. **精确控制**: LLM可以精确指定需要的更改
4. **可追溯性**: 完整的更新历史记录
5. **灵活性**: 支持多种操作类型的组合使用