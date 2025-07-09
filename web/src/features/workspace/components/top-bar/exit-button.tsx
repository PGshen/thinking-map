/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/top-bar/exit-button.tsx
 */
'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';

export function ExitButton() {
  const router = useRouter();
  const [showConfirm, setShowConfirm] = useState(false);
  const { hasUnsavedChanges } = useWorkspaceStore();

  const handleExit = () => {
    if (hasUnsavedChanges) {
      setShowConfirm(true);
    } else {
      router.push('/');
    }
  };

  const confirmExit = () => {
    setShowConfirm(false);
    router.push('/');
  };

  return (
    <>
      <Button
        variant="ghost"
        size="sm"
        onClick={handleExit}
        className="flex items-center gap-2 px-3 py-2 h-9 rounded-md hover:bg-accent transition-colors"
        aria-label="退出工作区"
      >
        <ArrowLeft className="w-4 h-4" />
        <span className="text-sm font-medium">退出</span>
      </Button>

      <AlertDialog open={showConfirm} onOpenChange={setShowConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认退出</AlertDialogTitle>
            <AlertDialogDescription>
              您有未保存的更改，确定要退出工作区吗？未保存的更改将会丢失。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={confirmExit} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              确定退出
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}

export default ExitButton;