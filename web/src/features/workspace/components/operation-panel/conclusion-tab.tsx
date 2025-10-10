/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/conclusion-tab.tsx
 */
'use client';

import React, { useState, useEffect, useCallback } from 'react';
import { Save, RotateCcw, CheckCircle, AlertCircle, Clock, FileText, ChevronDown, ChevronUp, Edit, Eye } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Card, CardContent } from '@/components/ui/card';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { toast } from 'sonner';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';
import { conclusion } from '@/api/node';

// 导入新的Notion编辑器
import EditorClient from '@/components/editor-client';

interface ConclusionTabProps {
  nodeID: string;
  node: any; // TODO: 使用正确的节点类型
}

interface ExecutionLog {
  id: string;
  timestamp: string;
  type: 'info' | 'success' | 'warning' | 'error';
  message: string;
  details?: string;
}



export function ConclusionTab({ nodeID, node }: ConclusionTabProps) {
  const nodeData = node.data as any;
  const [hasChanges, setHasChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);
  const [executionLogs, setExecutionLogs] = useState<ExecutionLog[]>([]);
  const [executionProgress, setExecutionProgress] = useState(0);
  const [isLogsCollapsed, setIsLogsCollapsed] = useState(false);
  const [editorContent, setEditorContent] = useState(nodeData?.conclusion?.content || '');
  const [isEditing, setIsEditing] = useState(false);
  const { actions } = useWorkspaceStore();

  // 处理编辑器内容变化
  const handleEditorChange = useCallback((content: string) => {
    setEditorContent(content);
    setHasChanges(content !== (nodeData?.conclusion || ''));
  }, [nodeData?.conclusion]);


  
  // 监听节点数据变化
  useEffect(() => {
    if (nodeData) {
      setEditorContent(nodeData.conclusion?.content || '');
      setHasChanges(false);
    }
    
    // 模拟加载执行日志
    const status = nodeData?.status || 'pending';
    if (status === 'running' || status === 'completed') {
      loadExecutionLogs();
    }
  }, [node]);

  // 加载执行日志
  const loadExecutionLogs = () => {
    // TODO: 从API获取真实的执行日志
    const mockLogs: ExecutionLog[] = [
      {
        id: '1',
        timestamp: '2025-01-27 10:30:00',
        type: 'info',
        message: '开始执行任务',
        details: '初始化执行环境'
      },
      {
        id: '2',
        timestamp: '2025-01-27 10:30:15',
        type: 'info',
        message: '分析问题要素',
        details: '正在分析问题的核心组成部分'
      },
      {
        id: '3',
        timestamp: '2025-01-27 10:31:20',
        type: 'success',
        message: '问题分析完成',
        details: '已识别出3个关键要素'
      },
      {
        id: '4',
        timestamp: '2025-01-27 10:32:10',
        type: 'info',
        message: '制定解决方案',
        details: '基于分析结果制定可行方案'
      }
    ];
    
    const nodeData = node.data as any;
    const status = nodeData?.status || 'pending';
    if (status === 'completed') {
      mockLogs.push({
        id: '5',
        timestamp: '2025-01-27 10:35:00',
        type: 'success',
        message: '任务执行完成',
        details: '所有子任务已成功完成'
      });
      setExecutionProgress(100);
    } else {
      setExecutionProgress(75);
    }
    
    setExecutionLogs(mockLogs);
  };

  const handleSave = async () => {
    if (!hasChanges) return;
    
    setIsSaving(true);
    try {
      // TODO: 调用API保存结论
      // await updateNodeConclusion(nodeID, editorContent);
      
      // 更新本地状态
      const nodeData = node.data as any;
      const currentStatus = nodeData?.status || 'pending';
      actions.updateNode(nodeID, { 
        data: {
          ...nodeData,
          conclusion: {
            ...nodeData?.conclusion,
            content: editorContent,
          },
          status: editorContent.trim() ? 'completed' : currentStatus
        }
      });
      
      toast.success('结论已保存');
      
      setHasChanges(false);
    } catch (error) {
      toast.error('保存失败，请重试');
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    const nodeData = node.data as any;
    setEditorContent(nodeData?.conclusion?.content || '');
    setHasChanges(false);
  };

  // 开始结论生成
  const handleStartConclusion = async () => {
    if (!nodeID) {
      toast.error('请先选择一个节点');
      return;
    }

    setIsGenerating(true);
    try {
      // 更新节点状态为运行中
      const nodeData = node.data as any;
      actions.updateNode(nodeID, { 
        data: {
          ...nodeData,
          status: 'in_conclusion'
        }
      });

      // 调用结论生成API
        const response = await conclusion(nodeID, editorContent, '请基于当前内容生成结论');
      
      if (response.code === 200) {
        toast.success('结论生成成功');
        // 加载执行日志
        loadExecutionLogs();
      } else {
        toast.error(response.message || '结论生成失败');
        actions.updateNode(nodeID, { 
          data: {
            ...nodeData,
            status: 'error'
          }
        });
      }
    } catch (error) {
      console.error('结论生成失败:', error);
      toast.error('结论生成失败');
      const nodeData = node.data as any;
      actions.updateNode(nodeID, { 
        data: {
          ...nodeData,
          status: 'error'
        }
      });
    } finally {
      setIsGenerating(false);
    }
  };
  
  const toggleLogsCollapse = () => {
    setIsLogsCollapsed(!isLogsCollapsed);
  };

  // 获取日志类型图标
  const getLogIcon = (type: string) => {
    switch (type) {
      case 'success': return <CheckCircle className="w-4 h-4 text-green-600" />;
      case 'warning': return <AlertCircle className="w-4 h-4 text-yellow-600" />;
      case 'error': return <AlertCircle className="w-4 h-4 text-red-600" />;
      default: return <Clock className="w-4 h-4 text-blue-600" />;
    }
  };

  // 获取日志类型样式
  const getLogStyle = (type: string) => {
    switch (type) {
      case 'success': return 'border-l-green-500 bg-green-50';
      case 'warning': return 'border-l-yellow-500 bg-yellow-50';
      case 'error': return 'border-l-red-500 bg-red-50';
      default: return 'border-l-blue-500 bg-blue-50';
    }
  };

  return (
    <div className="h-full flex flex-col">
      {/* 上半部分：状态和结论编辑器 */}
      <div className="flex-1 flex flex-col min-h-0 overflow-hidden">
        {/* 结论编辑器 */}
        <div className="flex-1 flex flex-col min-h-0 overflow-hidden">          
          {/* Notion风格编辑器 */}
          <div className="flex-1 mt-1 border rounded-md overflow-hidden">
            <EditorClient
              initContent={editorContent}
              placeholder="请输入结论..."
              onChange={handleEditorChange}
              editable={isEditing && (nodeData?.status || 'pending') !== 'running'}
              className={`min-h-[200px] ${isEditing ? 'p-4' : 'px-2 py-4'}`}
              hideToolbar={!isEditing}
              isEditing={isEditing}
            />
          </div>
          
          {/* 保存操作 */}
          {node.status !== 'running' && (
            <div className="flex gap-2 mt-3">
              <Button
                onClick={() => setIsEditing(!isEditing)}
                variant={isEditing ? "default" : "outline"}
                size="sm"
              >
                {isEditing ? <Eye className="w-4 h-4 mr-2" /> : <Edit className="w-4 h-4 mr-2" />}
                {isEditing ? '预览' : '编辑'}
              </Button>
              
              <Button
                onClick={handleReset}
                disabled={!hasChanges}
                variant="outline"
                className="flex-1"
                size="sm"
              >
                <RotateCcw className="w-4 h-4 mr-2" />
                重置
              </Button>

              <Button
                onClick={handleStartConclusion}
                disabled={isGenerating || !nodeID}
                variant="default"
                className="flex-1"
                size="sm"
              >
                <CheckCircle className="w-4 h-4 mr-2" />
                {isGenerating ? '生成中...' : '开始结论'}
              </Button>

              <Button
                onClick={handleSave}
                disabled={!hasChanges || isSaving}
                className="flex-1"
                size="sm"
              >
                <Save className="w-4 h-4 mr-2" />
                {isSaving ? '保存中...' : '保存结论'}
              </Button>
            </div>
          )}
        </div>
      </div>
      
      <Separator className="my-4" />
      
      {/* 下半部分：吸附在底部的可折叠执行日志 */}
      <div className="sticky bottom-0 bg-background border-t">
        <Collapsible
          open={!isLogsCollapsed}
          onOpenChange={toggleLogsCollapse}
          className="w-full"
        >
        <div className="flex items-center justify-between">
          <CollapsibleTrigger asChild>
            <Button variant="ghost" size="sm" className="flex items-center gap-1 p-0">
              <h4 className="font-medium">执行日志</h4>
              {isLogsCollapsed ? <ChevronDown className="w-4 h-4" /> : <ChevronUp className="w-4 h-4" />}
            </Button>
          </CollapsibleTrigger>
        </div>
        
        <CollapsibleContent className="mt-2">
          {executionLogs.length > 0 ? (
            <div className="space-y-3 max-h-[300px] overflow-y-auto">
              {executionLogs.map((log) => (
                <div
                  key={log.id}
                  className={`p-3 border-l-4 rounded-r-md ${getLogStyle(log.type)}`}
                >
                  <div className="flex items-start gap-3">
                    {getLogIcon(log.type)}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <p className="text-sm font-medium">{log.message}</p>
                        <span className="text-xs text-muted-foreground">
                          {log.timestamp}
                        </span>
                      </div>
                      {log.details && (
                        <p className="text-xs text-muted-foreground mt-1">
                          {log.details}
                        </p>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <Card>
              <CardContent className="flex items-center justify-center py-8">
                <div className="text-center">
                  <FileText className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
                  <p className="text-sm text-muted-foreground">
                    {node.status === 'pending' ? '任务尚未开始执行' : '暂无执行日志'}
                  </p>
                </div>
              </CardContent>
            </Card>
          )}
        </CollapsibleContent>
      </Collapsible>
      </div>
    </div>
  );
}

export default ConclusionTab;