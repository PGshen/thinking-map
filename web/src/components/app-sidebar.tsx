"use client"

import * as React from "react"
import {
  BookOpen,
  Bot,
  Frame,
  History,
  Map,
  PieChart,
  Settings2,
  SquareTerminal,
} from "lucide-react"

import { NavMain } from "@/components/nav-main"
import { NavProjects } from "@/components/nav-projects"
import { NavUser } from "@/components/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar"
import { useGlobalStore } from "@/store/globalStore"
import { Logo } from "@/components/logo"

// This is sample data.
const data = {
  user: {
    name: "shadcn",
    email: "m@example.com",
    avatar: "/avatar.jpg",
  },
  navMain: [
    {
      title: "思考一下",
      url: "/",
      icon: SquareTerminal,
      isActive: true
    },
    {
      title: "历史对话",
      url: "/history",
      icon: History
    },
    {
      title: "工具",
      url: "#",
      icon: Bot,
      items: [
        {
          title: "Debug",
          url: "/tool/debug",
        },
        {
          title: "Explorer",
          url: "#",
        },
        {
          title: "Quantum",
          url: "#",
        },
      ],
    },
    {
      title: "文档",
      url: "#",
      icon: BookOpen,
      items: [
        {
          title: "Introduction",
          url: "#",
        },
        {
          title: "Get Started",
          url: "#",
        },
        {
          title: "Tutorials",
          url: "#",
        },
        {
          title: "Changelog",
          url: "#",
        },
      ],
    },
    {
      title: "设置",
      url: "#",
      icon: Settings2,
      items: [
        {
          title: "General",
          url: "#",
        },
        {
          title: "Team",
          url: "#",
        },
        {
          title: "Billing",
          url: "#",
        },
        {
          title: "Limits",
          url: "#",
        },
      ],
    },
  ],
  projects: [
    {
      name: "Design Engineering",
      url: "#",
      icon: Frame,
    },
    {
      name: "Sales & Marketing",
      url: "#",
      icon: PieChart,
    },
    {
      name: "Travel",
      url: "#",
      icon: Map,
    },
  ],
}

export function AppSidebar({ onLogout, ...props }: React.ComponentProps<typeof Sidebar> & { onLogout?: () => void }) {
  const user = useGlobalStore((s) => s.user);
  const { toggleSidebar } = useSidebar();
  
  // 兼容NavUser需要的user结构
  const navUser = user
    ? {
        name: user.username || user.fullName || "用户",
        email: user.email || "",
        avatar: "/avatar.jpg", // 可根据实际user对象调整
      }
    : {
        name: "未登录",
        email: "",
        avatar: "/avatar.jpg",
      };
  return (
    <Sidebar collapsible="icon" {...props}>
      <SidebarHeader>
        <div 
          className="hover:opacity-80 transition-opacity"
          onClick={toggleSidebar}
        >
          <Logo />
        </div>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavProjects projects={data.projects} />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={navUser} onLogout={onLogout} />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
