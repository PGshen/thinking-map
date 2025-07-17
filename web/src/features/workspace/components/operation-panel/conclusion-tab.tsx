/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/conclusion-tab.tsx
 */
'use client';

import React, { useState, useEffect } from 'react';
import { Save, RotateCcw, CheckCircle, AlertCircle, Clock, FileText, Download } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { useToast } from '@/hooks/use-toast';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';

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
  const [conclusion, setConclusion] = useState(nodeData?.conclusion || '');
  const [hasChanges, setHasChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [executionLogs, setExecutionLogs] = useState<ExecutionLog[]>([]);
  const [executionProgress, setExecutionProgress] = useState(0);
  
  const { toast } = useToast();
  const { actions } = useWorkspaceStore();

  // 监听节点数据变化
  useEffect(() => {
    const nodeData = node.data as any;
    setConclusion(nodeData?.conclusion || '');
    setHasChanges(false);
    
    // 模拟加载执行日志
    const status = nodeData?.status || 'pending';
    if (status === 'running' || status === 'completed') {
      loadExecutionLogs();
    }
  }, [node]);

  // 检查是否有未保存的更改
  useEffect(() => {
    const nodeData = node.data as any;
    const changed = conclusion !== (nodeData?.conclusion || '');
    setHasChanges(changed);
  }, [conclusion, node]);

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
      // await updateNodeConclusion(nodeID, conclusion);
      
      // 更新本地状态
      const nodeData = node.data as any;
      const currentStatus = nodeData?.status || 'pending';
      actions.updateNode(nodeID, { 
        data: {
          ...nodeData,
          conclusion,
          status: conclusion.trim() ? 'completed' : currentStatus
        }
      });
      
      toast({
        title: '保存成功',
        description: '结论已更新',
      });
      
      setHasChanges(false);
    } catch (error) {
      toast({
        title: '保存失败',
        description: '更新结论时出错，请重试',
        variant: 'destructive',
      });
    } finally {
      setIsSaving(false);
    }
  };

  const handleReset = () => {
    const nodeData = node.data as any;
    setConclusion(nodeData?.conclusion || '');
    setHasChanges(false);
  };

  const handleExportLogs = () => {
    // TODO: 实现日志导出功能
    const logsText = executionLogs
      .map(log => `[${log.timestamp}] ${log.type.toUpperCase()}: ${log.message}${log.details ? ` - ${log.details}` : ''}`)
      .join('\n');
    
    const blob = new Blob([logsText], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `execution-logs-${nodeID}.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    toast({
      title: '导出成功',
      description: '执行日志已导出',
    });
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
    <div className="h-full flex flex-col space-y-6">
      {/* 执行状态 */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold">执行状态</h3>
          <Badge 
            variant={(nodeData?.status || 'pending') === 'completed' ? 'default' : 'secondary'}
            className={(nodeData?.status || 'pending') === 'completed' ? 'bg-green-100 text-green-800' : 'bg-blue-100 text-blue-800'}
          >
            {(nodeData?.status || 'pending') === 'completed' ? '已完成' : '执行中'}
          </Badge>
        </div>
        
        {/* 进度条 */}
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <span>执行进度</span>
            <span>{executionProgress}%</span>
          </div>
          <Progress value={executionProgress} className="w-full" />
        </div>
      </div>
      
      <Separator />
      
      {/* 结论编辑 */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <Label htmlFor="conclusion">执行结论</Label>
          {(nodeData?.status || 'pending') === 'completed' && (
            <CheckCircle className="w-5 h-5 text-green-600" />
          )}
        </div>
        
        <Textarea
          id="conclusion"
          value={conclusion}
          onChange={(e) => setConclusion(e.target.value)}
          placeholder={(nodeData?.status || 'pending') === 'running' ? '任务执行中，结论将在完成后生成...' : '请输入执行结论和总结...'}
          className="min-h-[120px] resize-none"
          disabled={(nodeData?.status || 'pending') === 'running'}
        />
        
        {/* 保存操作 */}
        {node.status !== 'running' && (
          <div className="flex gap-2">
            <Button
              onClick={handleSave}
              disabled={!hasChanges || isSaving}
              className="flex-1"
            >
              <Save className="w-4 h-4 mr-2" />
              {isSaving ? '保存中...' : '保存结论'}
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
        )}
      </div>
      
      <Separator />
      
      {/* 执行日志 */}
      <div className="space-y-4 flex-1 overflow-hidden">
        <div className="flex items-center justify-between">
          <h4 className="font-medium">执行日志</h4>
          {executionLogs.length > 0 && (
            <Button
              onClick={handleExportLogs}
              variant="outline"
              size="sm"
            >
              <Download className="w-4 h-4 mr-2" />
              导出
            </Button>
          )}
        </div>
        
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
      </div>
      
      {/* 子任务完成情况 */}
      {node.children && node.children.length > 0 && (
        <>
          <Separator />
          <div className="space-y-3">
            <h4 className="font-medium">子任务完成情况</h4>
            <div className="space-y-2">
              {node.children.map((child: any, index: number) => (
                <div key={child.id} className="flex items-center justify-between p-2 bg-gray-50 rounded-md">
                  <span className="text-sm font-medium">#{index + 1} {child.question}</span>
                  <Badge 
                    variant={child.status === 'completed' ? 'default' : 'secondary'}
                    className={child.status === 'completed' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}
                  >
                    {child.status === 'completed' ? '已完成' : '未完成'}
                  </Badge>
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}

export default ConclusionTab;