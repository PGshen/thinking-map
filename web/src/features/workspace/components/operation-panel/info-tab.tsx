/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/info-tab.tsx
 */
'use client';

import React, { useState, useEffect } from 'react';
import { Save, RotateCcw, Trash2, Play, AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { useToast } from '@/hooks/use-toast';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';

interface InfoTabProps {
  nodeId: string;
  node: any; // TODO: 使用正确的节点类型
}

export function InfoTab({ nodeId, node }: InfoTabProps) {
  const nodeData = node.data as any;
  const [formData, setFormData] = useState({
    question: nodeData?.question || '',
    target: nodeData?.target || '',
    context: nodeData?.context || '',
  });
  const [hasChanges, setHasChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  
  const { toast } = useToast();
  const { actions } = useWorkspaceStore();

  // 监听节点数据变化
  useEffect(() => {
    const nodeData = node.data as any;
    setFormData({
      question: nodeData?.question || '',
      target: nodeData?.target || '',
      context: nodeData?.context || '',
    });
    setHasChanges(false);
  }, [node]);

  // 检查是否有未保存的更改
  useEffect(() => {
    const nodeData = node.data as any;
    const changed = 
      formData.question !== (nodeData?.question || '') ||
      formData.target !== (nodeData?.target || '') ||
      formData.context !== (nodeData?.context || '');
    setHasChanges(changed);
  }, [formData, node]);

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const handleSave = async () => {
    if (!hasChanges) return;
    
    setIsSaving(true);
    try {
      // TODO: 调用API保存节点信息
      // await updateNodeInfo(nodeId, formData);
      
      // 更新本地状态
      actions.updateNode(nodeId, { data: { ...node.data, ...formData } });
      
      toast({
        title: '保存成功',
        description: '节点信息已更新',
      });
      
      setHasChanges(false);
    } catch (error) {
      toast({
        title: '保存失败',
        description: '更新节点信息时出错，请重试',
        variant: 'destructive',
      });
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    const nodeData = node.data as any;
    setFormData({
      question: nodeData?.question || '',
      target: nodeData?.target || '',
      context: nodeData?.context || '',
    });
    setHasChanges(false);
  };

  const handleDelete = async () => {
    try {
      // TODO: 调用API删除节点
      // await deleteNode(nodeId);
      
      // 更新本地状态
      actions.deleteNode(nodeId);
      actions.closePanel();
      
      toast({
        title: '删除成功',
        description: '节点已删除',
      });
    } catch (error) {
      toast({
        title: '删除失败',
        description: '删除节点时出错，请重试',
        variant: 'destructive',
      });
    }
  };

  const handleStartExecution = async () => {
    try {
      // TODO: 调用API开始执行节点
      // const result = await startNodeExecution(nodeId);
      
      toast({
        title: '开始执行',
        description: '节点执行已开始',
      });
      
      // 根据后端返回结果决定跳转到哪个Tab
      // if (result.needsDecomposition) {
      //   // 跳转到拆解Tab
      // } else {
      //   // 跳转到结论Tab
      // }
    } catch (error) {
      toast({
        title: '执行失败',
        description: '开始执行时出错，请重试',
        variant: 'destructive',
      });
    }
  };

  // 检查依赖是否满足
  const nodeData = node.data as any;
  const unmetDependencies = nodeData?.dependencies?.filter(
    (dep: any) => dep.status !== 'completed'
  ) || [];
  const canExecute = unmetDependencies.length === 0 && (nodeData?.status || 'pending') === 'pending';

  return (
    <div className="h-full flex flex-col space-y-6">
      {/* 节点基础信息 */}
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="question">当前问题</Label>
          <Textarea
            id="question"
            value={formData.question}
            onChange={(e) => handleInputChange('question', e.target.value)}
            placeholder="描述当前需要解决的问题..."
            className="min-h-[80px] resize-none"
          />
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="target">目标描述</Label>
          <Textarea
            id="target"
            value={formData.target}
            onChange={(e) => handleInputChange('target', e.target.value)}
            placeholder="描述期望达到的目标..."
            className="min-h-[80px] resize-none"
          />
        </div>
        
        <div className="space-y-2">
          <Label htmlFor="context">上下文背景</Label>
          <Textarea
            id="context"
            value={formData.context}
            onChange={(e) => handleInputChange('context', e.target.value)}
            placeholder="提供相关的背景信息和上下文..."
            className="min-h-[100px] resize-none"
          />
        </div>
        
        {/* 结论内容（仅在已完成状态显示） */}
        {(nodeData?.status === 'completed') && nodeData?.conclusion && (
          <div className="space-y-2">
            <Label>结论内容</Label>
            <div className="p-3 bg-green-50 border border-green-200 rounded-md">
              <p className="text-sm text-green-800">{nodeData.conclusion}</p>
            </div>
          </div>
        )}
      </div>
      
      <Separator />
      
      {/* 依赖检查 */}
      {nodeData?.dependencies && nodeData.dependencies.length > 0 && (
        <div className="space-y-3">
          <Label>依赖状态</Label>
          <div className="space-y-2">
            {nodeData.dependencies.map((dep: any, index: number) => (
              <div key={index} className="flex items-center justify-between p-2 bg-gray-50 rounded-md">
                <span className="text-sm font-medium">{dep.name}</span>
                <Badge 
                  variant={dep.status === 'completed' ? 'default' : 'secondary'}
                  className={dep.status === 'completed' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'}
                >
                  {dep.status === 'completed' ? '已完成' : '未完成'}
                </Badge>
              </div>
            ))}
          </div>
          
          {unmetDependencies.length > 0 && (
            <div className="flex items-center gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
              <AlertCircle className="w-4 h-4 text-yellow-600" />
              <p className="text-sm text-yellow-800">
                有 {unmetDependencies.length} 个依赖未满足，无法开始执行
              </p>
            </div>
          )}
        </div>
      )}
      
      <Separator />
      
      {/* 操作按钮 */}
      <div className="space-y-3">
        {/* 执行控制 */}
        {(nodeData?.status || 'pending') === 'pending' && (
          <Button
            onClick={handleStartExecution}
            disabled={!canExecute}
            className="w-full"
            size="lg"
          >
            <Play className="w-4 h-4 mr-2" />
            开始执行
          </Button>
        )}
        
        {/* 编辑操作 */}
        <div className="flex gap-2">
          <Button
            onClick={handleSave}
            disabled={!hasChanges || isSaving}
            className="flex-1"
            variant="default"
          >
            <Save className="w-4 h-4 mr-2" />
            {isSaving ? '保存中...' : '保存修改'}
          </Button>
          
          <Button
            onClick={handleReset}
            disabled={!hasChanges}
            variant="outline"
            className="flex-1"
          >
            <RotateCcw className="w-4 h-4 mr-2" />
            重置
          </Button>
        </div>
        
        {/* 删除操作 */}
        <AlertDialog>
          <AlertDialogTrigger asChild>
            <Button variant="destructive" className="w-full">
              <Trash2 className="w-4 h-4 mr-2" />
              删除节点
            </Button>
          </AlertDialogTrigger>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>确认删除</AlertDialogTitle>
              <AlertDialogDescription>
                确定要删除这个节点吗？此操作无法撤销，相关的子节点也会被删除。
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>取消</AlertDialogCancel>
              <AlertDialogAction onClick={handleDelete} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
                确定删除
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </div>
  );
}

export default InfoTab;