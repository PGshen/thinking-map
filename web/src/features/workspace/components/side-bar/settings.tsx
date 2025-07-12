"use client"

import * as React from "react"
import { useToast } from "@/hooks/use-toast"
import { Button } from "@/components/ui/button"

export default function Settings() {
  const { toast } = useToast()

  const handleExport = () => {
    toast({
      title: '导出功能',
      description: '导出功能正在开发中...',
    })
  }

  const handleLayoutSettings = () => {
    toast({
      title: '布局设置',
      description: '布局设置功能正在开发中...',
    })
  }

  const handleDisplayOptions = () => {
    toast({
      title: '显示选项',
      description: '显示选项功能正在开发中...',
    })
  }

  const handleCollaboration = () => {
    toast({
      title: '协作设置',
      description: '协作设置功能正在开发中...',
    })
  }

  const handleHelp = () => {
    window.open('/help', '_blank')
  }

  return (
    <div className="grid gap-4">
      <div className="space-y-2">
        <h4 className="font-medium leading-none">设置</h4>
        <p className="text-sm text-muted-foreground">
          调整导图的显示和操作设置
        </p>
      </div>
      <div className="grid gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={handleExport}
        >
          导出导图
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={handleLayoutSettings}
        >
          布局设置
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={handleDisplayOptions}
        >
          显示选项
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={handleCollaboration}
        >
          协作设置
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={handleHelp}
        >
          帮助文档
        </Button>
      </div>
    </div>
  )
}