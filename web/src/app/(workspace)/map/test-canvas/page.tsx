import React from 'react';
import { CustomNodeTestCanvas } from '@/features/map/components/custom-node-test-canvas';

export default function MapTestCanvasPage() {
  return (
    <main>
      <h1 className="text-xl font-bold mb-4">自定义节点可视化测试</h1>
      <CustomNodeTestCanvas />
    </main>
  );
} 