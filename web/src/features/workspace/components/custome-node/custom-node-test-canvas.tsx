"use client"
import React, { JSX, useCallback, useMemo } from 'react';
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
import { CustomNode } from './custom-node';
import type { CustomNodeModel } from '@/types/node';

// 示例节点数据
const initialNodes: Node<CustomNodeModel>[] = [
  {
    id: '1',
    type: 'custom',
    position: { x: 100, y: 100 },
    data: {
      id: '1',
      nodeType: 'problem',
      question: '如何设计用户友好的移动应用界面？',
      target: '制定完整的UI/UX设计方案',
      status: 'running',
      selected: false,
    },
  },
  {
    id: '2',
    type: 'custom',
    position: { x: 100, y: 250 },
    data: {
      id: '2',
      parentID: '1',
      nodeType: 'analysis',
      question: '用户研究与需求分析',
      target: '团队内部保持一致：最重要的是团队约定统一考虑后端风格：如果后端使用下划线，前端可以考虑保持一致使用转换工具',
      status: 'completed',
      conclusion: '团队内部保持一致：最重要的是团队约定统一考虑后端风格：如果后端使用下划线，前端可以考虑保持一致使用转换工具：可以使用 lodash 等工具库进行风格转换文档明确规定：在 API 文档中明确说明使用的命名风格',
      dependencies: [
      ],
      selected: false,
    },
  },
  {
    id: '3',
    type: 'custom',
    position: { x: 350, y: 250 },
    data: {
      id: '3',
      parentID: '1',
      nodeType: 'information',
      question: '界面设计与原型',
      target: '设计界面布局和元素',
      status: 'pending',
      selected: false,
    },
  },
];

const initialEdges: Edge[] = [
  { id: 'e1-2', source: '1', target: '2', type: 'smoothstep' },
  { id: 'e1-3', source: '1', target: '3', type: 'smoothstep' },
];


const nodeTypes = {
  'custom': CustomNode
}

export function CustomNodeTestCanvas(): JSX.Element {
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
      <div style={{ width: '100%', height: 850, background: '#f8fafc' }}>
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