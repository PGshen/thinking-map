import './globals.css';
import type { Metadata } from 'next';
import Script from 'next/script';
import { Toaster } from "@/components/ui/sonner"


export const metadata: Metadata = {
  title: 'ThinkingMap',
  description: 'AI 问题解决可视化助手',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  // 预留 Providers，可按需引入 Zustand、Theme 等
  return (
    <html lang="zh-CN">
      <head>
        {/* 修复 markdown-it 在 Next.js 15 + Turbopack 环境下的 isSpace 函数未定义问题 */}
        <Script id="markdown-it-isSpace-fix" strategy="beforeInteractive">
          {`
            if (typeof globalThis.isSpace === 'undefined') {
              globalThis.isSpace = function(code) {
                return code === 0x20 || code === 0x09 || code === 0x0A || code === 0x0B || code === 0x0C || code === 0x0D;
              };
            }
          `}
        </Script>
      </head>
      <body>
        <main>
        {children}
        </main>
        <Toaster richColors position="top-center"/>
      </body>
    </html>
  );
}
