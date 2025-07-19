/*
 * @Date: 2025-07-01 23:48:29
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-02 00:41:08
 * @FilePath: /thinking-map/web/src/layouts/sidebar-layout.tsx
 */
"use client"

import React from 'react';
import { AppSidebar } from '@/components/app-sidebar';
import { SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import { useRouter } from 'next/navigation';
import { logout } from '@/api/auth';
import { removeToken, getToken } from '@/lib/auth';
import { useGlobalStore } from '@/store/globalStore';
import { toast } from 'sonner';

export default function SidebarLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const setUser = useGlobalStore((state) => state.setUser);

  const handleLogout = async () => {
    try {
      await logout();
      removeToken();
      setUser(null); // 清除store中的用户信息
      toast.info("已登出！")
      router.push('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <AppSidebar onLogout={handleLogout} />
        {/* <div className="relative">
          <SidebarTrigger className="absolute left-0 top-4 -translate-y-1/2 z-20" />
        </div> */}
        <main className="flex-1 min-w-0 bg-background">{children}</main>
      </div>
    </SidebarProvider>
  );
}