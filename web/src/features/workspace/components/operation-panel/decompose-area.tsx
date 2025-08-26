"use client";
import {
  ChatMessageAvatar,
  ChatMessageContent,
  ChatMessage,
} from "@/components/ui/chat-message";
import { ChatMessageArea } from "@/components/ui/chat-message-area";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Action, MessageResponse, NoticeType } from "@/types/message";
import { ChevronDown, ChevronRight, Loader, Brain, Check, Clock, Pause, X, FileText, Waypoints } from "lucide-react";
import { useState } from "react";

interface DecomposeAreaProps {
  loading: boolean;
  messages: MessageResponse[];
  clickAction: (action: Action) => void;
}

export function DecomposeArea({ loading, messages, clickAction }: DecomposeAreaProps) {
  // 用于管理每个思考消息的折叠状态
  const [collapsedStates, setCollapsedStates] = useState<Record<string, boolean>>({});
  // 用于管理执行计划步骤的展开状态
  const [expandedSteps, setExpandedSteps] = useState<Record<string, Set<number>>>({});

  // 切换指定消息的折叠状态
  const toggleCollapse = (messageId: string) => {
    setCollapsedStates(prev => ({
      ...prev,
      [messageId]: !prev[messageId]
    }));
  };

  // 切换执行计划步骤的展开状态
  const toggleStepExpanded = (messageId: string, stepIndex: number) => {
    setExpandedSteps(prev => {
      const messageSteps = prev[messageId] || new Set();
      const newSteps = new Set(messageSteps);
      if (newSteps.has(stepIndex)) {
        newSteps.delete(stepIndex);
      } else {
        newSteps.add(stepIndex);
      }
      return {
        ...prev,
        [messageId]: newSteps
      };
    });
  };

  return (
    <div className="h-full">
      <ChatMessageArea scrollButtonAlignment="center" className="px-2 py-2 space-y-4 text-sm">
        {messages.map((message) => {
          if (message.content === undefined) {
            return
          }
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
                  {/* <ChatMessageAvatar /> */}
                  <ChatMessageContent content={message.content.text!} />
                </ChatMessage>
              );
            }
          }
          // 思考消息
          if (message.messageType === 'thought') {
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
            if (message.content.notice === undefined) {
              return
            }
            const isCollapsed = collapsedStates[message.id] || false;
            // 根据通知类型获取颜色主题
            const getNoticeTheme = (type: NoticeType) => {
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
            const notice = message.content.notice;
            const theme = getNoticeTheme(notice.type);

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
                      <Brain className="h-4 w-4 text-blue-600 flex-shrink-0" />
                      <span className="text-sm font-medium text-blue-800 flex-1">消息通知</span>
                      {isCollapsed ? (
                        <ChevronRight className="h-4 w-4 text-blue-600 flex-shrink-0" />
                      ) : (
                        <ChevronDown className="h-4 w-4 text-blue-600 flex-shrink-0" />
                      )}
                    </CollapsibleTrigger>
                    <CollapsibleContent className="mt-1">
                      <div className="pl-2 border-l-2 border-blue-200">
                        <div className="space-y-1">
                          <div className={`p-2 ${theme.bg} border ${theme.border} rounded-md`}>
                            <div className="flex items-center gap-1.5 mb-0.5">
                              <span className={`px-1.5 py-0.5 ${theme.tagBg} ${theme.tagText} text-xs rounded font-medium`}>
                                {notice.type}
                              </span>
                              <span className={`font-medium ${theme.nameText} text-sm`}>{notice.name}</span>
                            </div>
                            <p className={`${theme.contentText} text-xs leading-relaxed`}>{notice.content}</p>
                          </div>
                        </div>
                      </div>
                    </CollapsibleContent>
                  </Collapsible>
                </div>
              </ChatMessage>
            );
          }
          if (message.messageType === 'action') {
            if (message.content.action === undefined) {
              return
            }
            const isCollapsed = collapsedStates[message.id] || false;

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
                      <Brain className="h-4 w-4 text-blue-600 flex-shrink-0" />
                      <span className="text-sm font-medium text-blue-800 flex-1">执行动作</span>
                      {isCollapsed ? (
                        <ChevronRight className="h-4 w-4 text-blue-600 flex-shrink-0" />
                      ) : (
                        <ChevronDown className="h-4 w-4 text-blue-600 flex-shrink-0" />
                      )}
                    </CollapsibleTrigger>
                    <CollapsibleContent className="mt-2">
                      <div className="pl-2 border-l-2 border-blue-200">
                        <div className="flex flex-wrap gap-2">
                          {message.content.action?.map((action, index) => {

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
                      </div>
                    </CollapsibleContent>
                  </Collapsible>
                </div>
              </ChatMessage>
            );
          }
          // 计划步骤
          if (message.messageType === 'plan') {
            if (message.content.plan === undefined) {
              return
            }
            const isCollapsed = collapsedStates[message.id] || false;

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
                case 'running':
                case 'in_progress':
                case '进行中':
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
                case 'skipped':
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
                      <Waypoints className="h-4 w-4 text-blue-600" />
                      <span className="text-sm font-medium text-blue-800">执行计划</span>
                      {isCollapsed ? (
                        <ChevronRight className="h-4 w-4 text-blue-600 flex-shrink-0" />
                      ) : (
                        <ChevronDown className="h-4 w-4 text-blue-600 flex-shrink-0" />
                      )}
                    </CollapsibleTrigger>
                    <CollapsibleContent className="mt-2">
                      <div className="space-y-2 pl-2 border-l-2 border-blue-200">
                        {message.content.plan?.steps?.map((step, index) => {

                          const theme = getPlanTheme(step.status);
                          const messageSteps = expandedSteps[message.id] || new Set();
                          const isExpanded = messageSteps.has(index);

                          return (
                            <div key={index} className={`relative p-3 ${theme.bg} border ${theme.border} rounded-lg transition-all duration-200 hover:shadow-sm`}>
                              {/* 步骤序号 */}
                              <div className="absolute -left-1.5 -top-1.5 w-5 h-5 bg-white border-2 border-purple-200 rounded-full flex items-center justify-center text-xs font-bold text-purple-600">
                                {index + 1}
                              </div>

                              <div className="flex items-start gap-2.5">
                                {/* 状态图标 */}
                                <div className="flex-shrink-0 mt-0.5">
                                  <theme.IconComponent className={`h-3.5 w-3.5 ${theme.iconColor}`} />
                                </div>

                                <div className="flex-1 min-w-0">
                                  {/* 标题和状态 - 可点击展开 */}
                                  <div
                                    className="flex items-center gap-2 cursor-pointer hover:opacity-80 transition-opacity"
                                    onClick={() => toggleStepExpanded(message.id, index)}
                                  >
                                    <span className={`font-medium ${theme.nameText} text-sm`}>{step.name}</span>
                                    <span className={`px-1.5 py-0.5 ${theme.tagBg} ${theme.tagText} text-xs rounded-full font-medium`}>
                                      {step.status}
                                    </span>
                                    <ChevronDown className={`h-3 w-3 ${theme.iconColor} transition-transform duration-200 ${isExpanded ? 'rotate-180' : ''}`} />
                                  </div>

                                  {/* 描述 - 可展开收起 */}
                                  {isExpanded && (
                                    <div className="mt-2 animate-in slide-in-from-top-1 duration-200">
                                      <p className={`${theme.contentText} text-xs leading-relaxed`}>{step.description}</p>
                                    </div>
                                  )}
                                </div>
                              </div>

                              {/* 连接线（除了最后一个步骤） */}
                              {index < (message.content.plan?.steps?.length || 0) - 1 && (
                                <div className="absolute left-3 bottom-0 w-0.5 h-3 bg-purple-200 transform translate-y-full"></div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    </CollapsibleContent>
                  </Collapsible>
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
            {/* <ChatMessageAvatar /> */}
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
