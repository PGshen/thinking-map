/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/visualization-area/visualization-area.tsx
 */
'use client';

import React, { useCallback, useEffect } from 'react';
import ReactFlow, {
  ReactFlowProvider,
  Background,
  Controls,
  MiniMap,
  useNodesState,
  useEdgesState,
  addEdge,
  Connection,
  Edge,
  Node,
  NodeTypes,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Loader } from 'lucide-react';

import { CustomNode } from '@/features/workspace/components/custome-node/custom-node';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { useWorkspaceData } from '@/features/workspace/hooks/use-workspace-data';
import { useNodeSelection } from '@/features/workspace/hooks/use-node-selection';
import { useNodeOperations } from '@/features/workspace/hooks/use-node-operations';
import { useSSEConnection } from '@/hooks/use-sse-connection';
import { SSEStatusIndicator } from '@/components/sse-status-indicator';
import { CustomNodeModel } from '@/types/node';
import { NodeCreatedEvent, NodeUpdatedEvent } from '@/types/sse';
interface VisualizationAreaProps {
  mapID: string;
}

// 定义节点类型
const nodeTypes: NodeTypes = {
  custom: CustomNode,
};

function MapCanvas({ mapID }: VisualizationAreaProps) {
  const { selectedNodeIDs, actions } = useWorkspaceStore();
  const { nodes: nodesData, edges: edgesData, isLoading } = useWorkspaceData(mapID);
  const { handleNodeClick, handleNodeDoubleClick, handleNodeContextMenu } = useNodeSelection();
  const { handleNodeEdit, handleNodeDelete, handleAddChild, handleNodeUpdateID } = useNodeOperations();

  const [nodes, setNodes, onNodesChange] = useNodesState<Node<CustomNodeModel>[]>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge[]>([]);

  // SSE事件处理函数 - 使用useCallback优化，避免不必要的重新渲染
  const handleNodeCreated = useCallback((event: NodeCreatedEvent) => {
    console.log('Received node created event:', event);
    
    // 创建新节点
    const newNode: Node<CustomNodeModel> = {
      id: event.nodeID,
      type: 'custom',
      position: event.position,
      data: {
        id: event.nodeID,
        parentID: event.parentID,
        nodeType: event.nodeType,
        question: event.question,
        target: event.target,
        status: 'pending',
        selected: false,
      },
    };

    // 如果有父节点，创建连接边
    if (event.parentID) {
      actions.addChildNode(event.parentID, newNode);
    } else {
      // 同步到store
      actions.addNode(newNode);
    }
  }, [actions]);

  const handleNodeUpdated = useCallback((event: NodeUpdatedEvent) => {
    console.log('Received node updated event:', event);
    
    // 根据更新模式处理数据
    let processedUpdates = event.updates;
    
    if (event.mode === 'append') {
      // 对于append模式，需要先获取当前节点数据进行字符串追加
      const currentNodes = useWorkspaceStore.getState().nodes;
      const currentNode = currentNodes.find(node => node.id === event.nodeID);
      
      if (currentNode) {
        processedUpdates = { ...event.updates };
        Object.keys(event.updates).forEach((key) => {
          const currentValue = (currentNode.data as any)[key];
          const updateValue = event.updates[key];
          
          // 对字符串类型进行追加操作
          if (typeof currentValue === 'string' && typeof updateValue === 'string') {
            processedUpdates[key] = currentValue + updateValue;
          }
          // 其他类型直接替换
        });
      }
    }
    // replace模式直接使用原始updates
    
    // 通过store更新节点数据，工作区会自动监听store变化更新画布
    actions.updateNode(event.nodeID, { data: processedUpdates} as Partial<CustomNodeModel> );
  }, [actions]);

  const handleConnectionEstablished = useCallback((data: any) => {
    console.log('SSE connection established:', data);
    // 可以在这里显示连接状态或执行其他初始化操作
  }, []);

  // 建立SSE连接
  const { isConnected } = useSSEConnection({
    mapID,
    onNodeCreated: handleNodeCreated,
    onNodeUpdated: handleNodeUpdated,
    onConnectionEstablished: handleConnectionEstablished,
  });

  // 节点拖动结束时同步到 store
  const onNodeDragStop = useCallback((event: any, draggedNode: any) => {
    actions.updateNode(draggedNode.id, { position: draggedNode.position });
    actions.addChangedNodePosition(draggedNode.id);
  }, [actions]);

  // 注入事件到节点 data
  const nodesWithHandlers = useCallback((nodeData: any[]): Node[] => {
    return nodeData.map((node) => ({
      ...node,
      data: {
        ...node.data,
        onEdit: handleNodeEdit,
        onDelete: handleNodeDelete,
        onAddChild: handleAddChild,
        onSelect: handleNodeClick,
        onDoubleClick: handleNodeDoubleClick,
        onContextMenu: handleNodeContextMenu,
        onUpdateID: handleNodeUpdateID,
      } as any,
    }));
  }, []);

  // 从state中同步
  useEffect(() => {
    if (nodesData.length > 0) {
      setNodes(nodesWithHandlers(nodesData));
      setEdges(edgesData);
    }
  }, [nodesData, edgesData]);

  // 当选中状态变化时更新节点
  useEffect(() => {
    setNodes((nds) =>
      nds.map((node) => ({
        ...node,
        data: {
          ...node.data,
          selected: selectedNodeIDs.includes(node.id),
        },
      }))
    );
  }, [selectedNodeIDs]);

  const onConnect = useCallback(
    (params: Connection) => {
      const newEdge = {
        ...params,
        id: `${params.source}-${params.target}`,
        type: 'smoothstep',
        animated: false,
      };
      setEdges((eds) => addEdge(newEdge, eds));

      // TODO: 同步到后端
      console.log('New connection:', newEdge);
    },
    []
  );

  const onPaneClick = useCallback(() => {
    // 点击空白区域取消选中
    actions.clearSelection();
  }, [actions]);

  if (isLoading) {
    return (
      <div className="w-full h-full flex items-center justify-center bg-gray-50">
        <div className="text-center">
          <Loader className="h-8 w-8 animate-spin text-primary mx-auto mb-3" />
          <p className="text-sm text-muted-foreground">加载思维导图中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full h-full bg-gray-50 relative">
      {/* SSE连接状态指示器 */}
      <div className="absolute top-4 right-4 z-10">
        <SSEStatusIndicator isConnected={isConnected} />
      </div>
      
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onPaneClick={onPaneClick}
        onNodeDragStop={onNodeDragStop}
        nodeTypes={nodeTypes}
        fitView
        attributionPosition="bottom-left"
        className="bg-gray-50"
      >
        <Background color="#e2e8f0" gap={20} size={1} />
        <Controls className="bg-white border border-gray-200 rounded-lg shadow-sm" />
        <MiniMap />
      </ReactFlow>
    </div>
  );
}

export function VisualizationArea({ mapID }: VisualizationAreaProps) {
  return (
    <ReactFlowProvider>
      <MapCanvas mapID={mapID} />
    </ReactFlowProvider>
  );
}

export default VisualizationArea;