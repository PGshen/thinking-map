/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/workspace-layout.tsx
 */
'use client';

import React from 'react';
import { TopBar } from './top-bar/top-bar';
import { VisualizationArea } from './visualization-area/visualization-area';
import { OperationPanel } from './operation-panel/operation-panel';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';

interface WorkspaceLayoutProps {
  taskId: string;
  children?: React.ReactNode;
}

export function WorkspaceLayout({ taskId }: WorkspaceLayoutProps) {
  const { panelOpen, panelWidth } = useWorkspaceStore();

  return (
    <div className="h-screen flex flex-col bg-background">
      {/* 顶部固定栏 */}
      <TopBar taskId={taskId} />
      
      {/* 主体区域：可视化区域 + 操作面板 */}
      <div className="flex-1 flex relative overflow-hidden">
        {/* 可视化区域 */}
        <div 
          className="flex-1 transition-all duration-300 ease-in-out"
          style={{
            width: panelOpen ? `calc(100% - ${panelWidth}px)` : '100%'
          }}
        >
          <VisualizationArea taskId={taskId} />
        </div>
        
        {/* 操作面板 */}
        <OperationPanel />
      </div>
    </div>
  );
}

export default WorkspaceLayout;