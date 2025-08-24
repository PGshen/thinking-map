/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/utils/layout-utils.ts
 */

import dagre from 'dagre';
import { Node, Edge } from 'reactflow';
import { CustomNodeModel } from '@/types/node';
import { measureNodeContentSizeWithCache } from './dom-measurement';

// 布局配置
export interface LayoutConfig {
  direction: 'TB' | 'BT' | 'LR' | 'RL'; // 布局方向
  nodeWidth: number; // 默认节点宽度
  nodeHeight: number; // 默认节点高度
  rankSep: number; // 层级间距
  nodeSep: number; // 节点间距
  edgeSep: number; // 边间距
}

// 默认布局配置
export const DEFAULT_LAYOUT_CONFIG: LayoutConfig = {
  direction: 'TB',
  nodeWidth: 200,
  nodeHeight: 100,
  rankSep: 120,
  nodeSep: 50,
  edgeSep: 20,
};

// 获取节点的实际尺寸
export const getNodeDimensions = (
  node: Node<CustomNodeModel>, 
  nodeSizesMap?: Map<string, { width: number; height: number }>
) => {
  // 优先使用实际DOM测量的尺寸
  if (nodeSizesMap && nodeSizesMap.has(node.id)) {
    const actualSize = nodeSizesMap.get(node.id)!;
    return {
      width: actualSize.width,
      height: actualSize.height,
    };
  }
  
  // 回退到DOM预测量
  const questionText = node.data.question || '';
  const targetText = node.data.target || '';
  const conclusionText = (node.data.status === 'completed' && node.data.conclusion?.content) 
    ? node.data.conclusion.content 
    : '';
  
  // 使用DOM测量获取实际尺寸
  const measuredSize = measureNodeContentSizeWithCache(
    questionText,
    targetText,
    conclusionText,
    360 // 最大宽度对应 max-w-[360px]
  );
  
  return {
    width: measuredSize.width,
    height: measuredSize.height,
  };
};

// 全局重新布局
export const applyGlobalLayout = (
  nodes: Node<CustomNodeModel>[],
  edges: Edge[],
  config: Partial<LayoutConfig> = {},
  nodeSizesMap?: Map<string, { width: number; height: number }>
): { nodes: Node<CustomNodeModel>[]; edges: Edge[] } => {
  const layoutConfig = { ...DEFAULT_LAYOUT_CONFIG, ...config };
  
  // 创建dagre图
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));
  dagreGraph.setGraph({
    rankdir: layoutConfig.direction,
    ranksep: layoutConfig.rankSep,
    nodesep: layoutConfig.nodeSep,
    edgesep: layoutConfig.edgeSep,
  });

  // 添加节点到dagre图
  nodes.forEach((node) => {
    const dimensions = getNodeDimensions(node, nodeSizesMap);
    dagreGraph.setNode(node.id, {
      width: dimensions.width,
      height: dimensions.height,
    });
  });

  // 添加边到dagre图
  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  // 执行布局
  dagre.layout(dagreGraph);

  // 更新节点位置
  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    const dimensions = getNodeDimensions(node, nodeSizesMap);
    
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - dimensions.width / 2,
        y: nodeWithPosition.y - dimensions.height / 2,
      },
    };
  });

  return {
    nodes: layoutedNodes,
    edges,
  };
};

// 局部布局更新 - 为新节点找到合适位置
export const applyLocalLayout = (
  nodes: Node<CustomNodeModel>[],
  edges: Edge[],
  newNodeId: string,
  config: Partial<LayoutConfig> = {},
  nodeSizesMap?: Map<string, { width: number; height: number }>
): { nodes: Node<CustomNodeModel>[]; edges: Edge[] } => {
  const layoutConfig = { ...DEFAULT_LAYOUT_CONFIG, ...config };
  console.log('layoutConfig', layoutConfig);
  
  // 找到新节点
  const newNode = nodes.find(node => node.id === newNodeId);
  if (!newNode) {
    return { nodes, edges };
  }

  // 找到新节点的父节点
  const parentEdge = edges.find(edge => edge.target === newNodeId);
  if (!parentEdge) {
    // 如果没有父节点，使用全局布局
    return applyGlobalLayout(nodes, edges, config, nodeSizesMap);
  }

  const parentNode = nodes.find(node => node.id === parentEdge.source);
  if (!parentNode) {
    return { nodes, edges };
  }

  // 找到父节点的所有子节点
  const siblingEdges = edges.filter(edge => edge.source === parentNode.id);
  const siblingNodes = siblingEdges.map(edge => 
    nodes.find(node => node.id === edge.target)
  ).filter(Boolean) as Node<CustomNodeModel>[];

  // 重新排列所有兄弟节点的位置，实现均匀分布
  const parentDimensions = getNodeDimensions(parentNode, nodeSizesMap);
  
  // 计算所有兄弟节点的新位置
  const updatedNodes = nodes.map(node => {
    if (!siblingNodes.find(sibling => sibling.id === node.id)) {
      return node; // 不是兄弟节点，保持原位置
    }
    
    return node;
  });
  
  // 重新计算兄弟节点的排列位置
  if (layoutConfig.direction === 'TB' || layoutConfig.direction === 'BT') {
    // 垂直布局：兄弟节点水平排列
    const yOffset = layoutConfig.direction === 'TB' ? 
      parentNode.position.y + parentDimensions.height + layoutConfig.rankSep :
      parentNode.position.y - layoutConfig.rankSep;
    
    // 计算所有兄弟节点的总宽度
    const totalWidth = siblingNodes.reduce((sum, node) => {
      const dims = getNodeDimensions(node, nodeSizesMap);
      return sum + dims.width + layoutConfig.nodeSep;
    }, -layoutConfig.nodeSep);
    
    // 计算起始X位置，使兄弟节点居中对齐
    const startX = parentNode.position.x + parentDimensions.width / 2 - totalWidth / 2;
    let currentX = startX;
    
    // 为每个兄弟节点分配新位置
    siblingNodes.forEach((siblingNode) => {
      const siblingDims = getNodeDimensions(siblingNode, nodeSizesMap);
      const nodeIndex = updatedNodes.findIndex(n => n.id === siblingNode.id);
      
      if (nodeIndex !== -1) {
        const newY = layoutConfig.direction === 'TB' ? yOffset : yOffset - siblingDims.height;
        updatedNodes[nodeIndex] = {
          ...updatedNodes[nodeIndex],
          position: { x: currentX, y: newY }
        };
      }
      
      currentX += siblingDims.width + layoutConfig.nodeSep;
    });
  } else {
    // 水平布局：兄弟节点垂直排列
    const xOffset = layoutConfig.direction === 'LR' ? 
      parentNode.position.x + parentDimensions.width + layoutConfig.rankSep :
      parentNode.position.x - layoutConfig.rankSep;
    
    // 计算所有兄弟节点的总高度
    const totalHeight = siblingNodes.reduce((sum, node) => {
      const dims = getNodeDimensions(node, nodeSizesMap);
      return sum + dims.height + layoutConfig.nodeSep;
    }, -layoutConfig.nodeSep);
    
    // 计算起始Y位置，使兄弟节点居中对齐
    const startY = parentNode.position.y + parentDimensions.height / 2 - totalHeight / 2;
    let currentY = startY;
    
    // 为每个兄弟节点分配新位置
    siblingNodes.forEach((siblingNode) => {
      const siblingDims = getNodeDimensions(siblingNode, nodeSizesMap);
      const nodeIndex = updatedNodes.findIndex(n => n.id === siblingNode.id);
      
      if (nodeIndex !== -1) {
        const newX = layoutConfig.direction === 'LR' ? xOffset : xOffset - siblingDims.width;
        updatedNodes[nodeIndex] = {
          ...updatedNodes[nodeIndex],
          position: { x: newX, y: currentY }
        };
      }
      
      currentY += siblingDims.height + layoutConfig.nodeSep;
    });
  }

  return {
    nodes: updatedNodes,
    edges,
  };
};

// 检查节点是否重叠
export const hasNodeOverlap = (
  node1: Node<CustomNodeModel>,
  node2: Node<CustomNodeModel>,
  margin: number = 10
): boolean => {
  const dims1 = getNodeDimensions(node1);
  const dims2 = getNodeDimensions(node2);
  
  const rect1 = {
    left: node1.position.x - margin,
    right: node1.position.x + dims1.width + margin,
    top: node1.position.y - margin,
    bottom: node1.position.y + dims1.height + margin,
  };
  
  const rect2 = {
    left: node2.position.x - margin,
    right: node2.position.x + dims2.width + margin,
    top: node2.position.y - margin,
    bottom: node2.position.y + dims2.height + margin,
  };
  
  return !(
    rect1.right < rect2.left ||
    rect2.right < rect1.left ||
    rect1.bottom < rect2.top ||
    rect2.bottom < rect1.top
  );
};

// 布局类型
export type LayoutType = 'global' | 'local';

// 应用布局的主函数
export const applyLayout = (
  nodes: Node<CustomNodeModel>[],
  edges: Edge[],
  layoutType: LayoutType,
  newNodeId?: string,
  config?: Partial<LayoutConfig>,
  nodeSizesMap?: Map<string, { width: number; height: number }>
): { nodes: Node<CustomNodeModel>[]; edges: Edge[] } => {
  if (layoutType === 'local' && newNodeId) {
    return applyLocalLayout(nodes, edges, newNodeId, config, nodeSizesMap);
  }
  return applyGlobalLayout(nodes, edges, config, nodeSizesMap);
};