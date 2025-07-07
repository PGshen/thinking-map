import React from 'react';
import type { CustomNodeModel } from './CustomNodeModel';

interface NodeStatusIconProps {
  status: CustomNodeModel['status'];
}

export const NodeStatusIcon: React.FC<NodeStatusIconProps> = ({ status }) => {
  switch (status) {
    case 'pending':
      return <span title="待执行" className="text-gray-400">⏸</span>;
    case 'running':
      return <span title="执行中" className="text-blue-400 animate-spin">🔄</span>;
    case 'completed':
      return <span title="已完成" className="text-green-500">✅</span>;
    case 'error':
      return <span title="错误" className="text-red-500">⚠️</span>;
    default:
      return null;
  }
}; 