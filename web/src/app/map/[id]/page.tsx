/*
 * @Date: 2025-07-01 23:49:12
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-02 00:30:25
 * @FilePath: /thinking-map/web/src/app/map/[id]/page.tsx
 */
import WorkspaceLayout from '@/layouts/workspace-layout';

export default function MapWorkspacePage({ params }: { params: { id: string } }) {
  return (
    <WorkspaceLayout>
      <div className="p-8">工作区内容区（TODO），mapId: {params.id}</div>
    </WorkspaceLayout>
  );
} 