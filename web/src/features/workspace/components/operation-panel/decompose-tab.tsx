/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/decompose-tab.tsx
 */
'use client';

import React, { useState, useEffect } from 'react';
import { Plus, Trash2, Edit3, Save, X, GitBranch, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
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

interface DecomposeTabProps {
  nodeId: string;
  node: any; // TODO: 使用正确的节点类型
}

interface SubNode {
  id: string;
  question: string;
  target: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  isEditing?: boolean;
}

export function DecomposeTab({ nodeId, node }: DecomposeTabProps) {
  const [subNodes, setSubNodes] = useState<SubNode[]>([]);
  const [isDecomposing, setIsDecomposing] = useState(false);
  const [newSubNode, setNewSubNode] = useState({ question: '', target: '' });
  const [isAddingManual, setIsAddingManual] = useState(false);
  
  const { toast } = useToast();
  const { actions } = useWorkspaceStore();

  // 加载子节点数据
  useEffect(() => {
    // TODO: 从API或store获取子节点数据
    // const childNodes = getChildNodes(nodeId);
    // setSubNodes(childNodes);
  }, [nodeId]);

  // 自动拆解
  const handleAutoDecompose = async () => {
    setIsDecomposing(true);
    try {
      // TODO: 调用AI拆解API
      // const result = await autoDecomposeNode(nodeId, {
      //   question: node.question,
      //   target: node.target,
      //   context: node.context
      // });
      
      // 模拟AI拆解结果
      const mockSubNodes: SubNode[] = [
        {
          id: `${nodeId}-sub-1`,
          question: '分析当前问题的核心要素',
          target: '识别问题的关键组成部分',
          status: 'pending'
        },
        {
          id: `${nodeId}-sub-2`,
          question: '制定解决方案',
          target: '基于分析结果制定可行的解决方案',
          status: 'pending'
        },
        {
          id: `${nodeId}-sub-3`,
          question: '验证方案可行性',
          target: '确保方案能够有效解决问题',
          status: 'pending'
        }
      ];
      
      setSubNodes(mockSubNodes);
      
      // 更新节点状态
      actions.updateNode(nodeId, { data: { ...node.data, status: 'running' } });
      
      toast({
        title: '拆解完成',
        description: `已生成 ${mockSubNodes.length} 个子任务`,
      });
    } catch (error) {
      toast({
        title: '拆解失败',
        description: '自动拆解时出错，请重试',
        variant: 'destructive',
      });
    } finally {
      setIsDecomposing(false);
    }
  };

  // 手动添加子节点
  const handleAddManualSubNode = () => {
    if (!newSubNode.question.trim() || !newSubNode.target.trim()) {
      toast({
        title: '信息不完整',
        description: '请填写完整的问题和目标',
        variant: 'destructive',
      });
      return;
    }

    const subNode: SubNode = {
      id: `${nodeId}-manual-${Date.now()}`,
      question: newSubNode.question,
      target: newSubNode.target,
      status: 'pending'
    };

    setSubNodes(prev => [...prev, subNode]);
    setNewSubNode({ question: '', target: '' });
    setIsAddingManual(false);
    
    toast({
      title: '添加成功',
      description: '子任务已添加',
    });
  };

  // 编辑子节点
  const handleEditSubNode = (subNodeId: string, field: string, value: string) => {
    setSubNodes(prev => prev.map(subNode => 
      subNode.id === subNodeId 
        ? { ...subNode, [field]: value }
        : subNode
    ));
  };

  // 保存子节点编辑
  const handleSaveSubNode = async (subNodeId: string) => {
    try {
      // TODO: 调用API保存子节点
      // await updateSubNode(subNodeId, subNode);
      
      setSubNodes(prev => prev.map(subNode => 
        subNode.id === subNodeId 
          ? { ...subNode, isEditing: false }
          : subNode
      ));
      
      toast({
        title: '保存成功',
        description: '子任务已更新',
      });
    } catch (error) {
      toast({
        title: '保存失败',
        description: '更新子任务时出错，请重试',
        variant: 'destructive',
      });
    }
  };

  // 删除子节点
  const handleDeleteSubNode = (subNodeId: string) => {
    setSubNodes(prev => prev.filter(subNode => subNode.id !== subNodeId));
    toast({
      title: '删除成功',
      description: '子任务已删除',
    });
  };

  // 开始执行子节点
  const handleStartSubNode = async (subNodeId: string) => {
    try {
      // TODO: 调用API开始执行子节点
      // await startSubNodeExecution(subNodeId);
      
      setSubNodes(prev => prev.map(subNode => 
        subNode.id === subNodeId 
          ? { ...subNode, status: 'running' }
          : subNode
      ));
      
      toast({
        title: '开始执行',
        description: '子任务执行已开始',
      });
    } catch (error) {
      toast({
        title: '执行失败',
        description: '开始执行时出错，请重试',
        variant: 'destructive',
      });
    }
  };

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800';
      case 'running': return 'bg-blue-100 text-blue-800';
      case 'failed': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  // 获取状态文本
  const getStatusText = (status: string) => {
    switch (status) {
      case 'completed': return '已完成';
      case 'running': return '执行中';
      case 'failed': return '失败';
      default: return '待执行';
    }
  };

  return (
    <div className="h-full flex flex-col space-y-6">
      {/* 拆解控制 */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">任务拆解</h3>
          <Badge variant="outline">
            {subNodes.length} 个子任务
          </Badge>
        </div>
        
        {subNodes.length === 0 && (
          <div className="space-y-3">
            <Button
              onClick={handleAutoDecompose}
              disabled={isDecomposing}
              className="w-full"
              size="lg"
            >
              {isDecomposing ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <GitBranch className="w-4 h-4 mr-2" />
              )}
              {isDecomposing ? 'AI拆解中...' : 'AI智能拆解'}
            </Button>
            
            <div className="text-center text-sm text-muted-foreground">
              或
            </div>
            
            <Button
              onClick={() => setIsAddingManual(true)}
              variant="outline"
              className="w-full"
            >
              <Plus className="w-4 h-4 mr-2" />
              手动添加子任务
            </Button>
          </div>
        )}
      </div>
      
      {/* 手动添加子节点表单 */}
      {isAddingManual && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">添加子任务</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="new-question">问题描述</Label>
              <Textarea
                id="new-question"
                value={newSubNode.question}
                onChange={(e) => setNewSubNode(prev => ({ ...prev, question: e.target.value }))}
                placeholder="描述子任务需要解决的问题..."
                className="min-h-[60px] resize-none"
              />
            </div>
            
            <div className="space-y-2">
              <Label htmlFor="new-target">目标描述</Label>
              <Textarea
                id="new-target"
                value={newSubNode.target}
                onChange={(e) => setNewSubNode(prev => ({ ...prev, target: e.target.value }))}
                placeholder="描述子任务的期望目标..."
                className="min-h-[60px] resize-none"
              />
            </div>
            
            <div className="flex gap-2">
              <Button onClick={handleAddManualSubNode} className="flex-1">
                <Save className="w-4 h-4 mr-2" />
                添加
              </Button>
              <Button 
                onClick={() => {
                  setIsAddingManual(false);
                  setNewSubNode({ question: '', target: '' });
                }}
                variant="outline"
                className="flex-1"
              >
                <X className="w-4 h-4 mr-2" />
                取消
              </Button>
            </div>
          </CardContent>
        </Card>
      )}
      
      {/* 子节点列表 */}
      {subNodes.length > 0 && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h4 className="font-medium">子任务列表</h4>
            <Button
              onClick={() => setIsAddingManual(true)}
              variant="outline"
              size="sm"
            >
              <Plus className="w-4 h-4 mr-2" />
              添加
            </Button>
          </div>
          
          <div className="space-y-3 max-h-[400px] overflow-y-auto">
            {subNodes.map((subNode, index) => (
              <Card key={subNode.id} className="relative">
                <CardContent className="p-4">
                  <div className="space-y-3">
                    {/* 状态和操作 */}
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium">#{index + 1}</span>
                        <Badge className={getStatusColor(subNode.status)}>
                          {getStatusText(subNode.status)}
                        </Badge>
                      </div>
                      
                      <div className="flex items-center gap-1">
                        {!subNode.isEditing && (
                          <Button
                            onClick={() => setSubNodes(prev => prev.map(s => 
                              s.id === subNode.id ? { ...s, isEditing: true } : s
                            ))}
                            variant="ghost"
                            size="sm"
                          >
                            <Edit3 className="w-4 h-4" />
                          </Button>
                        )}
                        
                        <AlertDialog>
                          <AlertDialogTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <Trash2 className="w-4 h-4" />
                            </Button>
                          </AlertDialogTrigger>
                          <AlertDialogContent>
                            <AlertDialogHeader>
                              <AlertDialogTitle>确认删除</AlertDialogTitle>
                              <AlertDialogDescription>
                                确定要删除这个子任务吗？此操作无法撤销。
                              </AlertDialogDescription>
                            </AlertDialogHeader>
                            <AlertDialogFooter>
                              <AlertDialogCancel>取消</AlertDialogCancel>
                              <AlertDialogAction 
                                onClick={() => handleDeleteSubNode(subNode.id)}
                                className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
                              >
                                确定删除
                              </AlertDialogAction>
                            </AlertDialogFooter>
                          </AlertDialogContent>
                        </AlertDialog>
                      </div>
                    </div>
                    
                    {/* 问题和目标 */}
                    <div className="space-y-2">
                      <div>
                        <Label className="text-xs text-muted-foreground">问题</Label>
                        {subNode.isEditing ? (
                          <Textarea
                            value={subNode.question}
                            onChange={(e) => handleEditSubNode(subNode.id, 'question', e.target.value)}
                            className="mt-1 min-h-[60px] resize-none"
                          />
                        ) : (
                          <p className="text-sm mt-1">{subNode.question}</p>
                        )}
                      </div>
                      
                      <div>
                        <Label className="text-xs text-muted-foreground">目标</Label>
                        {subNode.isEditing ? (
                          <Textarea
                            value={subNode.target}
                            onChange={(e) => handleEditSubNode(subNode.id, 'target', e.target.value)}
                            className="mt-1 min-h-[60px] resize-none"
                          />
                        ) : (
                          <p className="text-sm mt-1">{subNode.target}</p>
                        )}
                      </div>
                    </div>
                    
                    {/* 编辑操作 */}
                    {subNode.isEditing && (
                      <div className="flex gap-2">
                        <Button
                          onClick={() => handleSaveSubNode(subNode.id)}
                          size="sm"
                          className="flex-1"
                        >
                          <Save className="w-4 h-4 mr-2" />
                          保存
                        </Button>
                        <Button
                          onClick={() => setSubNodes(prev => prev.map(s => 
                            s.id === subNode.id ? { ...s, isEditing: false } : s
                          ))}
                          variant="outline"
                          size="sm"
                          className="flex-1"
                        >
                          <X className="w-4 h-4 mr-2" />
                          取消
                        </Button>
                      </div>
                    )}
                    
                    {/* 执行操作 */}
                    {!subNode.isEditing && subNode.status === 'pending' && (
                      <Button
                        onClick={() => handleStartSubNode(subNode.id)}
                        size="sm"
                        className="w-full"
                      >
                        开始执行
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export default DecomposeTab;