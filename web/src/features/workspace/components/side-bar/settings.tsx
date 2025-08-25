"use client"

import * as React from "react"
import { useToast } from "@/hooks/use-toast"
import { Button } from "@/components/ui/button"
import { Label } from "@/components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Slider } from "@/components/ui/slider"
import { Switch } from "@/components/ui/switch"
import { Separator } from "@/components/ui/separator"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { useWorkspaceStore } from "@/features/workspace/store/workspace-store"
import { Globe, MapPin, ArrowUpDown, ArrowLeftRight, Settings as SettingsIcon, Zap, Eye } from "lucide-react"

export default function Settings() {
  const { toast } = useToast()
  const settings = useWorkspaceStore(state => state.settings)
  const actions = useWorkspaceStore(state => state.actions)

  const handleExport = () => {
    toast({
      title: '导出功能',
      description: '导出功能正在开发中...',
    })
  }

  const handleHelp = () => {
    window.open('/help', '_blank')
  }

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h4 className="font-medium leading-none flex items-center gap-2">
          <SettingsIcon className="h-4 w-4" />
          设置
        </h4>
        <p className="text-sm text-muted-foreground">
          调整导图的显示和操作设置
        </p>
      </div>

      {/* 布局设置 */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm flex items-center gap-2">
            <ArrowUpDown className="h-4 w-4" />
            布局设置
          </CardTitle>
          <CardDescription className="text-xs">
            控制节点的布局方式和更新策略
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 布局类型 */}
          <div className="space-y-2">
            <Label className="text-xs font-medium">布局类型</Label>
            <Select
              value={settings.layoutType}
              onValueChange={(value: 'global' | 'local') => 
                actions.updateSettings({ layoutType: value })
              }
            >
              <SelectTrigger className="h-8">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="global">
                  <div className="flex items-center gap-2">
                    <Globe className="h-3 w-3" />
                    全局更新
                  </div>
                </SelectItem>
                <SelectItem value="local">
                  <div className="flex items-center gap-2">
                    <MapPin className="h-3 w-3" />
                    局部更新
                  </div>
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* 布局方向 */}
          <div className="space-y-2">
            <Label className="text-xs font-medium">布局方向</Label>
            <Select
              value={settings.layoutConfig.direction}
              onValueChange={(value: 'TB' | 'LR' | 'BT' | 'RL') => 
                actions.updateSettings({ 
                  layoutConfig: { ...settings.layoutConfig, direction: value }
                })
              }
            >
              <SelectTrigger className="h-8">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="TB">
                  <div className="flex items-center gap-2">
                    <ArrowUpDown className="h-3 w-3" />
                    从上到下
                  </div>
                </SelectItem>
                <SelectItem value="LR">
                  <div className="flex items-center gap-2">
                    <ArrowLeftRight className="h-3 w-3" />
                    从左到右
                  </div>
                </SelectItem>
                {/* <SelectItem value="BT">
                  <div className="flex items-center gap-2">
                    <ArrowUpDown className="h-3 w-3 rotate-180" />
                    从下到上
                  </div>
                </SelectItem>
                <SelectItem value="RL">
                  <div className="flex items-center gap-2">
                    <ArrowLeftRight className="h-3 w-3 rotate-180" />
                    从右到左
                  </div>
                </SelectItem> */}
              </SelectContent>
            </Select>
          </div>

          {/* 节点间距 */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="text-xs font-medium">节点间距</Label>
              <span className="text-xs text-muted-foreground">{settings.layoutConfig.nodeSep}px</span>
            </div>
            <Slider
               value={[settings.layoutConfig.nodeSep]}
               onValueChange={([value]: number[]) => 
                 actions.updateSettings({
                   layoutConfig: { ...settings.layoutConfig, nodeSep: value }
                 })
               }
              max={200}
              min={20}
              step={10}
              className="w-full"
            />
          </div>

          {/* 层级间距 */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="text-xs font-medium">层级间距</Label>
              <span className="text-xs text-muted-foreground">{settings.layoutConfig.rankSep}px</span>
            </div>
            <Slider
               value={[settings.layoutConfig.rankSep]}
               onValueChange={([value]: number[]) => 
                 actions.updateSettings({
                   layoutConfig: { ...settings.layoutConfig, rankSep: value }
                 })
               }
              max={300}
              min={50}
              step={10}
              className="w-full"
            />
          </div>
        </CardContent>
      </Card>

      {/* 动画设置 */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm flex items-center gap-2">
            <Zap className="h-4 w-4" />
            动画设置
          </CardTitle>
          <CardDescription className="text-xs">
            控制布局变化时的动画效果
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 动画时长 */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <Label className="text-xs font-medium">动画时长</Label>
              <span className="text-xs text-muted-foreground">{settings.animationConfig.duration}ms</span>
            </div>
            <Slider
               value={[settings.animationConfig.duration]}
               onValueChange={([value]: number[]) => 
                 actions.updateSettings({
                   animationConfig: { ...settings.animationConfig, duration: value }
                 })
               }
              max={2000}
              min={100}
              step={100}
              className="w-full"
            />
          </div>
        </CardContent>
      </Card>

      {/* 显示设置 */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm flex items-center gap-2">
            <Eye className="h-4 w-4" />
            显示设置
          </CardTitle>
          <CardDescription className="text-xs">
            控制界面元素的显示
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* 显示小地图 */}
          <div className="flex items-center justify-between">
            <Label className="text-xs font-medium">显示小地图</Label>
            <Switch
               checked={settings.showMinimap}
               onCheckedChange={(checked: boolean) => 
                 actions.updateSettings({ showMinimap: checked })
               }
             />
          </div>

          {/* 显示控制面板 */}
          <div className="flex items-center justify-between">
            <Label className="text-xs font-medium">显示控制面板</Label>
            <Switch
               checked={settings.showControls}
               onCheckedChange={(checked: boolean) => 
                 actions.updateSettings({ showControls: checked })
               }
             />
          </div>

          {/* 自动保存 */}
          <div className="flex items-center justify-between">
            <Label className="text-xs font-medium">自动保存</Label>
            <Switch
               checked={settings.autoSave}
               onCheckedChange={(checked: boolean) => 
                 actions.updateSettings({ autoSave: checked })
               }
             />
          </div>
        </CardContent>
      </Card>

      <Separator />

      {/* 操作按钮 */}
      <div className="space-y-2">
        <Button
          variant="outline"
          size="sm"
          onClick={handleExport}
          className="w-full"
        >
          导出导图
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={handleHelp}
          className="w-full"
        >
          帮助文档
        </Button>
      </div>
    </div>
  )
}