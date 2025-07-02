/*
 * @Date: 2025-07-01 23:48:29
 * @LastEditors: peng pgs1108pgs@gmail.com
 * @LastEditTime: 2025-07-02 00:41:08
 * @FilePath: /thinking-map/web/src/layouts/sidebar-layout.tsx
 */
import React from 'react';
import { AppSidebar } from '../components/app-sidebar';
import { SidebarProvider, SidebarTrigger } from '../components/ui/sidebar';

export default function SidebarLayout({ children }: { children: React.ReactNode }) {
  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <AppSidebar />
        <div className="relative">
          <SidebarTrigger className="absolute left-0 top-4 -translate-y-1/2 z-20" />
        </div>
        <main className="flex-1 min-w-0 bg-background">{children}</main>
      </div>
    </SidebarProvider>
  );
} 