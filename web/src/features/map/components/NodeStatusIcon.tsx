import React from 'react';
import { Clock, Loader2, CheckCircle2, AlertCircle } from 'lucide-react';

interface NodeStatusIconProps {
  status: 'pending' | 'running' | 'completed' | 'error'
}

export const NodeStatusIcon: React.FC<NodeStatusIconProps> = ({ status }) => {
  switch (status) {
    case 'pending':
      return <Clock className="w-4 h-4 text-gray-400" />;
    case 'running':
      return <Loader2 className="w-4 h-4 text-blue-400 animate-spin" />;
    case 'completed':
      return <CheckCircle2 className="w-4 h-4 text-green-500" />;
    case 'error':
      return <AlertCircle className="w-4 h-4 text-red-500" />;
    default:
      return null;
  }
};