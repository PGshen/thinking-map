/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/top-bar/settings-button.tsx
 */
'use client';

import React from 'react';
import { Settings, Download, Layout, Eye, Users, HelpCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useToast } from '@/hooks/use-toast';

interface SettingsButtonProps {
  taskId: string;
}

export function SettingsButton({ taskId }: SettingsButtonProps) {
  const { toast } = useToast();

  const handleExport = () => {
    // TODO: 实现导出功能
    toast({
      title: '导出功能',
      description: '导出功能正在开发中...',
    });
  };

  const handleLayoutSettings = () => {
    // TODO: 实现布局设置
    toast({
      title: '布局设置',
      description: '布局设置功能正在开发中...',
    });
  };

  const handleDisplayOptions = () => {
    // TODO: 实现显示选项
    toast({
      title: '显示选项',
      description: '显示选项功能正在开发中...',
    });
  };

  const handleCollaboration = () => {
    // TODO: 实现协作设置
    toast({
      title: '协作设置',
      description: '协作设置功能正在开发中...',
    });
  };

  const handleHelp = () => {
    // TODO: 实现帮助文档
    window.open('/help', '_blank');
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          size="sm"
          className="h-9 w-9 p-0 rounded-md hover:bg-accent transition-colors"
          aria-label="设置"
        >
          <Settings className="w-4 h-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-56">
        <DropdownMenuLabel>工作区设置</DropdownMenuLabel>
        <DropdownMenuSeparator />
        
        <DropdownMenuItem onClick={handleExport} className="cursor-pointer">
          <Download className="w-4 h-4 mr-2" />
          导出思维导图
        </DropdownMenuItem>
        
        <DropdownMenuItem onClick={handleLayoutSettings} className="cursor-pointer">
          <Layout className="w-4 h-4 mr-2" />
          布局设置
        </DropdownMenuItem>
        
        <DropdownMenuItem onClick={handleDisplayOptions} className="cursor-pointer">
          <Eye className="w-4 h-4 mr-2" />
          显示选项
        </DropdownMenuItem>
        
        <DropdownMenuSeparator />
        
        <DropdownMenuItem onClick={handleCollaboration} className="cursor-pointer">
          <Users className="w-4 h-4 mr-2" />
          协作设置
        </DropdownMenuItem>
        
        <DropdownMenuSeparator />
        
        <DropdownMenuItem onClick={handleHelp} className="cursor-pointer">
          <HelpCircle className="w-4 h-4 mr-2" />
          帮助文档
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default SettingsButton;