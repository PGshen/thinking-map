import { useEffect, useRef, useCallback, useState } from 'react';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import { getToken } from '@/lib/auth';
import API_ENDPOINTS from '@/api/endpoints';

// SSE事件类型定义
type SSEEventType = 'nodeCreated' | 'nodeUpdated' | 'nodeDeleted' | 'nodeDependenciesUpdated' | 'connectionEstablished' | 'messageText' | 'messageConclusion' | 'messageThought' | 'messageNotice' | 'messageAction' | 'messagePlan' | 'error' | 'ping' | 'message';

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
  reconnectAttempts: number; // 重连尝试次数
  lastConnectedTime: number; // 最后连接成功时间
  connectionState: 'connecting' | 'connected' | 'disconnected' | 'error'; // 连接状态
}

// 全局连接管理器
class SSEConnectionManager {
  private connections = new Map<string, SSEConnection>();
  private callbackIdCounter = 0;
  private isPageVisible = true;
  private visibilityChangeHandler: (() => void) | null = null;
  private healthCheckInterval: NodeJS.Timeout | null = null;

  constructor() {
    this.initializeVisibilityListener();
    this.startHealthCheck();
  }

  // 启动连接健康检查
  private startHealthCheck() {
    // 每30秒检查一次连接健康状态
    this.healthCheckInterval = setInterval(() => {
      this.checkConnectionsHealth();
    }, 30000);
  }

  // 检查所有连接的健康状态
  private checkConnectionsHealth() {
    if (!this.isPageVisible) {
      return; // 页面不可见时不进行健康检查
    }

    this.connections.forEach((connection, mapID) => {
      const now = Date.now();
      const timeSinceLastConnection = now - connection.lastConnectedTime;
      
      // 如果连接超过5分钟没有活动且状态为错误，尝试重连
      if (connection.connectionState === 'error' && 
          timeSinceLastConnection > 300000 && // 5分钟
          connection.reconnectAttempts < 3) { // 限制健康检查重连次数
        
        console.log(`Health check: Attempting to reconnect stale connection for mapID: ${mapID}`);
        this.establishConnection(connection);
      }
    });
  }

  // 初始化页面可见性监听
  private initializeVisibilityListener() {
    if (typeof document !== 'undefined') {
      this.isPageVisible = !document.hidden;
      
      this.visibilityChangeHandler = () => {
        const wasVisible = this.isPageVisible;
        this.isPageVisible = !document.hidden;
        
        console.log(`Page visibility changed: ${this.isPageVisible ? 'visible' : 'hidden'}`);
        
        if (!wasVisible && this.isPageVisible) {
          // 页面从隐藏变为可见，检查并恢复连接
          this.handlePageVisible();
        } else if (wasVisible && !this.isPageVisible) {
          // 页面从可见变为隐藏，标记连接状态但不立即断开
          this.handlePageHidden();
        }
      };
      
      document.addEventListener('visibilitychange', this.visibilityChangeHandler);
    }
  }

  // 页面变为可见时的处理
  private handlePageVisible() {
    console.log('Page became visible, checking connections...');
    this.connections.forEach((connection, mapID) => {
      if (connection && !connection.isConnected && !connection.abortController.signal.aborted) {
        // 重置重连计数，给页面切换回来一个新的机会
        if (connection.connectionState === 'error') {
          connection.reconnectAttempts = Math.max(0, connection.reconnectAttempts - 2);
        }
        console.log(`Reconnecting SSE for mapID: ${mapID}`);
        this.establishConnection(connection);
      }
    });
  }

  // 页面变为隐藏时的处理
  private handlePageHidden() {
    console.log('Page became hidden, connections will be maintained');
    // 不立即断开连接，让浏览器自然处理
    // 只是标记页面状态，用于后续的重连判断
  }

  // 清理资源
  destroy() {
    if (this.visibilityChangeHandler && typeof document !== 'undefined') {
      document.removeEventListener('visibilitychange', this.visibilityChangeHandler);
    }
    if (this.healthCheckInterval) {
      clearInterval(this.healthCheckInterval);
    }
  }

  // 获取或创建连接
  getOrCreateConnection(mapID: string): { connection: SSEConnection; callbackId: string } {
    const callbackId = `callback_${++this.callbackIdCounter}`;
    
    let connection = this.connections.get(mapID);
    
    if (connection && !connection.abortController.signal.aborted) {
      // 复用现有连接
      connection.refCount++;
      // console.log(`SSE: Reusing existing connection for mapID ${mapID}, refCount: ${connection.refCount}`);
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
      reconnectAttempts: 0,
      lastConnectedTime: 0,
      connectionState: 'connecting',
    };
    
    this.connections.set(mapID, connection);
    // console.log(`SSE: Created new connection for mapID ${mapID}`);
    
    // 建立SSE连接
    this.establishConnection(connection);
    
    return { connection, callbackId };
  }
  
  // 注册回调
  registerCallbacks(mapID: string, callbackId: string, callbacks: SSECallbacks) {
    const connection = this.connections.get(mapID);
    if (connection) {
      connection.callbacks.set(callbackId, callbacks);
      // console.log(`SSE: Registered callbacks for mapID ${mapID}, callbackId ${callbackId}`);
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
      // console.log(`SSE: Removed callbacks for mapID ${mapID}, callbackId ${callbackId}, refCount: ${connection.refCount}`);
      
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
      connection.connectionState = 'disconnected';
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
    
    // 更新连接状态
    connection.connectionState = 'connecting';
    
    // 获取token
    const token = getToken();
    if (!token) {
      console.error('No authentication token found');
      connection.connectionState = 'error';
      return;
    }
    
    const url = API_ENDPOINTS.SSE.CONNECT.replace(':mapID', mapID);
    console.log('SSE connecting to:', url, `(attempt ${connection.reconnectAttempts + 1})`);
    
    // 保存this引用以在回调中使用
    const self = this;
    
    try {
      await fetchEventSource(url, {
        signal: abortController.signal,
        headers: {
          'Authorization': `Bearer ${token}`,
          'Accept': 'text/event-stream',
          'Cache-Control': 'no-cache',
        },
        // 添加连接保持配置
        openWhenHidden: true, // 页面隐藏时保持连接
        async onopen(response) {
          console.log('SSE onopen called:', { status: response.status, contentType: response.headers.get('content-type') });
          if (response.ok && response.headers.get('content-type')?.includes('text/event-stream')) {
            // 连接成功
            connection.isConnected = true;
            connection.connectionState = 'connected';
            connection.reconnectAttempts = 0; // 重置重连计数
            connection.lastConnectedTime = Date.now();
            
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
          let hasCallback = false;
          connection.callbacks.forEach((callbacks) => {
            const eventCallback = callbacks.eventCallbacks.get(eventData.event);
            if (eventCallback) {
              eventCallback(eventData);
              hasCallback = true;
            }
          });
          
          // 如果没有找到对应的callback，则把内容打印到控制台
          if (!hasCallback) {
            console.log('No callback found for event:', eventData.event, 'data:', eventData.data);
          }
        },
         onerror(error) {
           console.error('SSE connection error:', error);
           connection.isConnected = false;
           connection.connectionState = 'error';
           connection.reconnectAttempts++;
           
           // 广播错误事件到所有注册的回调
           connection.callbacks.forEach((callbacks) => {
             if (callbacks.onError) {
               callbacks.onError(error);
             }
           });
           
           // 智能重连逻辑：只在页面可见且连接未被主动取消时重连
           if (!abortController.signal.aborted && self.shouldReconnect(connection)) {
             const delay = self.getReconnectDelay(connection.reconnectAttempts);
             console.log(`SSE connection error, attempting to reconnect in ${delay}ms (attempt ${connection.reconnectAttempts})`);
             setTimeout(() => {
               if (!abortController.signal.aborted && self.shouldReconnect(connection)) {
                 self.establishConnection(connection);
               }
             }, delay);
           }
           
           throw error;
         },
      });
    } catch (error) {
       if (error instanceof Error && error.name !== 'AbortError') {
         console.error('fetchEventSource error:', error);
         connection.connectionState = 'error';
       }
     }
  }

  // 判断是否应该重连
  private shouldReconnect(connection: SSEConnection): boolean {
    // 页面可见时才重连，且重连次数不超过限制
    const maxReconnectAttempts = 10;
    const timeSinceLastConnection = Date.now() - connection.lastConnectedTime;
    
    // 如果最近刚连接过（30秒内），减少重连频率
    if (timeSinceLastConnection < 30000 && connection.reconnectAttempts > 3) {
      return false;
    }
    
    return this.isPageVisible && connection.reconnectAttempts < maxReconnectAttempts;
  }

  // 获取重连延迟时间（指数退避策略）
  private getReconnectDelay(attempts: number): number {
    // 使用指数退避策略，最大延迟30秒
    const baseDelay = this.isPageVisible ? 1000 : 5000;
    const maxDelay = 30000;
    
    // 前3次重连使用较短延迟，之后使用指数退避
    if (attempts <= 3) {
      return baseDelay;
    }
    
    const delay = Math.min(baseDelay * Math.pow(2, attempts - 3), maxDelay);
    return delay;
  }

  // 获取连接详细状态（新增方法）
  getConnectionDetails(mapID: string) {
    const connection = this.connections.get(mapID);
    if (!connection) {
      return null;
    }
    
    return {
      isConnected: connection.isConnected,
      connectionState: connection.connectionState,
      reconnectAttempts: connection.reconnectAttempts,
      lastConnectedTime: connection.lastConnectedTime,
      refCount: connection.refCount,
    };
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
    // console.log("eventCallbacks", eventCallbacks)
  }, [callbacks, onOpen, onError, mapID]);

  const connect = useCallback(() => {
    if (!mapID) {
      // console.log('SSE connect skipped: no mapID');
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
      // console.log('SSE callbacks removed');
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
      
      // console.log(`SSE: Registered dynamic event callbacks for mapID ${mapID}, callbackId ${callbackId}`);
      
      // 返回清理函数
      return () => {
        sseManager.removeCallbacks(mapID, callbackId);
        // console.log(`SSE: Removed dynamic event callbacks for mapID ${mapID}, callbackId ${callbackId}`);
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