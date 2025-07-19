"use client"

import { useEffect, useState, useRef } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { getToken, getUser } from '@/lib/auth';
import { useGlobalStore } from '@/store/globalStore';
import useToast from '@/hooks/use-toast';

interface AuthGuardProps {
  children: React.ReactNode;
}

// 不需要认证的页面路径
const PUBLIC_ROUTES = [
  '/login',
  '/register',
  '/404',
  '/not-found'
];

// 检查路径是否为公开路由
function isPublicRoute(pathname: string): boolean {
  return PUBLIC_ROUTES.some(route => pathname.startsWith(route));
}

export function AuthGuard({ children }: AuthGuardProps) {
  const router = useRouter();
  const pathname = usePathname();
  const [isLoading, setIsLoading] = useState(true);
  const [isMounted, setIsMounted] = useState(false);
  const setUser = useGlobalStore((state) => state.setUser);
  const user = useGlobalStore((state) => state.user);
  const hasTriedRestore = useRef(false);
  const { toast } = useToast();

  // Handle hydration
  useEffect(() => {
    setIsMounted(true);
  }, []);

  useEffect(() => {
    if (!isMounted) return;

    const checkAuth = () => {
      const token = getToken();
      
      // 如果是公开路由，直接允许访问
      if (isPublicRoute(pathname)) {
        setIsLoading(false);
        return;
      }

      // 如果没有token且不是公开路由，重定向到登录页
      if (!token) {
        toast({
          title: "请登录~",
          variant: "destructive"
        })
        router.replace('/login');
        return;
      }

      // 有token，允许访问
      setIsLoading(false);
    };

    checkAuth();
  }, [pathname, router, isMounted]);

  // 单独的effect处理用户信息恢复，避免循环依赖
  useEffect(() => {
    if (!isMounted) return;
    
    const token = getToken();
    if (token && !user && !hasTriedRestore.current) {
      hasTriedRestore.current = true;
      const savedUser = getUser();
      if (savedUser) {
        setUser(savedUser);
      }
    }
  }, [user, setUser, isMounted]);

  // 在检查认证状态时显示加载状态
  if (!isMounted || isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-900"></div>
      </div>
    );
  }

  return <>{children}</>;
}