"use client"
import React, { useCallback, useMemo } from 'react';
import ReactFlow, {
  ReactFlowProvider,
  addEdge,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  Node,
  Edge,
  Connection,
  useNodesState,
  useEdgesState,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { CustomNode } from './CustomNode';
import type { CustomNodeModel } from './CustomNodeModel';

// 示例节点数据
const initialNodes: Node<CustomNodeModel>[] = [
  {
    id: '1',
    type: 'custom',
    position: { x: 100, y: 100 },
    data: {
      id: '1',
      nodeType: '根节点',
      question: '如何设计用户友好的移动应用界面？',
      target: '制定完整的UI/UX设计方案',
      status: 'running',
      selected: false,
      childCount: 2,
    },
  },
  {
    id: '2',
    type: 'custom',
    position: { x: 100, y: 250 },
    data: {
      id: '2',
      parentId: '1',
      nodeType: '分析',
      question: '用户研究与需求分析',
      target: '了解用户需求和行为',
      status: 'completed',
      conclusion: '已完成用户调研',
      selected: false,
      childCount: 0,
    },
  },
  {
    id: '3',
    type: 'custom',
    position: { x: 350, y: 250 },
    data: {
      id: '3',
      parentId: '1',
      nodeType: '分析',
      question: '界面设计与原型',
      target: '设计界面布局和元素',
      status: 'pending',
      selected: false,
      childCount: 0,
      hasUnmetDependencies: true,
    },
  },
];

const initialEdges: Edge[] = [
  { id: 'e1-2', source: '1', target: '2', type: 'smoothstep' },
  { id: 'e1-3', source: '1', target: '3', type: 'smoothstep' },
];

// 组件外部常量，nodeTypes永远不变
export const nodeTypes = {
  custom: CustomNode,
};

export const CustomNodeTestCanvas: React.FC = () => {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // 事件处理函数全部用 useCallback，依赖项最小化
  const handleEdit = useCallback((id: string) => {
    alert(`编辑节点: ${id}`);
  }, []);
  const handleDelete = useCallback((id: string) => {
    alert(`删除节点: ${id}`);
  }, []);
  const handleAddChild = useCallback((id: string) => {
    alert(`添加子节点到: ${id}`);
  }, []);
  const handleSelect = useCallback((id: string) => {
    setNodes(nds => nds.map(n => ({ ...n, data: { ...n.data, selected: n.id === id } })));
  }, [setNodes]);
  const handleDoubleClick = useCallback((id: string) => {
    alert(`双击节点: ${id}`);
  }, []);
  const handleContextMenu = useCallback((id: string, e: React.MouseEvent) => {
    alert(`右键节点: ${id}`);
  }, []);

  // 注入事件到节点 data
  const nodesWithHandlers = nodes.map(node => ({
    ...node,
    data: {
      ...node.data,
      onEdit: handleEdit,
      onDelete: handleDelete,
      onAddChild: handleAddChild,
      onSelect: handleSelect,
      onDoubleClick: handleDoubleClick,
      onContextMenu: handleContextMenu,
    },
  }));

  return (
    <ReactFlowProvider>
      <div style={{ width: '100%', height: 500, background: '#f8fafc' }}>
        <ReactFlow
          nodes={nodesWithHandlers}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={params => setEdges(eds => addEdge(params, eds))}
          nodeTypes={nodeTypes}
          fitView
        >
          <MiniMap />
          <Controls />
          <Background variant={BackgroundVariant.Dots} gap={16} size={1} />
        </ReactFlow>
      </div>
    </ReactFlowProvider>
  );
}; 