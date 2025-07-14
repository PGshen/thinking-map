'use client';

import React, { useState } from 'react';
import { ChevronDown, ChevronUp, Edit2, Trash2, ArrowUp, ArrowDown } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from '@/components/ui/alert-dialog';
import { NodeContextItem } from '@/types/node';

interface DependencyItemProps {
  item: NodeContextItem;
  onUpdate: (updatedItem: NodeContextItem) => void;
  onDelete: () => void;
  onAddAbove: () => void;
  onAddBelow: () => void;
}

export function DependencyItem({ item, onUpdate, onDelete, onAddAbove, onAddBelow }: DependencyItemProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState({
    question: item.question,
    target: item.target,
    conclusion: item.conclusion || '',
    status: item.status
  });

  // 获取摘要内容
  const getSummary = () => {
    const content = item.conclusion || item.target;
    if (!content) return '';
    const lines = content.split('\n');
    if (lines.length <= 3) return content;
    return lines.slice(0, 3).join('\n') + '...';
  };

  const handleSave = () => {
    onUpdate({
      ...item,
      ...editForm
    });
    setIsEditing(false);
  };

  return (
    <div className="border rounded-lg overflow-hidden bg-white">
      {/* 标题栏 */}
      <div
        className="flex items-center justify-between p-3 cursor-pointer hover:bg-gray-50"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex-1">
          <h4 className="font-medium text-sm">{item.question}</h4>
          {!isExpanded && (
            <p className="text-xs text-gray-500 mt-1 line-clamp-3">{getSummary()}</p>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Badge 
            variant={item.status === 'completed' ? 'default' : 'secondary'}
            className={item.status === 'completed' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}
          >
            {item.status === 'completed' ? '已完成' : '未完成'}
          </Badge>
          {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
        </div>
      </div>

      {/* 展开内容 */}
      {isExpanded && (
        <div className="p-3 border-t bg-gray-50">
          <div className="space-y-2">
            <p className="text-sm text-gray-700">目标：{item.target}</p>
            {item.conclusion && (
              <p className="text-sm text-gray-700">结论：{item.conclusion}</p>
            )}
          </div>

          {/* 操作按钮 */}
          <div className="flex items-center gap-2 mt-4">
            <Popover open={isEditing} onOpenChange={setIsEditing}>
              <PopoverTrigger asChild>
                <Button variant="outline" size="sm" className="h-5 px-1 text-xs cursor-pointer">
                  <Edit2 className="w-2.5 h-2.5 text-blue-500" />
                  {/* 修改 */}
                </Button>
              </PopoverTrigger>
              <PopoverContent side="left" align="start" className="w-[400px]">
                <div className="space-y-4">
                  <div className="space-y-2">
                    <div className="space-y-2">
                      <Label>问题</Label>
                      <Textarea
                        value={editForm.question}
                        onChange={(e) => setEditForm({ ...editForm, question: e.target.value })}
                        placeholder="输入问题..."
                        className="min-h-[80px] resize-none"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>目标</Label>
                      <Textarea
                        value={editForm.target}
                        onChange={(e) => setEditForm({ ...editForm, target: e.target.value })}
                        placeholder="输入目标..."
                        className="min-h-[80px] resize-none"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>结论</Label>
                      <Textarea
                        value={editForm.conclusion}
                        onChange={(e) => setEditForm({ ...editForm, conclusion: e.target.value })}
                        placeholder="输入结论..."
                        className="min-h-[80px] resize-none"
                      />
                    </div>
                    <div className="space-y-2">
                      <Label>状态</Label>
                      <select
                        className="w-full h-9 rounded-md border border-input bg-background px-3 py-1 text-sm ring-offset-background"
                        value={editForm.status}
                        onChange={(e) => setEditForm({ ...editForm, status: e.target.value })}
                      >
                        <option value="pending">未完成</option>
                        <option value="completed">已完成</option>
                      </select>
                    </div>
                  </div>
                  <div className="flex justify-end gap-2">
                    <Button variant="outline" size="sm" className="h-7 px-2 text-xs" onClick={() => setIsEditing(false)}>
                      取消
                    </Button>
                    <Button size="sm" className="h-7 px-2 text-xs" onClick={handleSave}>
                      保存
                    </Button>
                  </div>
                </div>
              </PopoverContent>
            </Popover>

            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button variant="outline" size="sm" className="h-5 px-2 text-xs cursor-pointer">
                  <Trash2 className="w-3.5 h-3.5 text-red-500" />
                  {/* 删除 */}
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>确认删除</AlertDialogTitle>
                  <AlertDialogDescription>
                    确定要删除这个依赖项吗？此操作无法撤销。
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>取消</AlertDialogCancel>
                  <AlertDialogAction onClick={onDelete}>确认</AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>

            <Button
              variant="outline"
              size="sm"
              className="h-5 px-2 text-xs cursor-pointer"
              onClick={() => {
                onAddAbove();
                setIsEditing(true);
              }}
            >
              <ArrowUp className="w-3.5 h-3.5" />
              {/* 向上添加 */}
            </Button>

            <Button
              variant="outline"
              size="sm"
              className="h-5 px-2 text-xs cursor-pointer"
              onClick={() => {
                onAddBelow();
                setIsEditing(true);
              }}
            >
              <ArrowDown className="w-3.5 h-3.5" />
              {/* 向下添加 */}
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}