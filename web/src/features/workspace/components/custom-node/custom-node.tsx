/*
 * @Author: peng pgs1108pgs@gmail.com
 * @Date: 2025-07-07 22:05:27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-07 23:52:59
 * @FilePath: /thinking-map/web/src/features/map/components/CustomNode.tsx
 */
import React, { useState, useRef, useEffect, useMemo } from 'react';
import { HelpCircle, Search, Brain, Lightbulb, Scale, ChevronDown, ChevronUp } from 'lucide-react';
import type { CustomNodeModel } from '@/types/node';
import { NodeStatusIcon } from './node-status-icon';
import { NodeActionButtons } from './node-action-buttons';
import { Handle, Position, useEdges } from 'reactflow';
import { MarkdownContent } from '@/components/ui/markdown-content';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { useWorkspaceStoreData, useWorkspaceStore } from '../../store/workspace-store';

interface CustomNodeProps {
  data: CustomNodeModel & {
    isAnimating?: boolean;
    animationDuration?: number;
    animationEasing?: string;
    isSuggested?: boolean;
  }
}

export const CustomNode: React.FC<CustomNodeProps> = ({ data }) => {
  const { mapID } = useWorkspaceStoreData()
  const settings = useWorkspaceStore(state => state.settings)
  const edges = useEdges(); // 获取当前所有边的信息
  const nodeRef = useRef<HTMLDivElement>(null);
  const conclusionRef = useRef<HTMLDivElement>(null);
  const [editForm, setEditForm] = useState({
    question: data.question,
    target: data.target,
    nodeType: data.nodeType
  });

  // 结论折叠状态管理
  const [isConclusionCollapsed, setIsConclusionCollapsed] = useState(true);
  const [shouldShowCollapseButton, setShouldShowCollapseButton] = useState(false);

  // 检查当前节点是否有连接
  const nodeConnections = useMemo(() => {
    const nodeId = data.id;
    
    // 检查父子关系连接
    const hasParentTargetConnection = edges.some(edge => 
      edge.target === nodeId && edge.type !== 'dependency'
    );
    const hasParentSourceConnection = edges.some(edge => 
      edge.source === nodeId && edge.type !== 'dependency'
    );
    
    // 检查依赖关系连接
    const hasDependencyTargetConnection = edges.some(edge => 
      edge.target === nodeId && edge.type === 'dependency'
    );
    const hasDependencySourceConnection = edges.some(edge => 
      edge.source === nodeId && edge.type === 'dependency'
    );
    
    return {
      // 父子关系连接
      hasParentTargetConnection,
      hasParentSourceConnection,
      hasParentConnection: hasParentTargetConnection || hasParentSourceConnection,
      
      // 依赖关系连接
      hasDependencyTargetConnection,
      hasDependencySourceConnection,
      hasDependencyConnection: hasDependencyTargetConnection || hasDependencySourceConnection
    };
  }, [edges, data.id]);

  // 检查结论内容是否需要折叠按钮
  useEffect(() => {
    if (conclusionRef.current && data.conclusion?.content) {
      const element = conclusionRef.current;
      // 检查内容是否超过5行的高度（大约100px）
      const maxHeight = 100; // 约5行的高度
      setShouldShowCollapseButton(element.scrollHeight > maxHeight);
    }
  }, [data.conclusion?.content]);

  // 测量节点实际尺寸并通知布局系统
  useEffect(() => {
    if (nodeRef.current && data.onSizeChange) {
      const { clientWidth, clientHeight } = nodeRef.current;
      data.onSizeChange(data.id, { width: clientWidth, height: clientHeight });
    }
  }, [data.question, data.target, data.conclusion, data.isEditing, data.onSizeChange, data.id, isConclusionCollapsed]);

  const handleEdit = () => {
    data.onEdit?.(mapID, data.id, { isEditing: true });
  };

  const handleSave = async () => {
    if (data.status === 'initial') {
      // 新增节点时，提交到后端并更新节点ID
      const response = await data.onEdit?.(mapID, data.id, { ...editForm, status: 'pending', parentID: data.parentID, isEditing: false });
      if (response?.id) {
        // 更新节点ID
        data.onUpdateID?.(mapID, data.id, response.id);
      }
      return;
    }
    // 编辑已有节点
    data.onEdit?.(mapID, data.id, { ...editForm, isEditing: false });
  };

  const handleCancel = () => {
    if (data.status === 'initial') {
      // 如果是新增的节点，取消时删除该节点
      data.onDelete?.(mapID, data.id);
      return;
    }
    // 如果是编辑已有节点，恢复原始数据
    setEditForm({
      question: data.question,
      target: data.target,
      nodeType: data.nodeType
    });
    data.onEdit?.(mapID, data.id, { isEditing: false });
  };

  const handleAddChild = () => {
    data.onAddChild?.(data.id)
  }

  const handleDelete = () => {
    data.onDelete?.(mapID, data.id);
  }

  const toggleConclusionCollapse = () => {
    setIsConclusionCollapsed(!isConclusionCollapsed);
  };

  // 动画样式
  const animationStyle = data.isAnimating ? {
    transition: `all ${data.animationDuration || 500}ms ${data.animationEasing || 'cubic-bezier(0.4, 0, 0.2, 1)'}`
  } : {};

  return (
    <div
      ref={nodeRef}
      className={`rounded-lg shadow-lg bg-white border ${data.status === 'initial' ? 'border-dashed' : ''} px-3 py-2.5 min-w-[280px] max-w-[360px] select-none ${data.isAnimating ? '' : 'transition-all duration-200'} hover:shadow-xl ${data.selected ? 'ring-1 ring-blue-400' : ''} ${data.isSuggested ? 'ring-2 ring-yellow-400 shadow-yellow-100' : ''}`}
      style={animationStyle}
      onClick={() => data.onSelect?.(data.id)}
      onDoubleClick={() => data.onDoubleClick?.(data.id)}
      onContextMenu={e => { e.preventDefault(); data.onContextMenu?.(data.id, e); }}
    >
      {/* ReactFlow Handles for edge connection */}
      {/* 父子关系连接点 - 分别控制target和source的显示 */}
      {nodeConnections.hasParentTargetConnection && (
        <Handle 
          type="target" 
          position={settings.layoutConfig.direction === 'TB' ? Position.Top : Position.Left} 
          className={`!rounded-none !border-blue-500 !bg-blue-500 ${
            settings.layoutConfig.direction === 'TB' ? '!w-2 !h-1' : '!w-1 !h-2'
          }`}
        />
      )}
      {nodeConnections.hasParentSourceConnection && (
        <Handle 
          type="source" 
          position={settings.layoutConfig.direction === 'TB' ? Position.Bottom : Position.Right} 
          className={`!rounded-none !border-blue-500 !bg-blue-500 ${
            settings.layoutConfig.direction === 'TB' ? '!w-2 !h-1' : '!w-1 !h-2'
          }`}
        />
      )}
      
      {/* 依赖关系连接点 - 分别控制target和source的显示 */}
      {nodeConnections.hasDependencyTargetConnection && (
        <Handle 
          id="dependency-target"
          type="target" 
          position={settings.layoutConfig.direction === 'TB' ? Position.Left : Position.Top} 
          className={`!rounded-none !border-purple-500 !bg-purple-500 !w-1 !h-1`}
          style={{
            left: settings.layoutConfig.direction === 'TB' ? '0px' : '50%',
            top: settings.layoutConfig.direction === 'TB' ? '50%' : '0px'
          }}
        />
      )}
      {nodeConnections.hasDependencySourceConnection && (
        <Handle 
          id="dependency-source"
          type="source" 
          position={settings.layoutConfig.direction === 'TB' ? Position.Right : Position.Bottom} 
          className={`!rounded-none !border-purple-500 !bg-purple-500 !w-1 !h-1`}
          style={{
            right: settings.layoutConfig.direction === 'TB' ? '0px' : 'auto',
            left: settings.layoutConfig.direction === 'TB' ? 'auto' : '50%',
            bottom: settings.layoutConfig.direction === 'TB' ? 'auto' : '0px',
            top: settings.layoutConfig.direction === 'TB' ? '50%' : 'auto'
          }}
        />
      )}
      
      {/* Header: 类型+状态 */}
      <div className="flex items-center justify-between mb-2 pb-2">
        <div className="flex items-center gap-2">
          {nodeTypeIcons[data.nodeType] || <div className="p-1.5 rounded-lg bg-gray-400"><HelpCircle className="w-4 h-4 text-white" /></div>}
          <span className="text-sm font-medium text-gray-700">{data.nodeType}</span>
        </div>
        <NodeStatusIcon status={data.status} />
      </div>

      {/* Content: 问题/目标/结论 */}
      <div className="space-y-1 mb-2">
        {data.isEditing ? (
          <div className="space-y-2" onDoubleClick={e => e.stopPropagation()} onClick={(e) => e.stopPropagation()}>
            <div className="space-y-1.5">
              <label htmlFor="nodeType" className="text-sm font-medium text-gray-700">节点类型</label>
              <select
                id="nodeType"
                value={editForm.nodeType}
                onChange={(e) => setEditForm(prev => ({ ...prev, nodeType: e.target.value }))}
                className="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="problem">问题</option>
                <option value="information">信息</option>
                <option value="analysis">分析</option>
                <option value="generation">生成</option>
                <option value="evaluation">评估</option>
              </select>
            </div>
            <div className="space-y-1.5">
              <label htmlFor="question" className="text-sm font-medium text-gray-700">问题</label>
              <Textarea
                id="question"
                value={editForm.question}
                onChange={(e) => setEditForm(prev => ({ ...prev, question: e.target.value }))}
                placeholder="输入问题"
                className="text-sm min-h-[80px]"
              />
            </div>
            <div className="space-y-1.5">
              <label htmlFor="target" className="text-sm font-medium text-gray-700">目标</label>
              <Textarea
                id="target"
                value={editForm.target}
                onChange={(e) => setEditForm(prev => ({ ...prev, target: e.target.value }))}
                placeholder="输入目标"
                className="text-xs min-h-[100px]"
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button variant="outline" size="sm" onClick={handleCancel}>取消</Button>
              <Button size="sm" onClick={handleSave}>保存</Button>
            </div>
          </div>
        ) : (
          <>
            <div className="text-sm font-medium text-gray-900 line-clamp-2 overflow-hidden text-ellipsis">{data.question}</div>
            <div className="text-xs text-gray-600 overflow-hidden">
              <MarkdownContent 
                id={'target' + data.id}
                content={data.target} 
                className="max-w-full overflow-hidden text-ellipsis line-clamp-5"
              />
            </div>
            {data.status === 'completed' && data.conclusion && (
              <div className="relative">
                <div 
                  ref={conclusionRef}
                  className={`text-xs text-green-700 overflow-hidden leading-tight transition-all duration-300 ${
                    isConclusionCollapsed ? 'max-h-[150px]' : 'max-h-none'
                  }`}
                  style={{
                    maskImage: isConclusionCollapsed && shouldShowCollapseButton 
                      ? 'linear-gradient(to bottom, black 70%, transparent 100%)' 
                      : 'none'
                  }}
                >
                  <MarkdownContent 
                    id={'conclusion' + data.id}
                    content={data.conclusion.content} 
                    className={`max-w-full overflow-hidden ${
                      isConclusionCollapsed ? 'text-ellipsis' : ''
                    }`}
                  />
                </div>
                {shouldShowCollapseButton && (
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      toggleConclusionCollapse();
                    }}
                    className="flex items-center gap-1 mt-1 text-xs text-green-600 hover:text-green-800 transition-colors cursor-pointer"
                  >
                    {isConclusionCollapsed ? (
                      <>
                        <span>展开</span>
                        <ChevronDown className="w-3 h-3" />
                      </>
                    ) : (
                      <>
                        <span>收起</span>
                        <ChevronUp className="w-3 h-3" />
                      </>
                    )}
                  </button>
                )}
              </div>
            )}
          </>
        )}
      </div>

      {/* 操作按钮组（悬浮在节点上方） */}
      <div className="absolute right-0 -top-12 flex justify-end">
        <div 
          className={`
            transform transition-all duration-200 ease-out
            ${data.selected ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-2 pointer-events-none'}
          `}
        >
          <NodeActionButtons
            id={data.id}
            mapID={mapID}
            onEdit={handleEdit}
            onDelete={handleDelete}
            onAddChild={handleAddChild}
          />
        </div>
      </div>
    </div>
  );
};


const nodeTypeIcons: Record<string, React.ReactNode> = {
  problem: <div className="p-1.5 rounded-lg bg-purple-500"><HelpCircle className="w-4 h-4 text-white" /></div>,
  information: <div className="p-1.5 rounded-lg bg-blue-500"><Search className="w-4 h-4 text-white" /></div>,
  analysis: <div className="p-1.5 rounded-lg bg-cyan-500"><Brain className="w-4 h-4 text-white" /></div>,
  generation: <div className="p-1.5 rounded-lg bg-yellow-500"><Lightbulb className="w-4 h-4 text-white" /></div>,
  evaluation: <div className="p-1.5 rounded-lg bg-green-500"><Scale className="w-4 h-4 text-white" /></div>
};