/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/hooks/use-node-operations.ts
 */
import { useCallback } from 'react';
import { useWorkspaceStore } from '../store/workspace-store';
import { useToast } from '@/hooks/use-toast';

export const useNodeOperations = () => {
  const { actions } = useWorkspaceStore();
  const { toast } = useToast();

  const handleNodeEdit = useCallback(
    async (id: string) => {
      try {
        // 打开面板进行编辑
        actions.openPanel(id);
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
    async (id: string) => {
      try {
        // 删除节点及其相关边
        actions.deleteNode(id);
        toast({
          title: '删除成功',
          description: '节点已删除',
        });
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
    async (parentId: string) => {
      try {
        // 创建新节点
        const newNode = {
          id: `node-${Date.now()}`,
          type: 'custom',
          position: { x: 0, y: 0 }, // 位置将在store中根据父节点位置计算
          data: {
            question: '新问题',
            target: '新目标',
            status: 'pending' as const,
          },
        };

        // 添加节点和边
        actions.addChildNode(parentId, newNode);
        
        // 打开面板编辑新节点
        actions.openPanel(newNode.id);

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

  return {
    handleNodeEdit,
    handleNodeDelete,
    handleAddChild,
  };
};