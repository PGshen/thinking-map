package conclusion

import "github.com/cloudwego/eino/schema"

type UserMessage struct {
	Reference   string            `json:"reference"`
	Instruction string            `json:"instruction"`
	Conclusion  string            `json:"conclusion"`
	History     []*schema.Message `json:"history"`
}

// ConclusionAnalysisResult 结论分析结果
type ConclusionAnalysisResult struct {
	ConclusionType      string  `json:"conclusion_type"`
	GenerationStrategy  string  `json:"generation_strategy"`
	Reasoning          string  `json:"reasoning"`
	Confidence         float64 `json:"confidence"`
}

// LocalOptimizationRequest 局部优化请求
type LocalOptimizationRequest struct {
	TargetContent string `json:"target_content"` // 需要优化的具体内容
	Instruction   string `json:"instruction"`    // 优化指令
	Context       string `json:"context"`        // 上下文信息
}
