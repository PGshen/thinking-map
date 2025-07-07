/*
 * @Date: 2025-07-07 22:05:27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-07 23:52:59
 * @FilePath: /thinking-map/web/src/features/map/components/CustomNode.tsx
 */
import React from 'react';
import { HelpCircle, Search, Brain, Lightbulb, Scale } from 'lucide-react';
import type { CustomNodeModel } from '@/types/node';
import { NodeStatusIcon } from './NodeStatusIcon';
import { NodeActionButtons } from './NodeActionButtons';
import { NodeTooltip } from './NodeTooltip';
import { Handle, Position } from 'reactflow';
import { MarkdownContent } from '@/components/ui/markdown-content';
import { Badge } from '@/components/ui/badge';

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
      className={`rounded-lg shadow-lg bg-white border px-3 py-2.5 min-w-[280px] max-w-[360px] select-none transition-all duration-200 hover:shadow-xl ${data.selected ? 'ring-1 ring-blue-400' : ''}`}
      onClick={() => data.onSelect?.(data.id)}
      onDoubleClick={() => data.onDoubleClick?.(data.id)}
      onContextMenu={e => { e.preventDefault(); data.onContextMenu?.(data.id, e); }}
    >
      {/* ReactFlow Handles for edge connection */}
      <Handle type="target" position={Position.Top} />
      <Handle type="source" position={Position.Bottom} />
      
      {/* Header: 类型+状态 */}
      <div className="flex items-center justify-between mb-2 pb-2">
        <div className="flex items-center gap-2">
          {nodeTypeIcons[data.nodeType] || <div className="p-1.5 rounded-lg bg-gray-400"><HelpCircle className="w-4 h-4 text-white" /></div>}
          <span className="text-sm font-medium text-gray-700">{data.nodeType}</span>
        </div>
        <NodeStatusIcon status={data.status} />
      </div>

      {/* Content: 问题/目标/结论 */}
      <div className="space-y-2 mb-2">
        <div className="text-sm font-medium text-gray-900 line-clamp-2 overflow-hidden text-ellipsis">{data.question}</div>
        <div className="text-xs text-gray-600 overflow-hidden">
          <MarkdownContent 
            id={'conclusion' + data.id}
            content={data.target} 
            className="max-w-full overflow-hidden text-ellipsis line-clamp-5"
          />
        </div>
        {data.status === 'completed' && data.conclusion && (
          <div className="text-xs text-green-700 overflow-hidden">
            <MarkdownContent 
              id={'conclusion' + data.id}
              content={data.conclusion} 
              className="max-w-full overflow-hidden text-ellipsis line-clamp-5"
            />
          </div>
        )}
      </div>

      {/* Footer: 依赖/分支 */}
      <div className="flex items-center justify-between pt-1 border-t border-gray-100">
        <div className="flex items-center gap-1 flex-wrap">
          {data.dependencies && data.dependencies.length > 0 && (
            data.dependencies.map((dep, index) => {
              const { variant, className: statusClassName } = getDependencyStatusStyle(dep.status);
              return (
                <Badge
                  key={index}
                  variant={variant}
                  className={`text-xs px-1.5 py-0.5 h-5 ${statusClassName}`}
                >
                  {dep.name}
                </Badge>
              );
            })
          )}
        </div>
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
            onEdit={data.onEdit}
            onDelete={data.onDelete}
            onAddChild={data.onAddChild}
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

// 根据依赖状态返回对应的badge样式
const getDependencyStatusStyle = (status: string): { variant: 'default' | 'secondary' | 'destructive' | 'outline', className: string } => {
  switch (status.toLowerCase()) {
    case 'completed':
    case 'success':
    case 'resolved':
      return {
        variant: 'outline',
        className: 'bg-green-400 hover:bg-green-500 text-white'
      }; // 绿色，表示已完成
    case 'pending':
    case 'waiting':
    case 'blocked':
      return {
        variant: 'outline',
        className: 'bg-yellow-400 hover:bg-yellow-500 text-white'
      }; // 黄色，表示等待中
    case 'error':
    case 'failed':
    case 'unmet':
      return {
        variant: 'outline',
        className: 'bg-red-400 hover:bg-red-500 text-white'
      }; // 红色，表示错误或未满足
    case 'running':
    case 'processing':
    case 'active':
      return {
        variant: 'outline',
        className: 'bg-blue-400 hover:bg-blue-500 text-white'
      }; // 蓝色，表示进行中
    default:
      return {
        variant: 'outline',
        className: 'bg-gray-100 text-gray-800 border-gray-300 hover:bg-gray-200'
      }; // 默认为灰色
  }
};