/*
 * @Date: 2025-07-07 22:05:27
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-07 23:51:23
 * @FilePath: /thinking-map/web/src/features/map/components/NodeTooltip.tsx
 */
import React from 'react';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

interface NodeTooltipProps {
  content: string;
  children: React.ReactNode;
}

export const NodeTooltip: React.FC<NodeTooltipProps> = ({ content, children }) => {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          {children}
        </TooltipTrigger>
        {content && (
          <TooltipContent 
            className="
              whitespace-pre-line text-xs
              min-w-[160px] max-w-[320px] max-h-[240px] overflow-y-auto
              leading-relaxed
            "
          >
            {content}
          </TooltipContent>
        )}
      </Tooltip>
    </TooltipProvider>
  );
};