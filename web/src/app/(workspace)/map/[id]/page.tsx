"use client";
import React from "react";
import { WorkspaceLayout } from '@/features/workspace';
import { useParams } from "next/navigation";

export default function TestWorkspacePage() {
  // 从链接中获取mapId, 链接格式如下：http://localhost:3000/map/404c3dd8-83c7-4683-b0ca-8f4ebb5badf5
  // 404c3dd8-83c7-4683-b0ca-8f4ebb5badf5为mapId
  const { id: mapId } = useParams<{ id: string }>();
  console.log(mapId)
  
  return (
    <WorkspaceLayout mapId={mapId} />
  );
}