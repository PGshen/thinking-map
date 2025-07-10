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

import { CustomNode } from '@/features/workspace/components/custome-node/custom-node';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { useWorkspaceData } from '@/features/workspace/hooks/use-workspace-data';
import { useNodeSelection } from '@/features/workspace/hooks/use-node-selection';
import { useNodeOperations } from '@/features/workspace/hooks/use-node-operations';
import { CustomNodeModel } from '@/types/node';

interface VisualizationAreaProps {
  taskId: string;
}

// 定义节点类型
const nodeTypes: NodeTypes = {
  custom: CustomNode,
};

function MapCanvas({ taskId }: VisualizationAreaProps) {
  const { selectedNodeIds, actions } = useWorkspaceStore();
  const { nodes: nodesData, edges: edgesData, isLoading } = useWorkspaceData(taskId);
  const { handleNodeClick, handleNodeDoubleClick, handleNodeContextMenu } = useNodeSelection();
  const { handleNodeEdit, handleNodeDelete, handleAddChild } = useNodeOperations();
  
  const [nodes, setNodes, onNodesChange] = useNodesState<Node[]>([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge[]>([]);

  // 转换数据格式为ReactFlow需要的格式
  const convertToReactFlowNodes = useCallback((nodeData: any[]): Node[] => {
    return nodeData.map((node) => ({
      id: node.id,
      type: 'custom',
      position: node.position || { x: 0, y: 0 },
      data: {
        ...node,
        selected: selectedNodeIds.includes(node.id),
        onEdit: handleNodeEdit,
        onDelete: handleNodeDelete,
        onAddChild: handleAddChild,
        onSelect: handleNodeClick,
        onDoubleClick: handleNodeDoubleClick,
        onContextMenu: handleNodeContextMenu,
      } as CustomNodeModel,
    }));
  }, [selectedNodeIds, handleNodeClick, handleNodeDoubleClick, handleNodeContextMenu]);

  const convertToReactFlowEdges = useCallback((edgeData: any[]): Edge[] => {
    return edgeData.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      type: 'smoothstep',
      animated: false,
      style: {
        stroke: '#94a3b8',
        strokeWidth: 2,
      },
    }));
  }, []);

  // 更新节点和边数据
  useEffect(() => {
    if (nodesData && edgesData) {
      const reactFlowNodes = convertToReactFlowNodes(nodesData);
      const reactFlowEdges = convertToReactFlowEdges(edgesData);
      
      setNodes(reactFlowNodes);
      setEdges(reactFlowEdges);
      
      // 同步到store
      actions.setNodes(nodesData);
      actions.setEdges(edgesData);
    }
  }, [nodesData, edgesData, convertToReactFlowNodes, convertToReactFlowEdges, actions]);

  // 当选中状态变化时更新节点
  useEffect(() => {
    setNodes((nds) =>
      nds.map((node) => ({
        ...node,
        data: {
          ...node.data,
          selected: selectedNodeIds.includes(node.id),
        },
      }))
    );
  }, [selectedNodeIds]);

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
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-2"></div>
          <p className="text-sm text-muted-foreground">加载思维导图中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="w-full h-full bg-gray-50">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onPaneClick={onPaneClick}
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

export function VisualizationArea({ taskId }: VisualizationAreaProps) {
  return (
    <ReactFlowProvider>
      <MapCanvas taskId={taskId} />
    </ReactFlowProvider>
  );
}

export default VisualizationArea;