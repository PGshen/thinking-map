/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/decompose-tab.tsx
 */
'use client';

import React, { useState, useEffect } from 'react';
import { GitBranch } from 'lucide-react';
import { MessageArea } from './message-area';
import { Action, MessageResponse, MessageType, MessageContent } from '@/types/message';
import { CustomNodeModel } from '@/types/node';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { toast } from 'sonner';
import { ChatInput, ChatInputTextArea, ChatInputSubmit } from '@/components/ui/chat-input';
import { getMessages, decomposition, resetDecomposition } from '@/api/node';
import { useSSEConnection } from '@/hooks/use-sse-connection';
import { MessageActionEvent, MessageNoticeEvent, MessagePlanEvent, MessageTextEvent, MessageThoughtEvent, MessageRagEvent } from '@/types/sse';
import { Button } from '@/components/ui/button';
import { AlertDialog, AlertDialogTrigger, AlertDialogContent, AlertDialogHeader, AlertDialogTitle, AlertDialogDescription, AlertDialogFooter, AlertDialogCancel, AlertDialogAction } from '@/components/ui/alert-dialog';
import { useExecutableNodes } from '@/features/workspace/hooks/use-executable-nodes';
import { RefreshCcw } from 'lucide-react';

interface DecomposeTabProps {
  nodeID: string;
  nodeData: CustomNodeModel;
}

export function DecomposeTab({ nodeID, nodeData }: DecomposeTabProps) {
  const { mapID } = useWorkspaceStore();
  const [messages, setMessages] = useState<MessageResponse[]>([]);
  const [isDecomposed, setIsDecomposed] = useState(false);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState("");
  const { fetchExecutableNodes } = useExecutableNodes();

  const { actions } = useWorkspaceStore();

  // 使用useEffect来同步messages到workspace store，避免在渲染期间更新状态
  // useEffect(() => {
  //   if (messages.length > 0) {
  //     actions.updateNodeDecomposition(nodeID, {
  //       messages: messages,
  //     });
  //   }
  // }, [messages, nodeID, actions]);

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

  // 处理消息通知事件
  const handleMessageNoticeEvent = (data: MessageNoticeEvent) => {
    handleMessageEvent(data, 'notice', (eventData) => ({
      notice: eventData.notice
    }));
  };

  // 处理消息操作事件
  const handleMessageActionEvent = (data: MessageActionEvent) => {
    handleMessageEvent(data, 'action', (eventData) => ({
      action: eventData.actions
    }));
  };

  // 处理消息文本或思考事件
  const handleMessageTextOrThoughtEvent = (
    data: MessageTextEvent | MessageThoughtEvent, 
    messageType: 'text' | 'thought'
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

  // 处理规划消息事件
  const handleMessagePlanEvent = (data: MessagePlanEvent) => {
    handleMessageEvent(data, 'plan', (eventData) => ({
      plan: eventData.plan
    }));
  };

  // 处理RAG消息事件
  const handleMessageRagEvent = (data: MessageRagEvent) => {
    handleMessageEvent(data, 'rag', (eventData) => ({
      rag: eventData.ragRecord
    }));
  };

  // SSE连接处理 - 将useSSEConnection移到组件顶层
  const sseCallbacks = React.useMemo(() => {
    if (!mapID) return [];

    return [
      {
        eventType: 'messageText' as const,
        callback: (event: any) => {
          try {
            const data = JSON.parse(event.data) as MessageTextEvent;
            handleMessageTextOrThoughtEvent(data, 'text');
          } catch (error) {
            console.error('解析messageText事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'messageThought' as const,
        callback: (event: any) => {
          try {
            const data = JSON.parse(event.data) as MessageThoughtEvent;
            handleMessageTextOrThoughtEvent(data, 'thought');
          } catch (error) {
            console.error('解析messageThought事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'messageNotice' as const,
        callback: (event: any) => {
          try {
            const data = JSON.parse(event.data) as MessageNoticeEvent;
            handleMessageNoticeEvent(data);
          } catch (error) {
            console.error('解析messageNotice事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'messageAction' as const,
        callback: (event: any) => {
          try {
            const data = JSON.parse(event.data) as MessageActionEvent;
            handleMessageActionEvent(data);
          } catch (error) {
            console.error('解析messageAction事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'messagePlan' as const,
        callback: (event: any) => {
          try {
            const data = JSON.parse(event.data) as MessagePlanEvent;
            handleMessagePlanEvent(data);
          } catch (error) {
            console.error('解析messagePlan事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'messageRag' as const,
        callback: (event: any) => {
          try {
            const data = JSON.parse(event.data) as MessageRagEvent;
            console.log("rag data", data)
            handleMessageRagEvent(data);
          } catch (error) {
            console.error('解析messageRag事件失败:', error, event.data);
          }
        }
      }
    ];
  }, [mapID]);

  // 在组件顶层调用useSSEConnection
  useSSEConnection({
    mapID: mapID || '',
    callbacks: sseCallbacks
  });

  // 初始化拆解流程步骤                                                                                                                                                                                                                                                                                                                                                                      
  useEffect(() => {
    const initializeDecomposition = async () => {
      console.log("mapID", mapID)
      if (!mapID) {
        return;
      }
      console.log("messages", nodeData.decomposition?.messages)
      // if (nodeData.decomposition?.messages === undefined) {
        if (loading) {
          return;
        }
        // 初始化加载，如果为空，从后端加载
        setLoading(true);
        try {
          let res = await getMessages(mapID, nodeID, 'decomposition');
          console.log("res", res);
          if (res.code !== 200) {
            toast.error(`加载失败: ${res.message}`);
            setLoading(false);
            return;
          }
          actions.updateNodeDecomposition(nodeID, {
            messages: res.data,
          });
          setMessages(res.data);
        } catch (error) {
          toast.error('网络错误，请重试');
          console.error('加载拆解消息失败', error);
        } finally {
          setLoading(false);
        }
      // } else {
      //   setMessages(nodeData.decomposition.messages);
      //   setIsDecomposed(nodeData.decomposition.isDecomposed);
      // }
    };

    initializeDecomposition();
  }, [nodeID]);

  // 下一步
  const nextStep = () => {
    // 获取可执行节点和建议节点
    fetchExecutableNodes(nodeID);
  }

  // 提交
  const handleSubmit = (inputValue: string, newIsDecomposed?: boolean) => {
    if (loading) {
      return;
    }

    setLoading(true);
    const currentIsDecomposed = newIsDecomposed ?? isDecomposed;
    try {
      decomposition(nodeID, inputValue, currentIsDecomposed).then(res => {
        console.log("res", res);
        if (res.code !== 200) {
          toast.error(`加载失败: ${res.message}`);
          setLoading(false);
          return;
        }
      }).finally(() => {
        setLoading(false);
      });
    } catch (error) {
      toast.error('网络错误，请重试');
      console.error('加载拆解消息失败', error);
    }
  };

  const handleSubmitMessage = () => {
    if (loading) {
      return;
    }
    if (inputValue.trim() === "") {
      toast('请输入消息');
      return;
    }
    const userMessage: MessageResponse = {
      id: `user-${Date.now()}`,
      role: 'user',
      messageType: 'text',
      content: {
        text: inputValue
      }
    };
    const newMessages = [...messages, userMessage];
    setMessages(newMessages);
    handleSubmit(inputValue);
    setInputValue('');
  };

  const handleReset = async () => {
    if (!mapID) return;
    if (loading) return;
    setLoading(true);
    try {
      const res = await resetDecomposition(mapID, nodeID);
      if (res.code !== 200) {
        toast.error(`重置失败: ${res.message}`);
        return;
      }
      setMessages([]);
      setIsDecomposed(false);
      toast.success('重置成功');
    } catch (error) {
      toast.error('网络错误，请重试');
      console.error('重置拆解失败', error);
    } finally {
      setLoading(false);
    }
  };

  function clickAction(action: Action): void {
    if (action.name == '开始拆解') {
      setIsDecomposed(true);
      handleSubmit('开始拆解', true);
    } else if (action.name == '开始结论') {
      handleSubmit('开始结论');
    } else {
      throw new Error('Function not implemented.');
    }
  }

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 min-h-0 overflow-hidden">
        <MessageArea loading={loading} messages={messages} clickAction={clickAction} />
      </div>

      {/* 固定在底部的输入区域 */}
      <div className="flex-shrink-0 px-2 py-4 max-w-2xl mx-auto w-full">
        <ChatInput
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onSubmit={handleSubmitMessage}
          loading={loading}
          onStop={() => setLoading(false)}
        >
          <ChatInputTextArea variant='unstyled' placeholder="Type a message..." />
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => nextStep()}
              className="rounded-full! cursor-pointer"
            >
              下一步
            </Button>
            {/* 重置拆解按钮（确认弹窗） */}
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  className="rounded-full! cursor-pointer"
                >
                  <RefreshCcw className="w-3 h-3" />
                  重置拆解
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>确认重置</AlertDialogTitle>
                  <AlertDialogDescription>
                    确定要重置当前拆解吗？此操作将清空消息，且不可撤销。
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>取消</AlertDialogCancel>
                  <AlertDialogAction onClick={() => handleReset()}>
                    确认
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
            {/* 拆解识别按钮 */}
            <Button
              variant="outline"
              size="sm"
              disabled={nodeData.decomposition?.isDecomposed}
              onClick={() => handleSubmit("")}
              className="rounded-full! cursor-pointer"
            >
              <GitBranch className="w-3 h-3" />
              拆解分析
            </Button>
            <ChatInputSubmit />
          </div>
        </ChatInput>
      </div>
    </div>
  );
}

export default DecomposeTab;