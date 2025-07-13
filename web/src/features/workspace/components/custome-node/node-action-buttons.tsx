import React from 'react';
import { Pencil, Trash2, Plus } from 'lucide-react';
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

interface NodeActionButtonsProps {
  id: string;
  mapId: string | null;
  onEdit?: (id: string) => void;
  onDelete?: (mapId: string | null, id: string) => void;
  onAddChild?: (id: string) => void;
}

export const NodeActionButtons: React.FC<NodeActionButtonsProps> = ({ id, mapId, onEdit, onDelete, onAddChild }) => {
  return (
    <div className="flex gap-1.5 bg-white/80 backdrop-blur-sm rounded-lg p-1.5 shadow-lg border border-gray-100">
      <button 
        className="p-1.5 rounded-md hover:bg-blue-50 text-blue-500 transition-colors"
        onClick={e => { e.stopPropagation(); onEdit?.(id); }}
      >
        <Pencil className="w-3.5 h-3.5" />
      </button>
      <AlertDialog>
        <AlertDialogTrigger asChild>
          <button 
            className="p-1.5 rounded-md hover:bg-red-50 text-red-500 transition-colors"
            onClick={e => e.stopPropagation()}
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        </AlertDialogTrigger>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认删除</AlertDialogTitle>
            <AlertDialogDescription>
              确定要删除这个节点吗？删除后将无法恢复。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={e => e.stopPropagation()}>取消</AlertDialogCancel>
            <AlertDialogAction onClick={e => { e.stopPropagation(); onDelete?.(mapId, id); }}>删除</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <button 
        className="p-1.5 rounded-md hover:bg-green-50 text-green-500 transition-colors"
        onClick={e => { e.stopPropagation(); onAddChild?.(id); }}
      >
        <Plus className="w-3.5 h-3.5" />
      </button>
    </div>
  );
};