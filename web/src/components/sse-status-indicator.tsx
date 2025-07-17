'use client';

import React from 'react';
import { Badge } from '@/components/ui/badge';
import { Wifi, WifiOff } from 'lucide-react';

interface SSEStatusIndicatorProps {
  isConnected: boolean;
  className?: string;
}

export function SSEStatusIndicator({ isConnected, className = '' }: SSEStatusIndicatorProps) {
  return (
    <Badge 
      variant={isConnected ? 'default' : 'destructive'} 
      className={`flex items-center gap-1 ${isConnected ? 'bg-green-500! hover:bg-green-600! text-white!' : ''} ${className}`}
    >
      {isConnected ? (
        <>
          <Wifi className="h-3 w-3" />
          <span>已连接</span>
        </>
      ) : (
        <>
          <WifiOff className="h-3 w-3" />
          <span>连接断开</span>
        </>
      )}
    </Badge>
  );
}

export default SSEStatusIndicator;