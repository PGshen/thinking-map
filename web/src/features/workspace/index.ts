/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/index.ts
 */

// 主要组件
export { default as WorkspaceLayout } from './components/workspace-layout';

// 可视化区域组件
export { default as VisualizationArea } from './components/visualization-area/visualization-area';

export { default as Settings } from './components/side-bar/settings'
export { default as MapInfo } from './components/side-bar/map-info'

// 操作面板组件
export { default as OperationPanel } from './components/operation-panel/operation-panel';
export { default as PanelTabs } from './components/operation-panel/panel-tabs';
export { default as InfoTab } from './components/operation-panel/info-tab';
export { default as DecomposeTab } from './components/operation-panel/decompose-tab';
export { default as ConclusionTab } from './components/operation-panel/conclusion-tab';

// Store
export {
  useWorkspaceStore,
  useWorkspaceStoreData,
  usePanelState,
  useWorkspaceSettings,
} from './store/workspace-store';
export type { WorkspaceState, WorkspaceActions } from './store/workspace-store';

// Hooks
export { default as useWorkspaceData } from './hooks/use-workspace-data';

// 类型定义
export interface WorkspaceProps {
  taskID: string;
  className?: string;
}

export interface NodeActionEvent {
  nodeID: string;
  action: 'select' | 'doubleClick' | 'rightClick';
  event?: React.MouseEvent;
}

export interface PanelConfig {
  minWidth: number;
  maxWidth: number;
  defaultWidth: number;
}

// 常量
export const WORKSPACE_CONSTANTS = {
  PANEL: {
    MIN_WIDTH_RATIO: 0.25,
    MAX_WIDTH_RATIO: 0.5,
    DEFAULT_WIDTH: 400,
  },
  AUTOSAVE: {
    INTERVAL: 30000, // 30秒
  },
  VIEWPORT: {
    DEFAULT_ZOOM: 1,
    MIN_ZOOM: 0.1,
    MAX_ZOOM: 2,
  },
} as const;