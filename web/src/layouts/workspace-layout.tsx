/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/layouts/workspace-layout.tsx
 */
import React from 'react';
import { WorkspaceLayout as NewWorkspaceLayout } from '@/features/workspace';

interface WorkspaceLayoutProps {
  children?: React.ReactNode;
  taskID?: string;
}

export default function WorkspaceLayout({ children, taskID }: WorkspaceLayoutProps) {
  // 如果提供了taskID，使用新的工作区布局
  if (taskID) {
    return <NewWorkspaceLayout taskID={taskID} />;
  }
  
  // 否则保持原有的简单布局（向后兼容）
  return (
    <div className="w-full h-full min-h-screen bg-background flex flex-col">
      <main className="flex-1 min-w-0">{children}</main>
    </div>
  );
}