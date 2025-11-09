/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/hooks/use-executable-nodes.ts
 */
import { useCallback, useEffect } from 'react';
import { useWorkspaceStore } from '../store/workspace-store';
import { executableNodes } from '@/api/node';
import { useToast } from '@/hooks/use-toast';

/**
 * 获取可执行节点的hook
 * 根据当前地图ID和当前选中节点ID获取可执行节点列表和建议执行的节点
 */
export function useExecutableNodes() {
  const { toast } = useToast();
  const mapID = useWorkspaceStore(state => state.mapID);
  const activeNodeID = useWorkspaceStore(state => state.activeNodeID);
  const actions = useWorkspaceStore(state => state.actions);
  
  // 获取可执行节点
  const fetchExecutableNodes = useCallback(async (nodeID?: string) => {
    if (!mapID) return;
    
    try {
      const response = await executableNodes(mapID, nodeID);
      if (response && response.data && response.data.nodeIDs) {
        actions.setExecutableNodes(response.data.nodeIDs, response.data.suggestedNodeID);
      }
    } catch (error) {
      console.error('获取可执行节点失败:', error);
      toast({
        title: '获取可执行节点失败',
        description: '无法获取可执行节点列表，请重试',
        variant: 'destructive',
      });
    }
  }, [mapID, actions, toast]);
  
  // 当地图ID或活动节点ID变化时，获取可执行节点
  useEffect(() => {
    if (mapID) {
      fetchExecutableNodes(activeNodeID || undefined);
    }
  }, [mapID, activeNodeID]);
  
  return {
    fetchExecutableNodes,
  };
}