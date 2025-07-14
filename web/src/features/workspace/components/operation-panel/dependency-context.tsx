'use client';

import React from 'react';
import { DependencySection } from './dependency-section';
import { AlertCircle } from 'lucide-react';
import { DependentContext, NodeContextItem } from '@/types/node';
import { getUnmetDependenciesMessage } from '@/utils/dependency-utils';

interface DependencyContextProps {
  context: DependentContext;
  onContextChange?: (newContext: DependentContext) => void;
}

export function DependencyContext({ context, onContextChange }: DependencyContextProps) {


  return (
    <div className="space-y-4">
      {/* 祖先节点 */}
      {context.ancestor?.length > 0 && (
        <DependencySection
          title="祖先节点"
          items={context.ancestor}
          type="ancestor"
          onItemsChange={(newItems: NodeContextItem[]) => {
            if (onContextChange) {
              onContextChange({
                ...context,
                ancestor: newItems
              });
            }
          }}
        />
      )}

      {/* 同级节点 */}
      {context.prevSibling?.length > 0 && (
        <DependencySection
          title="同级依赖节点"
          items={context.prevSibling}
          type="prevSibling"
          onItemsChange={(newItems: NodeContextItem[]) => {
            if (onContextChange) {
              onContextChange({
                ...context,
                prevSibling: newItems
              });
            }
          }}
        />
      )}

      {/* 子节点 */}
      {context.children?.length > 0 && (
        <DependencySection
          title="子节点"
          items={context.children}
          type="children"
          onItemsChange={(newItems: NodeContextItem[]) => {
            if (onContextChange) {
              onContextChange({
                ...context,
                children: newItems
              });
            }
          }}
        />
      )}

      {/* 未满足依赖提示 */}
      <div className="flex items-center gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
        <AlertCircle className="w-4 h-4 text-yellow-600" />
        <p className="text-sm text-yellow-800">
          {getUnmetDependenciesMessage(context)}
        </p>
      </div>
    </div>
  );
}