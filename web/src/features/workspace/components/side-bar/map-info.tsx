"use client"

import {useState, useEffect} from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { useToast } from "@/hooks/use-toast"
import useWorkspaceData from "../../hooks/use-workspace-data"
import { useWorkspaceStore } from "../../store/workspace-store"

export default function MapInfo() {
  const { mapId } = useWorkspaceStore()
  const { mapInfo, saveMap } = useWorkspaceData()

  const { toast } = useToast()
  const [isEditing, setIsEditing] = useState(false)
  const [formData, setFormData] = useState({
    title: mapInfo?.title || '',
    problem: mapInfo?.problem || '',
    target: mapInfo?.target || '',
    keyPoints: mapInfo?.keyPoints || [],
    constraints: mapInfo?.constraints || [],
  })

  useEffect(() => {
    if (mapInfo) {
      setFormData({
        title: mapInfo.title || '',
        problem: mapInfo.problem || '',
        target: mapInfo.target || '',
        keyPoints: mapInfo.keyPoints || [],
        constraints: mapInfo.constraints || [],
      })
    }
  }, [mapInfo])

  const handleSave = async () => {
    try {
      await saveMap(mapId, {
        ...formData
      })
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
    if (field === 'keyPoints' || field === 'constraints') {
      setFormData(prev => ({
        ...prev,
        [field]: e.target.value.split('\n')
      }))
    } else {
      setFormData(prev => ({
        ...prev,
        [field]: e.target.value
      }))
    }
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
                value={formData.keyPoints.join('\n')}
                className="min-h-[60px]"
                onChange={handleInputChange('keyPoints')}
              />
            </div>
            <div className="grid gap-1">
              <label htmlFor="constraints" className="text-sm font-medium leading-none">约束</label>
              <Textarea
                id="constraints"
                value={formData.constraints.join('\n')}
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
            <div className="grid gap-3">
              <div className="grid gap-1">
                <span className="text-sm font-semibold text-foreground">标题</span>
                <p className="text-base font-medium text-primary pl-2 border-l-2 border-primary/40 bg-muted/40 py-1">{formData.title}</p>
              </div>
              <div className="grid gap-1">
                <span className="text-sm font-semibold text-foreground">问题</span>
                <p className="text-sm text-muted-foreground pl-2 border-l border-muted/60 bg-muted/20 rounded-sm py-1 whitespace-pre-line">{formData.problem}</p>
              </div>
              <div className="grid gap-1">
                <span className="text-sm font-semibold text-foreground">目标</span>
                <p className="text-sm text-muted-foreground pl-2 border-l border-muted/60 bg-muted/20 rounded-sm py-1 whitespace-pre-line">{formData.target}</p>
              </div>
              <div className="grid gap-1">
                <span className="text-sm font-semibold text-foreground">关键点</span>
                <ul className="text-sm text-muted-foreground list-disc pl-7 bg-muted/10 rounded-sm py-1 min-h-[28px]">
                  {formData.keyPoints?.length && formData.keyPoints.some(Boolean)
                    ? formData.keyPoints.filter(Boolean).map((item, idx) => <li key={idx} className="mb-0.5">{item}</li>)
                    : <li className="text-muted-foreground">无</li>}
                </ul>
              </div>
              <div className="grid gap-1">
                <span className="text-sm font-semibold text-foreground">约束</span>
                <ul className="text-sm text-muted-foreground list-disc pl-7 bg-muted/10 rounded-sm py-1 min-h-[28px]">
                  {formData.constraints?.length && formData.constraints.some(Boolean)
                    ? formData.constraints.filter(Boolean).map((item, idx) => <li key={idx} className="mb-0.5">{item}</li>)
                    : <li className="text-muted-foreground">无</li>}
                </ul>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="justify-self-end mt-2"
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