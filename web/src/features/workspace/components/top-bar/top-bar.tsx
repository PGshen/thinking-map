/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/top-bar/top-bar.tsx
 */
'use client';

import React from 'react';
import { ExitButton } from './exit-button';
import { TaskTitle } from './task-title';
import { SettingsButton } from './settings-button';
import { MapInfo } from './map-info';
import { useWorkspaceData } from '@/features/workspace/hooks/use-workspace-data';

interface TopBarProps {
  taskId: string;
}

export function TopBar({ taskId }: TopBarProps) {
  const { mapTitle, isLoading } = useWorkspaceData(taskId);

  return (
    <header className="h-12 flex items-center justify-between px-6 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 sticky top-0 z-50">
      {/* 左侧区域 */}
      <div className="flex items-center gap-4">
        <ExitButton />
        <TaskTitle 
          title={mapTitle || '加载中...'} 
          taskId={taskId}
          isLoading={isLoading}
        />
      </div>
      
      {/* 右侧区域 */}
      <div className="flex items-center gap-2">
        <MapInfo taskId={taskId} />
        <SettingsButton taskId={taskId} />
      </div>
    </header>
  );
}

export default TopBar;