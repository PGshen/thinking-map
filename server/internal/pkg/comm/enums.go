/*
 * @Date: 2025-06-18 23:15:18
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-06-24 09:13:14
 * @FilePath: /thinking-map/server/internal/pkg/comm/enums.go
 */
package comm

// 用户状态
const (
	UserStatusActive   = 1 // 正常
	UserStatusInactive = 0 // 禁用
)

// 思维导图状态
const (
	MapStatusInitial   = "initial"   // 初始
	MapStatusRunning   = "running"   // 进行中
	MapStatusCompleted = "completed" // 已完成
	MapStatusDeleted   = "deleted"   // 删除
)

const (
	ProblemTypeResearch   string = "research"   // 研究型
	ProblemTypeCreative   string = "creative"   // 创意型
	ProblemTypeAnalytical string = "analytical" // 分析型
	ProblemTypePlanning   string = "planning"   // 规划型
	ProblemTypeGeneral    string = "general"    // 通用型
)

// 思维节点类型
const (
	NodeTypeProblem        string = "problem"     // 问题
	NodeTypeInfoCollection string = "information" // 信息收集
	NodeTypeAnalysis       string = "analysis"    // 分析
	NodeTypeGeneration     string = "generation"  // 生成
	NodeTypeEvaluation     string = "evaluation"  // 评估
)

// 思维节点状态
const (
	NodeStatusInitial   = "initial"   // 初始
	NodeStatusPending   = "pending"   // 待执行
	NodeStatusRunning   = "running"   // 执行中
	NodeStatusCompleted = "completed" // 已完成
	NodeStatusError     = "error"     // 错误
)

// 节点详情类型
const (
	DetailTypeInfo       = "info"       // 信息
	DetailTypeDecompose  = "decompose"  // 拆解
	DetailTypeConclusion = "conclusion" // 结论
)

// 消息类型
const (
	MessageTypeText   = "text"   // 文本
	MessageTypeRag    = "rag"    // RAG
	MessageTypeNotice = "notice" // 通知
	MessageTypeAction = "action" // 操作
)

// RAG 记录状态
const (
	RAGStatusActive   = "active"   // 正常
	RAGStatusArchived = "archived" // 归档
	RAGStatusDeleted  = "deleted"  // 删除
)

const (
	EventText = "text"
	EventJson = "json"
	EventData = "data"
	EventID   = "id"
)
