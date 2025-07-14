'use client';

import React from 'react';
import { Label } from '@/components/ui/label';
import { DependencyItem } from './dependency-item';
import { NodeContextItem } from '@/types/node';

interface DependencySectionProps {
  title: string;
  items: NodeContextItem[];
  type: 'ancestor' | 'prevSibling' | 'children';
  onItemsChange: (items: NodeContextItem[]) => void;
}

export function DependencySection({ title, items, type, onItemsChange }: DependencySectionProps) {
  const handleUpdateItem = (index: number, updatedItem: NodeContextItem) => {
    const newItems = [...items];
    newItems[index] = updatedItem;
    onItemsChange(newItems);
  };

  const handleDeleteItem = (index: number) => {
    const newItems = items.filter((_, i) => i !== index);
    onItemsChange(newItems);
  };

  const handleAddItem = (index: number, position: 'above' | 'below') => {
    const newItem: NodeContextItem = {
      question: '',
      target: '',
      status: 'pending'
    };

    const newItems = [...items];
    const insertIndex = position === 'above' ? index : index + 1;
    newItems.splice(insertIndex, 0, newItem);
    onItemsChange(newItems);
  };

  return (
    <div className="space-y-3">
      <Label>{title}</Label>
      <div className="space-y-2">
        {items.map((item, index) => (
          <DependencyItem
            key={index}
            item={item}
            onUpdate={(updatedItem) => handleUpdateItem(index, updatedItem)}
            onDelete={() => handleDeleteItem(index)}
            onAddAbove={() => handleAddItem(index, 'above')}
            onAddBelow={() => handleAddItem(index, 'below')}
          />
        ))}
      </div>
    </div>
  );
}