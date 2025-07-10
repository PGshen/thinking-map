/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/panel-tabs.tsx
 */
'use client';

import React, { useState } from 'react';
import { Info, GitBranch, CheckCircle, X } from 'lucide-react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { InfoTab } from './info-tab';
import { DecomposeTab } from './decompose-tab';
import { ConclusionTab } from './conclusion-tab';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { Button } from '@/components/ui/button';

interface PanelTabsProps {
  nodeId: string;
}

export function PanelTabs({ nodeId }: PanelTabsProps) {
  const [activeTab, setActiveTab] = useState('info');
  const { nodes, actions } = useWorkspaceStore();

  // 获取当前节点数据
  const currentNode = nodes.find(node => node.id === nodeId);

  if (!currentNode) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-muted-foreground">节点不存在</p>
      </div>
    );
  }

  // 根据节点状态判断Tab可用性
  const nodeData = currentNode.data as any;
  const status = nodeData?.status || 'pending';
  const canDecompose = status === 'pending' || status === 'running';
  const canConclude = status === 'completed' ||
    (status === 'running' && !canDecompose);

  return (
    <div className="h-full flex flex-col">
      <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
        {/* Tab导航 */}
        <div className="flex items-center justify-between mt-2 px-4">
          <TabsList className="grid grid-cols-3">
            <TabsTrigger value="info" className="flex items-center gap-2">
              <Info className="w-4 h-4" />
              <span className="hidden sm:inline">信息</span>
            </TabsTrigger>
            <TabsTrigger
              value="decompose"
              disabled={!canDecompose}
              className="flex items-center gap-2"
            >
              <GitBranch className="w-4 h-4" />
              <span className="hidden sm:inline">拆解</span>
            </TabsTrigger>
            <TabsTrigger
              value="conclusion"
              disabled={!canConclude}
              className="flex items-center gap-2"
            >
              <CheckCircle className="w-4 h-4" />
              <span className="hidden sm:inline">结论</span>
            </TabsTrigger>
          </TabsList>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => actions.closePanel()}
            className="h-8 w-8 p-0"
            aria-label="关闭面板"
          >
            <X className="w-4 h-4" />
          </Button>
        </div>

        {/* Tab内容 */}
        <div className="flex-1 overflow-hidden">
          <TabsContent value="info" className="h-full m-0 p-4">
            <InfoTab nodeId={nodeId} node={currentNode} />
          </TabsContent>

          <TabsContent value="decompose" className="h-full m-0 p-4">
            <DecomposeTab nodeId={nodeId} node={currentNode} />
          </TabsContent>

          <TabsContent value="conclusion" className="h-full m-0 p-4">
            <ConclusionTab nodeId={nodeId} node={currentNode} />
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}

export default PanelTabs;