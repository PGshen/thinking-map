import { useEffect, useRef, useCallback, useState } from 'react';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import { getToken } from '@/lib/auth';
import API_ENDPOINTS from '@/api/endpoints';

// SSE事件类型定义
type SSEEventType = 'nodeCreated' | 'nodeUpdated' | 'connection_established' | 'messageText' | 'messageThought' | 'messageAction' | 'messagePlan' | 'error' | 'ping' | 'message';

// SSE事件数据结构
interface SSEEventData {
  event: SSEEventType;
  data: string;
  id?: string;
  retry?: number;
}

// 事件回调函数类型
type SSEEventCallback = (eventData: SSEEventData) => void;

// 回调注册配置 - 每个事件类型只能有一个回调
interface SSECallbackConfig {
  eventType: SSEEventType;
  callback: SSEEventCallback;
}

interface SSEConnectionOptions {
  mapID: string;
  callbacks?: SSECallbackConfig[];
  onOpen?: () => void;
  onError?: (error: any) => void;
}

interface SSECallbacks {
  eventCallbacks: Map<SSEEventType, SSEEventCallback>;
  onOpen?: () => void;
  onError?: (error: any) => void;
}

interface SSEConnection {
  mapID: string;
  abortController: AbortController;
  isConnected: boolean;
  callbacks: Map<string, SSECallbacks>;
  refCount: number;
}

// 全局连接管理器
class SSEConnectionManager {
  private connections = new Map<string, SSEConnection>();
  private callbackIdCounter = 0;

  // 获取或创建连接
  getOrCreateConnection(mapID: string): { connection: SSEConnection; callbackId: string } {
    const callbackId = `callback_${++this.callbackIdCounter}`;
    
    let connection = this.connections.get(mapID);
    
    if (connection && !connection.abortController.signal.aborted) {
      // 复用现有连接
      connection.refCount++;
      console.log(`SSE: Reusing existing connection for mapID ${mapID}, refCount: ${connection.refCount}`);
      return { connection, callbackId };
    }
    
    // 创建新连接
    const abortController = new AbortController();
    connection = {
      mapID,
      abortController,
      isConnected: false,
      callbacks: new Map(),
      refCount: 1,
    };
    
    this.connections.set(mapID, connection);
    console.log(`SSE: Created new connection for mapID ${mapID}`);
    
    // 建立SSE连接
    this.establishConnection(connection);
    
    return { connection, callbackId };
  }
  
  // 注册回调
  registerCallbacks(mapID: string, callbackId: string, callbacks: SSECallbacks) {
    const connection = this.connections.get(mapID);
    if (connection) {
      connection.callbacks.set(callbackId, callbacks);
      console.log(`SSE: Registered callbacks for mapID ${mapID}, callbackId ${callbackId}`);
    }
  }
  
  // 更新回调
  updateCallbacks(mapID: string, callbackId: string, callbacks: SSECallbacks) {
    const connection = this.connections.get(mapID);
    if (connection) {
      connection.callbacks.set(callbackId, callbacks);
    }
  }

  // 注册事件回调
  registerEventCallbacks(mapID: string, callbackId: string, configs: SSECallbackConfig[]) {
    const connection = this.connections.get(mapID);
    if (connection) {
      const eventCallbacks = new Map<SSEEventType, SSEEventCallback>();
      
      configs.forEach(config => {
        // 每个事件类型只能有一个回调，如果重复注册会覆盖之前的
        eventCallbacks.set(config.eventType, config.callback);
      });
      
      const callbacks: SSECallbacks = {
        eventCallbacks,
        onOpen: undefined,
        onError: undefined
      };
      
      connection.callbacks.set(callbackId, callbacks);
    }
  }
  
  // 移除回调并可能断开连接
  removeCallbacks(mapID: string, callbackId: string) {
    const connection = this.connections.get(mapID);
    if (connection) {
      connection.callbacks.delete(callbackId);
      connection.refCount--;
      console.log(`SSE: Removed callbacks for mapID ${mapID}, callbackId ${callbackId}, refCount: ${connection.refCount}`);
      
      // 如果没有更多的引用，断开连接
      if (connection.refCount <= 0) {
        this.disconnectConnection(mapID);
      }
    }
  }
  
  // 断开特定连接
  disconnectConnection(mapID: string) {
    const connection = this.connections.get(mapID);
    if (connection) {
      connection.abortController.abort();
      this.connections.delete(mapID);
      console.log(`SSE: Disconnected connection for mapID ${mapID}`);
    }
  }
  
  // 获取连接状态
  getConnectionStatus(mapID: string): boolean {
    const connection = this.connections.get(mapID);
    return connection ? connection.isConnected : false;
  }
  
  // 检查连接是否存在
  hasConnection(mapID: string): boolean {
    const connection = this.connections.get(mapID);
    return connection ? !connection.abortController.signal.aborted : false;
  }
  
  // 建立SSE连接
  private async establishConnection(connection: SSEConnection) {
    const { mapID, abortController } = connection;
    
    // 获取token
    const token = getToken();
    if (!token) {
      console.error('No authentication token found');
      return;
    }
    
    const url = API_ENDPOINTS.SSE.CONNECT.replace(':mapID', mapID);
    console.log('SSE connecting to:', url);
    
    try {
      await fetchEventSource(url, {
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
            connection.isConnected = true;
            // 广播连接打开事件到所有注册的回调
            connection.callbacks.forEach((callbacks) => {
              if (callbacks.onOpen) {
                callbacks.onOpen();
              }
            });
            return;
          } else if (response.status >= 400 && response.status < 500 && response.status !== 429) {
            throw new Error(`Client error: ${response.status}`);
          } else {
            throw new Error(`Server error: ${response.status}`);
          }
        },
        onmessage(event) {
          const eventData: SSEEventData = {
            event: (event.event as SSEEventType) || 'message',
            data: event.data,
            id: event.id,
            retry: event.retry
          };
          
          // 根据事件类型触发相应的回调
          connection.callbacks.forEach((callbacks) => {
            const eventCallback = callbacks.eventCallbacks.get(eventData.event);
            if (eventCallback) {
              eventCallback(eventData);
            }
          });
        },
         onerror(error) {
           console.error('SSE connection error:', error);
           connection.isConnected = false;
           
           // 广播错误事件到所有注册的回调
           connection.callbacks.forEach((callbacks) => {
             if (callbacks.onError) {
               callbacks.onError(error);
             }
           });
           
           // 如果连接失败，尝试重连
           if (!abortController.signal.aborted) {
             console.log('SSE connection error, attempting to reconnect...');
             setTimeout(() => {
               if (!abortController.signal.aborted) {
                 sseManager.establishConnection(connection);
               }
             }, 3000);
           }
           
           throw error;
         },
      });
    } catch (error) {
       if (error instanceof Error && error.name !== 'AbortError') {
         console.error('fetchEventSource error:', error);
       }
     }
  }
}

// 全局SSE连接管理器实例
const sseManager = new SSEConnectionManager();

export function useSSEConnection({
  mapID,
  callbacks,
  onOpen,
  onError,
}: SSEConnectionOptions) {
  const [isConnected, setIsConnected] = useState(false);
  const callbackIdRef = useRef<string | null>(null);
  
  // 使用useRef存储回调函数，避免依赖变化导致连接重建
  const callbacksRef = useRef<SSECallbacks>({
    eventCallbacks: new Map(),
    onOpen,
    onError,
  });
  
  // 更新回调引用
  useEffect(() => {
    const eventCallbacks = new Map<SSEEventType, SSEEventCallback>();
    
    if (callbacks) {
      callbacks.forEach(config => {
        // 每个事件类型只能有一个回调，如果重复注册会覆盖之前的
        eventCallbacks.set(config.eventType, config.callback);
      });
    }
    
    callbacksRef.current = {
      eventCallbacks,
      onOpen,
      onError,
    };
    
    // 如果已经有连接，更新回调
    if (callbackIdRef.current) {
      sseManager.updateCallbacks(mapID, callbackIdRef.current, callbacksRef.current);
    }
  }, [callbacks, onOpen, onError, mapID]);

  const connect = useCallback(() => {
    if (!mapID) {
      console.log('SSE connect skipped: no mapID');
      return () => {};
    }

    // 获取或创建连接
    const { connection, callbackId } = sseManager.getOrCreateConnection(mapID);
    callbackIdRef.current = callbackId;
    
    // 注册回调
    sseManager.registerCallbacks(mapID, callbackId, callbacksRef.current);
    
    // 更新连接状态
    setIsConnected(connection.isConnected);
    
    // 监听连接状态变化
    const checkConnectionStatus = () => {
      const status = sseManager.getConnectionStatus(mapID);
      setIsConnected(status);
    };
    
    // 定期检查连接状态
    const statusInterval = setInterval(checkConnectionStatus, 1000);
    
    return () => {
      clearInterval(statusInterval);
    };
  }, [mapID]);

  // 动态注册事件回调
  const registerEventCallbacks = useCallback((configs: SSECallbackConfig[]) => {
    if (callbackIdRef.current) {
      sseManager.registerEventCallbacks(mapID, callbackIdRef.current, configs);
    }
  }, [mapID]);

  const disconnect = useCallback(() => {
    if (callbackIdRef.current) {
      sseManager.removeCallbacks(mapID, callbackIdRef.current);
      callbackIdRef.current = null;
      setIsConnected(false);
      console.log('SSE callbacks removed');
    }
  }, [mapID]);

  // 组件挂载时连接，卸载时断开
  useEffect(() => {
    const cleanup = connect();
    
    return () => {
      if (cleanup) cleanup();
      disconnect();
    };
  }, [connect, disconnect]);

  return {
    connect,
    disconnect,
    isConnected,
    registerEventCallbacks, // 新增：支持动态注册事件回调
  };
}

// 新增：独立的回调注册hook，支持随时为指定mapID注册事件回调
export function useSSECallbackRegistration() {
  const registerEventCallbacksForMap = useCallback((mapID: string, configs: SSECallbackConfig[]) => {
    if (sseManager.hasConnection(mapID)) {
      // 为现有连接注册新的事件回调
      const callbackId = `dynamic_callback_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
      sseManager.registerEventCallbacks(mapID, callbackId, configs);
      
      console.log(`SSE: Registered dynamic event callbacks for mapID ${mapID}, callbackId ${callbackId}`);
      
      // 返回清理函数
      return () => {
        sseManager.removeCallbacks(mapID, callbackId);
        console.log(`SSE: Removed dynamic event callbacks for mapID ${mapID}, callbackId ${callbackId}`);
      };
    } else {
      console.warn(`SSE: No active connection found for mapID ${mapID}`);
      return () => {}; // 返回空的清理函数
    }
  }, []);
  
  const getConnectionStatus = useCallback((mapID: string) => {
    return sseManager.getConnectionStatus(mapID);
  }, []);
  
  const disconnectMap = useCallback((mapID: string) => {
    sseManager.disconnectConnection(mapID);
  }, []);
  
  return {
    registerEventCallbacksForMap,
    getConnectionStatus,
    disconnectMap,
  };
}

// 导出连接管理器实例，供高级用法使用
export { sseManager };