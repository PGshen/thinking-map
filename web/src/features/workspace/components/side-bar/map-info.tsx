"use client"

import * as React from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { useToast } from "@/hooks/use-toast"
import { useWorkspaceStore } from "@/features/workspace/store/workspace-store"

export default function MapInfo() {
  const { mapDetail } = useWorkspaceStore()
  const { toast } = useToast()
  const [isEditing, setIsEditing] = React.useState(false)
  const [formData, setFormData] = React.useState({
    title: mapDetail?.title || '',
    problem: mapDetail?.problem || '',
    target: mapDetail?.target || '',
    keyPoints: mapDetail?.keyPoints || '',
    constraints: mapDetail?.constraints || ''
  })

  React.useEffect(() => {
    if (mapDetail) {
      setFormData({
        title: mapDetail.title || '',
        problem: mapDetail.problem || '',
        target: mapDetail.target || '',
        keyPoints: mapDetail.keyPoints || '',
        constraints: mapDetail.constraints || ''
      })
    }
  }, [mapDetail])

  const handleSave = async () => {
    try {
      // TODO: 调用API保存导图信息
      setIsEditing(false)
      toast({
        title: '保存成功',
        description: '导图信息已更新',
      })
    } catch (error) {
      toast({
        title: '保存失败',
        description: '更新导图信息时出错，请重试',
        variant: 'destructive',
      })
    }
  }

  const handleInputChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData(prev => ({
      ...prev,
      [field]: e.target.value
    }))
  }

  return (
    <div className="grid gap-4">
      <div className="space-y-2">
        <h4 className="font-medium leading-none">导图信息</h4>
        <p className="text-sm text-muted-foreground">
          查看和修改当前导图的基本信息
        </p>
      </div>
      <div className="grid gap-2">
        {isEditing ? (
          <>
            <div className="grid gap-1">
              <label htmlFor="title" className="text-sm font-medium leading-none">标题</label>
              <Input
                id="title"
                value={formData.title}
                className="h-8"
                onChange={handleInputChange('title')}
              />
            </div>
            <div className="grid gap-1">
              <label htmlFor="problem" className="text-sm font-medium leading-none">问题</label>
              <Textarea
                id="problem"
                value={formData.problem}
                className="min-h-[60px]"
                onChange={handleInputChange('problem')}
              />
            </div>
            <div className="grid gap-1">
              <label htmlFor="target" className="text-sm font-medium leading-none">目标</label>
              <Textarea
                id="target"
                value={formData.target}
                className="min-h-[60px]"
                onChange={handleInputChange('target')}
              />
            </div>
            <div className="grid gap-1">
              <label htmlFor="keyPoints" className="text-sm font-medium leading-none">关键点</label>
              <Textarea
                id="keyPoints"
                value={formData.keyPoints}
                className="min-h-[60px]"
                onChange={handleInputChange('keyPoints')}
              />
            </div>
            <div className="grid gap-1">
              <label htmlFor="constraints" className="text-sm font-medium leading-none">约束</label>
              <Textarea
                id="constraints"
                value={formData.constraints}
                className="min-h-[60px]"
                onChange={handleInputChange('constraints')}
              />
            </div>
            <div className="flex justify-end gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setIsEditing(false)}
              >
                取消
              </Button>
              <Button size="sm" onClick={handleSave}>
                保存
              </Button>
            </div>
          </>
        ) : (
          <>
            <div className="grid gap-1">
              <span className="text-sm font-medium leading-none">标题</span>
              <p className="text-sm">{formData.title}</p>
            </div>
            <div className="grid gap-1">
              <span className="text-sm font-medium leading-none">问题</span>
              <p className="text-sm">{formData.problem}</p>
            </div>
            <div className="grid gap-1">
              <span className="text-sm font-medium leading-none">目标</span>
              <p className="text-sm">{formData.target}</p>
            </div>
            <div className="grid gap-1">
              <span className="text-sm font-medium leading-none">关键点</span>
              <p className="text-sm">{formData.keyPoints}</p>
            </div>
            <div className="grid gap-1">
              <span className="text-sm font-medium leading-none">约束</span>
              <p className="text-sm">{formData.constraints}</p>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="justify-self-end"
              onClick={() => setIsEditing(true)}
            >
              编辑
            </Button>
          </>
        )}
      </div>
    </div>
  )
}