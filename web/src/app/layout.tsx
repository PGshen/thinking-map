import './globals.css';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'ThinkingMap',
  description: 'AI 问题解决可视化助手',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  // 预留 Providers，可按需引入 Zustand、Theme 等
  return (
    <html lang="zh-CN">
      <body>{children}</body>
    </html>
  );
}
