/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/decompose-tab.tsx
 */
'use client';

import React, { useState, useEffect } from 'react';
import { GitBranch, Loader2, CheckCircle, Clock, AlertCircle } from 'lucide-react';
import { DecomposeArea } from './decompose-area';
import { ChatMsg } from '@/types/message';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { toast } from 'sonner';
import { ChatInput, ChatInputTextArea, ChatInputSubmit } from '@/components/ui/chat-input';

interface DecomposeTabProps {
  nodeID: string;
  node: any; // TODO: 使用正确的节点类型
}

interface DecomposeStep {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  description: string;
}

interface SubProblem {
  id: string;
  title: string;
  description: string;
  status: 'suggested' | 'confirmed' | 'rejected';
}

export function DecomposeTab({ nodeID, node }: DecomposeTabProps) {
  const [messages, setMessages] = useState<ChatMsg[]>([]);
  const [isDecomposing, setIsDecomposing] = useState(false);
  const [decomposeSteps, setDecomposeSteps] = useState<DecomposeStep[]>([]);
  const [progress, setProgress] = useState(0);
  const [loading, setLoading] = useState(false);
  const [inputValue, setInputValue] = useState("");

  const { actions } = useWorkspaceStore();

  // 初始化拆解流程步骤
  useEffect(() => {
    const steps: DecomposeStep[] = [
      {
        id: 'rag-search',
        name: 'RAG知识检索',
        status: 'pending',
        description: '搜索相关知识和案例'
      },
      {
        id: 'ai-analysis',
        name: 'AI分析',
        status: 'pending',
        description: '分析问题并生成拆解建议'
      },
      {
        id: 'node-creation',
        name: '节点创建',
        status: 'pending',
        description: '创建子问题节点'
      }
    ];
    setDecomposeSteps(steps);

    // 初始化消息
    const initialMessages: ChatMsg[] = [
      {
        type: 'text',
        textMsg: {
          id: 'welcome',
          role: 'assistant',
          content: '我将帮您分析这个问题并进行智能拆解。点击开始拆解按钮启动流程，或者直接与我对话调整拆解建议。'
        }
      }
    ];
    setMessages(initialMessages);
  }, [nodeID]);

  // 开始拆解流程
  const handleStartDecompose = async () => {
    setIsDecomposing(true);
    setProgress(0);

    try {
      // 步骤1: RAG检索
      updateStepStatus('rag-search', 'running');
      addSystemMessage('🔍 开始RAG知识检索...');
      setProgress(20);

      // 模拟RAG检索
      await new Promise(resolve => setTimeout(resolve, 1500));
      updateStepStatus('rag-search', 'completed');
      addSystemMessage('✅ RAG检索完成，找到相关知识');
      setProgress(40);

      // 步骤2: AI分析
      updateStepStatus('ai-analysis', 'running');
      addSystemMessage('🤖 AI正在分析问题...');
      setProgress(60);

      // 模拟AI分析
      await new Promise(resolve => setTimeout(resolve, 2000));
      updateStepStatus('ai-analysis', 'completed');

      // 添加AI分析结果
      const analysisMessage: ChatMsg = {
        type: 'text',
        textMsg: {
          id: `analysis-${Date.now()}`,
          role: 'assistant',
          content: `基于RAG检索的知识，我建议将"${node.data?.question || '当前问题'}"拆解为以下几个子问题：\n1. 需求分析与用户研究\n2. 技术方案设计\n3. 实现与测试\n4. 部署与维护\n您可以通过对话调整这些建议，或者直接确认创建子节点。`
        }
      };
      setMessages(prev => [...prev, analysisMessage]);

      const analysisMessage2: ChatMsg = {
        type: 'text',
        textMsg: {
          id: `analysis2-${Date.now()}`,
          role: 'assistant',
          content: `基于RAG检索的知识，我建议将"${node.data?.question || '当前问题'}"拆解为以下几个子问题：\n\n1. 需求分析与用户研究\n2. 技术方案设计\n3. 实现与测试\n4. 部署与维护\n\n您可以通过对话调整这些建议，或者直接确认创建子节点。`
        }
      };
      setMessages(prev => [...prev, analysisMessage2]);
      setProgress(80);

      // 步骤3: 节点创建准备
      updateStepStatus('node-creation', 'running');
      addSystemMessage('📝 子问题建议已生成，等待您的确认');
      setProgress(100);

      updateStepStatus('node-creation', 'completed');

    } catch (error) {
      toast('拆解过程中出现错误，请重试');
      setDecomposeSteps(prev => prev.map(step =>
        step.status === 'running' ? { ...step, status: 'failed' } : step
      ));
    } finally {
      setIsDecomposing(false);
    }
  };

  // 更新步骤状态
  const updateStepStatus = (stepId: string, status: DecomposeStep['status']) => {
    setDecomposeSteps(prev => prev.map(step =>
      step.id === stepId ? { ...step, status } : step
    ));
  };

  // 添加系统消息
  const addSystemMessage = (content: string) => {
    const systemMessage: ChatMsg = {
      type: 'action',
      actionMsg: {
        id: `system-${Date.now()}-${Math.random().toString(36).slice(2)}`,
        role: 'system',
        actions: [{
          name: 'system_notification',
          url: '',
          arguments: content
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
      return;
    }
    messages.push({
      type: 'text',
      textMsg: {
        id: `user-${Date.now()}`,
        role: 'user',
        content: inputValue
      }
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
            {!isDecomposing && decomposeSteps.every(step => step.status === 'pending') && (
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