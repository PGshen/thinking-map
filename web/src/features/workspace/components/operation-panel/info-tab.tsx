/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/info-tab.tsx
 */
'use client';

import React, { useState, useEffect } from 'react';
import { Save, RotateCcw, Trash2, Play } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Separator } from '@/components/ui/separator';
import { DependencyContext } from './dependency-context';
import { getUnmetDependenciesMessage, checkAllDependenciesMet } from '@/utils/dependency-utils';
import { useToast } from '@/hooks/use-toast';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { CustomNodeModel, DependentContext, NodeContextItem } from '@/types/node';
import { resetNodeContext, updateNode, updateNodeContext } from '@/api/node';

interface InfoTabProps {
  nodeID: string;
  nodeData: CustomNodeModel;
}

export function InfoTab({ nodeID, nodeData }: InfoTabProps) {
  const { mapID } = useWorkspaceStore();
  const defaultContext: DependentContext = {
    ancestor: [],
    prevSibling: [],
    children: []
  };

  const [formData, setFormData] = useState({
    question: nodeData?.question || '',
    target: nodeData?.target || '',
    context: nodeData?.context || defaultContext,
  });
  const [hasChanges, setHasChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  
  const { toast } = useToast();
  const { actions } = useWorkspaceStore();

  // 监听节点数据变化
  useEffect(() => {
    setFormData({
      question: nodeData?.question || '',
      target: nodeData?.target || '',
      context: nodeData?.context || {
        ancestor: [],
        prevSibling: [],
        children: []
      } as DependentContext,
    });
    setHasChanges(false);
  }, [nodeData]);



  // 检查是否有未保存的更改
  useEffect(() => {
    const changed = 
      formData.question !== (nodeData?.question || '') ||
      formData.target !== (nodeData?.target || '') ||
      JSON.stringify(formData.context) !== JSON.stringify(nodeData?.context || {
        ancestor: [],
        prevSibling: [],
        children: []
      });
    setHasChanges(changed);
  }, [formData, nodeData]);

  const handleInputChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const handleSave = async () => {
    if (!hasChanges) return;
    if (!mapID) return
    
    setIsSaving(true);
    try {
      // 调用API保存节点上下文 
      let res = await updateNodeContext(mapID, nodeID, formData);
      if (res.code !== 200) {
        throw new Error(res.message);
      }  
      res = await updateNode(mapID, nodeID, formData);
      if (res.code !== 200) {
        throw new Error(res.message);
      }
      actions.updateNode(nodeID, { data: { ...nodeData, ...formData } });    
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

  const handleReset = async () => {
    if (!mapID) return
    try {
      const res = await resetNodeContext(mapID, nodeID);
      if (res.code !== 200) {
        throw new Error(res.message);
      }
      // 重置上下文
      actions.updateNode(nodeID, { data: { ...nodeData, context: res.data.context } });
      
      toast({
        title: '重置成功',
        description: '节点上下文已重置',
      });
      setHasChanges(false);
    } catch (error) {
      toast({
        title: '重置失败',
        description: '重置节点上下文时出错，请重试',
        variant: 'destructive',
      });
    }
  };

  const handleStartExecution = async () => {
    try {
      // TODO: 调用API开始执行节点
      // const result = await startNodeExecution(nodeID);
      
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
  const canExecute = (() => {
    if (!nodeData?.context || nodeData?.status !== 'pending') return false;
    return checkAllDependenciesMet(nodeData.context);
  })();

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
        
        {/* 结论内容（仅在已完成状态显示） */}
        {(nodeData?.status === 'completed') && nodeData?.conclusion && (
          <div className="space-y-2">
            <Label>结论内容</Label>
            <div className="p-3 bg-green-50 border border-green-200 rounded-md">
              <p className="text-sm text-green-800">{nodeData.conclusion.content}</p>
            </div>
          </div>
        )}
      </div>
      
      <Separator />
      
      {/* 依赖上下文 */}
      <div className="space-y-4">
        <Label>依赖上下文</Label>
        <DependencyContext
          context={formData.context}
          onContextChange={(newContext) => {
            setFormData(prev => ({
              ...prev,
              context: newContext
            }));
          }}
        />
      </div>
      
      <Separator />
      
      {/* 操作按钮 */}
      <div className="space-y-3">
        {/* 执行控制 */}
        {(nodeData?.status || 'pending') === 'pending' && (
          <Button
            onClick={handleStartExecution}
            disabled={!canExecute}
            className="w-full cursor-pointer"
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
            className="flex-1 cursor-pointer"
            variant="default"
          >
            <Save className="w-4 h-4 mr-2" />
            {isSaving ? '保存中...' : '保存修改'}
          </Button>
          
          <Button
            onClick={handleReset}
            variant="outline"
            className="flex-1 cursor-pointer"
          >
            <RotateCcw className="w-4 h-4 mr-2" />
            重置
          </Button>
        </div>
      </div>
    </div>
  );
}

export default InfoTab;