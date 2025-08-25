/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/hooks/use-auto-layout.ts
 */

import { useCallback, useRef, useState } from 'react';
import { Node, Edge } from 'reactflow';
import { CustomNodeModel } from '@/types/node';
import {
  applyLayout,
  LayoutType,
  LayoutConfig,
} from '@/utils/layout-utils';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';

// 动画配置接口（从store导入）
interface AnimationConfig {
  duration: number; // 动画持续时间（毫秒）
  easing: string; // 缓动函数
}

// Hook返回类型
interface UseAutoLayoutReturn {
  isLayouting: boolean;
  layoutConfig: LayoutConfig;
  animationConfig: AnimationConfig;
  applyAutoLayout: (
    nodes: Node<CustomNodeModel>[],
    edges: Edge[],
    layoutType: LayoutType,
    newNodeId?: string,
    nodeSizesMap?: Map<string, { width: number; height: number }>
  ) => Promise<{ 
    nodes: Node<CustomNodeModel>[]; 
    edges: Edge[];
    animationState?: AnimationState;
  }>;
  finishAnimation: () => void;
  updateLayoutConfig: (config: Partial<LayoutConfig>) => void;
  updateAnimationConfig: (config: Partial<AnimationConfig>) => void;
}

// 动画状态类型
interface AnimationState {
  isAnimating: boolean;
  fromPositions: Map<string, { x: number; y: number }>;
  toPositions: Map<string, { x: number; y: number }>;
}

export const useAutoLayout = (): UseAutoLayoutReturn => {
  const [isLayouting, setIsLayouting] = useState(false);
  const animationRef = useRef<AbortController | null>(null);
  
  // 从store获取配置
  const { settings, actions } = useWorkspaceStore();
  const { layoutConfig, animationConfig } = settings;
  
  const updateLayoutConfig = useCallback((config: Partial<LayoutConfig>) => {
    actions.updateSettings({
      layoutConfig: { ...layoutConfig, ...config }
    });
  }, [layoutConfig, actions]);

  const updateAnimationConfig = useCallback((config: Partial<AnimationConfig>) => {
    actions.updateSettings({
      animationConfig: { ...animationConfig, ...config }
    });
  }, [animationConfig, actions]);

  const finishAnimation = useCallback(() => {
    setIsLayouting(false);
  }, []);

  const applyAutoLayout = useCallback(async (
    nodes: Node<CustomNodeModel>[],
    edges: Edge[],
    layoutType: LayoutType,
    newNodeId?: string,
    nodeSizesMap?: Map<string, { width: number; height: number }>
  ): Promise<{ 
    nodes: Node<CustomNodeModel>[]; 
    edges: Edge[];
    animationState?: AnimationState;
  }> => {
    // 如果正在进行布局，取消之前的动画
    if (animationRef.current) {
      animationRef.current.abort();
    }
    
    setIsLayouting(true);
    animationRef.current = new AbortController();
    
    try {
      // 计算新的布局
      const layoutResult = applyLayout(nodes, edges, layoutType, newNodeId, layoutConfig, nodeSizesMap);
      
      // 检查是否有位置变化
      const fromPositions = new Map<string, { x: number; y: number }>();
      const toPositions = new Map<string, { x: number; y: number }>();
      let hasPositionChanges = false;
      
      layoutResult.nodes.forEach((newNode) => {
        const oldNode = nodes.find(n => n.id === newNode.id);
        if (oldNode) {
          const positionChanged = 
            Math.abs(newNode.position.x - oldNode.position.x) > 1 ||
            Math.abs(newNode.position.y - oldNode.position.y) > 1;
          
          if (positionChanged) {
            fromPositions.set(newNode.id, oldNode.position);
            toPositions.set(newNode.id, newNode.position);
            hasPositionChanges = true;
          }
        }
      });
      
      // 如果没有位置变化，直接返回
      if (!hasPositionChanges) {
        setIsLayouting(false);
        return layoutResult;
      }
      
      // 返回布局结果和动画状态
      const animationState: AnimationState = {
        isAnimating: true,
        fromPositions,
        toPositions
      };
      
      return {
        ...layoutResult,
        animationState
      };
      
    } catch (error) {
      setIsLayouting(false);
      if (error instanceof Error && error.message === 'Animation aborted') {
        // 动画被取消，返回原始数据
        return { nodes, edges };
      }
      throw error;
    }
  }, [layoutConfig, animationConfig]);

  return {
    isLayouting,
    layoutConfig,
    animationConfig,
    applyAutoLayout,
    finishAnimation,
    updateLayoutConfig,
    updateAnimationConfig,
  };
};