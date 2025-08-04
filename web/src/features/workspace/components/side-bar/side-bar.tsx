"use client"

import * as React from "react"
import { ArrowLeft, Settings as SettingsIcon, GalleryVerticalEnd } from "lucide-react"

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar"
import { useRouter } from "next/navigation"
import { useWorkspaceStore, useWorkspaceStoreData } from "@/features/workspace/store/workspace-store"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import MapInfo from "./map-info"
import Settings from "./settings"
import useWorkspaceData from "../../hooks/use-workspace-data"

export function InfoSidebar({ mapID, ...props }: React.ComponentProps<typeof Sidebar> & { mapID: string }) {
  const router = useRouter()
  const [showConfirm, setShowConfirm] = React.useState(false)
  const { changedNodePositions, actions } = useWorkspaceStore()
  const { savePosition } = useWorkspaceData()
  const { setOpen, toggleSidebar } = useSidebar()
  const [activePanel, setActivePanel] = React.useState<'map-info' | 'settings'>('map-info')

  const handleExit = () => {
    if (changedNodePositions.length > 0) {
      setShowConfirm(true)
    } else {
      // 清理工作区状态
      actions.reset()
      router.push('/')
    }
  }

  const confirmExit = () => {
    setShowConfirm(false)
    savePosition()
    // 清理工作区状态
    actions.reset()
    router.push('/')
  }

  const handlePanelChange = (panel: 'map-info' | 'settings') => {
    setActivePanel(panel)
    setOpen(true)
  }

  return (
    <Sidebar
      collapsible="icon"
      className="overflow-hidden *:data-[sidebar=sidebar]:flex-row"
      {...props}
    >
      {/* 第一个侧边栏 - 图标栏 */}
      <Sidebar
        collapsible="none"
        className="w-[calc(var(--sidebar-width-icon)+1px)]! border-r"
      >
        <SidebarHeader>
          <div
            className="hover:opacity-80 transition-opacity"
            onClick={toggleSidebar}
          >
            <SidebarMenu>
              <SidebarMenuItem>
                <div className="cursor-pointer bg-sidebar-primary text-sidebar-primary-foreground flex aspect-square size-8 items-center justify-center rounded-lg">
                  <GalleryVerticalEnd className="size-4" />
                </div>
              </SidebarMenuItem>
            </SidebarMenu>
          </div>
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupContent className="px-1.5 md:px-0">
              <SidebarMenu>
                {/* 退出按钮 */}
                <SidebarMenuItem>
                  <SidebarMenuButton
                    tooltip={{
                      children: "退出工作区",
                      hidden: false,
                    }}
                    onClick={handleExit}
                    className="px-2.5 md:px-2"
                  >
                    <ArrowLeft className="size-4" />
                    <span>退出</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>

                {/* 导图信息按钮 */}
                <SidebarMenuItem>
                  <SidebarMenuButton
                    tooltip={{
                      children: "导图信息",
                      hidden: false,
                    }}
                    onClick={() => handlePanelChange('map-info')}
                    className="px-2.5 md:px-2"
                  >
                    <GalleryVerticalEnd className="size-4" />
                    <span>导图信息</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>

                {/* 设置按钮 */}
                <SidebarMenuItem>
                  <SidebarMenuButton
                    tooltip={{
                      children: "设置",
                      hidden: false,
                    }}
                    onClick={() => handlePanelChange('settings')}
                    className="px-2.5 md:px-2"
                  >
                    <SettingsIcon className="size-4" />
                    <span>设置</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        </SidebarContent>
      </Sidebar>

      {/* 第二个侧边栏 - 内容区 */}
      <Sidebar collapsible="none" className="hidden flex-1 md:flex">
        <SidebarContent>
          <SidebarGroup className="px-4">
            {activePanel === 'map-info' ? <MapInfo /> : <Settings />}
          </SidebarGroup>
        </SidebarContent>
      </Sidebar>

      {/* 退出确认对话框 */}
      <AlertDialog open={showConfirm} onOpenChange={setShowConfirm}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>确认退出</AlertDialogTitle>
            <AlertDialogDescription>
              您有未保存的更改，确定要退出工作区吗？未保存的更改将会丢失。
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>取消</AlertDialogCancel>
            <AlertDialogAction onClick={confirmExit} className="bg-destructive text-destructive-foreground hover:bg-destructive/90">
              确定退出
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Sidebar>
  )
}
