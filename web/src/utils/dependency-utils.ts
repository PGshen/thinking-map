import { DependentContext } from '@/types/node';

/**
 * 获取未满足依赖的提示信息
 * @param context 依赖上下文
 * @returns 提示信息
 */
export function getUnmetDependenciesMessage(context: DependentContext): string {
  const unmetDependencies = {
    ancestor: context.ancestor?.filter(item => item.status !== 'completed') || [],
    prevSibling: context.prevSibling?.filter(item => item.status !== 'completed') || [],
    children: context.children?.filter(item => item.status !== 'completed') || []
  };

  const totalUnmet = unmetDependencies.ancestor.length + 
                     unmetDependencies.prevSibling.length + 
                     unmetDependencies.children.length;

  if (totalUnmet === 0) {
    return '所有依赖节点已完成，可以开始执行';
  }

  const messages = [];
  if (unmetDependencies.ancestor.length > 0) {
    messages.push(`${unmetDependencies.ancestor.length} 个祖先节点`);
  }
  if (unmetDependencies.prevSibling.length > 0) {
    messages.push(`${unmetDependencies.prevSibling.length} 个同级节点`);
  }
  if (unmetDependencies.children.length > 0) {
    messages.push(`${unmetDependencies.children.length} 个子节点`);
  }

  return `有 ${messages.join('、')} 未完成，无法开始执行`;
}

/**
 * 检查是否所有依赖都已满足
 * @param context 依赖上下文
 * @returns 是否所有依赖都已满足
 */
export function checkAllDependenciesMet(context: DependentContext): boolean {
  if (!context) return true;
  
  const { ancestor, prevSibling, children } = context;
  const hasUnmetDependencies = [
    ...(ancestor || []),
    ...(prevSibling || []),
    ...(children || [])
  ].some(item => item.status !== 'completed');
  
  return !hasUnmetDependencies;
}