"use client";
import React from "react";
import WorkspaceLayout from '@/layouts/workspace-layout';

export default function MapWorkspacePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = React.use(params);

  return (
    <WorkspaceLayout>
      <div className="p-8">工作区内容区（TODO），mapId: {id}</div>
    </WorkspaceLayout>
  );
} 