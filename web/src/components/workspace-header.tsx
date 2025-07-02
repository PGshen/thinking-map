'use client';
import { useRouter } from 'next/navigation';
import { ArrowLeftIcon } from '@radix-ui/react-icons';

export default function WorkspaceHeader() {
  const router = useRouter();
  return (
    <header className="h-14 flex items-center px-4 border-b bg-background">
      <button
        className="flex items-center gap-2 px-3 py-1 rounded hover:bg-accent transition-colors"
        onClick={() => router.push('/')}
        aria-label="返回首页"
      >
        <ArrowLeftIcon />
        <span>返回</span>
      </button>
      <div className="flex-1 text-center font-bold text-lg">工作区</div>
    </header>
  );
} 