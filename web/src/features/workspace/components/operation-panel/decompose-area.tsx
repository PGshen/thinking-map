"use client";
import {
  ChatMessageAvatar,
  ChatMessageContent,
  ChatMessage,
} from "@/components/ui/chat-message";
import { ChatMessageArea } from "@/components/ui/chat-message-area";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Action, MessageResponse } from "@/types/message";
import { ChevronDown, ChevronRight, Loader, Brain, Check, Clock, Pause, X, FileText } from "lucide-react";
import { useState } from "react";

interface DecomposeAreaProps {
  loading: boolean;
  messages: MessageResponse[];
  clickAction: (action: Action) => void;
}

export function DecomposeArea({ loading, messages, clickAction }: DecomposeAreaProps) {
  // 用于管理每个思考消息的折叠状态
  const [collapsedStates, setCollapsedStates] = useState<Record<string, boolean>>({});

  // 切换指定消息的折叠状态
  const toggleCollapse = (messageId: string) => {
    setCollapsedStates(prev => ({
      ...prev,
      [messageId]: !prev[messageId]
    }));
  };

  return (
    <div className="h-full">
      <ChatMessageArea scrollButtonAlignment="center" className="px-2 py-2 space-y-4 text-sm">
        {messages.map((message) => {
          // 文本消息
          if (message.messageType === 'text') {
            if (message.content.text === undefined) {
              return
            }
            if (message.role === 'user') {
              return (
                <ChatMessage
                  key={message.id}
                  id={message.id}
                  variant="bubble"
                  type="outgoing"
                >
                  <ChatMessageContent content={message.content.text!} />
                </ChatMessage>
              );
            } else {
              return (
                <ChatMessage
                  key={message.id}
                  id={message.id}
                  type="incoming"
                >
                  <ChatMessageAvatar />
                  <ChatMessageContent content={message.content.text!} />
                </ChatMessage>
              );
            }
          }
          // 思考消息
          if (message.messageType === 'thought') {
            const isCollapsed = collapsedStates[message.id] || false;
            
            return (
              <ChatMessage
                key={message.id}
                id={message.id}
                type="incoming"
              >
                <ChatMessageAvatar />
                <div className="flex-1">
                  <Collapsible open={!isCollapsed} onOpenChange={() => toggleCollapse(message.id)}>
                    <CollapsibleTrigger className="cursor-pointer flex items-center gap-2 text-left p-1 rounded-md bg-blue-50 hover:bg-blue-100 transition-colors border border-blue-200">
                      <Brain className="h-4 w-4 text-blue-600 flex-shrink-0" />
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
          }
          if (message.messageType === 'notice') {
            // 根据通知类型获取颜色主题
            const getNoticeTheme = (type: string) => {
              switch (type.toLowerCase()) {
                case 'error':
                  return {
                    bg: 'bg-red-50',
                    border: 'border-red-200',
                    tagBg: 'bg-red-100',
                    tagText: 'text-red-800',
                    nameText: 'text-red-900',
                    contentText: 'text-red-800'
                  };
                case 'warning':
                  return {
                    bg: 'bg-yellow-50',
                    border: 'border-yellow-200',
                    tagBg: 'bg-yellow-100',
                    tagText: 'text-yellow-800',
                    nameText: 'text-yellow-900',
                    contentText: 'text-yellow-800'
                  };
                case 'success':
                  return {
                    bg: 'bg-green-50',
                    border: 'border-green-200',
                    tagBg: 'bg-green-100',
                    tagText: 'text-green-800',
                    nameText: 'text-green-900',
                    contentText: 'text-green-800'
                  };
                case 'info':
                  return {
                    bg: 'bg-blue-50',
                    border: 'border-blue-200',
                    tagBg: 'bg-blue-100',
                    tagText: 'text-blue-800',
                    nameText: 'text-blue-900',
                    contentText: 'text-blue-800'
                  };
                default:
                  return {
                    bg: 'bg-gray-50',
                    border: 'border-gray-200',
                    tagBg: 'bg-gray-100',
                    tagText: 'text-gray-800',
                    nameText: 'text-gray-900',
                    contentText: 'text-gray-800'
                  };
              }
            };

            return (
              <ChatMessage
                key={message.id}
                id={message.id}
                type="incoming"
              >
                <ChatMessageAvatar />
                <div className="space-y-2">
                  {message.content.notice?.map((notice, index) => {
                    const theme = getNoticeTheme(notice.type);
                    return (
                      <div key={index} className={`p-3 ${theme.bg} border ${theme.border} rounded-md`}>
                        <div className="flex items-center gap-2 mb-1">
                          <span className={`px-2 py-1 ${theme.tagBg} ${theme.tagText} text-xs rounded-full font-medium`}>
                            {notice.type}
                          </span>
                          <span className={`font-medium ${theme.nameText}`}>{notice.name}</span>
                        </div>
                        <p className={`${theme.contentText} text-sm`}>{notice.content}</p>
                      </div>
                    );
                  })}
                </div>
              </ChatMessage>
            );
          }
          if (message.messageType === 'action') {
            return (
              <ChatMessage
                key={message.id}
                id={message.id}
                type="incoming"
              >
                <ChatMessageAvatar />
                <div className="flex flex-wrap gap-2">
                  {message.content.action?.map((action, index) => {
                    // 根据HTTP方法设置不同的样式主题
                    const getMethodTheme = (method: string) => {
                      switch (method?.toUpperCase()) {
                        case 'GET':
                          return {
                            bg: 'bg-green-50',
                            hover: 'hover:bg-green-100',
                            text: 'text-green-700',
                            border: 'border-green-200',
                            methodBg: 'bg-green-100',
                            methodText: 'text-green-600'
                          };
                        case 'POST':
                          return {
                            bg: 'bg-blue-50',
                            hover: 'hover:bg-blue-100',
                            text: 'text-blue-700',
                            border: 'border-blue-200',
                            methodBg: 'bg-blue-100',
                            methodText: 'text-blue-600'
                          };
                        case 'PUT':
                          return {
                            bg: 'bg-orange-50',
                            hover: 'hover:bg-orange-100',
                            text: 'text-orange-700',
                            border: 'border-orange-200',
                            methodBg: 'bg-orange-100',
                            methodText: 'text-orange-600'
                          };
                        case 'DELETE':
                          return {
                            bg: 'bg-red-50',
                            hover: 'hover:bg-red-100',
                            text: 'text-red-700',
                            border: 'border-red-200',
                            methodBg: 'bg-red-100',
                            methodText: 'text-red-600'
                          };
                        case 'PATCH':
                          return {
                            bg: 'bg-purple-50',
                            hover: 'hover:bg-purple-100',
                            text: 'text-purple-700',
                            border: 'border-purple-200',
                            methodBg: 'bg-purple-100',
                            methodText: 'text-purple-600'
                          };
                        default:
                          return {
                            bg: 'bg-gray-50',
                            hover: 'hover:bg-gray-100',
                            text: 'text-gray-700',
                            border: 'border-gray-200',
                            methodBg: 'bg-gray-100',
                            methodText: 'text-gray-600'
                          };
                      }
                    };

                    const theme = getMethodTheme(action.method);

                    return (
                      <button
                        key={index}
                        className={`cursor-pointer px-3 py-2 ${theme.bg} ${theme.hover} ${theme.text} text-sm rounded-md border ${theme.border} transition-colors duration-200 text-left flex-shrink-0`}
                        onClick={() => {
                          clickAction(action);
                        }}
                      >
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{action.name}</span>
                          <span className={`text-xs ${theme.methodBg} ${theme.methodText} px-1.5 py-0.5 rounded uppercase font-mono`}>
                            {action.method}
                          </span>
                        </div>
                        {action.url && (
                          <div className={`text-xs ${theme.text} mt-1 opacity-75 font-mono`}>
                            {action.url}
                          </div>
                        )}
                      </button>
                    );
                  })}
                </div>
              </ChatMessage>
            );
          }
          // 计划步骤
          if (message.messageType === 'plan') {
            return (
              <ChatMessage
                key={message.id}
                id={message.id}
                type="incoming"
              >
                <ChatMessageAvatar />
                <div className="space-y-3">
                  <div className="flex items-center gap-2 mb-2">
                    <Brain className="h-4 w-4 text-purple-600" />
                    <span className="text-sm font-medium text-purple-800">执行计划</span>
                  </div>
                  {message.content.plan?.map((plan, index) => {
                    // 根据状态设置不同的样式主题
                    const getPlanTheme = (status: string) => {
                      switch (status?.toLowerCase()) {
                        case 'completed':
                        case '已完成':
                        case 'done':
                          return {
                            bg: 'bg-green-50',
                            border: 'border-green-200',
                            tagBg: 'bg-green-100',
                            tagText: 'text-green-800',
                            nameText: 'text-green-900',
                            contentText: 'text-green-700',
                            iconColor: 'text-green-600',
                            IconComponent: Check
                          };
                        case 'in_progress':
                        case '进行中':
                        case 'running':
                          return {
                            bg: 'bg-blue-50',
                            border: 'border-blue-200',
                            tagBg: 'bg-blue-100',
                            tagText: 'text-blue-800',
                            nameText: 'text-blue-900',
                            contentText: 'text-blue-700',
                            iconColor: 'text-blue-600',
                            IconComponent: Clock
                          };
                        case 'pending':
                        case '待执行':
                        case 'waiting':
                          return {
                            bg: 'bg-yellow-50',
                            border: 'border-yellow-200',
                            tagBg: 'bg-yellow-100',
                            tagText: 'text-yellow-800',
                            nameText: 'text-yellow-900',
                            contentText: 'text-yellow-700',
                            iconColor: 'text-yellow-600',
                            IconComponent: Pause
                          };
                        case 'failed':
                        case '失败':
                        case 'error':
                          return {
                            bg: 'bg-red-50',
                            border: 'border-red-200',
                            tagBg: 'bg-red-100',
                            tagText: 'text-red-800',
                            nameText: 'text-red-900',
                            contentText: 'text-red-700',
                            iconColor: 'text-red-600',
                            IconComponent: X
                          };
                        default:
                          return {
                            bg: 'bg-gray-50',
                            border: 'border-gray-200',
                            tagBg: 'bg-gray-100',
                            tagText: 'text-gray-800',
                            nameText: 'text-gray-900',
                            contentText: 'text-gray-700',
                            iconColor: 'text-gray-600',
                            IconComponent: FileText
                          };
                      }
                    };

                    const theme = getPlanTheme(plan.status);
                    
                    return (
                      <div key={index} className={`relative p-4 ${theme.bg} border ${theme.border} rounded-lg transition-all duration-200 hover:shadow-sm`}>
                        {/* 步骤序号 */}
                        <div className="absolute -left-2 -top-2 w-6 h-6 bg-white border-2 border-purple-200 rounded-full flex items-center justify-center text-xs font-bold text-purple-600">
                          {index + 1}
                        </div>
                        
                        <div className="flex items-start gap-3">
                          {/* 状态图标 */}
                          <div className="flex-shrink-0 mt-0.5">
                            <theme.IconComponent className={`h-4 w-4 ${theme.iconColor}`} />
                          </div>
                          
                          <div className="flex-1 min-w-0">
                            {/* 标题和状态 */}
                            <div className="flex items-center gap-2 mb-2">
                              <span className={`font-semibold ${theme.nameText} text-base`}>{plan.name}</span>
                              <span className={`px-2 py-1 ${theme.tagBg} ${theme.tagText} text-xs rounded-full font-medium`}>
                                {plan.status}
                              </span>
                            </div>
                            
                            {/* 描述 */}
                            <p className={`${theme.contentText} text-sm leading-relaxed`}>{plan.description}</p>
                          </div>
                        </div>
                        
                        {/* 连接线（除了最后一个步骤） */}
                        {index < (message.content.plan?.length || 0) - 1 && (
                          <div className="absolute left-4 bottom-0 w-0.5 h-4 bg-purple-200 transform translate-y-full"></div>
                        )}
                      </div>
                    );
                  })}
                </div>
              </ChatMessage>
            );
          }
        })}
        
        {/* Loading状态 */}
        {loading && (
          <ChatMessage
            id="loading-message"
            type="incoming"
          >
            <ChatMessageAvatar />
            <div className="flex items-center gap-2 text-muted-foreground">
              <Loader className="h-4 w-4 animate-spin" />
              <span className="text-sm">正在思考中...</span>
            </div>
          </ChatMessage>
        )}
      </ChatMessageArea>
    </div>
  );
}
