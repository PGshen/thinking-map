"use client"

import { useState, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { SseJsonStreamParser } from '@/lib/sse-parser'
import SidebarLayout from '@/layouts/sidebar-layout'

interface StreamData {
  [key: string]: any
}

export default function Page() {
  const [input, setInput] = useState('')
  const [method, setMethod] = useState('POST')
  const [uri, setUri] = useState('/api/problem/decompose')
  const [params, setParams] = useState('')
  const [keys, setKeys] = useState('')
  const [output, setOutput] = useState<Record<string, string[]>>({})
  const [isStreaming, setIsStreaming] = useState(false)

  const handleSubmit = useCallback(() => {
    if (!keys || !uri) return

    setIsStreaming(true)
    setOutput({})

    const jsonStream = new SseJsonStreamParser(true)
    keys.split(',').map(key => key.trim()).forEach(key => {
      jsonStream.on(key, (data, path) => {
        setOutput(prev => ({
          ...prev,
          [key]: [String(data)]
        }))
      })
    })

    let requestParams = { message: input }
    try {
      if (params) {
        requestParams = { ...requestParams, ...JSON.parse(params) }
      }
    } catch (e) {
      console.error('Invalid JSON parameters:', e)
    }

    jsonStream.connect(method as "GET" | "POST" | "PUT" | "DELETE", uri, requestParams, () => {
      console.log('close')
      setIsStreaming(false)
    })
  }, [input, keys, method, uri, params])

  return (
    <SidebarLayout>
      <div className="flex h-full gap-4 p-4">
        <div className="w-1/2 border rounded-lg p-4 flex flex-col gap-4">
          <div className="flex flex-col gap-2">
            <Label htmlFor="method">HTTP方法</Label>
            <Select value={method} onValueChange={setMethod}>
              <SelectTrigger>
                <SelectValue placeholder="选择HTTP方法" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="GET">GET</SelectItem>
                <SelectItem value="POST">POST</SelectItem>
                <SelectItem value="PUT">PUT</SelectItem>
                <SelectItem value="DELETE">DELETE</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="uri">请求URI</Label>
            <Input
              id="uri"
              value={uri}
              onChange={(e) => setUri(e.target.value)}
              placeholder="例如: /api/debug"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="params">请求参数 (JSON格式)</Label>
            <Textarea
              id="params"
              value={params}
              onChange={(e) => setParams(e.target.value)}
              placeholder='{"key": "value"}'
              className="h-[100px]"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="input">输入内容</Label>
            <Textarea
              id="input"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              placeholder="请输入要发送给LLM的内容"
              className="h-[200px]"
            />
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="keys">解析字段 (用逗号分隔)</Label>
            <Input
              id="keys"
              value={keys}
              onChange={(e) => setKeys(e.target.value)}
              placeholder="例如: content,tokens"
            />
          </div>

          <Button
            onClick={handleSubmit}
            disabled={!keys || !uri || isStreaming}
          >
            {isStreaming ? '接收中...' : '发送'}
          </Button>
        </div>

        <div className="w-1/2 border rounded-lg p-4 overflow-auto">
          {Object.entries(output).map(([key, values]) => (
            <div key={key} className="mb-4">
              <h3 className="font-medium mb-2">{key}:</h3>
              <div className="whitespace-pre-wrap">
                {values.join('')}
              </div>
            </div>
          ))}
        </div>
      </div>
    </SidebarLayout>
  )
}