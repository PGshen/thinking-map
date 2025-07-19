"use client";
import React from "react";
import { WorkspaceLayout } from '@/features/workspace';
import { useParams } from "next/navigation";
import { AuthGuard } from "@/components/auth-guard";

export default function TestWorkspacePage() {
  // 从链接中获取mapID, 链接格式如下：http://localhost:3000/map/404c3dd8-83c7-4683-b0ca-8f4ebb5badf5
  // 404c3dd8-83c7-4683-b0ca-8f4ebb5badf5为mapID
  const { id: mapID } = useParams<{ id: string }>();
  console.log(mapID)

  return (
    <AuthGuard>
      <WorkspaceLayout mapID={mapID} />
    </AuthGuard>
  );
}