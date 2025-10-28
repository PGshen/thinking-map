"use client";

/* eslint-disable unicorn/no-null */
/* eslint-disable quotes */
import { useCallback, useState, useEffect, useRef, useMemo } from "react"

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
import { Markdown } from 'tiptap-markdown';
import { renderToMarkdown } from '@tiptap/static-renderer/pm/markdown'

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
    // characterCount: {
    //   limit: 50_000,
    // },
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
  Markdown.configure({
    html: true, // 允许 HTML 输入/输出
    tightLists: true, // Markdown 输出中 <li> 内没有 <p>
    linkify: false, // 从 "https://..." 文本创建链接
    breaks: false, // Markdown 输入中的换行符转换为 <br>
    transformPastedText: true, // 允许粘贴 markdown 文本
    transformCopiedText: true, // 复制的文本转换为 markdown
  }) as any,
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
  const [theme, setTheme] = useState('light')
  const extensionsList = useMemo(() => createExtensions(placeholder), [placeholder])
  
  // 内容版本控制
  const [internalContent, setInternalContent] = useState(initContent || '')
  const lastExternalContent = useRef(initContent || '')
  const lastInternalContent = useRef(initContent || '')
  const editorKey = useRef(Math.random().toString(36).substr(2, 9))
  
  // 检测外部内容变化
  const hasExternalChange = initContent !== lastExternalContent.current
  
  // 当外部内容发生变化时，同步到内部状态
  useEffect(() => {
    if (hasExternalChange) {
      const newContent = initContent || ''
      lastExternalContent.current = newContent
      
      // 只有当内部内容与新的外部内容不同时才更新
      if (internalContent !== newContent) {
        setInternalContent(newContent)
        lastInternalContent.current = newContent
        // 强制重新渲染编辑器以同步内容
        editorKey.current = Math.random().toString(36).substr(2, 9)
      }
    }
  }, [initContent, internalContent, hasExternalChange])

  const debouncedOnChange = useCallback(
    debounce((value: any) => {
      onChange?.(value)
    }, 300),
    [onChange],
  )

  const onValueChange = useCallback(
    (value: any) => {
      // 更新内部状态
      setInternalContent(value)
      lastInternalContent.current = value
      
      // 将 JSON 内容序列化为 Markdown，并仅在有实际变化时触发 onChange
      try {
        const markdownValue = renderToMarkdown({
          extensions: extensionsList as any,
          content: value,
        })
        if (markdownValue !== lastExternalContent.current) {
          debouncedOnChange(markdownValue)
        }
      } catch (error) {
        // 序列化失败时回退为文本
        if (value !== lastExternalContent.current) {
          debouncedOnChange(typeof value === 'string' ? value : JSON.stringify(value))
        }
      }
    },
    [debouncedOnChange, extensionsList],
  )

  useEffect(() => {
    locale.setLang('zh_CN')
  }, [])

  return (
    <main>
      <div>
        <RichTextEditor
          key={editorKey.current}
          output="json"
          content={internalContent as any}
          onChangeContent={onValueChange}
          extensions={extensionsList}
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
      </div>
    </main>
  )
}

export default Editor
