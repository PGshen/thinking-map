/*
 * @Date: 2025-01-27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-10 09:44:51
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/operation-panel.tsx
 */
'use client';

import React, { useEffect, useRef } from 'react';
import { X, GripVertical } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { PanelTabs } from './panel-tabs';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { cn } from '@/lib/utils';

export function OperationPanel() {
  const { 
    panelOpen, 
    panelWidth, 
    activeNodeID, 
    actions 
  } = useWorkspaceStore();
  
  const panelRef = useRef<HTMLDivElement>(null);
  const resizerRef = useRef<HTMLDivElement>(null);
  const isDragging = useRef(false);

  // 处理面板宽度调整
  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!isDragging.current || typeof window === 'undefined') return;
      
      const newWidth = window.innerWidth - e.clientX;
      const minWidth = window.innerWidth * 0.25; // 最小25%
      const maxWidth = window.innerWidth * 0.5;  // 最大50%
      
      const clampedWidth = Math.max(minWidth, Math.min(maxWidth, newWidth));
      actions.setPanelWidth(clampedWidth);
    };

    const handleMouseUp = () => {
      isDragging.current = false;
      document.body.style.cursor = '';
      document.body.style.userSelect = '';
    };

    const handleMouseDown = (e: MouseEvent) => {
      if (e.target === resizerRef.current) {
        isDragging.current = true;
        document.body.style.cursor = 'col-resize';
        document.body.style.userSelect = 'none';
        e.preventDefault();
      }
    };

    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
    document.addEventListener('mousedown', handleMouseDown);

    return () => {
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
      document.removeEventListener('mousedown', handleMouseDown);
    };
  }, [actions]);

  // 键盘快捷键支持
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && panelOpen) {
        actions.closePanel();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [panelOpen, actions]);

  if (!panelOpen || !activeNodeID) {
    return null;
  }

  return (
    <>
      {/* 遮罩层 */}
      <div 
        className="fixed inset-0 bg-black/20 z-40 lg:hidden"
        onClick={() => actions.closePanel()}
      />
      
      {/* 面板主体 */}
      <div
        ref={panelRef}
        className={cn(
          "fixed h-screen right-0 top-16 bottom-0 bg-background border-l shadow-lg z-50",
          "lg:relative lg:top-0 lg:shadow-none",
          panelOpen ? "translate-x-0" : "translate-x-full lg:translate-x-0"
        )}
        style={{
          width: `${panelWidth}px`,
          minWidth: typeof window !== 'undefined' ? `${window.innerWidth * 0.25}px` : '300px',
          maxWidth: typeof window !== 'undefined' ? `${window.innerWidth * 0.5}px` : '600px',
        }}
      >
        {/* 调整大小手柄 */}
        <div
          ref={resizerRef}
          className="absolute left-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-primary/20 transition-colors group hidden lg:block z-20"
        >
          <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 transition-opacity">
            <GripVertical className="w-3 h-3 text-muted-foreground" />
          </div>
        </div>
        
        {/* 面板内容 */}
        <div className="h-full flex flex-col overflow-hidden">
          <PanelTabs nodeID={activeNodeID} />
        </div>
      </div>
    </>
  );
}

export default OperationPanel;