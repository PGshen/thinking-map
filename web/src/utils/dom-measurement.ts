/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/utils/dom-measurement.ts
 */

// DOM测量工具函数

/**
 * 测量文本在指定样式下的实际尺寸
 */
export const measureTextSize = (
  text: string,
  styles: {
    fontSize?: string;
    fontWeight?: string;
    fontFamily?: string;
    lineHeight?: string;
    maxWidth?: number;
    className?: string;
  } = {}
): { width: number; height: number } => {
  if (!text || typeof window === 'undefined') {
    return { width: 0, height: 0 };
  }

  // 创建临时测量元素
  const measureElement = document.createElement('div');
  measureElement.style.position = 'absolute';
  measureElement.style.visibility = 'hidden';
  measureElement.style.pointerEvents = 'none';
  measureElement.style.top = '-9999px';
  measureElement.style.left = '-9999px';
  measureElement.style.whiteSpace = 'pre-wrap';
  measureElement.style.wordBreak = 'break-word';
  
  // 应用样式
  if (styles.fontSize) measureElement.style.fontSize = styles.fontSize;
  if (styles.fontWeight) measureElement.style.fontWeight = styles.fontWeight;
  if (styles.fontFamily) measureElement.style.fontFamily = styles.fontFamily;
  if (styles.lineHeight) measureElement.style.lineHeight = styles.lineHeight;
  if (styles.maxWidth) measureElement.style.maxWidth = `${styles.maxWidth}px`;
  if (styles.className) measureElement.className = styles.className;
  
  // 设置文本内容
  measureElement.textContent = text;
  
  // 添加到DOM进行测量
  document.body.appendChild(measureElement);
  
  const rect = measureElement.getBoundingClientRect();
  const width = rect.width;
  const height = rect.height;
  
  // 清理临时元素
  document.body.removeChild(measureElement);
  
  return { width, height };
};

/**
 * 测量Markdown内容的实际尺寸
 */
export const measureMarkdownSize = (
  content: string,
  styles: {
    fontSize?: string;
    fontWeight?: string;
    fontFamily?: string;
    lineHeight?: string;
    maxWidth?: number;
    className?: string;
  } = {}
): { width: number; height: number } => {
  if (!content || typeof window === 'undefined') {
    return { width: 0, height: 0 };
  }

  // 创建临时测量元素
  const measureElement = document.createElement('div');
  measureElement.style.position = 'absolute';
  measureElement.style.visibility = 'hidden';
  measureElement.style.pointerEvents = 'none';
  measureElement.style.top = '-9999px';
  measureElement.style.left = '-9999px';
  
  // 应用样式
  if (styles.fontSize) measureElement.style.fontSize = styles.fontSize;
  if (styles.fontWeight) measureElement.style.fontWeight = styles.fontWeight;
  if (styles.fontFamily) measureElement.style.fontFamily = styles.fontFamily;
  if (styles.lineHeight) measureElement.style.lineHeight = styles.lineHeight;
  if (styles.maxWidth) measureElement.style.maxWidth = `${styles.maxWidth}px`;
  if (styles.className) measureElement.className = styles.className;
  
  // 简单的Markdown渲染（基础支持）
  let htmlContent = content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>') // 粗体
    .replace(/\*(.*?)\*/g, '<em>$1</em>') // 斜体
    .replace(/`(.*?)`/g, '<code>$1</code>') // 行内代码
    .replace(/\n/g, '<br>'); // 换行
  
  measureElement.innerHTML = htmlContent;
  
  // 添加到DOM进行测量
  document.body.appendChild(measureElement);
  
  const rect = measureElement.getBoundingClientRect();
  const width = rect.width;
  const height = rect.height;
  
  // 清理临时元素
  document.body.removeChild(measureElement);
  
  return { width, height };
};

/**
 * 测量节点内容的总尺寸
 */
export const measureNodeContentSize = (
  question: string,
  target: string,
  conclusion: string = '',
  maxWidth: number = 360
): { width: number; height: number } => {
  if (typeof window === 'undefined') {
    return { width: 280, height: 120 }; // 服务端渲染时返回默认值
  }

  let totalHeight = 0;
  let maxContentWidth = 0;
  
  // 基础间距和边距
  const basePadding = 24; // px-3 py-2.5 的总和
  const headerHeight = 40; // 头部高度
  const contentSpacing = 8; // space-y-1 mb-2
  
  totalHeight += basePadding + headerHeight + contentSpacing;
  
  // 测量问题文本 (text-sm font-medium, line-clamp-2)
  if (question) {
    const questionSize = measureTextSize(question, {
      fontSize: '14px',
      fontWeight: '500',
      maxWidth: maxWidth - 24, // 减去padding
      className: 'line-clamp-2'
    });
    totalHeight += Math.min(questionSize.height, 40); // line-clamp-2 限制
    maxContentWidth = Math.max(maxContentWidth, questionSize.width);
  }
  
  // 测量目标文本 (text-xs, line-clamp-5)
  if (target) {
    const targetSize = measureMarkdownSize(target, {
      fontSize: '12px',
      maxWidth: maxWidth - 24,
      className: 'line-clamp-5'
    });
    totalHeight += Math.min(targetSize.height, 80); // line-clamp-5 限制
    maxContentWidth = Math.max(maxContentWidth, targetSize.width);
  }
  
  // 测量结论文本 (text-xs, line-clamp-5)
  if (conclusion) {
    const conclusionSize = measureMarkdownSize(conclusion, {
      fontSize: '12px',
      maxWidth: maxWidth - 24,
      className: 'line-clamp-5'
    });
    totalHeight += Math.min(conclusionSize.height, 80); // line-clamp-5 限制
    maxContentWidth = Math.max(maxContentWidth, conclusionSize.width);
  }
  
  // 计算最终尺寸
  const finalWidth = Math.max(280, Math.min(360, maxContentWidth + 24)); // 加上padding
  const finalHeight = Math.max(100, totalHeight);
  
  return {
    width: finalWidth,
    height: finalHeight
  };
};

/**
 * 缓存测量结果以提高性能
 */
const measurementCache = new Map<string, { width: number; height: number }>();

/**
 * 带缓存的节点尺寸测量
 */
export const measureNodeContentSizeWithCache = (
  question: string,
  target: string,
  conclusion?: string,
  maxWidth: number = 360
): { width: number; height: number } => {
  // 创建缓存键
  const cacheKey = `${question}|${target}|${conclusion || ''}|${maxWidth}`;
  
  // 检查缓存
  if (measurementCache.has(cacheKey)) {
    return measurementCache.get(cacheKey)!;
  }
  
  // 测量尺寸
  const size = measureNodeContentSize(question, target, conclusion ?? '', maxWidth);
  
  // 缓存结果
  measurementCache.set(cacheKey, size);
  
  // 限制缓存大小
  if (measurementCache.size > 100) {
    const firstKey = measurementCache.keys().next().value;
    if (firstKey) {
      measurementCache.delete(firstKey);
    }
  }
  
  return size;
};

/**
 * 清理测量缓存
 */
export const clearMeasurementCache = () => {
  measurementCache.clear();
};