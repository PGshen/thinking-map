/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/hooks/use-workspace-data.ts
 */
import { useEffect, useCallback } from 'react';
import { useWorkspaceStore } from '../store/workspace-store';
import { useToast } from '@/hooks/use-toast';

// TODO: 替换为实际的API调用
import { Map } from '@/types/map';

// 节点数据接口

interface NodeData {
  id: string;
  question: string;
  target: string;
  context?: string;
  conclusion?: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  dependencies?: Array<{
    id: string;
    name: string;
    status: string;
  }>;
  children?: NodeData[];
  position: { x: number; y: number };
}

// 模拟API调用
const mockApi = {
  // 获取思维导图信息
  getMap: async (mapId: string): Promise<Map> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    return {
      id: mapId,
      title: '示例思维导图',
      problem: '这是一个示例思维导图，用于演示工作区功能',
      status: 1,
      createdAt: '2025-01-27T10:00:00Z',
      updatedAt: '2025-01-27T10:30:00Z',
    };
  },
  
  // 获取节点数据
  getNodes: async (mapId: string): Promise<NodeData[]> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    return [
      {
        id: 'root',
        question: '如何提高团队协作效率？',
        target: '建立高效的团队协作机制',
        context: '当前团队在项目协作中存在沟通不畅、任务分配不明确等问题',
        status: 'running',
        position: { x: 250, y: 50 },
        dependencies: [],
        children: [
          {
            id: 'sub-1',
            question: '分析当前协作问题',
            target: '识别协作中的主要障碍',
            status: 'completed',
            position: { x: 100, y: 200 },
            conclusion: '主要问题包括：沟通渠道混乱、任务优先级不明确、进度跟踪困难',
          },
          {
            id: 'sub-2',
            question: '制定改进方案',
            target: '设计可行的协作改进策略',
            status: 'running',
            position: { x: 400, y: 200 },
          },
        ],
      },
    ];
  },
  
  // 更新思维导图详细信息
  updateMap: async (mapId: string, info: Partial<Map>): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    // 模拟API调用
  },
  
  // 更新节点信息
  updateNode: async (nodeId: string, updates: Partial<NodeData>): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    // 模拟API调用
  },
  
  // 保存工作区状态
  saveWorkspace: async (mapId: string, nodes: NodeData[]): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    // 模拟API调用
  },
};

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
      const mapInfo = await mockApi.getMap(id);
      actions.setMap(mapInfo.id, mapInfo.title, mapInfo.problem);
      actions.updateMap(mapInfo);
    } catch (error) {
      toast({
        title: '加载失败',
        description: '无法加载思维导图信息，请重试',
        variant: 'destructive',
      });
    } finally {
      actions.setLoading(false);
    }
  }, []); // 移除actions和toast依赖，避免无限循环

  // 加载节点数据
  const loadNodes = useCallback(async (id: string) => {
    if (!id) return;
    
    actions.setLoading(true);
    try {
      const nodeData = await mockApi.getNodes(id);
      
      // 转换为ReactFlow格式
      const reactFlowNodes = nodeData.flatMap(node => {
        const nodes = [{
          id: node.id,
          type: 'custom',
          position: node.position,
          data: {
            question: node.question,
            target: node.target,
            context: node.context,
            conclusion: node.conclusion,
            status: node.status,
            dependencies: node.dependencies,
          },
        }];
        
        // 添加子节点
        if (node.children) {
          nodes.push(...node.children.map(child => ({
            id: child.id,
            type: 'custom',
            position: child.position,
            data: {
              question: child.question,
              target: child.target,
              context: child.context,
              conclusion: child.conclusion,
              status: child.status,
              dependencies: child.dependencies,
            },
          })));
        }
        
        return nodes;
      });
      
      // 生成边
      const reactFlowEdges = nodeData.flatMap(node => {
        if (!node.children) return [];
        return node.children.map(child => ({
          id: `${node.id}-${child.id}`,
          source: node.id,
          target: child.id,
          type: 'default',
        }));
      });
      
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
  }, []); // 移除actions和toast依赖，避免无限循环

  // 保存思维导图详细信息
  const saveMap = useCallback(async (info: Partial<Omit<Map, 'id' | 'createdAt' | 'updatedAt'>>) => {
    if (!mapId || !mapInfo) return false;
    
    try {
      await mockApi.updateMap(mapId, info);
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
  }, [mapId, mapInfo]); // 移除actions和toast依赖，避免无限循环

  // 保存节点信息
  const saveNodeInfo = useCallback(async (nodeId: string, updates: any) => {
    try {
      await mockApi.updateNode(nodeId, updates);
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
  }, []); // 移除actions和toast依赖，避免无限循环

  // 保存整个工作区
  const saveWorkspace = useCallback(async () => {
    if (!mapId || !hasUnsavedChanges) return true;
    
    actions.setLoading(true);
    try {
      // 转换节点数据格式
      const nodeData = nodes.map(node => ({
        id: node.id,
        question: node.data.question,
        target: node.data.target,
        context: node.data.context,
        conclusion: node.data.conclusion,
        status: node.data.status,
        position: node.position,
        dependencies: node.data.dependencies,
      }));
      
      await mockApi.saveWorkspace(mapId, nodeData);
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

  // 自动保存
  // useEffect(() => {
  //   if (!hasUnsavedChanges || !mapId) return;
    
  //   const autoSaveTimer = setTimeout(() => {
  //     saveWorkspace();
  //   }, 30000); // 30秒后自动保存
    
  //   return () => clearTimeout(autoSaveTimer);
  // }, [hasUnsavedChanges, mapId]);

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