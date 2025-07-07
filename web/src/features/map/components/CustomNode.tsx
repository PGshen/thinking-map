import React from 'react';
import type { CustomNodeModel } from './CustomNodeModel';
import { NodeStatusIcon } from './NodeStatusIcon';
import { NodeActionButtons } from './NodeActionButtons';
import { NodeTooltip } from './NodeTooltip';
import { Handle, Position } from 'reactflow';

interface CustomNodeProps {
  data: CustomNodeModel
}

const statusColor: Record<CustomNodeModel['status'], string> = {
  pending: 'border-gray-300',
  running: 'border-blue-400 animate-pulse',
  completed: 'border-green-400',
  error: 'border-red-400',
};

export const CustomNode: React.FC<CustomNodeProps> = ({ data }) => {
  return (
    <div
      className={`rounded-xl shadow-md bg-white border-2 px-4 py-2 min-w-[220px] max-w-[320px] select-none transition-all duration-200 ${statusColor[data.status]} ${data.selected ? 'ring-2 ring-blue-300' : ''}`}
      onClick={() => data.onSelect?.(data.id)}
      onDoubleClick={() => data.onDoubleClick?.(data.id)}
      onContextMenu={e => { e.preventDefault(); data.onContextMenu?.(data.id, e); }}
    >
      {/* ReactFlow Handles for edge connection */}
      <Handle type="target" position={Position.Top} />
      <Handle type="source" position={Position.Bottom} />
      {/* Header: 类型+状态 */}
      <div className="flex items-center justify-between mb-1">
        <div className="flex items-center gap-1">
          {/* TODO: 类型图标，可根据 nodeType 渲染 */}
          <span className="text-lg">🎯</span>
          <span className="text-xs font-semibold text-gray-600">{data.nodeType}</span>
        </div>
        <NodeStatusIcon status={data.status} />
      </div>
      {/* Content: 问题/目标/结论 */}
      <div className="mb-1">
        <NodeTooltip content={data.question}>
          <div className="text-sm font-medium text-gray-900 truncate cursor-help">{data.question}</div>
        </NodeTooltip>
        <NodeTooltip content={data.target}>
          <div className="text-xs text-gray-500 truncate cursor-help">{data.target}</div>
        </NodeTooltip>
        {data.status === 'completed' && data.conclusion && (
          <NodeTooltip content={data.conclusion}>
            <div className="text-xs text-green-700 truncate cursor-help">{data.conclusion}</div>
          </NodeTooltip>
        )}
      </div>
      {/* Footer: 依赖/分支 */}
      <div className="flex items-center justify-between mt-1">
        <div className="flex items-center gap-1">
          {data.hasUnmetDependencies && <span className="text-red-400 text-xs" title="有未完成依赖">⛓️</span>}
        </div>
        <div className="text-xs text-gray-400">{data.childCount ? `${data.childCount}分支` : ''}</div>
      </div>
      {/* 操作按钮组（悬浮/选中时显示） */}
      {data.selected && (
        <NodeActionButtons
          id={data.id}
          onEdit={data.onEdit}
          onDelete={data.onDelete}
          onAddChild={data.onAddChild}
        />
      )}
    </div>
  );
}; 