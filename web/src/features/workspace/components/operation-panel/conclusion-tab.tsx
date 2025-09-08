/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/features/workspace/components/operation-panel/conclusion-tab.tsx
 */
'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Save, RotateCcw, CheckCircle, AlertCircle, Clock, FileText, Download, ChevronDown, ChevronUp } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Card, CardContent } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { toast } from 'sonner';
import { useWorkspaceStore } from '@/features/workspace/store/workspace-store';

// Tiptap imports
import { useEditor, EditorContent } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import BubbleMenu from '@tiptap/extension-bubble-menu';

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

interface TextSelectionPopupProps {
  text: string;
  onSubmit: (text: string) => Promise<void>;
  onClose: () => void;
}

// 文本选择弹出框组件
function TextSelectionPopup({ text, onSubmit, onClose }: TextSelectionPopupProps) {
  const [inputText, setInputText] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  
  const handleSubmit = async () => {
    if (!inputText.trim()) return;
    
    setIsLoading(true);
    try {
      await onSubmit(inputText);
      setInputText('');
      onClose();
    } catch (error) {
      console.error('Error processing text:', error);
      toast.error('处理文本时出错，请重试');
    } finally {
      setIsLoading(false);
    }
  };
  
  return (
    <div className="flex flex-col p-2 gap-2 bg-white dark:bg-gray-800 border rounded-md shadow-md min-w-[200px]">
      <div className="text-sm font-medium mb-1 truncate max-w-[200px]">
        选中文本: {text}
      </div>
      <input
        type="text"
        value={inputText}
        onChange={(e) => setInputText(e.target.value)}
        placeholder="输入提示词..."
        className="px-2 py-1 border rounded text-sm dark:bg-gray-700 dark:border-gray-600"
        autoFocus
      />
      <div className="flex gap-2">
        <Button 
          size="sm" 
          onClick={handleSubmit} 
          disabled={isLoading || !inputText.trim()}
          className="flex-1"
        >
          {isLoading ? '处理中...' : '提交'}
        </Button>
        <Button 
          size="sm" 
          variant="outline" 
          onClick={onClose}
          className="flex-1"
        >
          取消
        </Button>
      </div>
    </div>
  );
}

export function ConclusionTab({ nodeID, node }: ConclusionTabProps) {
  const nodeData = node.data as any;
  const [hasChanges, setHasChanges] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [executionLogs, setExecutionLogs] = useState<ExecutionLog[]>([]);
  const [executionProgress, setExecutionProgress] = useState(0);
  const [isLogsCollapsed, setIsLogsCollapsed] = useState(false);
  const [selectedText, setSelectedText] = useState('');
  const [showTextSelectionPopup, setShowTextSelectionPopup] = useState(false);
  const [selectionPosition, setSelectionPosition] = useState({ top: 0, left: 0 });
  
  const editorRef = useRef<HTMLDivElement>(null);
  const { actions } = useWorkspaceStore();

  // 初始化Tiptap编辑器
  const editor = useEditor({
    extensions: [StarterKit],
    content: nodeData?.conclusion || '',
    onUpdate: ({ editor }) => {
      setHasChanges(editor.getHTML() !== (nodeData?.conclusion || ''));
    },
    onSelectionUpdate: ({ editor }) => {
      const { from, to } = editor.state.selection;
      const text = editor.state.doc.textBetween(from, to, ' ');
      
      if (text && text.trim()) {
        setSelectedText(text);
        
        // 计算选择位置
        if (editorRef.current) {
          const selection = window.getSelection();
          if (selection && selection.rangeCount > 0) {
            const range = selection.getRangeAt(0);
            const rect = range.getBoundingClientRect();
            const editorRect = editorRef.current.getBoundingClientRect();
            
            setSelectionPosition({
              top: rect.bottom - editorRect.top,
              left: rect.left - editorRect.left
            });
            
            setShowTextSelectionPopup(true);
          }
        }
      } else {
        setShowTextSelectionPopup(false);
      }
    },
    immediatelyRender: false, // 解决SSR渲染问题
  });

  // 处理文本选择后的API请求
  const handleTextSelection = async (prompt: string) => {
    if (!editor) return;
    
    const { from, to } = editor.state.selection;
    const selectedText = editor.state.doc.textBetween(from, to, ' ');
    
    if (!selectedText.trim()) {
      toast.error('请先选择文本');
      return;
    }
    
    try {
      // TODO: 调用API处理选中的文本
      // const response = await processSelectedText(selectedText, prompt);
      
      // 模拟API响应
      const mockResponse = `基于"${prompt}"处理的结果: ${selectedText} 的分析结果`;
      
      // 将处理结果插入到编辑器中
      editor
        .chain()
        .focus()
        .deleteRange({ from, to })
        .insertContent(mockResponse)
        .run();
        
      toast.success('文本处理成功');
    } catch (error) {
      console.error('Error processing text:', error);
      toast.error('处理文本时出错');
    }
  };
  
  // 监听节点数据变化
  useEffect(() => {
    if (editor && nodeData) {
      editor.commands.setContent(nodeData.conclusion || '');
      setHasChanges(false);
    }
    
    // 模拟加载执行日志
    const status = nodeData?.status || 'pending';
    if (status === 'running' || status === 'completed') {
      loadExecutionLogs();
    }
  }, [node, editor]);

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
    if (!hasChanges || !editor) return;
    
    const editorContent = editor.getHTML();
    
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
          conclusion: editorContent,
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
    if (!editor) return;
    
    const nodeData = node.data as any;
    editor.commands.setContent(nodeData?.conclusion || '');
    setHasChanges(false);
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
          {/* Tiptap编辑器 */}
          <div className="flex-1 border rounded-md overflow-hidden flex flex-col relative" ref={editorRef}>
            {editor && (
              <>
                {showTextSelectionPopup && selectedText && (
                  <div 
                    className="absolute z-10"
                    style={{
                      top: `${selectionPosition.top}px`,
                      left: `${selectionPosition.left}px`,
                    }}
                  >
                    <TextSelectionPopup 
                      text={selectedText}
                      onClose={() => setShowTextSelectionPopup(false)}
                      onSubmit={handleTextSelection}
                    />
                  </div>
                )}
                
                <EditorContent 
                  editor={editor} 
                  className="flex-1 p-3 overflow-y-auto prose prose-sm max-w-none dark:prose-invert"
                  disabled={(nodeData?.status || 'pending') === 'running'}
                />
              </>
            )}
          </div>
          
          {/* 保存操作 */}
          {node.status !== 'running' && (
            <div className="flex gap-2 mt-3">
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