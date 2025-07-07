export interface CustomNodeModel {
  id: string;
  parentId?: string;
  nodeType: string;
  question: string;
  target: string;
  conclusion?: string;
  status: 'pending' | 'running' | 'completed' | 'error';
  dependencies?: any[];
  context?: any;
  metadata?: any;
  selected?: boolean;
  childCount?: number;
  hasUnmetDependencies?: boolean;
  // 交互事件（由外部注入，非持久数据）
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
  onAddChild?: (id: string) => void;
  onSelect?: (id: string) => void;
  onDoubleClick?: (id: string) => void;
  onContextMenu?: (id: string, e: React.MouseEvent) => void;
} 