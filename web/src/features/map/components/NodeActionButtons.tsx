import React from 'react';
import { Pencil, Trash2, Plus } from 'lucide-react';

interface NodeActionButtonsProps {
  id: string;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
  onAddChild?: (id: string) => void;
}

export const NodeActionButtons: React.FC<NodeActionButtonsProps> = ({ id, onEdit, onDelete, onAddChild }) => {
  return (
    <div className="flex gap-1.5 bg-white/80 backdrop-blur-sm rounded-lg p-1.5 shadow-lg border border-gray-100">
      <button 
        className="p-1.5 rounded-md hover:bg-blue-50 text-blue-500 transition-colors"
        onClick={e => { e.stopPropagation(); onEdit?.(id); }}
      >
        <Pencil className="w-3.5 h-3.5" />
      </button>
      <button 
        className="p-1.5 rounded-md hover:bg-red-50 text-red-500 transition-colors"
        onClick={e => { e.stopPropagation(); onDelete?.(id); }}
      >
        <Trash2 className="w-3.5 h-3.5" />
      </button>
      <button 
        className="p-1.5 rounded-md hover:bg-green-50 text-green-500 transition-colors"
        onClick={e => { e.stopPropagation(); onAddChild?.(id); }}
      >
        <Plus className="w-3.5 h-3.5" />
      </button>
    </div>
  );
};