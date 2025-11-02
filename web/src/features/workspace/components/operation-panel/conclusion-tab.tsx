/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/conclusion-tab.tsx
 */
'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Save, RotateCcw, CheckCircle, FileText, ChevronDown, Edit, Eye, Square, Brain, ChevronRight, Sparkle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ChatMessageArea } from '@/components/ui/chat-message-area';
import { ChatMessage, ChatMessageContent } from '@/components/ui/chat-message';
import { toast } from 'sonner';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { conclusion } from '@/api/node';
import { getMessages, saveNodeConclusion, resetNodeConclusion } from '@/api/node';

// 导入新的Notion编辑器
import EditorClient from '@/components/editor-client';
import { MessageResponse, MessageType, MessageContent } from '@/types/message';
import { MessageConclusionEvent, MessageThoughtEvent } from '@/types/sse';
import { useSSEConnection } from '@/hooks/use-sse-connection';
import { CustomNodeModel } from '@/types/node';
import { Node } from 'reactflow';
import { MarkdownContent } from '@/components/ui/markdown-content';

interface ConclusionTabProps {
  nodeID: string;
  node: Node<CustomNodeModel>;
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
  const [activeTab, setActiveTab] = useState<'thinking' | 'conclusion'>('thinking');
  const [collapsedStates, setCollapsedStates] = useState<Record<string, boolean>>({});
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
  const handleEditorChange = useCallback((newContent: string) => {
    // 只有当内容真正发生变化时才更新状态
    setContent((prevContent: string) => {
      if (prevContent !== newContent) {
        setHasChanges(true)
        return newContent
      }
      return prevContent
    })
  }, []);

  // 监听节点数据变化
  useEffect(() => {
    if (nodeData) {
      const nodeConclusion = nodeData.conclusion?.content || '';
      // 直接设置内容，让编辑器的内部同步机制处理
      setContent(nodeConclusion);
      setHasChanges(false);
    }
  }, [nodeData?.conclusion?.content]); // 更精确的依赖

  // 初始化加载消息和结论
  const [loading, setLoading] = useState(false);
  
  useEffect(() => {
    const initializeConclusion = async () => {
      console.log("mapID", mapID);
      if (!mapID) {
        return;
      }
      console.log("messages", nodeData.conclusion?.messages);
      if (nodeData.conclusion?.messages === undefined) {
        if (loading) {
          return;
        }
        // 初始化加载，如果为空，从后端加载
        setLoading(true);
        try {
          let res = await getMessages(mapID, nodeID, 'conclusion');
          console.log("res", res);
          if (res.code !== 200) {
            toast.error(`加载失败: ${res.message}`);
            setLoading(false);
            return;
          }
          actions.updateNodeConclusion(nodeID, {
            messages: res.data,
          });
          setMessages(res.data);
        } catch (error) {
          toast.error('网络错误，请重试');
          console.error('加载结论消息失败', error);
        } finally {
          setLoading(false);
        }
      } else {
        setMessages(nodeData.conclusion.messages);
        // 如果有保存的结论内容，也要加载
        if (nodeData.conclusion.content) {
          setContent(nodeData.conclusion.content);
        }
      }
    };

    initializeConclusion();
   }, [nodeID]);

   // 使用useEffect来同步messages到workspace store，避免在渲染期间更新状态
  //  useEffect(() => {
  //    if (messages.length > 0) {
  //      actions.updateNodeConclusion(nodeID, {
  //        messages: messages,
  //      });
  //    }
  //  }, [messages, nodeID, actions]);


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
    // 自动切换到思考tab
    setActiveTab('thinking');
    
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
    if (data.message == '') {
      return
    }
    // 自动切换到结论tab
    setActiveTab('conclusion');
    
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
    if (!hasChanges || !mapID) return;
    
    setIsSaving(true);
    try {
      const res = await saveNodeConclusion(mapID, nodeID, content);
      if (res.code !== 200) {
        throw new Error(res.message);
      }
      
      // 更新本地状态
      const nodeData = node.data as any;
      const currentStatus = nodeData?.status || 'pending';
      actions.updateNode(nodeID, { 
        data: {
          ...nodeData,
          conclusion: {
            ...nodeData?.conclusion,
            messages: messages,
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

  const handleReset = async () => {
    if (!mapID) return;
    try {
      const res = await resetNodeConclusion(mapID, nodeID);
      if (res.code !== 200) {
        throw new Error(res.message);
      }
      
      // 重置本地状态
      setContent('');
      setMessages([]);
      actions.updateNode(nodeID, { 
        data: { 
          ...nodeData, 
          conclusion: {
            ...nodeData?.conclusion,
            content: '',
          }
        }
      });
      
      toast.success('结论已重置');
      setHasChanges(false);
    } catch (error) {
      toast.error('重置失败，请重试');
    }
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
    // 切换到预览状态
    setIsEditing(false);
    
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
  
  // 切换指定消息的折叠状态
  const toggleCollapse = (messageId: string) => {
    setCollapsedStates(prev => ({
      ...prev,
      [messageId]: !prev[messageId]
    }));
  };

  return (
    <div className="flex flex-col h-full">
      {/* Tab布局：思考日志和结论编辑器 */}
      <div className="flex-1 min-h-0 overflow-hidden">
        <Tabs value={activeTab} onValueChange={(value) => setActiveTab(value as 'thinking' | 'conclusion')} orientation='vertical' className="h-full flex">
          <TabsList className="flex flex-col h-fit w-auto p-1 gap-1">
            <TabsTrigger value="thinking" className="flex items-center gap-2 w-full justify-start px-3 py-2 text-sm cursor-pointer">
              <Brain className="w-4 h-4" />
            </TabsTrigger>
            <TabsTrigger value="conclusion" className="flex items-center gap-2 w-full justify-start px-3 py-2 text-sm cursor-pointer">
              <FileText className="w-4 h-4" />
            </TabsTrigger>
          </TabsList>
          
          <TabsContent value="thinking" className="flex-1 ml-2 mt-0 data-[state=inactive]:hidden">
            <div className="h-full">
              <ChatMessageArea scrollButtonAlignment="center" className="px-2 py-2 space-y-4 text-sm">
                 {messages.map((message) => {
                    if (message.content.thought === undefined) {
                      return
                    }
                    const isCollapsed = collapsedStates[message.id] || false;

                    return (
                      <ChatMessage
                        key={message.id}
                        id={message.id}
                        type="incoming"
                      >
                        {/* <ChatMessageAvatar /> */}
                        <div className="flex-1">
                          <Collapsible open={!isCollapsed} onOpenChange={() => toggleCollapse(message.id)}>
                            <CollapsibleTrigger className="cursor-pointer flex items-center gap-2 text-left p-1 rounded-md bg-blue-50 hover:bg-blue-100 transition-colors border border-blue-200">
                              <Sparkle className="h-4 w-4 text-blue-600 flex-shrink-0" />
                              <span className="text-sm font-medium text-blue-800 flex-1">思考过程</span>
                              {isCollapsed ? (
                                <ChevronRight className="h-4 w-4 text-blue-600 flex-shrink-0" />
                              ) : (
                                <ChevronDown className="h-4 w-4 text-blue-600 flex-shrink-0" />
                              )}
                            </CollapsibleTrigger>
                            <CollapsibleContent className="mt-2">
                              <div className="pl-2 border-l-2 border-blue-200">
                                <ChatMessageContent content={message.content.thought!} />
                              </div>
                            </CollapsibleContent>
                          </Collapsible>
                        </div>
                      </ChatMessage>
                    );
                 })}
               </ChatMessageArea>
            </div>
          </TabsContent>
          
          <TabsContent value="conclusion" className="flex-1 ml-2 mt-0 data-[state=inactive]:hidden">
            <div className="h-full overflow-auto">
              {isEditing ? (
                <EditorClient
                  initContent={content}
                  placeholder="请输入结论..."
                  onChange={handleEditorChange}
                  editable={(nodeData?.status || 'pending') !== 'running'}
                  className="h-full p-4"
                  hideToolbar={false}
                  isEditing={true}
                />
              ) : (
                <div className="h-full px-2 py-4">
                  <MarkdownContent 
                    id={`conclusion-${nodeID}`}
                    content={content || ''}
                    className="text-sm"
                  />
                </div>
              )}
            </div>
          </TabsContent>
        </Tabs>
      </div>
      
      {/* 固定在底部的操作按钮区域 */}
      <div className="flex-shrink-0 px-2 py-4 max-w-2xl mx-auto w-full">
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
              disabled={!nodeID || (nodeData?.conclusion?.content && nodeData.conclusion.content.trim() !== '')}
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
    </div>
  );
}

export default ConclusionTab;