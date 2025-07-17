import { useCallback } from 'react';
import { useWorkspaceStore } from '../store/workspace-store';

export const useNodeSelection = () => {
  const { actions } = useWorkspaceStore();

  const handleNodeClick = useCallback(
    (nodeID: string) => {
      actions.selectNode(nodeID);
    },
    [actions]
  );

  const handleNodeDoubleClick = useCallback(
    (nodeID: string) => {
      // 双击节点时展开侧边面板
      actions.openPanel(nodeID);
    },
    [actions]
  );

  const handleNodeContextMenu = useCallback(
    (nodeID: string, event: React.MouseEvent) => {
      event.preventDefault();
      // TODO: 实现右键菜单功能
      console.log('Context menu for node:', nodeID);
    },
    []
  );

  return {
    handleNodeClick,
    handleNodeDoubleClick,
    handleNodeContextMenu,
  };
};