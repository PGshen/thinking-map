/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/top-bar/task-title.tsx
 */
'use client';

import React, { useState, useRef, useEffect } from 'react';
import { Edit2, Check, X, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useToast } from '@/hooks/use-toast';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';

interface TaskTitleProps {
  title: string;
  taskId: string;
  isLoading?: boolean;
}

export function TaskTitle({ title, taskId, isLoading }: TaskTitleProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editValue, setEditValue] = useState(title);
  const [isSaving, setIsSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const { toast } = useToast();
  const { actions } = useWorkspaceStore();

  useEffect(() => {
    setEditValue(title);
  }, [title]);

  useEffect(() => {
    if (isEditing && inputRef.current) {
      inputRef.current.focus();
      inputRef.current.select();
    }
  }, [isEditing]);

  const handleStartEdit = () => {
    if (!isLoading) {
      setIsEditing(true);
    }
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditValue(title);
  };

  const handleSaveEdit = async () => {
    if (editValue.trim() === '' || editValue === title) {
      handleCancelEdit();
      return;
    }

    setIsSaving(true);
    try {
      // TODO: 调用API更新任务标题
      // await updateTaskTitle(taskId, editValue.trim());
      
      // 更新本地状态
      actions.updateTaskTitle(editValue.trim());
      setIsEditing(false);
      
      toast({
        title: '保存成功',
        description: '任务标题已更新',
      });
    } catch (error) {
      toast({
        title: '保存失败',
        description: '更新任务标题时出错，请重试',
        variant: 'destructive',
      });
    } finally {
      setIsSaving(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSaveEdit();
    } else if (e.key === 'Escape') {
      handleCancelEdit();
    }
  };

  const truncatedTitle = title.length > 50 ? `${title.slice(0, 50)}...` : title;

  if (isEditing) {
    return (
      <div className="flex items-center gap-2 max-w-md">
        <Input
          ref={inputRef}
          value={editValue}
          onChange={(e) => setEditValue(e.target.value)}
          onKeyDown={handleKeyDown}
          className="h-8 text-lg font-semibold"
          placeholder="输入任务标题"
          disabled={isSaving}
        />
        <Button
          size="sm"
          variant="ghost"
          onClick={handleSaveEdit}
          disabled={isSaving}
          className="h-8 w-8 p-0"
        >
          {isSaving ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : (
            <Check className="w-4 h-4 text-green-600" />
          )}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          onClick={handleCancelEdit}
          disabled={isSaving}
          className="h-8 w-8 p-0"
        >
          <X className="w-4 h-4 text-red-600" />
        </Button>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-2 group">
      <h1 
        className="text-sm text-foreground cursor-pointer hover:text-primary transition-colors"
        onClick={handleStartEdit}
        title={title}
      >
        {isLoading ? (
          <div className="flex items-center gap-2">
            <Loader2 className="w-4 h-4 animate-spin" />
            <span>加载中...</span>
          </div>
        ) : (
          truncatedTitle
        )}
      </h1>
      {!isLoading && (
        <Button
          size="sm"
          variant="ghost"
          onClick={handleStartEdit}
          className="h-6 w-6 p-0 opacity-0 group-hover:opacity-100 transition-opacity"
          aria-label="编辑任务标题"
        >
          <Edit2 className="w-3 h-3" />
        </Button>
      )}
    </div>
  );
}

export default TaskTitle;