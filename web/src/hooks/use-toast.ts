/*
 * @Date: 2025-01-27
 * @LastEditors: AI Assistant
 * @LastEditTime: 2025-01-27
 * @FilePath: /thinking-map/web/src/hooks/use-toast.ts
 */
import { toast as sonnerToast } from 'sonner';

interface ToastProps {
  title?: string;
  description?: string;
  variant?: 'default' | 'info' | 'destructive';
  duration?: number;
}

export function useToast() {
  const toast = ({ title, description, variant = 'default', duration }: ToastProps) => {
    const message = title || description || '';
    const fullMessage = title && description ? `${title}: ${description}` : message;
    
    if (variant === 'destructive') {
      sonnerToast.error(fullMessage, {
        duration: duration || 5000,
      });
    } else if (variant === 'info') {
      sonnerToast.info(fullMessage, {
        duration: duration || 3000,
      })
    } else {
      sonnerToast.success(fullMessage, {
        duration: duration || 3000,
      });
    }
  };

  return { toast };
}

export { useToast as default };