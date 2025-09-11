"use client";

/* eslint-disable unicorn/no-null */
/* eslint-disable quotes */
import { useCallback, useState, useEffect } from "react"

import RichTextEditor, { BaseKit } from "reactjs-tiptap-editor"

import { locale } from 'reactjs-tiptap-editor/locale-bundle'

import { Blockquote } from "reactjs-tiptap-editor/blockquote";
import { Bold } from "reactjs-tiptap-editor/bold";
import { BulletList } from "reactjs-tiptap-editor/bulletlist";
import { Clear } from "reactjs-tiptap-editor/clear";
import { Code } from "reactjs-tiptap-editor/code";
import { CodeBlock } from "reactjs-tiptap-editor/codeblock";
import { Heading } from "reactjs-tiptap-editor/heading";
import { History } from "reactjs-tiptap-editor/history";
import { HorizontalRule } from "reactjs-tiptap-editor/horizontalrule";
import { Image } from "reactjs-tiptap-editor/image";
import { Italic } from "reactjs-tiptap-editor/italic";
import { Link } from "reactjs-tiptap-editor/link";
import { OrderedList } from "reactjs-tiptap-editor/orderedlist";
import { Strike } from "reactjs-tiptap-editor/strike";
import { Table } from "reactjs-tiptap-editor/table";
import { TaskList } from "reactjs-tiptap-editor/tasklist";
import { SlashCommand } from 'reactjs-tiptap-editor/slashcommand';

import "reactjs-tiptap-editor/style.css"
import "prism-code-editor-lightweight/layout.css";
import "prism-code-editor-lightweight/themes/github-dark.css"

import "katex/dist/katex.min.css"

const createExtensions = (placeholder?: string) => [
  BaseKit.configure({
    placeholder: {
      showOnlyCurrent: true,
      placeholder: placeholder || '开始输入...',
    },
    characterCount: {
      limit: 50_000,
    },
  }),
  SlashCommand,
  History,
  Clear,
  Heading.configure({ spacer: true }),
  Bold,
  Italic,
  Strike,
  BulletList,
  OrderedList,
  TaskList.configure({
    spacer: true,
    taskItem: {
      nested: true,
    },
  }),
  Link,
  Image.configure({
    upload: (files: File) => {
      return new Promise((resolve) => {
        setTimeout(() => {
          resolve(URL.createObjectURL(files))
        }, 500)
      })
    },
  }),
  Blockquote,
  HorizontalRule,
  Code.configure({
    toolbar: false,
  }),
  CodeBlock,
  Table,
]

function debounce(func: any, wait: number) {
  let timeout: NodeJS.Timeout
  return function (...args: any[]) {
    clearTimeout(timeout)
    // @ts-ignore
    timeout = setTimeout(() => func.apply(this, args), wait)
  }
}

interface EditorProps {
  initContent?: string
  placeholder?: string
  onChange?: (value: string) => void
  editable?: boolean
  className?: string
  hideToolbar?: boolean
  isEditing?: boolean
}

function Editor({ initContent, placeholder, onChange, editable = true, className, hideToolbar = false, isEditing = true }: EditorProps) {
  const [content, setContent] = useState(initContent)
  const [theme, setTheme] = useState('light')

  const debouncedOnChange = useCallback(
    debounce((value: any) => {
      onChange?.(value)
    }, 300),
    [onChange],
  )

  const onValueChange = useCallback(
    (value: any) => {
      setContent(value)
      debouncedOnChange(value)
    },
    [debouncedOnChange],
  )

  useEffect(() => {
    locale.setLang('zh_CN')
  }, [initContent])

  return (
    <main>
      <div>

        <RichTextEditor
          output="html"
          content={content as any}
          onChangeContent={onValueChange}
          extensions={createExtensions(placeholder)}
          dark={theme === 'dark'}
          disabled={!editable}
          contentClass={`prosemirror-custom-padding ${isEditing ? 'editing-mode' : 'preview-mode'} ${className || ''}`}
          hideToolbar={hideToolbar}
          bubbleMenu={{
            render({ extensionsNames, editor, disabled }, bubbleDefaultDom) {
              return bubbleDefaultDom
            },
          }}
        />

        {/* {typeof content === 'string' && (
          <textarea
            className="textarea"
            readOnly
            style={{
              marginTop: 20,
              height: 500,
              width: '100%',
              borderRadius: 4,
              padding: 10,
            }}
            value={content}
          />
        )} */}
      </div>
    </main>
  )
}

export default Editor
