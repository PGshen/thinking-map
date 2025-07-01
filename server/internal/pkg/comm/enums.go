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
	MapStatusExecuting = 1 // 进行中
	MapStatusCompleted = 2 // 已完成
	MapStatusDeleted   = 3 // 删除
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
	NodeStatusPending   = 1 // 待执行
	NodeStatusExecuting = 2 // 执行中
	NodeStatusCompleted = 3 // 已完成
)

// 节点详情类型
const (
	DetailTypeInfo       = "info"       // 信息
	DetailTypeDecompose  = "decompose"  // 拆解
	DetailTypeConclusion = "conclusion" // 结论
)

// 节点详情状态
const (
	DetailStatusPending   = 1 // 待执行
	DetailStatusExecuting = 2 // 执行中
	DetailStatusCompleted = 3 // 已完成
)

// 消息类型
const (
	MessageTypeText   = "text"   // 文本
	MessageTypeRag    = "rag"    // RAG
	MessageTypeNotice = "notice" // 通知
)

// RAG 记录状态
const (
	RAGStatusActive   = 1 // 正常
	RAGStatusArchived = 0 // 归档
	RAGStatusDeleted  = 2 // 删除
)
