import path from "node:path";
import { notFound } from "next/navigation";
import { MarkdownContent } from "@/components/ui/markdown-content";
import { promises as fs } from "node:fs";

interface DocsPageProps {
  params: { slug?: string[] };
}

export default async function DocsPage({ params }: DocsPageProps) {
  const docsRoot = path.resolve(process.cwd(), "../docs");
  const segments = Array.isArray(params.slug) ? params.slug : [];

  // 允许嵌套路径（如 blog/filename），强制 .md 扩展名
  let targetPath = path.join(docsRoot, ...segments);
  if (!targetPath.endsWith(".md")) {
    targetPath += ".md";
  }

  const resolved = path.resolve(targetPath);
  // 防止目录穿越
  if (!resolved.startsWith(docsRoot)) {
    notFound();
  }

  let content = "";
  try {
    content = await fs.readFile(resolved, "utf-8");
  } catch (e) {
    notFound();
  }

  return (
    <div className="px-6 py-6 max-w-4xl mx-auto">
      <MarkdownContent content={content} id={segments.join("/")} />
    </div>
  );
}