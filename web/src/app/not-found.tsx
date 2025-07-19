import Link from 'next/link';

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-background">
      <div className="text-center">
        <h1 className="text-6xl font-bold text-gray-900 dark:text-gray-100">404</h1>
        <h2 className="mt-4 text-2xl font-semibold text-gray-700 dark:text-gray-300">
          页面未找到
        </h2>
        <p className="mt-2 text-gray-600 dark:text-gray-400">
          抱歉，您访问的页面不存在。
        </p>
        <div className="mt-6">
          <Link
            href="/"
            className="inline-flex items-center px-4 py-2 text-sm font-medium text-white bg-black border border-transparent rounded-md shadow-sm hover:bg-black focus:outline-none focus:ring-2 focus:ring-offset-2"
          >
            返回首页
          </Link>
        </div>
      </div>
    </div>
  );
}