/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/store/workspace-store.ts
 */
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { Node, Edge } from 'reactflow';
import { Map } from '@/types/map';

// 扩展Node类型以包含自定义属性
interface CustomNodeData {
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
}

// 工作区状态接口
interface WorkspaceState {
  // 面板状态
  isPanelOpen: boolean;
  panelOpen: boolean; // 别名，保持兼容性
  panelWidth: number;
  activeNodeId: string | null;
  selectedNodeIds: string[];
  
  // 侧边栏状态
  sidebarOpen: boolean;
  
  // 节点和边数据
  nodes: Node[];
  edges: Edge[];
  nodesData: Node[]; // 别名，保持兼容性
  edgesData: Edge[]; // 别名，保持兼容性
  
  // 思维导图信息
  mapId: string | null;
  mapInfo: Map | null;
  
  // UI状态
  isLoading: boolean;
  hasUnsavedChanges: boolean;
  
  // 视图状态
  viewportState: {
    x: number;
    y: number;
    zoom: number;
  };
  
  // 设置
  settings: {
    autoSave: boolean;
    showMinimap: boolean;
    showControls: boolean;
    layoutDirection: 'TB' | 'LR' | 'BT' | 'RL';
    nodeSpacing: number;
    edgeType: 'default' | 'straight' | 'step' | 'smoothstep';
  };
}

// 操作接口
interface WorkspaceActions {
  // 面板操作
  openPanel: (nodeId: string) => void;
  closePanel: () => void;
  setPanelWidth: (width: number) => void;
  
  // 侧边栏操作
  toggleSidebar: () => void;
  
  // 节点操作
  setNodes: (nodes: Node[]) => void;
  addNode: (node: Node) => void;
  addChildNode: (parentId: string, childNode: Node) => void;
  updateNode: (nodeId: string, updates: Partial<Node & CustomNodeData>) => void;
  deleteNode: (nodeId: string) => void;
  selectNode: (nodeId: string | null) => void;
  clearSelection: () => void;
  
  // 边操作
  setEdges: (edges: Edge[]) => void;
  addEdge: (edge: Edge) => void;
  deleteEdge: (edgeId: string) => void;
  
  // 思维导图操作
  setMap: (mapId: string, title: string, problem?: string) => void;
  updateMapTitle: (title: string) => void;
  updateMap: (info: WorkspaceState['mapInfo']) => void;
  
  // 状态操作
  setLoading: (loading: boolean) => void;
  setUnsavedChanges: (hasChanges: boolean) => void;
  
  // 视图操作
  setViewportState: (viewport: { x: number; y: number; zoom: number }) => void;
  
  // 设置操作
  updateSettings: (settings: Partial<WorkspaceState['settings']>) => void;
  
  // 重置操作
  reset: () => void;
}

// 初始状态
const initialState: WorkspaceState = {
  // 面板状态
  isPanelOpen: false,
  panelOpen: false,
  panelWidth: 400,
  activeNodeId: null,
  selectedNodeIds: [],
  
  // 侧边栏状态
  sidebarOpen: true,
  
  // 节点和边数据
  nodes: [],
  edges: [],
  nodesData: [],
  edgesData: [],
  
  // 思维导图信息
  mapId: null,
  mapInfo: null,
  
  // UI状态
  isLoading: false,
  hasUnsavedChanges: false,
  
  // 视图状态
  viewportState: {
    x: 0,
    y: 0,
    zoom: 1,
  },
  
  // 设置
  settings: {
    autoSave: true,
    showMinimap: true,
    showControls: true,
    layoutDirection: 'TB',
    nodeSpacing: 100,
    edgeType: 'default',
  },
};

// 创建store
export const useWorkspaceStore = create<WorkspaceState & { actions: WorkspaceActions }>()(
  devtools(
    (set, get) => ({
      ...initialState,
      
      actions: {
        // 侧边栏操作
        toggleSidebar: () => {
          set(
            (state) => ({
              sidebarOpen: !state.sidebarOpen,
            }),
            false,
            'toggleSidebar'
          );
        },

        // 面板操作
        openPanel: (nodeId: string) => {
          set(
            (state) => ({
              isPanelOpen: true,
              panelOpen: true,
              activeNodeId: nodeId,
            }),
            false,
            'openPanel'
          );
        },
        
        closePanel: () => {
          set(
            (state) => ({
              isPanelOpen: false,
              panelOpen: false,
              activeNodeId: null,
            }),
            false,
            'closePanel'
          );
        },
        
        setPanelWidth: (width: number) => {
          // 限制面板宽度在25%-50%之间
          const minWidth = window.innerWidth * 0.25;
          const maxWidth = window.innerWidth * 0.5;
          const clampedWidth = Math.max(minWidth, Math.min(maxWidth, width));
          
          set(
            (state) => ({ panelWidth: clampedWidth }),
            false,
            'setPanelWidth'
          );
        },
        
        // 节点操作
        setNodes: (nodes: Node[]) => {
          set(
            (state) => ({ nodes, nodesData: nodes, hasUnsavedChanges: true }),
            false,
            'setNodes'
          );
        },
        
        addNode: (node: Node) => {
          set(
            (state) => ({
              nodes: [...state.nodes, node],
              hasUnsavedChanges: true,
            }),
            false,
            'addNode'
          );
        },

        addChildNode: (parentId: string, childNode: Node) => {
          const parentNode = get().nodes.find(node => node.id === parentId);
          if (!parentNode) return;

          // 计算子节点位置（在父节点右下方）
          const childPosition = {
            x: parentNode.position.x + 200,
            y: parentNode.position.y + 100
          };

          // 创建新节点
          const newNode = {
            ...childNode,
            position: childPosition,
          };

          // 创建连接边
          const newEdge = {
            id: `${parentId}-${childNode.id}`,
            source: parentId,
            target: childNode.id,
            type: 'smoothstep',
          };

          set(
            (state) => ({
              nodes: [...state.nodes, newNode],
              edges: [...state.edges, newEdge],
              hasUnsavedChanges: true,
            }),
            false,
            'addChildNode'
          );
        },
        
        updateNode: (nodeId: string, updates: Partial<Node>) => {
          set(
            (state) => ({
              nodes: state.nodes.map((node) =>
                node.id === nodeId ? { ...node, ...updates } : node
              ),
              hasUnsavedChanges: true,
            }),
            false,
            'updateNode'
          );
        },
        
        deleteNode: (nodeId: string) => {
          set(
            (state) => ({
              nodes: state.nodes.filter((node) => node.id !== nodeId),
              edges: state.edges.filter(
                (edge) => edge.source !== nodeId && edge.target !== nodeId
              ),
              activeNodeId: state.activeNodeId === nodeId ? null : state.activeNodeId,
              isPanelOpen: state.activeNodeId === nodeId ? false : state.isPanelOpen,
              hasUnsavedChanges: true,
            }),
            false,
            'deleteNode'
          );
        },
        
        selectNode: (nodeId: string | null) => {
          set(
            (state) => ({
              nodes: state.nodes.map((node) => ({
                ...node,
                selected: node.id === nodeId,
              })),
              selectedNodeIds: nodeId ? [nodeId] : [],
            }),
            false,
            'selectNode'
          );
        },
        
        clearSelection: () => {
          set(
            (state) => ({
              nodes: state.nodes.map((node) => ({
                ...node,
                selected: false,
              })),
              selectedNodeIds: [],
            }),
            false,
            'clearSelection'
          );
        },
        
        // 边操作
        setEdges: (edges: Edge[]) => {
          set(
            (state) => ({ edges, edgesData: edges, hasUnsavedChanges: true }),
            false,
            'setEdges'
          );
        },
        
        addEdge: (edge: Edge) => {
          set(
            (state) => ({
              edges: [...state.edges, edge],
              hasUnsavedChanges: true,
            }),
            false,
            'addEdge'
          );
        },
        
        deleteEdge: (edgeId: string) => {
          set(
            (state) => ({
              edges: state.edges.filter((edge) => edge.id !== edgeId),
              hasUnsavedChanges: true,
            }),
            false,
            'deleteEdge'
          );
        },
        
        // 思维导图操作
        setMap: (mapId: string, title: string, problem = '') => {
          set(
            (state) => ({
              ...state,
              mapId,
              mapTitle: title,
              mapProblem: problem,
            }),
            false,
            'setMap'
          );
        },
        
        updateMap: (info: Map) => {
          set(
            (state) => ({
              ...state,
              mapId: info.id,
              mapTitle: info.title,
              mapProblem: info.problem,
              mapInfo: info,
              hasUnsavedChanges: true,
            }),
            false,
            'updateMap'
          );
        },
        
        // 状态操作
        setLoading: (loading: boolean) => {
          set(
            (state) => ({ isLoading: loading }),
            false,
            'setLoading'
          );
        },
        
        setUnsavedChanges: (hasChanges: boolean) => {
          set(
            (state) => ({ hasUnsavedChanges: hasChanges }),
            false,
            'setUnsavedChanges'
          );
        },
        
        // 视图操作
        setViewportState: (viewport: { x: number; y: number; zoom: number }) => {
          set(
            (state) => ({ viewportState: viewport }),
            false,
            'setViewportState'
          );
        },
        
        // 设置操作
        updateSettings: (newSettings: Partial<WorkspaceState['settings']>) => {
          set(
            (state) => ({
              settings: { ...state.settings, ...newSettings },
              hasUnsavedChanges: true,
            }),
            false,
            'updateSettings'
          );
        },
        
        // 重置操作
        reset: () => {
          set(
            () => ({ ...initialState }),
            false,
            'reset'
          );
        },
      },
    }),
    {
      name: 'workspace-store',
      partialize: (state: WorkspaceState & { actions: WorkspaceActions }) => ({
        // 只持久化设置和视图状态
        settings: state.settings,
        viewportState: state.viewportState,
        panelWidth: state.panelWidth,
      }),
    }
  )
);

// 导出类型
export type { WorkspaceState, WorkspaceActions };

// 选择器函数
export const useWorkspaceStoreData = () => {
  const store = useWorkspaceStore();
  return {
    mapId: store.mapId,
    mapInfo: store.mapInfo,
    nodes: store.nodes,
    edges: store.edges,
    isLoading: store.isLoading,
    hasUnsavedChanges: store.hasUnsavedChanges,
  };
};

export const usePanelState = () => {
  const store = useWorkspaceStore();
  return {
    isPanelOpen: store.isPanelOpen,
    panelWidth: store.panelWidth,
    activeNodeId: store.activeNodeId,
  };
};

export const useWorkspaceSettings = () => {
  const store = useWorkspaceStore();
  return {
    settings: store.settings,
    viewportState: store.viewportState,
  };
};