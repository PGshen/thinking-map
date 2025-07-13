/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/hooks/use-workspace-data.ts
 */
import { useEffect, useCallback } from 'react';
import { useWorkspaceStore } from '../store/workspace-store';
import { useToast } from '@/hooks/use-toast';
import { Map } from '@/types/map';
import { getMap, updateMap } from '@/api/map';
import { getMapNodes, updateNode, updateNodeContext } from '@/api/node';

// 工作区数据管理hook
export function useWorkspaceData(mapId?: string) {
  const {
    mapInfo,
    nodes,
    edges,
    isLoading,
    hasUnsavedChanges,
    actions,
  } = useWorkspaceStore();
  
  const { toast } = useToast();

  // 加载思维导图信息
  const loadMap = useCallback(async (id: string) => {
    if (!id) return;
    
    actions.setLoading(true);
    try {
      const mapInfoResp = await getMap(id);
      actions.updateMap(mapInfoResp.data); // 修复：取data字段
    } catch (error) {
      toast({
        title: '加载失败',
        description: '无法加载思维导图信息，请重试',
        variant: 'destructive',
      });
    } finally {
      actions.setLoading(false);
    }
  }, []);

  // 加载节点数据
  const loadNodes = useCallback(async (id: string) => {
    if (!id) return;
    
    actions.setLoading(true);
    try {
      const nodeDataResp = await getMapNodes(id);
      const nodes = nodeDataResp.data.nodes;
      
      // 转换为ReactFlow格式节点
      const reactFlowNodes = nodes.map((node) => ({
        id: node.id,
        type: 'custom',
        position: node.position,
        data: {
          ...node
        },
      }));
      
      // 生成边（parentID存在且非空）
      const reactFlowEdges = nodes
        .filter(node => node.parentID)
        .map(node => ({
          id: `${node.parentID}-${node.id}`,
          source: node.parentID,
          target: node.id,
          type: 'default',
        }));
      
      actions.setNodes(reactFlowNodes);
      actions.setEdges(reactFlowEdges);
      actions.setUnsavedChanges(false);
    } catch (error) {
      toast({
        title: '加载失败',
        description: '无法加载节点数据，请重试',
        variant: 'destructive',
      });
    } finally {
      actions.setLoading(false);
    }
  }, []);

  // 保存思维导图详细信息
  const saveMap = useCallback(async (info: Partial<Omit<Map, 'id' | 'createdAt' | 'updatedAt'>>) => {
    if (!mapId || !mapInfo) return false;
    
    try {
      await updateMap(mapId, info);
      actions.updateMap({
        ...mapInfo,
        ...info
      });
      actions.setUnsavedChanges(false);
      
      toast({
        title: '保存成功',
        description: '思维导图信息已更新',
      });
      
      return true;
    } catch (error) {
      toast({
        title: '保存失败',
        description: '更新思维导图信息时出错，请重试',
        variant: 'destructive',
      });
      return false;
    }
  }, [mapId, mapInfo]);

  // 保存节点信息
  const saveNodeInfo = useCallback(async (nodeId: string, updates: any) => {
    if (!mapId) return false;
    
    try {
      await updateNode(mapId, nodeId, updates);
      actions.updateNode(nodeId, { data: { ...updates } });
      
      toast({
        title: '保存成功',
        description: '节点信息已更新',
      });
      
      return true;
    } catch (error) {
      toast({
        title: '保存失败',
        description: '更新节点信息时出错，请重试',
        variant: 'destructive',
      });
      return false;
    }
  }, [mapId]);

  // 保存整个工作区
  const saveWorkspace = useCallback(async () => {
    if (!mapId || !hasUnsavedChanges) return true;
    
    actions.setLoading(true);
    try {
      // 转换节点数据格式并保存每个节点
      const savePromises = nodes.map(async node => {
        const nodeData = {
          question: node.data.question,
          target: node.data.target,
          context: node.data.context,
          conclusion: node.data.conclusion,
          status: node.data.status,
          position: node.position,
          dependencies: node.data.dependencies,
        };
        
        await updateNode(mapId, node.id, nodeData);
      });
      
      await Promise.all(savePromises);
      actions.setUnsavedChanges(false);
      
      toast({
        title: '保存成功',
        description: '工作区已保存',
      });
      
      return true;
    } catch (error) {
      toast({
        title: '保存失败',
        description: '保存工作区时出错，请重试',
        variant: 'destructive',
      });
      return false;
    } finally {
      actions.setLoading(false);
    }
  }, [mapId, hasUnsavedChanges, nodes]);

  // 初始化数据
  useEffect(() => {
    if (mapId) {
      loadMap(mapId);
      loadNodes(mapId);
    }
  }, [mapId]);

  // 页面卸载前保存
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (hasUnsavedChanges) {
        e.preventDefault();
        e.returnValue = '您有未保存的更改，确定要离开吗？';
      }
    };
    
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [hasUnsavedChanges]);

  return {
    // 数据
    mapInfo,
    nodes,
    edges,
    isLoading,
    hasUnsavedChanges,
    
    // 操作
    loadMap,
    loadNodes,
    saveMap,
    saveNodeInfo,
    saveWorkspace,
    
    // Store actions
    actions,
  };
}

export default useWorkspaceData;