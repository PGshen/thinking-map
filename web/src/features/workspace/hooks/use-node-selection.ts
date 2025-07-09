import { useCallback } from 'react';
import { useWorkspaceStore } from '../store/workspace-store';

export const useNodeSelection = () => {
  const { actions } = useWorkspaceStore();

  const handleNodeClick = useCallback(
    (nodeId: string) => {
      actions.selectNode(nodeId);
    },
    [actions]
  );

  const handleNodeDoubleClick = useCallback(
    (nodeId: string) => {
      // TODO: 实现节点编辑功能
      console.log('Double click node:', nodeId);
    },
    []
  );

  const handleNodeContextMenu = useCallback(
    (nodeId: string, event: React.MouseEvent) => {
      event.preventDefault();
      // TODO: 实现右键菜单功能
      console.log('Context menu for node:', nodeId);
    },
    []
  );

  return {
    handleNodeClick,
    handleNodeDoubleClick,
    handleNodeContextMenu,
  };
};