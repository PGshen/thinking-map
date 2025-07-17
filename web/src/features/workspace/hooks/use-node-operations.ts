/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/hooks/use-node-operations.ts
 */
import { useCallback } from 'react';
import { useWorkspaceStore } from '../store/workspace-store';
import { useToast } from '@/hooks/use-toast';
import { CreateNodeResponse, CustomNodeModel } from '@/types/node';
import { updateNode, createNode, deleteNode } from '@/api/node';

export const useNodeOperations = () => {
  const { actions } = useWorkspaceStore();
  const { toast } = useToast();

  const handleNodeEdit = useCallback(
    async (mapID: string, nodeID: string, updates: Partial<CustomNodeModel>) => {
      if (!mapID) return;
      try {
        // If updates only contains isEditing and selected, skip backend API call
        const updateKeys = Object.keys(updates);
        const skipUpdate = updateKeys.length <= 2 && 
          updateKeys.every(key => key === 'isEditing' || key === 'selected');
        
        if (!skipUpdate) {
          if (nodeID.startsWith('temp-')) {
            // 如果是临时ID，说明是新增节点
            const response = await createNode(mapID, {
              nodeType: updates.nodeType || 'problem',
              question: updates.question || '',
              target: updates.target || '',
              mapID: mapID,
              parentID: updates.parentID || '',
              position: { x: 0, y: 0 } // 位置将在store中计算
            });
            if (response.code === 200) {
              actions.updateNodeID(nodeID, response.data.id);
              nodeID = response.data.id;  // 更新node为后端返回的ID
              toast({
                title: '创建成功',
                variant: 'info'
              });
            }
          } else {
            // 如果不是临时ID，说明是更新已有节点
            const response = await updateNode(mapID, nodeID, {
              question: updates.question || '',
              target: updates.target || ''
            });
            if (response.code === 200) {
              toast({
                title: '更新成功',
                variant: 'info'
              });
            }
          }
        }
        actions.updateNode(nodeID, { "data": { ...updates } } as Partial<CustomNodeModel>);
      } catch (error) {
        toast({
          title: '操作失败',
          description: '编辑节点时出现错误',
          variant: 'destructive',
        });
      }
    },
    [actions, toast]
  );

  const handleNodeDelete = useCallback(
    async (mapID: string, id: string) => {
      try {
        // 删除节点及其相关边
        const response = await deleteNode(mapID, id);
        if (response.code === 200) {
          actions.deleteNode(id);
          toast({
            title: '删除成功',
            description: '节点已删除',
          });
        }
      } catch (error) {
        toast({
          title: '操作失败',
          description: '删除节点时出现错误',
          variant: 'destructive',
        });
      }
    },
    [actions, toast]
  );

  const handleAddChild = useCallback(
    async (parentID: string) => {
      console.log("parentID",parentID)
      try {
        // 创建新节点
        const newNode = {
          id: `temp-${Date.now()}`,
          type: 'custom',
          position: { x: 0, y: 0 }, // 位置将在store中根据父节点位置计算
          data: {
            id: `temp-${Date.now()}`,
            parentID: parentID,
            nodeType: 'custom',
            question: '新问题',
            target: '新目标',
            status: 'initial' as const,
          },
        };

        // 添加节点和边
        actions.addChildNode(parentID, newNode);
        
        actions.selectNode(newNode.id);
        actions.setEditing(newNode.id); // 进入编辑状态

        toast({
          title: '添加成功',
          description: '新节点已创建',
        });
      } catch (error) {
        toast({
          title: '操作失败',
          description: '添加子节点时出现错误',
          variant: 'destructive',
        });
      }
    },
    [actions, toast]
  );

  const handleNodeUpdateID = useCallback(
    async (mapID: string | null, oldID: string, newID: string) => {
      console.log("mapID",mapID,"oldID",oldID,"newID",newID)
      try {
        actions.updateNodeID(oldID, newID);
        toast({
          title: '节点更新成功',
          variant: 'info'
        });
      } catch (error) {
        toast({
          title: '操作失败',
          description: '更新节点ID时出现错误',
          variant: 'destructive',
        });
      }
    },
    [actions, toast]
  );

  return {
    handleNodeEdit,
    handleNodeDelete,
    handleAddChild,
    handleNodeUpdateID,
  };
};