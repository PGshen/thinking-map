/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/decompose-tab.tsx
 */
'use client';

import React, { useState, useEffect } from 'react';
import { GitBranch } from 'lucide-react';
import { DecomposeArea } from './decompose-area';
import { Action, MessageResponse } from '@/types/message';
import { CustomNodeModel } from '@/types/node';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { toast } from 'sonner';
import { ChatInput, ChatInputTextArea, ChatInputSubmit } from '@/components/ui/chat-input';
import { getMessages, decomposition } from '@/api/node';
import { useSSEConnection } from '@/hooks/use-sse-connection';
import { MessageActionEvent, MessagePlanEvent, MessageTextEvent, MessageThoughtEvent } from '@/types/sse';

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

  const { actions } = useWorkspaceStore();

  // 使用useEffect来同步messages到workspace store，避免在渲染期间更新状态
  useEffect(() => {
    if (messages.length > 0) {
      actions.updateNodeDecomposition(nodeID, {
        messages: messages,
      });
    }
  }, [messages, nodeID, actions]);

  // 处理消息操作事件
  const handleMessageActionEvent = (data: MessageActionEvent) => {
    setMessages(prevMessages => {
      const existingMessageIndex = prevMessages.findIndex(msg => msg.id === data.messageID);
      let updatedMessages: MessageResponse[];

      if (existingMessageIndex !== -1) {
        // 消息已存在，更新操作内容
        updatedMessages = [...prevMessages];
        const existingMessage = updatedMessages[existingMessageIndex];

        updatedMessages[existingMessageIndex] = {
          ...existingMessage,
          content: {
            ...existingMessage.content,
            action: data.actions
          },
          updatedAt: data.timestamp
        };
      } else {
        // 消息不存在，创建新的操作消息
        const newMessage: MessageResponse = {
          id: data.messageID,
          messageType: 'action',
          role: 'assistant',
          content: {
            action: data.actions
          },
          createdAt: data.timestamp,
          updatedAt: data.timestamp
        };

        updatedMessages = [...prevMessages, newMessage];
      }

      return updatedMessages;
    });
  };

  // 处理消息文本事件
  const handleMessageTextOrThoughtEvent = (data: MessageTextEvent|MessageThoughtEvent, messageType: 'text' | 'thought') => {
    setMessages(prevMessages => {
      const existingMessageIndex = prevMessages.findIndex(msg => msg.id === data.messageID);
      let updatedMessages: MessageResponse[];

      if (existingMessageIndex !== -1) {
        // 消息已存在，根据mode更新
        updatedMessages = [...prevMessages];
        const existingMessage = updatedMessages[existingMessageIndex];

        if (data.mode === 'replace') {
          // 替换模式：直接替换文本内容
          updatedMessages[existingMessageIndex] = {
            ...existingMessage,
            content: {
              ...existingMessage.content,
              [messageType]: data.message
            },
            updatedAt: data.timestamp
          };
        } else if (data.mode === 'append') {
          // 追加模式：在现有文本后追加
          const currentText = existingMessage.content[messageType] || '';
          updatedMessages[existingMessageIndex] = {
            ...existingMessage,
            content: {
              ...existingMessage.content,
              [messageType]: currentText + data.message
            },
            updatedAt: data.timestamp
          };
        }
      } else {
        // 消息不存在，创建新消息
        const newMessage: MessageResponse = {
          id: data.messageID,
          messageType: messageType,
          role: 'assistant',
          content: {
            [messageType]: data.message
          },
          createdAt: data.timestamp,
          updatedAt: data.timestamp
        };

        updatedMessages = [...prevMessages, newMessage];
      }

      return updatedMessages;
    });
  };

  // 处理规划消息
  const handleMessagePlanEvent = (data: MessagePlanEvent) => {
    setMessages(prevMessages => {
      const existingMessageIndex = prevMessages.findIndex(msg => msg.id === data.messageID);
      let updatedMessages: MessageResponse[];

      if (existingMessageIndex !== -1) {
        // 消息已存在，直接覆盖更新 plan 内容
        updatedMessages = [...prevMessages];
        const existingMessage = updatedMessages[existingMessageIndex];
        
        updatedMessages[existingMessageIndex] = {
          ...existingMessage,
          content: {
            ...existingMessage.content,
            plan: data.plan
          },
          updatedAt: data.timestamp
        };
      } else {
        // 消息不存在，创建新的规划消息
        const newMessage: MessageResponse = {
          id: data.messageID,
          messageType: 'plan',
          role: 'assistant',
          content: {
            plan: data.plan
          },
          createdAt: data.timestamp,
          updatedAt: data.timestamp
        };

        updatedMessages = [...prevMessages, newMessage];
      }

      return updatedMessages;
    });
  }

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
      if (nodeData.decomposition?.messages === undefined) {
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
      } else {
        setMessages(nodeData.decomposition.messages);
        setIsDecomposed(nodeData.decomposition.isDecomposed);
      }
    };

    initializeDecomposition();
  }, []);

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
        <DecomposeArea loading={loading} messages={messages} clickAction={clickAction} />
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
            {/* 拆解识别按钮 */}
            <button
              onClick={() => handleSubmit("")}
              className="px-3 py-1.5 bg-primary cursor-pointer text-primary-foreground rounded-full text-sm font-medium hover:bg-primary/90 transition-colors flex items-center gap-1.5 shrink-0"
            >
              <GitBranch className="w-3 h-3" />
              拆解识别
            </button>
            <ChatInputSubmit />
          </div>
        </ChatInput>
      </div>
    </div>
  );
}

export default DecomposeTab;