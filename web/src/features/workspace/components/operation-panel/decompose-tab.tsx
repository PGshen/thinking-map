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
import { MessageResponse } from '@/types/message';
import { CustomNodeModel } from '@/types/node';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { toast } from 'sonner';
import { ChatInput, ChatInputTextArea, ChatInputSubmit } from '@/components/ui/chat-input';
import { getMessages } from '@/api/node';

interface DecomposeTabProps {
  nodeID: string;
  nodeData: CustomNodeModel;
}

export function DecomposeTab({ nodeID, nodeData }: DecomposeTabProps) {
  const { mapID } = useWorkspaceStore();
  const [messages, setMessages] = useState<MessageResponse[]>([]);
  const [isDecomposing, setIsDecomposing] = useState(false);
  const [progress, setProgress] = useState(0);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState("");

  const { actions } = useWorkspaceStore();

  // 初始化拆解流程步骤                                                                                                                                                                                                                                                                                                                                                                      
  useEffect(() => {
    const initializeDecomposition = async () => {
      console.log("mapID", mapID)
      if (!mapID) {
        return;
      }
      console.log("messages", nodeData.decomposition?.messages)
      if (nodeData.decomposition?.messages === undefined) {
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
      }
    };

    initializeDecomposition();
  }, [nodeData]);

  // 开始拆解流程
  const handleStartDecompose = async () => {
    setIsDecomposing(true);
    setProgress(0);

    try {
      // 步骤1: RAG检索
      addSystemMessage('🔍 开始RAG知识检索...');
      setProgress(20);

      // 模拟RAG检索
      await new Promise(resolve => setTimeout(resolve, 1500));
      addSystemMessage('✅ RAG检索完成，找到相关知识');
      setProgress(40);

      // 步骤2: AI分析
      addSystemMessage('🤖 AI正在分析问题...');
      setProgress(60);

      // 模拟AI分析
      await new Promise(resolve => setTimeout(resolve, 2000));

      // 添加AI分析结果
      const analysisMessage: MessageResponse = {
        id: `analysis-${Date.now()}`,
        role: 'assistant',
        messageType: 'text',
        content: {
          text: `基于RAG检索的知识，我建议将"${nodeData?.question || '当前问题'}"拆解为以下几个子问题：\n1. 需求分析与用户研究\n2. 技术方案设计\n3. 实现与测试\n4. 部署与维护\n您可以通过对话调整这些建议，或者直接确认创建子节点。`
        }
      };
      setMessages(prev => [...prev, analysisMessage]);

      const analysisMessage2: MessageResponse = {
        id: `analysis2-${Date.now()}`,
        role: 'assistant',
        messageType: 'text',
        content: {
          text: `基于RAG检索的知识，我建议将"${nodeData?.question || '当前问题'}"拆解为以下几个子问题：\n\n1. 需求分析与用户研究\n2. 技术方案设计\n3. 实现与测试\n4. 部署与维护\n\n您可以通过对话调整这些建议，或者直接确认创建子节点。`
        }
      };
      setMessages(prev => [...prev, analysisMessage2]);
      setProgress(80);

      // 步骤3: 节点创建准备
      addSystemMessage('📝 子问题建议已生成，等待您的确认');
      setProgress(100);
    } catch (error) {
      toast('拆解过程中出现错误，请重试');
    } finally {
      setIsDecomposing(false);
    }
  };

  // 添加系统消息
  const addSystemMessage = (content: string) => {
    const systemMessage: MessageResponse = {
      id: `system-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      role: 'system',
      messageType: 'notice',
      content: {
        notice: [{
          type: '',
          name: 'system_notification',
          content: content
        }]
      }
    };
    setMessages(prev => [...prev, systemMessage]);
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
    setMessages(prev => [...prev, userMessage]);
    actions.updateNodeDecomposition(nodeID, {
      messages: [...messages, userMessage],
    });
    setInputValue('');
    // handleSubmit();
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 min-h-0 overflow-hidden">
        <DecomposeArea messages={messages} />
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
            {/* 开始拆解按钮 */}
            {!isDecomposing && (
              <button
                onClick={handleStartDecompose}
                className="px-3 py-1.5 bg-primary cursor-pointer text-primary-foreground rounded-full text-sm font-medium hover:bg-primary/90 transition-colors flex items-center gap-1.5 shrink-0"
              >
                <GitBranch className="w-3 h-3" />
                拆解
              </button>
            )}
            <ChatInputSubmit />
          </div>
        </ChatInput>
      </div>
    </div>
  );
}

export default DecomposeTab;