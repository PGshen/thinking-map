/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/top-bar/map-info.tsx
 */
'use client';

import * as React from 'react';
import { Button } from '@/components/ui/button';
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { useWorkspaceStore } from '../../store/workspace-store';
import { useWorkspaceData } from '../../hooks/use-workspace-data';

interface MapInfoProps {
  taskId: string;
}

export function MapInfo({ taskId }: MapInfoProps) {
  const { mapDetail } = useWorkspaceStore();
  const { saveMapDetail } = useWorkspaceData(taskId);
  
  const [isEditing, setIsEditing] = React.useState(false);
  const [formData, setFormData] = React.useState({
    title: mapDetail?.title || '',
    problem: mapDetail?.problem || '',
    target: mapDetail?.target || '',
    keyPoints: mapDetail?.keyPoints || '',
    constraints: mapDetail?.constraints || ''
  });

  React.useEffect(() => {
    if (mapDetail) {
      setFormData({
        title: mapDetail.title || '',
        problem: mapDetail.problem || '',
        target: mapDetail.target || '',
        keyPoints: mapDetail.keyPoints || '',
        constraints: mapDetail.constraints || ''
      });
    }
  }, [mapDetail]);

  const handleSave = async () => {
    try {
      await saveMapDetail(formData);
      setIsEditing(false);
    } catch (error) {
      console.error('Failed to update map info:', error);
    }
  };

  const handleInputChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData(prev => ({
      ...prev,
      [field]: e.target.value
    }));
  };

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button variant="ghost" className="h-8 w-8 p-0">
          <span className="sr-only">查看导图信息</span>
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            className="h-4 w-4"
          >
            <path d="M21 12a9 9 0 1 1-9-9c2.52 0 4.93 1 6.74 2.74L21 8" />
            <path d="M21 3v5h-5" />
          </svg>
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-80">
        <div className="grid gap-4">
          <div className="space-y-2">
            <h4 className="font-medium leading-none">导图信息</h4>
            <p className="text-sm text-muted-foreground">
              查看和修改当前导图的基本信息
            </p>
          </div>
          <div className="grid gap-2">
            {isEditing ? (
              <>
                <div className="grid gap-1">
                  <label htmlFor="title" className="text-sm font-medium leading-none">标题</label>
                  <Input
                    id="title"
                    value={formData.title}
                    className="h-8"
                    onChange={handleInputChange('title')}
                  />
                </div>
                <div className="grid gap-1">
                  <label htmlFor="problem" className="text-sm font-medium leading-none">问题</label>
                  <Textarea
                    id="problem"
                    value={formData.problem}
                    className="min-h-[60px]"
                    onChange={handleInputChange('problem')}
                  />
                </div>
                <div className="grid gap-1">
                  <label htmlFor="target" className="text-sm font-medium leading-none">目标</label>
                  <Textarea
                    id="target"
                    value={formData.target}
                    className="min-h-[60px]"
                    onChange={handleInputChange('target')}
                  />
                </div>
                <div className="grid gap-1">
                  <label htmlFor="keyPoints" className="text-sm font-medium leading-none">关键点</label>
                  <Textarea
                    id="keyPoints"
                    value={formData.keyPoints}
                    className="min-h-[60px]"
                    onChange={handleInputChange('keyPoints')}
                  />
                </div>
                <div className="grid gap-1">
                  <label htmlFor="constraints" className="text-sm font-medium leading-none">约束</label>
                  <Textarea
                    id="constraints"
                    value={formData.constraints}
                    className="min-h-[60px]"
                    onChange={handleInputChange('constraints')}
                  />
                </div>
                <div className="flex justify-end gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setIsEditing(false)}
                  >
                    取消
                  </Button>
                  <Button size="sm" onClick={handleSave}>
                    保存
                  </Button>
                </div>
              </>
            ) : (
              <>
                <div className="grid gap-1">
                  <span className="text-sm font-medium leading-none">标题</span>
                  <p className="text-sm">{formData.title}</p>
                </div>
                <div className="grid gap-1">
                  <span className="text-sm font-medium leading-none">问题</span>
                  <p className="text-sm">{formData.problem}</p>
                </div>
                <div className="grid gap-1">
                  <span className="text-sm font-medium leading-none">目标</span>
                  <p className="text-sm">{formData.target}</p>
                </div>
                <div className="grid gap-1">
                  <span className="text-sm font-medium leading-none">关键点</span>
                  <p className="text-sm">{formData.keyPoints}</p>
                </div>
                <div className="grid gap-1">
                  <span className="text-sm font-medium leading-none">约束</span>
                  <p className="text-sm">{formData.constraints}</p>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  className="justify-self-end"
                  onClick={() => setIsEditing(true)}
                >
                  编辑
                </Button>
              </>
            )}
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}