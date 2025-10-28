/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/visualization-area/visualization-area.tsx
 */
'use client';

import React, { useCallback, useEffect, useRef, useState, useMemo } from 'react';
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
  EdgeTypes,
  getBezierPath,
  EdgeProps,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Loader, Layout, RotateCcw } from 'lucide-react';

import { CustomNode } from '@/features/workspace/components/custom-node/custom-node';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { useWorkspaceData } from '@/features/workspace/hooks/use-workspace-data';
import { useNodeSelection } from '@/features/workspace/hooks/use-node-selection';
import { useNodeOperations } from '@/features/workspace/hooks/use-node-operations';
import { useAutoLayout } from '@/features/workspace/hooks/use-auto-layout';
import { useExecutableNodes } from '@/features/workspace/hooks/use-executable-nodes';
import { useSSEConnection } from '@/hooks/use-sse-connection';
import { SSEStatusIndicator } from '@/components/sse-status-indicator';
import { Button } from '@/components/ui/button';
import { CustomNodeModel } from '@/types/node';
import { NodeCreatedEvent, NodeUpdatedEvent, NodeDeletedEvent, NodeDependenciesUpdatedEvent } from '@/types/sse';
import { LayoutType } from '@/utils/layout-utils';
import { DependencyEdge } from '@/features/workspace/components/custom-edge/custom-edge';

interface VisualizationAreaProps {
  mapID: string;
}


// 定义节点类型
const nodeTypes: NodeTypes = {
  custom: CustomNode,
};

// 定义边类型
const edgeTypes: EdgeTypes = {
  dependency: DependencyEdge,
};

function MapCanvas({ mapID }: VisualizationAreaProps) {
  const selectedNodeIDs = useWorkspaceStore(state => state.selectedNodeIDs);
  const settings = useWorkspaceStore(state => state.settings);
  const actions = useWorkspaceStore(state => state.actions);
  const { nodes: nodesData, edges: edgesData, isLoading } = useWorkspaceData(mapID);
  const { handleNodeClick, handleNodeDoubleClick, handleNodeContextMenu } = useNodeSelection();
  const { handleNodeEdit, handleNodeDelete, handleAddChild, handleNodeUpdateID } = useNodeOperations();
  const { isLayouting, applyAutoLayout, layoutConfig, updateLayoutConfig, finishAnimation } = useAutoLayout();
  const { fetchExecutableNodes } = useExecutableNodes();

  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [isDragging, setIsDragging] = useState(false);
  const layoutType = useWorkspaceStore(state => state.settings.layoutType);
  const [pendingLayout, setPendingLayout] = useState<{ nodeId?: string; type: LayoutType } | null>(null);
  const triggerLayoutTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const nodeSizesRef = useRef<Map<string, { width: number; height: number }>>(new Map());
  // 节点尺寸变化处理函数
  const handleNodeSizeChange = useCallback((id: string, size: { width: number; height: number }) => {
    nodeSizesRef.current.set(id, size);
    // console.log(`节点 ${id} 尺寸更新:`, size);
  }, []);

  // 使用稳定的处理函数引用
  const stableHandlers = useRef({
    handleNodeEdit,
    handleNodeDelete,
    handleAddChild,
    handleNodeClick,
    handleNodeDoubleClick,
    handleNodeContextMenu,
    handleNodeUpdateID,
    handleNodeSizeChange,
  });
  
  // 只在组件挂载时设置一次
  useEffect(() => {
    stableHandlers.current = {
      handleNodeEdit,
      handleNodeDelete,
      handleAddChild,
      handleNodeClick,
      handleNodeDoubleClick,
      handleNodeContextMenu,
      handleNodeUpdateID,
      handleNodeSizeChange,
    };
  }, [handleNodeSizeChange]);

  // 手动触发全局布局
  const handleGlobalLayout = useCallback(async () => {
    if (isLayouting || nodes.length === 0) {
      return;
    }
    
    try {
      const layoutResult = await applyAutoLayout(
        nodes as Node<CustomNodeModel>[],
        edges,
        layoutType,
        undefined,
        nodeSizesRef.current
      );
      
      if (layoutResult.animationState) {
        // 第一步：添加动画样式并设置到目标位置
        setNodes((nds) => 
          nds.map((node) => {
            const layoutedNode = layoutResult.nodes.find(n => n.id === node.id);
            if (layoutedNode) {
              return { 
                ...node, 
                position: layoutedNode.position,
                data: {
                  ...node.data,
                  isAnimating: true,
                  animationDuration: 500,
                  animationEasing: 'cubic-bezier(0.4, 0, 0.2, 1)'
                }
              };
            }
            return node;
          })
        );
        
        // 第二步：动画完成后移除动画样式并同步到store
        setTimeout(() => {
          setNodes((nds) => 
            nds.map((node) => {
              const layoutedNode = layoutResult.nodes.find(n => n.id === node.id);
              if (layoutedNode) {
                // 同步位置到store
                actions.updateNode(node.id, { position: layoutedNode.position });
                // 记录位置变更的节点，用于后续提交给后端
                actions.addChangedNodePosition(node.id);
                return { 
                  ...node, 
                  data: {
                    ...node.data,
                    isAnimating: false,
                    animationDuration: undefined,
                    animationEasing: undefined
                  }
                };
              }
              return node;
            })
          );
          finishAnimation();
        }, 500);
      } else {
        // 无动画状态，直接更新位置
        setNodes((nds) => 
          nds.map((node) => {
            const layoutedNode = layoutResult.nodes.find(n => n.id === node.id);
            if (layoutedNode) {
              actions.updateNode(node.id, { position: layoutedNode.position });
              // 记录位置变更的节点，用于后续提交给后端
              actions.addChangedNodePosition(node.id);
              return { ...node, position: layoutedNode.position };
            }
            return node;
          })
        );
      }
    } catch (error) {
      console.error('全局布局失败:', error);
    }
  }, [isLayouting, nodes, edges, applyAutoLayout, setNodes, actions, settings.layoutConfig.direction, layoutType, updateLayoutConfig, finishAnimation]);

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
    
    // 标记需要进行布局更新
    setPendingLayout({
      nodeId: event.nodeID,
      type: layoutType
    });
  }, [actions, layoutType]);

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

  const handleNodeDeleted = useCallback((event: NodeDeletedEvent) => {
    console.log('Received node deleted event:', event);
    
    // 从store中删除节点，这会自动删除相关的边
    actions.deleteNode(event.nodeID);
    
    // 如果删除的是当前选中的节点，清除选择状态
    const { activeNodeID } = useWorkspaceStore.getState();
    if (activeNodeID === event.nodeID) {
      actions.closePanel();
    }
  }, [actions]);

  const handleNodeDependenciesUpdated = useCallback((event: NodeDependenciesUpdatedEvent) => {
    console.log('Received node dependencies updated event:', event);
    
    // 更新节点的依赖关系，这会自动更新依赖边
    actions.updateNodeDependencies(event.nodeID, event.dependencies);
  }, [actions]);

  const handleConnectionEstablished = useCallback((data: any) => {
    console.log('SSE connection established:', data);
    // 可以在这里显示连接状态或执行其他初始化操作
  }, []);

  // 建立SSE连接
  const { isConnected } = useSSEConnection({
    mapID,
    callbacks: [
      {
        eventType: 'nodeCreated',
        callback: (event) => {
          try {
            const data = JSON.parse(event.data) as NodeCreatedEvent;
            handleNodeCreated(data);
            console.log(data)
          } catch (error) {
            console.error('解析nodeCreated事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'nodeUpdated',
        callback: (event) => {
          try {
            const data = JSON.parse(event.data) as NodeUpdatedEvent;
            handleNodeUpdated(data);
            console.log(data)
          } catch (error) {
            console.error('解析nodeUpdated事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'nodeDeleted',
        callback: (event) => {
          try {
            const data = JSON.parse(event.data) as NodeDeletedEvent;
            handleNodeDeleted(data);
            console.log(data)
          } catch (error) {
            console.error('解析nodeDeleted事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'nodeDependenciesUpdated',
        callback: (event) => {
          try {
            const data = JSON.parse(event.data) as NodeDependenciesUpdatedEvent;
            handleNodeDependenciesUpdated(data);
            console.log(data)
          } catch (error) {
            console.error('解析nodeDependenciesUpdated事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'connectionEstablished',
        callback: (event) => {
          try {
            const data = JSON.parse(event.data);
            handleConnectionEstablished(data);
          } catch (error) {
            console.error('解析connectionEstablished事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'error',
        callback: (event) => {
          try {
            const data = JSON.parse(event.data);
            console.error('SSE业务错误:', data);
          } catch (error) {
            console.error('解析error事件失败:', error, event.data);
          }
        }
      },
      {
        eventType: 'ping',
        callback: (event) => {
          // 心跳事件，通常不需要处理
          console.log('收到ping事件');
        }
      }
    ],
    onOpen: () => {
      console.log('SSE连接已建立');
    },
    onError: (error) => {
      console.error('SSE连接错误:', error);
    }
  });

  // 节点拖拽开始
  const onNodeDragStart = useCallback((event: any, draggedNode: any) => {
    setIsDragging(true);
    // 拖拽时禁用节点动画，避免移动延迟
    setNodes((nds) => 
      nds.map((node) => {
        if (node.id === draggedNode.id) {
          return {
            ...node,
            data: {
              ...node.data,
              isAnimating: false,
              animationDuration: undefined,
              animationEasing: undefined
            }
          };
        }
        return node;
      })
    );
  }, [setNodes]);

  // 节点拖动结束时同步到 store
  const onNodeDragStop = useCallback((event: any, draggedNode: any) => {
    setIsDragging(false);
    actions.updateNode(draggedNode.id, { position: draggedNode.position });
    actions.addChangedNodePosition(draggedNode.id);
    
    // 拖拽结束后恢复节点的正常状态（不强制开启动画）
    setNodes((nds) => 
      nds.map((node) => {
        if (node.id === draggedNode.id) {
          return {
            ...node,
            data: {
              ...node.data,
              isAnimating: false,
              animationDuration: undefined,
              animationEasing: undefined
            }
          };
        }
        return node;
      })
    );
  }, [actions, setNodes]);

  // 从state中同步
  useEffect(() => {
    if (nodesData.length > 0) {
      const handlers = stableHandlers.current;
      setNodes((currentNodes) => {
        return nodesData.map((node) => {
          // 查找当前节点以保留动画状态
          const currentNode = currentNodes.find(n => n.id === node.id);
          const currentAnimationState = currentNode?.data ? {
            isAnimating: currentNode.data.isAnimating,
            animationDuration: currentNode.data.animationDuration,
            animationEasing: currentNode.data.animationEasing
          } : {};
          
          return {
            ...node,
            data: {
              ...node.data,
              ...currentAnimationState, // 保留动画状态
              onEdit: handlers.handleNodeEdit,
              onDelete: handlers.handleNodeDelete,
              onAddChild: handlers.handleAddChild,
              onSelect: handlers.handleNodeClick,
              onDoubleClick: handlers.handleNodeDoubleClick,
              onContextMenu: handlers.handleNodeContextMenu,
              onUpdateID: handlers.handleNodeUpdateID,
              onSizeChange: handlers.handleNodeSizeChange,
              selected: selectedNodeIDs.includes(node.id),
            } as any,
          };
        });
      });
    }
  }, [nodesData, selectedNodeIDs]);
  
  // 单独同步边数据
  useEffect(() => {
    setEdges(edgesData);
  }, [edgesData]);
  
  // 监听节点数据变化，触发自动布局
  useEffect(() => {
    if (nodesData.length > 0 && pendingLayout && !isDragging && !isLayouting) {
      
      if (triggerLayoutTimeoutRef.current) {
        clearTimeout(triggerLayoutTimeoutRef.current);
      }
      
      const { nodeId, type } = pendingLayout;
      setPendingLayout(null);
      
      triggerLayoutTimeoutRef.current = setTimeout(async () => {
        try {
          const layoutResult = await applyAutoLayout(
            nodesData as Node<CustomNodeModel>[],
            edgesData,
            type,
            nodeId,
            nodeSizesRef.current
          );
          
          if (layoutResult.animationState) {
            // 第一步：添加动画样式并设置到目标位置
            setNodes((nds) => 
              nds.map((node) => {
                const layoutedNode = layoutResult.nodes.find(n => n.id === node.id);
                if (layoutedNode) {
                  return { 
                    ...node, 
                    position: layoutedNode.position,
                    data: {
                      ...node.data,
                      isAnimating: true,
                      animationDuration: 500,
                      animationEasing: 'cubic-bezier(0.4, 0, 0.2, 1)'
                    }
                  };
                }
                return node;
              })
            );
            
            // 第二步：动画完成后移除动画样式并同步到store
            setTimeout(() => {
              setNodes((nds) => 
                nds.map((node) => {
                  const layoutedNode = layoutResult.nodes.find(n => n.id === node.id);
                  if (layoutedNode) {
                    // 同步位置到store
                    actions.updateNode(node.id, { position: layoutedNode.position });
                    // 记录位置变更的节点，用于后续提交给后端
                    actions.addChangedNodePosition(node.id);
                    return { 
                      ...node, 
                      data: {
                        ...node.data,
                        isAnimating: false,
                        animationDuration: undefined,
                        animationEasing: undefined
                      }
                    };
                  }
                  return node;
                })
              );
              finishAnimation();
            }, 500);
          } else {
            // 无动画状态，直接更新位置
            setNodes((nds) => 
              nds.map((node) => {
                const layoutedNode = layoutResult.nodes.find(n => n.id === node.id);
                if (layoutedNode) {
                  actions.updateNode(node.id, { position: layoutedNode.position });
                  // 记录位置变更的节点，用于后续提交给后端
                  actions.addChangedNodePosition(node.id);
                  return { ...node, position: layoutedNode.position };
                }
                return node;
              })
            );
          }
        } catch (error) {
          console.error('自动布局失败:', error);
        }
      }, 100);
    }
  }, [nodesData, isDragging, isLayouting, pendingLayout, edgesData, applyAutoLayout, setNodes, actions]); // 添加必要的依赖
  
  // 清理定时器
  useEffect(() => {
    return () => {
      if (triggerLayoutTimeoutRef.current) {
        clearTimeout(triggerLayoutTimeoutRef.current);
      }
    };
  }, []);



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
    [setEdges]
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
      
      {/* 布局控制按钮 */}
      <div className="absolute top-4 left-4 z-10 flex gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={handleGlobalLayout}
          disabled={isLayouting || nodes.length === 0}
          className="bg-white/90 backdrop-blur-sm"
        >
          <Layout className="h-4 w-4 mr-1" />
          {isLayouting ? '布局中...' : '全局整理'}
        </Button>
        

      </div>
      
      <ReactFlow
        nodes={nodes.map(node => ({
          ...node,
          className: node.data.isAnimating ? 'animating' : ''
        }))}
        edges={edges.map(edge => ({
          ...edge,
          className: isDragging ? '' : 'animating'
        }))}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onPaneClick={onPaneClick}
        onNodeDragStart={onNodeDragStart}
        onNodeDragStop={onNodeDragStop}
        nodeTypes={nodeTypes}
        edgeTypes={edgeTypes}
        fitView
        attributionPosition="bottom-left"
        className="bg-gray-50"
        style={{
          '--rf-edge-transition': 'all 500ms cubic-bezier(0.4, 0, 0.2, 1)',
          '--rf-node-transition': 'all 500ms cubic-bezier(0.4, 0, 0.2, 1)'
        } as React.CSSProperties}
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