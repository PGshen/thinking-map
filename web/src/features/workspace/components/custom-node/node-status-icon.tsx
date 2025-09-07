import React from 'react';
import { Clock, Loader, CheckCircle2, AlertCircle, SquircleDashed } from 'lucide-react';

interface NodeStatusIconProps {
  status: 'initial' | 'pending' | 'in_decomposition' | 'in_conclusion' | 'completed' | 'error'
}

export const NodeStatusIcon: React.FC<NodeStatusIconProps> = ({ status }) => {
  switch (status) {
    case 'initial':
      return <SquircleDashed className="w-4 h-4 text-gray-400" />;
    case 'pending':
      return <Clock className="w-4 h-4 text-gray-400" />;
    case 'in_decomposition':
    case 'in_conclusion':
      return <Loader className="w-4 h-4 text-blue-400" />;
    case 'completed':
      return <CheckCircle2 className="w-4 h-4 text-green-500" />;
    case 'error':
      return <AlertCircle className="w-4 h-4 text-red-500" />;
    default:
      return null;
  }
};