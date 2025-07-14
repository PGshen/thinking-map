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
import { CustomNodeModel } from '@/types/node';

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
  nodes: Node<CustomNodeModel>[];
  edges: Edge[];
  
  // 思维导图信息
  mapId: string | null;
  mapInfo: Map | null;
  
  // UI状态
  isLoading: boolean;
  changedNodePositions: string[];
  
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
  setNodes: (nodes: Node<CustomNodeModel>[]) => void;
  addNode: (node: Node<CustomNodeModel>) => void;
  addChildNode: (parentId: string, childNode: Node<CustomNodeModel>) => void;
  updateNode: (nodeId: string, updates: Partial<Node<CustomNodeModel>>) => void;
  updateNodeId: (oldId: string, newId: string) => void;
  deleteNode: (nodeId: string) => void;
  selectNode: (nodeId: string | null) => void;
  setEditing: (nodeId: string | null) => void;
  clearSelection: () => void;
  
  // 边操作
  setEdges: (edges: Edge[]) => void;
  addEdge: (edge: Edge) => void;
  deleteEdge: (edgeId: string) => void;
  
  // 思维导图操作
  updateMap: (info: WorkspaceState['mapInfo']) => void;
  
  // 状态操作
  setLoading: (loading: boolean) => void;
  addChangedNodePosition: (nodeId: string) => void;
  getChangedNodePositions: () => string[];
  clearChangedNodePositions: () => void;
  
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
  panelWidth: 600,
  activeNodeId: null,
  selectedNodeIds: [],
  
  // 侧边栏状态
  sidebarOpen: true,
  
  // 节点和边数据
  nodes: [],
  edges: [],
  
  // 思维导图信息
  mapId: null,
  mapInfo: null,
  
  // UI状态
  isLoading: false,
  changedNodePositions: [],
  
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
        setNodes: (nodes: Node<CustomNodeModel>[]) => {
          set(
            (state) => ({ nodes: [...state.nodes, ...nodes] }),
            false,
            'setNodes'
          );
        },
        
        addNode: (node: Node<CustomNodeModel>) => {
          set(
            (state) => ({
              nodes: [...state.nodes, node],
            }),
            false,
            'addNode'
          );
        },

        addChildNode: (parentId: string, childNode: Node<CustomNodeModel>) => {
          const parentNode = get().nodes.find(node => node.id === parentId);
          if (!parentNode) return;

          // 计算子节点位置（在父节点右下方）
          const childPosition = {
            x: parentNode.position.x + (parentNode.width || 0) + 200,
            y: parentNode.position.y + (parentNode.height || 0) + 200
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
            type: 'default',
          };

          set(
            (state) => ({
              nodes: [...state.nodes, newNode],
              edges: [...state.edges, newEdge],
            }),
            false,
            'addChildNode'
          );
        },
        
        updateNode: (nodeId: string, updates: Partial<Node<CustomNodeModel>>) => {
          set(
            (state) => ({
              nodes: state.nodes.map((node) => {
                if (node.id !== nodeId) return node;
                
                // 处理 data 字段的局部更新
                const updatedNode = { ...node, ...updates };
                // console.log("updates",updates)
                // 如果 updates 中包含 data 字段，进行深度合并
                if (updates.data && node.data) {
                  updatedNode.data = {
                    ...node.data,
                    ...updates.data,
                  };
                }
                return updatedNode;
              }),
            }),
            false,
            'updateNode'
          );
        },

        updateNodeId: (oldId: string, newId: string) => {
          set(
            (state) => ({
              nodes: state.nodes.map((node) => {
                if (node.id !== oldId) return node;
                return {
                  ...node,
                  id: newId,
                  data: {
                    ...node.data,
                    id: newId
                  }
                };
              }),
              edges: state.edges.map((edge) => ({
                ...edge,
                source: edge.source === oldId ? newId : edge.source,
                target: edge.target === oldId ? newId : edge.target
              })),
              activeNodeId: state.activeNodeId === oldId ? newId : state.activeNodeId,
            }),
            false,
            'updateNodeId'
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
                data: {
                  ...node.data,
                  selected: node.id === nodeId,
                }
              })),
              selectedNodeIds: nodeId ? [nodeId] : [],
            }),
            false,
            'selectNode'
          );
        },

        setEditing: (nodeId: string | null) => {
          set(
            (state) => ({
              nodes: state.nodes.map((node) => ({
                ...node,
                data: {
                  ...node.data,
                  isEditing: node.id === nodeId ? !node.data.isEditing : false,
                },
              }))
            }),
            false,
            'setEditing'
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
            (state) => ({ edges: [...state.edges, ...edges] }),
            false,
            'setEdges'
          );
        },
        
        addEdge: (edge: Edge) => {
          set(
            (state) => ({
              edges: [...state.edges, edge],
            }),
            false,
            'addEdge'
          );
        },
        
        deleteEdge: (edgeId: string) => {
          set(
            (state) => ({
              edges: state.edges.filter((edge) => edge.id !== edgeId),
            }),
            false,
            'deleteEdge'
          );
        },
        
        updateMap: (info: Map) => {
          set(
            (state) => ({
              ...state,
              mapId: info.id,
              mapInfo: info,
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

        getChangedNodePositions: () => {
          return get().changedNodePositions;
        },
        
        addChangedNodePosition: (nodeId?: string) => {
          set(
            (state) => ({
              changedNodePositions: nodeId ? [...state.changedNodePositions, nodeId].filter((id): id is string => id !== undefined) : state.changedNodePositions
            }),
            false,
            'addChangedNodePosition'
          );
        },
        
        clearChangedNodePositions: () => {
          set(
            (state) => ({ changedNodePositions: [] }),
            false,
            'clearChangedNodePositions'
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