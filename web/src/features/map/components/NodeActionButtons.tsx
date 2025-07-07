import React from 'react';

interface NodeActionButtonsProps {
  id: string;
  onEdit?: (id: string) => void;
  onDelete?: (id: string) => void;
  onAddChild?: (id: string) => void;
}

export const NodeActionButtons: React.FC<NodeActionButtonsProps> = ({ id, onEdit, onDelete, onAddChild }) => {
  return (
    <div className="flex gap-2 mt-2 justify-end">
      <button className="text-blue-500 hover:underline text-xs" onClick={e => { e.stopPropagation(); onEdit?.(id); }}>✏️</button>
      <button className="text-red-500 hover:underline text-xs" onClick={e => { e.stopPropagation(); onDelete?.(id); }}>🗑️</button>
      <button className="text-green-500 hover:underline text-xs" onClick={e => { e.stopPropagation(); onAddChild?.(id); }}>➕</button>
    </div>
  );
}; 