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
      // 双击节点时展开侧边面板
      actions.openPanel(nodeId);
    },
    [actions]
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