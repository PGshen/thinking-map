import React from 'react';

interface NodeTooltipProps {
  content: string;
  children: React.ReactNode;
}

export const NodeTooltip: React.FC<NodeTooltipProps> = ({ content, children }) => {
  // 简单实现，后续可用 shadcn/ui Tooltip 替换
  return (
    <span className="relative group">
      {children}
      {content && (
        <span className="absolute z-10 left-1/2 -translate-x-1/2 mt-2 px-2 py-1 bg-black text-white text-xs rounded shadow-lg opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-pre-line min-w-[120px] max-w-[260px]">
          {content}
        </span>
      )}
    </span>
  );
}; 