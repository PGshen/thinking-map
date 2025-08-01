/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/workspace-layout.tsx
 */
'use client';

import React from 'react';
import { VisualizationArea } from './visualization-area/visualization-area';
import { OperationPanel } from './operation-panel/operation-panel';
import { InfoSidebar } from './side-bar/side-bar';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { SidebarProvider, SidebarInset } from '@/components/ui/sidebar';

interface WorkspaceLayoutProps {
  mapID: string;
  children?: React.ReactNode;
}

export function WorkspaceLayout({ mapID }: WorkspaceLayoutProps) {
  const { panelOpen, panelWidth } = useWorkspaceStore();

  return (
    <SidebarProvider
      style={
        {
          "--sidebar-width": "350px",
        } as React.CSSProperties
      }
    >
      <InfoSidebar mapID={mapID} />
      <SidebarInset>
        <div className="flex-1 flex relative overflow-hidden">
          {/* 可视化区域 */}
          <div
            className="flex-1 transition-all duration-300 ease-in-out"
            style={{
              width: panelOpen ? `calc(100% - ${panelWidth}px)` : '100%',
            }}
          >
            <VisualizationArea mapID={mapID} />
          </div>

          {/* 操作面板 */}
          <OperationPanel />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

export default WorkspaceLayout;