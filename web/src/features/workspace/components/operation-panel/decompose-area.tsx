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
import { ChatActionMessage, ChatMsg, ChatTextMessage, ChatToolMessage } from "@/types/message";
import { useState } from "react";

interface DecomposeAreaProps {
	messages: ChatMsg[];
}

export function DecomposeArea({ messages }: DecomposeAreaProps) {
	return (
		<div className="h-full">
			<ChatMessageArea scrollButtonAlignment="center" className="px-2 py-2 space-y-4 text-sm">
			{messages.map((message) => {
				// 文本消息
				if (message.type === 'text') {
					let msg: ChatTextMessage = message.textMsg!
					if (msg.role === 'user') {
						return (
							<ChatMessage
								key={msg.id}
								id={msg.id}
								variant="bubble"
								type="outgoing"
							>
								<ChatMessageContent content={msg.content} />
							</ChatMessage>
						);
					} else {
						return (
							<ChatMessage
								key={msg.id}
								id={msg.id}
								type="incoming"
							>
								<ChatMessageAvatar />
								<ChatMessageContent content={msg.content} />
							</ChatMessage>
						);
					}
				}
				// 工具消息
				if (message.type === 'tool') {
					let msg: ChatToolMessage = message.toolMsg!
					return (
						// todo 工具消息UI
						<div></div>
					);
				}
				if (message.type === 'action') {
					let msg: ChatActionMessage = message.actionMsg!
					return (
						<ChatMessage
							key={msg.id}
							id={msg.id}
							type="incoming"
						>
							<ChatMessageAvatar />
							<div key={msg.id} className="flex justify-center mb-2">
								<div className="px-3 py-1 bg-blue-50 text-blue-700 text-sm rounded-sm border border-blue-200">
									{msg.actions[0]?.arguments || '系统通知'}
								</div>
							</div>
						</ChatMessage>
					);
				}
			})}
			</ChatMessageArea>
		</div>
	);
}
