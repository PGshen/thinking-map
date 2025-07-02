/*
 * @Date: 2025-07-01 23:48:33
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-02 00:29:20
 * @FilePath: /thinking-map/web/src/layouts/WorkspaceLayout.tsx
 */
import React from 'react';
import WorkspaceHeader from '../components/workspace-header';

export default function WorkspaceLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="w-full h-full min-h-screen bg-background flex flex-col">
      <WorkspaceHeader />
      <main className="flex-1 min-w-0">{children}</main>
    </div>
  );
} 