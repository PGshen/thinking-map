import { useState } from 'react';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { Pencil, X } from "lucide-react";

interface EditableTextareaProps {
  value: string;
  onChange: (value: string) => void;
  onDelete?: () => void;
  className?: string;
  placeholder?: string;
  showDeleteButton?: boolean;
}

export function EditableTextarea({
  value,
  onChange,
  onDelete,
  className = '',
  placeholder = '',
  showDeleteButton = true
}: EditableTextareaProps) {
  const [isEditing, setIsEditing] = useState(false);

  return (
    <div className="relative group/item">
      <Textarea
        value={value}
        onFocus={() => setIsEditing(true)}
        onBlur={(e) => {
          setIsEditing(false);
          onChange(e.target.value);
        }}
        onChange={(e) => onChange(e.target.value)}
        className={`${showDeleteButton ? 'pr-14' : 'pr-7'} resize-none ${className}`}
        placeholder={placeholder}
        readOnly={!isEditing}
      />
      <div className="absolute top-2 right-2 opacity-0 group-hover/item:opacity-100 transition-opacity flex space-x-1">
        <Button
          variant="secondary"
          size="icon"
          className="h-6 w-6"
          onClick={() => setIsEditing(true)}
        >
          <Pencil className="h-4 w-4" />
        </Button>
        {showDeleteButton && onDelete && (
          <Button
            variant="secondary"
            size="icon"
            className="h-6 w-6 text-red-500 hover:text-red-600"
            onClick={onDelete}
          >
            <X className="h-4 w-4" />
          </Button>
        )}
      </div>
    </div>
  );
}