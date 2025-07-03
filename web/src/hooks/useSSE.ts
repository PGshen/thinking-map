import { useEffect, useRef, useState, useCallback } from 'react';
import type { SSEEvent, NodeCreatedEvent, NodeUpdatedEvent, ThinkingProgressEvent, ErrorEvent } from '@/types/sse';

export type SSEStatus = 'connecting' | 'open' | 'closed' | 'error';

interface UseSSEOptions {
  url: string;
  onEvent?: (event: SSEEvent) => void;
  onOpen?: () => void;
  onError?: (e: Event) => void;
  withCredentials?: boolean;
  retryInterval?: number; // ms
}

export function useSSE({ url, onEvent, onOpen, onError, withCredentials = true, retryInterval = 3000 }: UseSSEOptions) {
  const [status, setStatus] = useState<SSEStatus>('connecting');
  const eventSourceRef = useRef<EventSource | null>(null);
  const retryTimeout = useRef<NodeJS.Timeout | null>(null);

  const connect = useCallback(() => {
    setStatus('connecting');
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }
    const es = new window.EventSource(url, { withCredentials });
    eventSourceRef.current = es;

    es.onopen = () => {
      setStatus('open');
      onOpen?.();
    };
    es.onerror = (e) => {
      setStatus('error');
      onError?.(e);
      es.close();
      // 自动重连
      retryTimeout.current = setTimeout(connect, retryInterval);
    };
    es.onmessage = (e) => {
      try {
        const parsed = JSON.parse(e.data);
        onEvent?.(parsed);
      } catch (err) {
        // ignore
      }
    };
  }, [url, onEvent, onOpen, onError, withCredentials, retryInterval]);

  useEffect(() => {
    connect();
    return () => {
      if (eventSourceRef.current) eventSourceRef.current.close();
      if (retryTimeout.current) clearTimeout(retryTimeout.current);
    };
  }, [connect]);

  return { status };
} 