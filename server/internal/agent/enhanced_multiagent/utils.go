package enhanced_multiagent

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

// 提示构建函数

// BuildThinkingPrompt 构建思考提示
func BuildThinkingPrompt(messages []*schema.Message, state *EnhancedState) string {
	var prompt strings.Builder
	prompt.WriteString("请分析以下用户请求，并进行深度思考：\n\n")
	
	// 添加历史消息
	for _, msg := range messages {
		prompt.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
	}
	
	prompt.WriteString("\n请按照以下格式输出你的思考结果：\n")
	prompt.WriteString("```json\n")
	prompt.WriteString("{\n")
	prompt.WriteString("  \"thought\": \"你的详细思考过程\",\n")
	prompt.WriteString("  \"complexity\": \"simple|moderate|complex\",\n")
	prompt.WriteString("  \"reasoning\": \"复杂度判断的理由\",\n")
	prompt.WriteString("  \"next_action\": \"direct_answer|create_plan|execute_step|reflect\"\n")
	prompt.WriteString("}\n")
	prompt.WriteString("```")
	
	return prompt.String()
}

// BuildDirectAnswerPrompt 构建直接回答提示
func BuildDirectAnswerPrompt(messages []*schema.Message) string {
	var prompt strings.Builder
	prompt.WriteString("请直接回答以下用户问题：\n\n")
	
	for _, msg := range messages {
		prompt.WriteString(fmt.Sprintf("用户: %s\n", msg.Content))
	}
	
	prompt.WriteString("\n请提供清晰、准确的回答。")
	return prompt.String()
}

// BuildFinalAnswerPrompt 构建最终答案提示
func BuildFinalAnswerPrompt(messages []*schema.Message, state *EnhancedState) string {
	var prompt strings.Builder
	prompt.WriteString("基于以下信息提供最终答案：\n\n")
	
	// 添加原始问题
	for _, msg := range messages {
		prompt.WriteString(fmt.Sprintf("用户问题: %s\n", msg.Content))
	}
	
	// 添加执行历史摘要
	if len(state.ExecutionHistory) > 0 {
		prompt.WriteString("\n执行历史：\n")
		for i, record := range state.ExecutionHistory {
			prompt.WriteString(fmt.Sprintf("%d. 轮次 %d 的执行结果\n", i+1, record.Round))
		}
	}
	
	prompt.WriteString("\n请提供完整、准确的最终答案。")
	return prompt.String()
}

// 解析函数

// ParseThinkingResult 解析思考结果
func ParseThinkingResult(content string) (*ThinkingResult, error) {
	var result struct {
		Thought    string `json:"thought"`
		Complexity string `json:"complexity"`
		Reasoning  string `json:"reasoning"`
		NextAction string `json:"next_action"`
	}
	
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("解析思考结果失败: %w", err)
	}
	
	// 转换复杂度
	var complexity TaskComplexity
	switch result.Complexity {
	case "simple":
		complexity = TaskComplexitySimple
	case "moderate":
		complexity = TaskComplexityModerate
	case "complex":
		complexity = TaskComplexityComplex
	default:
		complexity = TaskComplexityModerate
	}
	
	// 转换下一步行动
	var nextAction ActionType
	switch result.NextAction {
	case "direct_answer":
		nextAction = ActionTypeDirectAnswer
	case "create_plan":
		nextAction = ActionTypeCreatePlan
	case "execute_step":
		nextAction = ActionTypeExecuteStep
	case "reflect":
		nextAction = ActionTypeReflect
	default:
		nextAction = ActionTypeDirectAnswer
	}
	
	return &ThinkingResult{
		Thought:          result.Thought,
		Complexity:       complexity,
		Reasoning:        result.Reasoning,
		NextAction:       nextAction,
		OriginalMessages: nil,
		Timestamp:        time.Now(),
	}, nil
}

// 业务逻辑函数

// DetermineComplexity 确定任务复杂度
func DetermineComplexity(content string) TaskComplexity {
	content = strings.ToLower(content)
	
	// 简单任务的关键词
	simpleKeywords := []string{"什么是", "定义", "解释", "简单", "基本"}
	for _, keyword := range simpleKeywords {
		if strings.Contains(content, keyword) {
			return TaskComplexitySimple
		}
	}
	
	// 复杂任务的关键词
	complexKeywords := []string{"分析", "设计", "实现", "优化", "架构", "系统"}
	for _, keyword := range complexKeywords {
		if strings.Contains(content, keyword) {
			return TaskComplexityComplex
		}
	}
	
	return TaskComplexityModerate
}

// CleanInput 清理输入内容
func CleanInput(input string) string {
	// 移除多余的空白字符
	input = strings.TrimSpace(input)
	// 移除连续的空行
	lines := strings.Split(input, "\n")
	var cleanLines []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			cleanLines = append(cleanLines, line)
		}
	}
	return strings.Join(cleanLines, "\n")
}

// ValidateJSON 验证JSON格式
func ValidateJSON(content string) error {
	var temp interface{}
	return json.Unmarshal([]byte(content), &temp)
}

// FormatDuration 格式化持续时间
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fm", d.Minutes())
}