"use client";
import {
	ChatInput,
	ChatInputSubmit,
	ChatInputTextArea,
} from "@/components/ui/chat-input";
import {
	ChatMessageAvatar,
	ChatMessageContent,
	ChatMessage,
} from "@/components/ui/chat-message";
import { ChatMessageArea } from "@/components/ui/chat-message-area";
import { MessageResponse } from "@/types/message";
import { useState } from "react";

interface DecomposeAreaProps {
	messages: MessageResponse[];
}

export function DecomposeArea({ messages }: DecomposeAreaProps) {
	return (
		<div className="h-full">
			<ChatMessageArea scrollButtonAlignment="center" className="px-2 py-2 space-y-4 text-sm">
			{messages.map((message) => {
				// 文本消息
				if (message.messageType === 'text') {
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
							<div className="space-y-2">
								{message.content.action?.map((action, index) => (
									<button
										key={index}
										className="w-full px-4 py-2 bg-blue-50 hover:bg-blue-100 text-blue-700 text-sm rounded-md border border-blue-200 transition-colors duration-200 text-left"
										onClick={() => {
											// TODO: 处理动作点击事件
											console.log('Action clicked:', action);
										}}
									>
										<div className="flex items-center justify-between">
											<span className="font-medium">{action.name}</span>
											<span className="text-xs text-blue-500 uppercase">{action.method}</span>
										</div>
										{action.url && (
											<div className="text-xs text-blue-600 mt-1 opacity-75">
												{action.url}
											</div>
										)}
									</button>
								))}
							</div>
						</ChatMessage>
					);
				}
			})}
			</ChatMessageArea>
		</div>
	);
}
