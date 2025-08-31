/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/hooks/use-workspace-data.ts
 */
import { useCallback, useEffect } from 'react';
import { useWorkspaceStore } from '../store/workspace-store';
import { useToast } from '@/hooks/use-toast';
import { Map, UpdateMapRequest } from '@/types/map';
import { getMap, updateMap } from '@/api/map';
import { getMapNodes, updateNode, updateNodeContext } from '@/api/node';
import { CustomNodeModel, Position } from '@/types/node';

// 工作区数据管理hook
export function useWorkspaceData(mapID?: string) {
  const {
    mapInfo,
    nodes,
    edges,
    isLoading,
    changedNodePositions,
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
      const reactFlowNodes = nodes.map((nodeData) => ({
        id: nodeData.id,
        type: 'custom',
        position: nodeData.position,
        data: {
          id: nodeData.id,
          parentID: nodeData.parentID,
          nodeType: nodeData.nodeType,
          question: nodeData.question,
          target: nodeData.target,
          decomposition: nodeData.decomposition,
          conclusion: nodeData.conclusion,
          status: nodeData.status,
          context: nodeData.context,
          metadata: nodeData.metadata,
        } as CustomNodeModel,
      }));
      
      // 生成父子关系边（parentID存在且非空）
      const parentChildEdges = nodes
        .filter(node => node.parentID)
        .map(node => ({
          id: `${node.parentID}-${node.id}`,
          source: node.parentID,
          target: node.id,
          type: 'default',
          style: { stroke: '#3b82f6' },
        }));
      
      // 生成依赖关系边（默认隐藏，只在选中节点时显示）
      const dependencyEdges = nodes
        .filter(node => node.dependencies && node.dependencies.length > 0)
        .flatMap(node => 
          node.dependencies!.map(depNodeID => ({
            id: `dep-${node.id}-${depNodeID}`,
            source: node.id,
            target: depNodeID,
            type: 'dependency',
            style: { strokeDasharray: '5,5', stroke: '#8b5cf6' },
            animated: true,
            sourceHandle: 'dependency-source',
            targetHandle: 'dependency-target',
            hidden: true, // 默认隐藏
          }))
        );
      
      const reactFlowEdges = [...parentChildEdges, ...dependencyEdges];
      
      actions.setNodes(reactFlowNodes);
      actions.setEdges(reactFlowEdges);
      actions.clearChangedNodePositions();
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
  const saveMap = useCallback(async (mapID: string | null, info: Partial<UpdateMapRequest>) => {
    console.log(mapID)
    if (!mapID || !mapInfo) return false;
    
    try {
      const resp = await updateMap(mapID, info);
      actions.updateMap({
        ...mapInfo,
        ...info
      });
      
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
  }, [mapID, mapInfo]);

  // 定时保存节点位置的间隔（毫秒）
  const AUTO_SAVE_INTERVAL = 30000; // 30秒

  // 保存提交位置到后端
  const savePosition = useCallback(async () => {
    if (!mapID || changedNodePositions.length === 0) return true;
    
    try {
      const changedNodes = nodes.filter(node => changedNodePositions.includes(node.id));
      const savePromises = changedNodes.map(async node => {
        const nodeData = {
          position: node.position,
        };
        
        await updateNode(mapID, node.id, nodeData);
      });
      
      await Promise.all(savePromises);
      // 所有节点都保存完成后，重置状态
      actions.clearChangedNodePositions();
      
      return true;
    } catch (error) {
      toast({
        title: '保存失败',
        description: '保存位置时出错',
        variant: 'destructive',
      });
      return false;
    }
  }, [mapID, changedNodePositions, nodes]);

  // 设置定时保存
  useEffect(() => {
    if (!mapID) return;

    const timer = setInterval(() => {
      if (changedNodePositions.length > 0) {
        savePosition();
      }
    }, AUTO_SAVE_INTERVAL);

    return () => clearInterval(timer);
  }, [mapID, changedNodePositions, savePosition]);

  // 初始化数据
  useEffect(() => {
    if (mapID) {
      console.log("init map", mapID)
      loadMap(mapID);
      loadNodes(mapID);
    }
  }, [mapID]);

  // 页面卸载前保存
  useEffect(() => {
    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      if (changedNodePositions.length > 0) {
        e.preventDefault();
        e.returnValue = '您有未保存的更改，确定要离开吗？';
      }
    };
    
    window.addEventListener('beforeunload', handleBeforeUnload);
    return () => window.removeEventListener('beforeunload', handleBeforeUnload);
  }, [changedNodePositions]);

  return {
    // 数据
    mapInfo,
    nodes,
    edges,
    isLoading,
    changedNodePositions,
    
    // 操作
    loadMap,
    loadNodes,
    saveMap,
    savePosition,
    
    // Store actions
    actions,
  };
}

export default useWorkspaceData;