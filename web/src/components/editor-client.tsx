'use client';

import dynamic from 'next/dynamic';
import React from 'react';

const Editor = dynamic(() => import("@/components/editor"), {
  ssr: false,
  loading: () => <p>Loading...</p>,
});

interface EditorClientProps {
  initContent?: string
  placeholder?: string
  onChange?: (value: string) => void
  editable?: boolean
  className?: string
  hideToolbar?: boolean
  isEditing?: boolean
}

function EditorClient({ initContent, placeholder, onChange, editable, className, hideToolbar, isEditing }: EditorClientProps) {
  return (
    <Editor
      initContent={initContent}
      placeholder={placeholder}
      onChange={onChange}
      editable={editable}
      className={className}
      hideToolbar={hideToolbar}
      isEditing={isEditing}
    />
  )
}



export default EditorClient;