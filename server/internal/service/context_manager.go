/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @FilePath: /thinking-map/server/internal/service/context_manager.go
 */
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/PGshen/thinking-map/server/internal/model"
	"github.com/PGshen/thinking-map/server/internal/model/dto"
	"github.com/PGshen/thinking-map/server/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContextManager 上下文管理器
// 职责：通过工程化方式管理思考导图的上下文信息，基于导图结构自动收集相关上下文
type ContextManager struct {
	nodeRepo    repository.ThinkingNode
	mapRepo     repository.ThinkingMap
	messageRepo repository.Message
}

// NewContextManager 创建新的上下文管理器实例
func NewContextManager(nodeRepo repository.ThinkingNode, mapRepo repository.ThinkingMap, messageRepo repository.Message) *ContextManager {
	return &ContextManager{
		nodeRepo:    nodeRepo,
		mapRepo:     mapRepo,
		messageRepo: messageRepo,
	}
}

// ContextInfo 节点完整上下文信息
type ContextInfo struct {
	MapInfo             *model.ThinkingMap     `json:"mapInfo"`
	NodeInfo            *model.ThinkingNode    `json:"nodeInfo"`
	AncestorsContext    []NodeContextInfo      `json:"ancestorsContext,omitempty"`
	DependencyContext   []NodeContextInfo      `json:"dependencyContext,omitempty"`
	ChildrenContext     []NodeContextInfo      `json:"childrenContext,omitempty"`
	ConversationContext []ConversationMessage  `json:"conversationContext,omitempty"`
	UserContext         map[string]interface{} `json:"userContext,omitempty"`
}

// NodeContextInfo 节点上下文信息
type NodeContextInfo struct {
	NodeID     string `json:"nodeID"`
	Question   string `json:"question"`
	Target     string `json:"target"`
	Conclusion string `json:"conclusion,omitempty"`
	Status     string `json:"status"`
}

// ConversationMessage 对话消息
type ConversationMessage struct {
	MessageID string    `json:"messageID"`
	ParentID  string    `json:"parentID"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// GetContextInfo 获取节点的完整上下文
func (cm *ContextManager) GetContextInfo(ctx context.Context, nodeID string) (*ContextInfo, error) {
	// 获取当前节点
	node, err := cm.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to find node %s: %w", nodeID, err)
	}

	// 获取导图信息
	mapInfo, err := cm.mapRepo.FindByID(ctx, node.MapID)
	if err != nil {
		return nil, fmt.Errorf("failed to find map %s: %w", node.MapID, err)
	}

	// 构建完整上下文
	contextInfo := &ContextInfo{
		MapInfo:  mapInfo,
		NodeInfo: node,
	}

	// 获取祖先节点上下文（递归到根节点）
	if ancestorsContext, err := cm.getAncestorsContext(ctx, nodeID); err == nil {
		contextInfo.AncestorsContext = ancestorsContext
	}

	// 获取依赖节点上下文
	if dependencyContext, err := cm.getDependencyContext(ctx, nodeID); err == nil {
		contextInfo.DependencyContext = dependencyContext
	}

	// 获取子节点上下文
	if childrenContext, err := cm.getChildrenContext(ctx, nodeID); err == nil {
		contextInfo.ChildrenContext = childrenContext
	}

	// 获取用户自定义上下文
	if userContext := cm.getUserContext(nodeID); userContext != nil {
		contextInfo.UserContext = userContext
	}

	return contextInfo, nil
}

// GetNodeContextWithConversation 获取包含对话历史的节点完整上下文
func (cm *ContextManager) GetNodeContextWithConversation(ctx context.Context, nodeID string, parentMsgID string) (*ContextInfo, error) {
	// 先获取基础上下文
	contextInfo, err := cm.GetContextInfo(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	// 获取对话历史上下文
	if conversationContext, err := cm.getConversationContext(ctx, nodeID, parentMsgID); err == nil {
		contextInfo.ConversationContext = conversationContext
	}

	return contextInfo, nil
}

// getAncestorsContext 递归获取祖先节点上下文（到根节点）
func (cm *ContextManager) getAncestorsContext(ctx context.Context, nodeID string) ([]NodeContextInfo, error) {
	var ancestors []NodeContextInfo

	currentNodeID := nodeID
	for {
		node, err := cm.nodeRepo.FindByID(ctx, currentNodeID)
		if err != nil {
			return ancestors, nil // 如果找不到节点，返回已收集的祖先
		}

		// 如果没有父节点，说明到达根节点
		if node.ParentID == "" || node.ParentID == uuid.Nil.String() {
			break
		}

		// 获取父节点信息
		parent, err := cm.nodeRepo.FindByID(ctx, node.ParentID)
		if err != nil {
			break // 如果找不到父节点，停止递归
		}

		// 添加父节点到祖先列表
		ancestors = append([]NodeContextInfo{{
			NodeID:     parent.ID,
			Question:   parent.Question,
			Target:     parent.Target,
			Conclusion: parent.Conclusion.Content,
			Status:     parent.Status,
		}}, ancestors...)

		// 继续向上查找
		currentNodeID = node.ParentID
	}

	return ancestors, nil
}

// getDependencyContext 获取同级依赖节点的问题目标和结论
func (cm *ContextManager) getDependencyContext(ctx context.Context, nodeID string) ([]NodeContextInfo, error) {
	node, err := cm.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	if len(node.Dependencies) == 0 {
		return nil, nil
	}

	// 获取依赖节点的详细信息
	dependencyNodes, err := cm.nodeRepo.FindByIDs(ctx, node.Dependencies)
	if err != nil {
		return nil, err
	}

	var dependencies []NodeContextInfo
	for _, depNode := range dependencyNodes {
		dependencies = append(dependencies, NodeContextInfo{
			NodeID:     depNode.ID,
			Question:   depNode.Question,
			Target:     depNode.Target,
			Conclusion: depNode.Conclusion.Content,
			Status:     depNode.Status,
		})
	}

	return dependencies, nil
}

// getChildrenContext 获取直接子节点的问题目标和结论
func (cm *ContextManager) getChildrenContext(ctx context.Context, nodeID string) ([]NodeContextInfo, error) {
	children, err := cm.nodeRepo.FindByParentID(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	var childrenContext []NodeContextInfo
	for _, child := range children {
		childrenContext = append(childrenContext, NodeContextInfo{
			NodeID:     child.ID,
			Question:   child.Question,
			Target:     child.Target,
			Conclusion: child.Conclusion.Content,
			Status:     child.Status,
		})
	}

	return childrenContext, nil
}

// getConversationContext 获取节点的对话历史上下文
func (cm *ContextManager) getConversationContext(ctx context.Context, nodeID string, parentMsgID string) ([]ConversationMessage, error) {
	if cm.messageRepo == nil {
		return nil, nil
	}

	// 获取最近的对话历史，用于问题拆解和结论生成的对话框交互
	recentMessages, err := cm.getRecentNodeConversation(ctx, nodeID, parentMsgID, 10)
	if err != nil {
		return nil, err
	}

	return recentMessages, nil
}

// getRecentNodeConversation 获取节点最近的对话记录
func (cm *ContextManager) getRecentNodeConversation(ctx context.Context, nodeID string, parentMsgID string, limit int) ([]ConversationMessage, error) {
	if parentMsgID == "" {
		return nil, nil
	}

	// 通过MessageService获取对话历史
	msgService := NewMessageManager(cm.messageRepo, cm.nodeRepo)
	messages, err := msgService.GetMessageChain(ctx, parentMsgID)
	if err != nil {
		return nil, err
	}

	// 转换为ConversationMessage格式，并限制数量
	var conversationMessages []ConversationMessage
	for i, msg := range messages {
		if i >= limit {
			break
		}
		conversationMessages = append(conversationMessages, ConversationMessage{
			MessageID: msg.ID,
			ParentID:  msg.ParentID,
			Role:      string(msg.Role),
			Content:   msg.Content.String(),
			Timestamp: msg.CreatedAt,
		})
	}

	return conversationMessages, nil
}

// getUserContext 获取用户手动添加或修改的上下文信息
func (cm *ContextManager) getUserContext(nodeID string) map[string]interface{} {
	// 这里可以从数据库或其他存储中获取用户自定义的上下文
	// 暂时返回空，具体实现根据需求定义
	return nil
}

// FormatContextForAgent 格式化上下文信息供Agent使用
func (cm *ContextManager) FormatContextForAgent(contextInfo *ContextInfo) string {
	if contextInfo == nil {
		return ""
	}

	// 添加导图信息
	prompt := fmt.Sprintf(`导图信息：
- 标题：%s
- 问题：%s
- 目标：%s
`,
		contextInfo.MapInfo.Title,
		contextInfo.MapInfo.Problem,
		contextInfo.MapInfo.Target)

	if len(contextInfo.MapInfo.Constraints) > 0 {
		prompt += fmt.Sprintf("- 约束条件：%v\n", contextInfo.MapInfo.Constraints)
	}

	// 添加当前节点信息
	prompt += fmt.Sprintf(`\n当前节点信息：
- 问题：%s
- 目标：%s
- 状态：%s
`,
		contextInfo.NodeInfo.Question,
		contextInfo.NodeInfo.Target,
		contextInfo.NodeInfo.Status)

	// 添加祖先节点上下文
	if len(contextInfo.AncestorsContext) > 0 {
		prompt += "\n祖先节点上下文：\n"
		for i, ancestor := range contextInfo.AncestorsContext {
			prompt += fmt.Sprintf("%d. 问题：%s，目标：%s，结论：%s，状态：%s\n",
				i+1, ancestor.Question, ancestor.Target, ancestor.Conclusion, ancestor.Status)
		}
	}

	// 添加依赖节点上下文
	if len(contextInfo.DependencyContext) > 0 {
		prompt += "\n依赖节点上下文：\n"
		for i, dep := range contextInfo.DependencyContext {
			prompt += fmt.Sprintf("%d. 问题：%s，目标：%s，结论：%s，状态：%s\n",
				i+1, dep.Question, dep.Target, dep.Conclusion, dep.Status)
		}
	}

	// 添加子节点上下文
	if len(contextInfo.ChildrenContext) > 0 {
		prompt += "\n子节点上下文：\n"
		for i, child := range contextInfo.ChildrenContext {
			prompt += fmt.Sprintf("%d. 问题：%s，目标：%s，结论：%s，状态：%s\n",
				i+1, child.Question, child.Target, child.Conclusion, child.Status)
		}
	}

	// 添加对话历史上下文
	if len(contextInfo.ConversationContext) > 0 {
		prompt += "\n对话历史：\n"
		for _, msg := range contextInfo.ConversationContext {
			prompt += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		}
	}
	return prompt
}

// UpdateNodeDependencies 更新节点依赖关系
func (cm *ContextManager) UpdateNodeDependencies(ctx context.Context, nodeID string, dependencies []string) error {
	node, err := cm.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return err
	}

	node.Dependencies = dependencies
	return cm.nodeRepo.Update(ctx, node)
}

// RefreshNodeContext 刷新节点上下文（重新计算所有上下文信息）
func (cm *ContextManager) RefreshNodeContext(ctx *gin.Context, nodeID string) (*dto.NodeResponse, error) {
	node, err := cm.nodeRepo.FindByID(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	// 重新计算上下文
	contextInfo, err := cm.GetContextInfo(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	// 更新节点的DependentContext
	node.Context = cm.convertToNodeDependentContext(contextInfo)

	// 保存更新
	if err := cm.nodeRepo.Update(ctx, node); err != nil {
		return nil, err
	}

	resp := dto.ToNodeResponse(node)
	return &resp, nil
}

// convertToNodeDependentContext 将ContextInfo转换为model.DependentContext
func (cm *ContextManager) convertToNodeDependentContext(contextInfo *ContextInfo) model.DependentContext {
	var dependentContext model.DependentContext

	// 转换祖先节点上下文
	for _, ancestor := range contextInfo.AncestorsContext {
		dependentContext.Ancestor = append(dependentContext.Ancestor, model.NodeContext{
			Question:   ancestor.Question,
			Target:     ancestor.Target,
			Conclusion: ancestor.Conclusion,
			Status:     ancestor.Status,
		})
	}

	// 转换依赖节点上下文为前置兄弟节点上下文
	for _, dep := range contextInfo.DependencyContext {
		dependentContext.PrevSibling = append(dependentContext.PrevSibling, model.NodeContext{
			Question:   dep.Question,
			Target:     dep.Target,
			Conclusion: dep.Conclusion,
			Status:     dep.Status,
		})
	}

	// 转换子节点上下文
	for _, child := range contextInfo.ChildrenContext {
		dependentContext.Children = append(dependentContext.Children, model.NodeContext{
			Question:   child.Question,
			Target:     child.Target,
			Conclusion: child.Conclusion,
			Status:     child.Status,
		})
	}

	return dependentContext
}
