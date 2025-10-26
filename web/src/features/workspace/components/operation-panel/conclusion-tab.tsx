/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/conclusion-tab.tsx
 */
'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Save, RotateCcw, CheckCircle, AlertCircle, Clock, FileText, ChevronDown, ChevronUp, Edit, Eye, Square } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Card, CardContent } from '@/components/ui/card';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { toast } from 'sonner';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { conclusion } from '@/api/node';
import { SseTextStreamParser } from '@/lib/sse-parser';
import { getMessages, decomposition } from '@/api/node';

// 导入新的Notion编辑器
import EditorClient from '@/components/editor-client';
import { MessageResponse, MessageType, MessageContent } from '@/types/message';
import { MessageConclusionEvent, MessageTextEvent, MessageThoughtEvent } from '@/types/sse';
import { useSSEConnection } from '@/hooks/use-sse-connection';

interface ConclusionTabProps {
  nodeID: string;
  node: any; // TODO: 使用正确的节点类型
}

interface ExecutionLog {
  id: string;
  timestamp: string;
  type: 'info' | 'success' | 'warning' | 'error';
  message: string;
  details?: string;
}



export function ConclusionTab({ nodeID, node }: ConclusionTabProps) {
  const nodeData = node.data as any;
  const { mapID } = useWorkspaceStore();
  const [messages, setMessages] = useState<MessageResponse[]>([]);
  const [hasChanges, setHasChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [content, setContent] = useState(nodeData?.conclusion?.content || '');
  const [instruction, setInstruction] = useState('');
  const [reference, setReference] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  // optimize 模式的流式处理状态
  const [optimizeState, setOptimizeState] = useState<{
    isOptimizing: boolean;
    originalContent: string;
    referenceStartIndex: number;
    referenceEndIndex: number;
    accumulatedOptimizedContent: string;
  }>({
    isOptimizing: false,
    originalContent: '',
    referenceStartIndex: -1,
    referenceEndIndex: -1,
    accumulatedOptimizedContent: ''
  });
  const { actions } = useWorkspaceStore();

  // optimize 流式处理结束检测
  const optimizeTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // 处理编辑器内容变化
  const handleEditorChange = useCallback((content: string) => {
    setContent(content);
    setHasChanges(content !== (nodeData?.conclusion || ''));
  }, [nodeData?.conclusion]);


  
  // 监听节点数据变化
  useEffect(() => {
    if (nodeData) {
      setContent(nodeData.conclusion?.content || '');
      setHasChanges(false);
    }
  }, [node]);

  // 清理定时器
  useEffect(() => {
    return () => {
      if (optimizeTimeoutRef.current) {
        clearTimeout(optimizeTimeoutRef.current);
      }
    };
  }, []);

  // 通用消息处理函数
  const handleMessageEvent = <T extends { messageID: string; timestamp: string }>(
    data: T,
    messageType: MessageType,
    contentUpdater: (data: T, existingContent?: MessageContent, mode?: string) => Partial<MessageContent>
  ) => {
    setMessages(prevMessages => {
      const existingMessageIndex = prevMessages.findIndex(msg => msg.id === data.messageID);
      let updatedMessages: MessageResponse[];

      if (existingMessageIndex !== -1) {
        // 消息已存在，更新内容
        updatedMessages = [...prevMessages];
        const existingMessage = updatedMessages[existingMessageIndex];
        const mode = 'mode' in data ? (data as any).mode : undefined;
        
        const updatedContent = contentUpdater(data, existingMessage.content, mode);
        
        updatedMessages[existingMessageIndex] = {
          ...existingMessage,
          content: {
            ...existingMessage.content,
            ...updatedContent
          },
          updatedAt: data.timestamp
        };
      } else {
        // 消息不存在，创建新消息
        const newContent = contentUpdater(data);
        const newMessage: MessageResponse = {
          id: data.messageID,
          messageType,
          role: 'assistant',
          content: newContent,
          createdAt: data.timestamp,
          updatedAt: data.timestamp
        };

        updatedMessages = [...prevMessages, newMessage];
      }

      return updatedMessages;
    });
  };

  // 处理消息文本或思考事件
  const handleMessageThoughtEvent = (
    data: MessageThoughtEvent, 
    messageType: 'thought'
  ) => {
    if (data.message == "") {
      return;
    }
    handleMessageEvent(data, messageType, (eventData, existingContent, mode) => {
      if (mode === 'replace') {
        // 替换模式：直接替换文本内容
        return {
          [messageType]: eventData.message
        };
      } else if (mode === 'append') {
        // 追加模式：在现有文本后追加
        const currentText = existingContent?.[messageType] || '';
        return {
          [messageType]: currentText + eventData.message
        };
      } else {
        // 默认为替换模式
        return {
          [messageType]: eventData.message
        };
      }
    });
  };

  // 处理结论消息
  const handleMessageConclusionEvent = (data: MessageConclusionEvent) => {
    if (data.mode === 'generate') {
      // generate 模式：累积接收到的字符到结论 content
      setContent((prevContent: string) => prevContent + data.message)
    } else if (data.mode === 'optimize') {
      // optimize 模式：流式处理优化内容
      
      // 清除之前的超时定时器
      if (optimizeTimeoutRef.current) {
        clearTimeout(optimizeTimeoutRef.current);
      }
      
      setOptimizeState((prevState) => {
        if (!prevState.isOptimizing) {
          // 第一次接收到 optimize 消息，初始化状态
          const currentContent = content;
          if (reference && reference.trim() !== '') {
            const referenceStartIndex = currentContent.indexOf(reference);
            if (referenceStartIndex !== -1) {
              const referenceEndIndex = referenceStartIndex + reference.length;
              const newAccumulatedContent = data.message;
              
              // 立即更新 content
              const beforeReference = currentContent.substring(0, referenceStartIndex);
              const afterReference = currentContent.substring(referenceEndIndex);
              const newContent = beforeReference + newAccumulatedContent + afterReference;
              setContent(newContent);
              
              return {
                isOptimizing: true,
                originalContent: currentContent,
                referenceStartIndex,
                referenceEndIndex,
                accumulatedOptimizedContent: newAccumulatedContent
              };
            }
          }
          // 如果没有找到 reference，则直接追加
          setContent((prevContent: string) => prevContent + data.message);
          return prevState;
        } else {
          // 后续消息，累积优化内容并更新 content
          const newAccumulatedContent = prevState.accumulatedOptimizedContent + data.message;
          
          // 更新 content
          const beforeReference = prevState.originalContent.substring(0, prevState.referenceStartIndex);
          const afterReference = prevState.originalContent.substring(prevState.referenceEndIndex);
          const newContent = beforeReference + newAccumulatedContent + afterReference;
          setContent(newContent);
          
          return {
            ...prevState,
            accumulatedOptimizedContent: newAccumulatedContent
          };
        }
      });

      // 设置新的超时定时器，在 1 秒后重置 optimize 状态
      optimizeTimeoutRef.current = setTimeout(() => {
        setOptimizeState((prevState) => {
          if (prevState.isOptimizing) {
            // 更新 reference 为最终的优化内容
            setReference(prevState.accumulatedOptimizedContent);
            
            return {
              isOptimizing: false,
              originalContent: '',
              referenceStartIndex: -1,
              referenceEndIndex: -1,
              accumulatedOptimizedContent: ''
            };
          }
          return prevState;
        });
      }, 1000);
    } else {
      // 其他模式：追加到内容末尾
      setContent((prevContent: string) => prevContent + "\n" + data.message)
    }
  }

  const sseCallbacks =  React.useMemo(() => {
    if (!mapID) return [];

    return [
      {
        eventType: 'messageConclusion' as const,
        callback: (event: any) => {
          try {
            const data = JSON.parse(event.data) as MessageConclusionEvent;
            handleMessageConclusionEvent(data);
          } catch (error) {
            console.error('解析messageConclusion事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'messageThought' as const,
        callback: (event: any) => {
          try {
            const data = JSON.parse(event.data) as MessageThoughtEvent;
            handleMessageThoughtEvent(data, 'thought');
          } catch (error) {
            console.error('解析messageThought事件失败:', error, event.data);
          }
        }
      },
    ]
  }, [mapID])

  // 在组件顶层调用useSSEConnection
  useSSEConnection({
    mapID: mapID || '',
    callbacks: sseCallbacks
  });

  const handleSave = async () => {
    if (!hasChanges) return;
    
    setIsSaving(true);
    try {
      // TODO: 调用API保存结论
      // await updateNodeConclusion(nodeID, editorContent);
      
      // 更新本地状态
      const nodeData = node.data as any;
      const currentStatus = nodeData?.status || 'pending';
      actions.updateNode(nodeID, { 
        data: {
          ...nodeData,
          conclusion: {
            ...nodeData?.conclusion,
            content: content,
          },
          status: content.trim() ? 'completed' : currentStatus
        }
      });
      
      toast.success('结论已保存');
      
      setHasChanges(false);
    } catch (error) {
      toast.error('保存失败，请重试');
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    const nodeData = node.data as any;
    setContent(nodeData?.conclusion?.content || '');
    setHasChanges(false);
  };

  // 停止结论生成
  const handleStopConclusion = () => {
    setIsGenerating(false);
    
    // 更新节点状态
    const nodeData = node.data as any;
    actions.updateNode(nodeID, { 
      data: {
        ...nodeData,
        status: 'pending'
      }
    });
    
    toast.info('结论生成已停止');
  };

  // 开始结论生成
  const handleStartConclusion = async () => {
    if (!nodeID) {
      toast.error('请先选择一个节点');
      return;
    }

    setIsGenerating(true);
    
    try {
      conclusion(nodeID, reference, instruction).then(res => {
        console.log("res", res)
        if (res.code !== 200) {
          console.error('启动结论生成失败:', res.message);
          toast.error('启动结论生成失败');
          
          setIsGenerating(false);
        }
      }).finally(() => {
        setIsGenerating(false);
        setReference('')
        setInstruction('')
      })
    } catch (error) {
      console.error('启动结论生成失败:', error);
      toast.error('启动结论生成失败');
    }
  };

  return (
    <div className="h-full pt-1 relative">
      {/* 内容区域：结论编辑器 */}
      <div className="h-full pb-20 overflow-auto">
        {/* Notion风格编辑器 */}
        <div className="h-full border rounded-md overflow-auto">
          <EditorClient
            initContent={content}
            placeholder="请输入结论..."
            onChange={handleEditorChange}
            editable={isEditing && (nodeData?.status || 'pending') !== 'running'}
            className={`h-full ${isEditing ? 'p-4' : 'px-2 py-4'}`}
            hideToolbar={!isEditing}
            isEditing={isEditing}
          />
        </div>
      </div>

      {/* 日志区域：展示思考过程 */}
      
      {/* 固定在底部的操作按钮区域 */}
      {node.status !== 'running' && (
        <div className="absolute bottom-0 left-0 right-0 pb-1 border-t bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
          <div className="flex gap-2 p-3">
            <Button
              onClick={() => setIsEditing(!isEditing)}
              variant={isEditing ? "default" : "outline"}
              size="sm"
              className="cursor-pointer"
            >
              {isEditing ? <Eye className="w-4 h-4 mr-2" /> : <Edit className="w-4 h-4 mr-2" />}
              {isEditing ? '预览' : '编辑'}
            </Button>
            
            <Button
              onClick={handleReset}
              disabled={!hasChanges}
              variant="outline"
              className="flex-1 cursor-pointer"
              size="sm"
            >
              <RotateCcw className="w-4 h-4 mr-2" />
              重置
            </Button>

            {isGenerating ? (
              <Button
                onClick={handleStopConclusion}
                variant="destructive"
                className="flex-1 cursor-pointer"
                size="sm"
              >
                <Square className="w-4 h-4 mr-2" />
                停止生成
              </Button>
            ) : (
              <Button
                onClick={handleStartConclusion}
                disabled={!nodeID}
                variant="default"
                className="flex-1 cursor-pointer"
                size="sm"
              >
                <CheckCircle className="w-4 h-4 mr-2" />
                开始结论
              </Button>
            )}

            <Button
              onClick={handleSave}
              disabled={!hasChanges || isSaving}
              className="flex-1 cursor-pointer"
              size="sm"
            >
              <Save className="w-4 h-4 mr-2" />
              {isSaving ? '保存中...' : '保存结论'}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

export default ConclusionTab;