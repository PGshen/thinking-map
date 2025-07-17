import { useEffect, useRef, useCallback, useState } from 'react';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import { NodeCreatedEvent, NodeUpdatedEvent, ThinkingProgressEvent, ErrorEvent } from '@/types/sse';
import { getToken } from '@/lib/auth';
import API_ENDPOINTS from '@/api/endpoints';

interface SSEConnectionOptions {
  mapID: string;
  onNodeCreated?: (event: NodeCreatedEvent) => void;
  onNodeUpdated?: (event: NodeUpdatedEvent) => void;
  onThinkingProgress?: (event: ThinkingProgressEvent) => void;
  onError?: (event: ErrorEvent) => void;
  onConnectionEstablished?: (data: any) => void;
}

export function useSSEConnection({
  mapID,
  onNodeCreated,
  onNodeUpdated,
  onThinkingProgress,
  onError,
  onConnectionEstablished,
}: SSEConnectionOptions) {
  const abortControllerRef = useRef<AbortController | null>(null);
  const isConnectedRef = useRef(false);
  // 使用useState管理连接状态，确保组件能响应状态变化
  const [isConnected, setIsConnected] = useState(false);
  
  // 使用useRef存储回调函数，避免依赖变化导致连接重建
  const callbacksRef = useRef({
    onNodeCreated,
    onNodeUpdated,
    onThinkingProgress,
    onError,
    onConnectionEstablished,
  });
  
  // 更新回调引用
  useEffect(() => {
    callbacksRef.current = {
      onNodeCreated,
      onNodeUpdated,
      onThinkingProgress,
      onError,
      onConnectionEstablished,
    };
  });

  const connect = useCallback(() => {
    if (abortControllerRef.current || !mapID) {
      console.log('SSE connect skipped:', { hasController: !!abortControllerRef.current, mapID });
      return;
    }

    // 获取token
    const token = getToken();
    if (!token) {
      console.error('No authentication token found');
      return;
    }

    // 创建AbortController用于控制连接
    const abortController = new AbortController();
    abortControllerRef.current = abortController;

    // 建立SSE连接，通过header传递token
    const url = API_ENDPOINTS.SSE.CONNECT.replace(':mapID', mapID);
    console.log('SSE connecting to:', url);
    
    fetchEventSource(url, {
      signal: abortController.signal,
      headers: {
        'Authorization': `Bearer ${token}`,
        'Accept': 'text/event-stream',
        'Cache-Control': 'no-cache',
      },
      async onopen(response) {
        console.log('SSE onopen called:', { status: response.status, contentType: response.headers.get('content-type') });
        if (response.ok && response.headers.get('content-type')?.includes('text/event-stream')) {
          console.log('SSE connection opened successfully');
          isConnectedRef.current = true;
          setIsConnected(true);
          console.log('SSE isConnected set to true');
          return; // everything's good
        } else if (response.status >= 400 && response.status < 500 && response.status !== 429) {
          // client-side errors are usually non-retriable:
          throw new Error(`Client error: ${response.status}`);
        } else {
          throw new Error(`Server error: ${response.status}`);
        }
      },
      onmessage(event) {
        // Handle different event types based on event.type
        // 使用callbacksRef.current访问最新的回调函数
        const callbacks = callbacksRef.current;
        
        switch (event.event) {
          case 'connection_established':
            console.log('SSE connection established:', event.data);
            isConnectedRef.current = true;
            setIsConnected(true);
            if (callbacks.onConnectionEstablished) {
              try {
                const data = JSON.parse(event.data);
                callbacks.onConnectionEstablished(data);
              } catch (error) {
                console.error('Failed to parse connection_established data:', error);
              }
            }
            break;
          
          case 'nodeCreated':
            console.log('Node created event:', event.data);
            if (callbacks.onNodeCreated) {
              try {
                const data: NodeCreatedEvent = JSON.parse(event.data);
                callbacks.onNodeCreated(data);
              } catch (error) {
                console.error('Failed to parse nodeCreated data:', error);
              }
            }
            break;
          
          case 'nodeUpdated':
            console.log('Node updated event:', event.data);
            if (callbacks.onNodeUpdated) {
              try {
                const data: NodeUpdatedEvent = JSON.parse(event.data);
                callbacks.onNodeUpdated(data);
              } catch (error) {
                console.error('Failed to parse nodeUpdated data:', error);
              }
            }
            break;
          
          case 'thinkingProgress':
            console.log('Thinking progress event:', event.data);
            if (callbacks.onThinkingProgress) {
              try {
                const data: ThinkingProgressEvent = JSON.parse(event.data);
                callbacks.onThinkingProgress(data);
              } catch (error) {
                console.error('Failed to parse thinkingProgress data:', error);
              }
            }
            break;
          
          case 'error':
            console.error('SSE error event:', event.data);
            if (callbacks.onError) {
              try {
                const data: ErrorEvent = JSON.parse(event.data);
                callbacks.onError(data);
              } catch (error) {
                console.error('Failed to parse error data:', error);
              }
            }
            break;
          
          case 'ping':
            console.log('SSE ping event:', event.data);
            break;
          
          default:
            console.log('Unknown SSE event type:', event.event, event.data);
        }
      },
      onerror(error) {
        console.error('SSE connection error:', error);
        isConnectedRef.current = false;
        setIsConnected(false);
        console.log('SSE isConnected set to false due to error');
        
        // 如果连接失败，尝试重连
        if (!abortController.signal.aborted) {
          console.log('SSE connection error, attempting to reconnect...');
          setTimeout(() => {
            disconnect();
            connect();
          }, 3000); // 3秒后重连
        }
        
        throw error; // rethrow to stop the operation
      },
    }).catch((error) => {
      if (error.name !== 'AbortError') {
        console.error('fetchEventSource error:', error);
      }
    });



  }, [mapID]); // 只依赖mapID，避免回调函数变化导致重连

  const disconnect = useCallback(() => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
      isConnectedRef.current = false;
      setIsConnected(false);
      console.log('SSE connection disconnected');
    }
  }, []);

  // 组件挂载时连接，卸载时断开
  useEffect(() => {
    connect();
    
    return () => {
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    connect,
    disconnect,
    isConnected, // 直接返回状态值，而不是函数
  };
}