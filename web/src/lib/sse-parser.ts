import { EventSourceMessage, EventStreamContentType, fetchEventSource } from '@microsoft/fetch-event-source'
import { getToken } from "./auth"
import { toast } from 'sonner'
import { useGlobalStore } from '@/store/globalStore'

class RetriableError extends Error {}
class FatalError extends Error {}

/**
 * 路径匹配器的回调函数类型
 */
type PathMatcherCallback = (value: any, path: (string | number)[]) => void;

/**
 * 路径匹配模式类型
 */
type PathPattern = {
  tokens: (string | number)[];
  original: string;
  callback: PathMatcherCallback;
};

/**
 * 简化版 JSON 路径匹配系统
 */
class SimplePathMatcher {
  private patterns: PathPattern[] = []; // 存储所有注册的模式和回调

  /**
   * 注册一个路径模式和对应的回调函数
   * @param pattern - 路径模式，如 "users[*].name"
   * @param callback - 匹配时调用的回调函数
   */
  on(pattern: string, callback: PathMatcherCallback): this {
    // 解析路径模式为标记数组
    const parsedPattern = this.parsePath(pattern);
    this.patterns.push({
      tokens: parsedPattern,
      original: pattern,
      callback: callback
    });
    return this;
  }

  /**
   * 解析路径字符串为标记数组
   * @param path - 路径字符串
   * @return 解析后的标记数组
   */
  private parsePath(path: string): (string | number)[] {
    if (!path || path === '$') {
      return ['$'];
    }

    // 移除开头的 $ 和 . 符号
    if (path.startsWith('$')) {
      path = path.substring(1);
    }
    if (path.startsWith('.')) {
      path = path.substring(1);
    }

    // 分割路径
    const parts: (string | number)[] = [];
    let currentPart = '';
    let inBrackets = false;

    for (let i = 0; i < path.length; i++) {
      const char = path[i];
      
      if (char === '.' && !inBrackets) {
        if (currentPart) {
          parts.push(currentPart);
          currentPart = '';
        }
      } else if (char === '[') {
        if (currentPart) {
          parts.push(currentPart);
          currentPart = '';
        }
        inBrackets = true;
      } else if (char === ']') {
        if (currentPart === '*') {
          parts.push('*');
        } else if (!isNaN(Number(currentPart))) {
          parts.push(parseInt(currentPart, 10));
        }
        currentPart = '';
        inBrackets = false;
      } else {
        currentPart += char;
      }
    }

    if (currentPart) {
      parts.push(currentPart);
    }
    return parts;
  }

  /**
   * 检查当前路径是否匹配任何注册的模式
   * @param path - 当前路径
   * @param value - 当前节点的值
   */
  checkPatterns(path: (string | number)[], value: any): void {
    // 使用requestAnimationFrame来创建一个微小的延迟
    // requestAnimationFrame(() => {
      for (const pattern of this.patterns) {
        if (this.matchPath(path, pattern.tokens)) {
          // 如果匹配，调用回调函数
          pattern.callback(value, path);
        }
      }
    // });
  }

  /**
   * 检查路径是否匹配模式
   * @param path - 当前路径
   * @param pattern - 模式标记
   * @return 是否匹配
   */
  private matchPath(path: (string | number)[], pattern: (string | number)[]): boolean {
    // 如果模式比路径长，不可能匹配
    if (pattern.length > path.length) {
      return false;
    }

    // 从路径末尾开始匹配
    for (let i = 0; i < pattern.length; i++) {
      const patternPart = pattern[pattern.length - 1 - i];
      const pathPart = path[path.length - 1 - i];
      
      // 处理通配符
      if (patternPart === '*') {
        continue;
      }
      
      // 处理数组索引
      if (typeof patternPart === 'number' && typeof pathPart === 'number') {
        if (patternPart !== pathPart) {
          return false;
        }
        continue;
      }
      
      // 处理属性名
      if (patternPart !== pathPart) {
        return false;
      }
    }
    
    return true;
  }
}

/**
 * JSON 解析器的状态类型
 */
type ParserState = 
  | 'VALUE'
  | 'KEY_OR_END'
  | 'KEY'
  | 'COLON'
  | 'COMMA'
  | 'VALUE_OR_END'
  | 'NUMBER'
  | 'TRUE1'
  | 'TRUE2'
  | 'TRUE3'
  | 'FALSE1'
  | 'FALSE2'
  | 'FALSE3'
  | 'FALSE4'
  | 'NULL1'
  | 'NULL2'
  | 'NULL3';

/**
 * 真实的流式 JSON 解析器
 */
class StreamingJsonParser {
  private matcher: SimplePathMatcher;
  private realtime: boolean;
  private stack: any[] = [];
  private path: (string | number)[] = ['$'];
  private state: ParserState = 'VALUE';
  private buffer = '';
  private isEscaped = false;
  private isInString = false;
  private currentKey: string | null = null;
  private arrayIndexes: number[] = [];

  constructor(matcher: SimplePathMatcher, realtime: boolean) {
    this.matcher = matcher;
    this.realtime = realtime;
    this.reset();
  }

  /**
   * 重置解析器状态
   */
  reset(): void {
    this.stack = [];
    this.path = ['$'];
    this.state = 'VALUE';
    this.buffer = '';
    this.isEscaped = false;
    this.isInString = false;
    this.currentKey = null;
    this.arrayIndexes = [];
  }

  /**
   * 逐字符处理输入流
   * @param chunk - 输入的字符块
   */
  write(chunk: string): void {
    for (let i = 0; i < chunk.length; i++) {
      const char = chunk[i];
      if (char !== undefined) {
        this.processChar(char);
      }
    }
  }

  /**
   * 处理单个字符
   * @param char - 输入的字符
   */
  private processChar(char: string): void {
    // 处理字符串中的转义
    if (this.isInString) {
      if (this.isEscaped) {
        this.buffer += char;
        this.isEscaped = false;
        return;
      }
      
      if (char === '\\') {
        this.isEscaped = true;
        return;
      }
      
      if (char === '"') {
        this.isInString = false;
        
        if (this.state === 'KEY') {
          this.currentKey = this.buffer;
          this.buffer = '';
          this.state = 'COLON';
        } else if (this.state === 'VALUE') {
          this.addValue(this.buffer);
          this.buffer = '';
          this.state = 'COMMA';
        }
        
        return;
      }
      
      this.buffer += char;
      // 实时触发回调
      if (this.realtime && this.state === 'VALUE' && this.buffer !== '') {
        this.matcher.checkPatterns(this.path, this.buffer);
      }
      return;
    }

    // 处理非字符串状态
    switch (this.state) {
      case 'VALUE':
        if (char === '{') {
          // 开始对象
          const obj: Record<string, any> = {};
          this.addValue(obj);
          this.stack.push(obj);
          this.state = 'KEY_OR_END';
        } else if (char === '[') {
          // 开始数组
          const arr: any[] = [];
          this.addValue(arr);
          this.stack.push(arr);
          this.arrayIndexes.push(0);
          this.path.push(0);
          this.state = 'VALUE_OR_END';
        } else if (char === '"') {
          // 开始字符串
          this.isInString = true;
          this.buffer = '';
        } else if (char === 't') {
          // 可能是 true
          this.buffer = 't';
          this.state = 'TRUE1';
        } else if (char === 'f') {
          // 可能是 false
          this.buffer = 'f';
          this.state = 'FALSE1';
        } else if (char === 'n') {
          // 可能是 null
          this.buffer = 'n';
          this.state = 'NULL1';
        } else if (char === '-' || (char >= '0' && char <= '9')) {
          // 开始数字
          this.buffer = char;
          this.state = 'NUMBER';
        } else if (char !== ' ' && char !== '\t' && char !== '\n' && char !== '\r') {
          throw new Error(`Unexpected character in VALUE state: ${char}`);
        }
        break;
        
      case 'KEY_OR_END':
        if (char === '}') {
          // 结束对象
          this.endObject();
          this.state = 'COMMA';
        } else if (char === '"') {
          // 开始键名
          this.isInString = true;
          this.buffer = '';
          this.state = 'KEY';
        } else if (char !== ' ' && char !== '\t' && char !== '\n' && char !== '\r') {
          throw new Error(`Unexpected character in KEY_OR_END state: ${char}`);
        }
        break;
        
      case 'KEY':
        // 处理在 isInString 中
        if (char === '"') {
          // 开始字符串
          this.isInString = true;
          this.buffer = '';
        }
        break;
        
      case 'COLON':
        if (char === ':') {
          this.state = 'VALUE';
          // 更新路径
          if (Array.isArray(this.stack[this.stack.length - 1])) {
            const currentIndex = this.arrayIndexes[this.arrayIndexes.length - 1];
            if (currentIndex !== undefined) {
                this.path.push(currentIndex);
            }
          } else {
            this.path.push(this.currentKey!);
          }
        } else if (char !== ' ' && char !== '\t' && char !== '\n' && char !== '\r') {
          throw new Error(`Unexpected character in COLON state: ${char}`);
        }
        break;
        
      case 'COMMA':
        if (char === ',') {
          if (Array.isArray(this.stack[this.stack.length - 1])) {
            // 数组中的下一个元素
            const lastIndex = this.arrayIndexes[this.arrayIndexes.length - 1];
            if (lastIndex !== undefined) {
                this.arrayIndexes[this.arrayIndexes.length - 1] = lastIndex + 1;
                this.path[this.path.length - 1] = lastIndex + 1;
            }
            this.state = 'VALUE';
          } else {
            // 对象中的下一个键
            this.path.pop(); // 移除上一个键
            this.state = 'KEY';
          }
        } else if (char === '}') {
          // 结束对象
          this.endObject();
        } else if (char === ']') {
          // 结束数组
          this.endArray();
        } else if (char !== ' ' && char !== '\t' && char !== '\n' && char !== '\r') {
          throw new Error(`Unexpected character in COMMA state: ${char}`);
        }
        break;
        
      case 'VALUE_OR_END':
        if (char === ']') {
          // 空数组
          this.endArray();
          this.state = 'COMMA';
        } else {
          // 回到 VALUE 状态处理这个字符
          this.state = 'VALUE';
          this.processChar(char);
        }
        break;
        
      case 'NUMBER':
        if ((char >= '0' && char <= '9') || char === '.' || char === 'e' || char === 'E' || char === '+' || char === '-') {
          this.buffer += char;
        } else {
          // 数字结束
          this.addValue(parseFloat(this.buffer));
          this.buffer = '';
          this.state = 'COMMA';
          // 重新处理当前字符
          this.processChar(char);
        }
        break;
        
      case 'TRUE1':
        if (char === 'r') {
          this.buffer += char;
          this.state = 'TRUE2';
        } else {
          throw new Error(`Unexpected character in TRUE1 state: ${char}`);
        }
        break;
        
      case 'TRUE2':
        if (char === 'u') {
          this.buffer += char;
          this.state = 'TRUE3';
        } else {
          throw new Error(`Unexpected character in TRUE2 state: ${char}`);
        }
        break;
        
      case 'TRUE3':
        if (char === 'e') {
          this.addValue(true);
          this.buffer = '';
          this.state = 'COMMA';
        } else {
          throw new Error(`Unexpected character in TRUE3 state: ${char}`);
        }
        break;
        
      case 'FALSE1':
        if (char === 'a') {
          this.buffer += char;
          this.state = 'FALSE2';
        } else {
          throw new Error(`Unexpected character in FALSE1 state: ${char}`);
        }
        break;
        
      case 'FALSE2':
        if (char === 'l') {
          this.buffer += char;
          this.state = 'FALSE3';
        } else {
          throw new Error(`Unexpected character in FALSE2 state: ${char}`);
        }
        break;
        
      case 'FALSE3':
        if (char === 's') {
          this.buffer += char;
          this.state = 'FALSE4';
        } else {
          throw new Error(`Unexpected character in FALSE3 state: ${char}`);
        }
        break;
        
      case 'FALSE4':
        if (char === 'e') {
          this.addValue(false);
          this.buffer = '';
          this.state = 'COMMA';
        } else {
          throw new Error(`Unexpected character in FALSE4 state: ${char}`);
        }
        break;
        
      case 'NULL1':
        if (char === 'u') {
          this.buffer += char;
          this.state = 'NULL2';
        } else {
          throw new Error(`Unexpected character in NULL1 state: ${char}`);
        }
        break;
        
      case 'NULL2':
        if (char === 'l') {
          this.buffer += char;
          this.state = 'NULL3';
        } else {
          throw new Error(`Unexpected character in NULL2 state: ${char}`);
        }
        break;
        
      case 'NULL3':
        if (char === 'l') {
          this.addValue(null);
          this.buffer = '';
          this.state = 'COMMA';
        } else {
          throw new Error(`Unexpected character in NULL3 state: ${char}`);
        }
        break;
    }
  }

  /**
   * 添加值到当前容器
   * @param value - 要添加的值
   */
  private addValue(value: any): void {
    if (this.stack.length === 0) {
      // 根值
      this.stack.push(value);
      this.matcher.checkPatterns(this.path, value);
      return;
    }
    
    const parent = this.stack[this.stack.length - 1];
    
    if (Array.isArray(parent)) {
      // 添加到数组
      const index = this.arrayIndexes[this.arrayIndexes.length - 1];
      if (typeof index === 'number') {
        parent[index] = value;
      }
    } else {
      // 添加到对象
      if (this.currentKey) {
        (parent as Record<string, any>)[this.currentKey] = value;
      }
    }
    
    // 检查路径匹配
    this.matcher.checkPatterns(this.path, value);
  }

  /**
   * 结束对象处理
   */
  private endObject(): void {
    this.stack.pop();
    this.path.pop();
    this.state = 'COMMA';
  }

  /**
   * 结束数组处理
   */
  private endArray(): void {
    this.stack.pop();
    this.arrayIndexes.pop();
    this.path.pop();
    this.state = 'COMMA';
  }

  /**
   * 结束解析
   */
  end(): void {
    if (this.stack.length !== 1) {
      throw new Error('Unexpected end of input: JSON structure is incomplete');
    }
    console.log('JSON parsing complete:', JSON.stringify(this.stack[0], null, 2));
  }
}

/**
 * SSE 连接配置类型
 */
interface SseConfig {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  url: string;
  param: Record<string, unknown>;
  message?: (ev: EventSourceMessage) => void;
  close?: () => void;
  error?: (err: unknown) => void;
}

/**
 * 使用 SSE 的流式 JSON 解析器
 */
export class SseJsonStreamParser {
  private matcher: SimplePathMatcher;
  private parser: StreamingJsonParser;
  private abortController: AbortController | null = null;

  constructor(realtime: boolean) {
    this.matcher = new SimplePathMatcher();
    this.parser = new StreamingJsonParser(this.matcher, realtime);
  }

  /**
   * 注册路径模式
   * @param pattern - 路径模式
   * @param callback - 回调函数
   */
  on(pattern: string, callback: PathMatcherCallback): this {
    this.matcher.on(pattern, callback);
    return this;
  }

  /**
   * 连接到 SSE 端点
   */
  connect(config: SseConfig): this;
  connect(
    method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    url: string,
    param: Record<string, unknown>,
    message?: (ev: EventSourceMessage) => void,
    close?: () => void,
    error?: (err: unknown) => void,
  ): this;
  connect(
    methodOrConfig: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | SseConfig,
    url?: string,
    param?: Record<string, unknown>,
    message?: (ev: EventSourceMessage) => void,
    close?: () => void,
    error?: (err: unknown) => void,
  ): this {
    // 重置解析器
    this.parser.reset();
    
    let config: SseConfig;
    
    if (typeof methodOrConfig === 'object') {
      // 使用配置对象
      config = methodOrConfig;
    } else {
      // 使用参数形式
      config = {
        method: methodOrConfig,
        url: url!,
        param: param!,
        message,
        close,
        error,
      };
    }

    const token = getToken();
    this.abortController = new AbortController();
    let retryCount = 0;
    const maxRetries = 3;
    const parser = this.parser;

    // 处理 GET 请求的参数
    const finalUrl = config.method === 'GET' && config.param
      ? `${config.url}${config.url.includes('?') ? '&' : '?'}${new URLSearchParams(config.param as Record<string, string>)}`
      : config.url;

    fetchEventSource(finalUrl, {
      method: config.method || 'POST',
      headers: {
        'Accept': 'text/event-stream;charset=UTF-8',
        'Cache-Control': 'no-cache',
        Authorization: token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json',
        'Sec-Fetch-Mode': 'cors',
        'Sec-Fetch-Site': 'same-site',
        'Sec-Fetch-Dest': 'empty',
      },
      // 只在非 GET 请求时添加 body
      ...(config.method !== 'GET' && config.param ? { body: JSON.stringify(config.param) } : {}),
      signal: this.abortController.signal,
      onmessage(ev) {
        // 空格替换
        let data = ev.data;
        data = data.replace(/&nbsp;/g, ' ');
        if(ev.event == 'json') {  // 只解析json
          parser.write(data);
        }
        if (config.message) config.message(ev);
      },
      async onopen(response) {
        if (response.status === 401) {
          const { setError } = useGlobalStore.getState();
          setError('登录已过期，请重新登录');
          toast.error('登录已过期，请重新登录');
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
          throw new FatalError('Unauthorized');
        }
        const ct = response.headers.get('content-type');
        if (!response.ok || !ct || ct.indexOf('text/event-stream') !== 0) {
          throw new RetriableError('SSE connection failed');
        }
      },
      onclose() {
        parser.end();
        if (config.close) config.close();
      },
      onerror(err) {
        if (err instanceof FatalError) {
          throw err;
        }
        retryCount++;
        if (retryCount > maxRetries) {
          const { setError } = useGlobalStore.getState();
          setError('服务异常，请稍后再试');
          toast.error('服务异常，请稍后再试');
          throw new FatalError('Max retries exceeded');
        }
        if (config.error) config.error(err);
      }
    });
    
    return this;
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    if (this.abortController) {
      this.abortController.abort();
      this.abortController = null;
    }
  }
}

/**
 * Markdown 流解析器
 */
export class SseMarkdownStreamParser {
  private abortController: AbortController | null = null;

  /**
   * 连接到 SSE 端点
   */
  connect(config: SseConfig): this;
  connect(
    method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    url: string,
    param: Record<string, unknown>,
    message?: (ev: EventSourceMessage) => void,
    close?: () => void,
    error?: (err: unknown) => void,
  ): this;
  connect(
    methodOrConfig: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | SseConfig,
    url?: string,
    param?: Record<string, unknown>,
    message?: (ev: EventSourceMessage) => void,
    close?: () => void,
    error?: (err: unknown) => void,
  ): this {
    let config: SseConfig;
    
    if (typeof methodOrConfig === 'object') {
      // 使用配置对象
      config = methodOrConfig;
    } else {
      // 使用参数形式
      config = {
        method: methodOrConfig,
        url: url!,
        param: param!,
        message,
        close,
        error,
      };
    }

    const token = getToken();
    this.abortController = new AbortController();
    let retryCount = 0;
    const maxRetries = 3;

    // 处理 GET 请求的参数
    const finalUrl = config.method === 'GET' && config.param
      ? `${config.url}${config.url.includes('?') ? '&' : '?'}${new URLSearchParams(config.param as Record<string, string>)}`
      : config.url;

    fetchEventSource(finalUrl, {
      method: config.method || 'POST',
      headers: {
        'Accept': 'text/event-stream;charset=UTF-8',
        'Cache-Control': 'no-cache',
        Authorization: token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json',
        'Sec-Fetch-Mode': 'cors',
        'Sec-Fetch-Site': 'same-site',
        'Sec-Fetch-Dest': 'empty',
      },
      // 只在非 GET 请求时添加 body
      ...(config.method !== 'GET' && config.param ? { body: JSON.stringify(config.param) } : {}),
      signal: this.abortController.signal,
      onmessage(ev) {
        // 确保空字符串和换行符的正确处理
        let data = ev.data;
        data = data.replace(/&nbsp;/g, ' ');
        if (config.message) config.message(ev);
      },
      async onopen(response) {
        if (response.status === 401) {
          const { setError } = useGlobalStore.getState();
          setError('登录已过期，请重新登录');
          toast.error('登录已过期，请重新登录');
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
          throw new FatalError('Unauthorized');
        }
        const ct = response.headers.get('content-type');
        if (!response.ok || ct !== EventStreamContentType) {
          throw new RetriableError('SSE connection failed');
        }
      },
      onclose() {
        if (config.close) config.close();
      },
      onerror(err) {
        if (err instanceof FatalError) {
          throw err;
        }
        retryCount++;
        if (retryCount > maxRetries) {
          const { setError } = useGlobalStore.getState();
          setError('服务异常，请稍后再试');
          toast.error('服务异常，请稍后再试');
          throw new FatalError('Max retries exceeded');
        }
        if (config.error) config.error(err);
      }
    });
    
    return this;
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    if (this.abortController) {
      this.abortController.abort();
      this.abortController = null;
    }
  }
}

/**
 * 纯文本流解析器配置接口
 */
interface SseTextConfig extends Omit<SseConfig, 'message'> {
  onText?: (text: string, fullText: string) => void; // 接收到新文本片段时的回调
  onComplete?: (fullText: string) => void; // 文本接收完成时的回调
  message?: (ev: EventSourceMessage) => void; // 原始消息回调
}

/**
 * 纯文本流解析器
 */
export class SseTextStreamParser {
  private abortController: AbortController | null = null;
  private fullText: string = '';

  /**
   * 连接到 SSE 端点
   */
  connect(config: SseTextConfig): this;
  connect(
    method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE',
    url: string,
    param: Record<string, unknown>,
    onText?: (text: string, fullText: string) => void,
    onComplete?: (fullText: string) => void,
    close?: () => void,
    error?: (err: unknown) => void,
  ): this;
  connect(
    methodOrConfig: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | SseTextConfig,
    url?: string,
    param?: Record<string, unknown>,
    onText?: (text: string, fullText: string) => void,
    onComplete?: (fullText: string) => void,
    close?: () => void,
    error?: (err: unknown) => void,
  ): this {
    let config: SseTextConfig;
    
    if (typeof methodOrConfig === 'object') {
      // 使用配置对象
      config = methodOrConfig;
    } else {
      // 使用参数形式
      config = {
        method: methodOrConfig,
        url: url!,
        param: param!,
        onText,
        onComplete,
        close,
        error,
      };
    }

    const token = getToken();
    this.abortController = new AbortController();
    this.fullText = '';
    let retryCount = 0;
    const maxRetries = 3;

    // 处理 GET 请求的参数
    const finalUrl = config.method === 'GET' && config.param
      ? `${config.url}${config.url.includes('?') ? '&' : '?'}${new URLSearchParams(config.param as Record<string, string>)}`
      : config.url;

    fetchEventSource(finalUrl, {
      method: config.method || 'POST',
      headers: {
        'Accept': 'text/event-stream;charset=UTF-8',
        'Cache-Control': 'no-cache',
        Authorization: token ? `Bearer ${token}` : '',
        'Content-Type': 'application/json',
        'Sec-Fetch-Mode': 'cors',
        'Sec-Fetch-Site': 'same-site',
        'Sec-Fetch-Dest': 'empty',
      },
      // 只在非 GET 请求时添加 body
      ...(config.method !== 'GET' && config.param ? { body: JSON.stringify(config.param) } : {}),
      signal: this.abortController.signal,
      onmessage: (ev) => {
        // 处理接收到的文本数据
        let data = ev.data;
        
        // 处理特殊字符
        data = data.replace(/&nbsp;/g, ' ');
        
        // 累积文本
        this.fullText += data;
        
        // 调用文本回调
        if (config.onText) {
          config.onText(data, this.fullText);
        }
        
        // 调用原始消息回调
        if (config.message) {
          config.message(ev);
        }
      },
      async onopen(response) {
        if (response.status === 401) {
          const { setError } = useGlobalStore.getState();
          setError('登录已过期，请重新登录');
          toast.error('登录已过期，请重新登录');
          if (typeof window !== 'undefined') {
            window.location.href = '/login';
          }
          throw new FatalError('Unauthorized');
        }
        const ct = response.headers.get('content-type');
        if (!response.ok || ct !== EventStreamContentType) {
          throw new RetriableError('SSE connection failed');
        }
      },
      onclose: () => {
        // 连接关闭时调用完成回调
        if (config.onComplete) {
          config.onComplete(this.fullText);
        }
        if (config.close) {
          config.close();
        }
      },
      onerror: (err) => {
        if (err instanceof FatalError) {
          throw err;
        }
        retryCount++;
        if (retryCount > maxRetries) {
          const { setError } = useGlobalStore.getState();
          setError('服务异常，请稍后再试');
          toast.error('服务异常，请稍后再试');
          throw new FatalError('Max retries exceeded');
        }
        if (config.error) {
          config.error(err);
        }
      }
    });
    
    return this;
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    if (this.abortController) {
      this.abortController.abort();
      this.abortController = null;
    }
  }

  /**
   * 获取当前累积的完整文本
   */
  getFullText(): string {
    return this.fullText;
  }

  /**
   * 清空累积的文本
   */
  clearText(): void {
    this.fullText = '';
  }
}

// 使用示例
/**
 * 演示如何使用纯文本流解析器
 */
function demoTextStreamParser(): void {
  const textParser = new SseTextStreamParser();
  
  // 方式1：使用配置对象
  textParser.connect({
    method: 'POST',
    url: '/api/stream/text',
    param: { prompt: 'Hello, world!' },
    onText: (newText, fullText) => {
      console.log('新文本片段:', newText);
      console.log('完整文本:', fullText);
    },
    onComplete: (fullText) => {
      console.log('文本接收完成:', fullText);
    },
    close: () => {
      console.log('连接已关闭');
    },
    error: (err) => {
      console.error('连接错误:', err);
    }
  });
  
  // 方式2：使用参数形式
  // textParser.connect(
  //   'POST',
  //   '/api/stream/text',
  //   { prompt: 'Hello, world!' },
  //   (newText, fullText) => console.log('新文本:', newText),
  //   (fullText) => console.log('完成:', fullText),
  //   () => console.log('关闭'),
  //   (err) => console.error('错误:', err)
  // );
  
  // 获取当前累积的文本
  setTimeout(() => {
    console.log('当前文本:', textParser.getFullText());
  }, 5000);
  
  // 10秒后断开连接
  setTimeout(() => {
    textParser.disconnect();
  }, 10000);
}

function demoWithSSE(): void {
  const jsonStream = new SseJsonStreamParser(true);
  
  // 注册路径模式
  jsonStream.on('users[*].name', (value: any, path: (string | number)[]) => {
    console.log('Found user name:', value, 'at path:', path.join('.'));
    // 可以在这里更新 UI
  });
  
  jsonStream.on('ips[*]', (value: any, path: (string | number)[]) => {
    console.log('Found user email:', value, 'at path:', path.join('.'));
  });
  
  jsonStream.on('metadata.count', (value: any, path: (string | number)[]) => {
    console.log('Found metadata count:', value, 'at path:', path.join('.'));
  });
  
  // 连接到 SSE 端点
  jsonStream.connect('GET', '/api/debug', {
    'message': '{"ips":["127.0.0.1","192.168.0.1","10.10.0.1"]}'
  });
}

// 模拟 SSE 流的测试函数
function testWithMockStream(): void {
  const matcher = new SimplePathMatcher();
  const parser = new StreamingJsonParser(matcher, true);
  
  // 注册路径模式
  matcher.on('users[*].name', (value: any, path: (string | number)[]) => {
    console.log('Found user name:', value, 'at path:', path.join('.'));
  });
  
  matcher.on('users[*].email', (value: any, path: (string | number)[]) => {
    console.log('Found user email:', value, 'at path:', path.join('.'));
  });
  
  matcher.on('metadata.count', (value: any, path: (string | number)[]) => {
    console.log('Found metadata count:', value, 'at path:', path.join('.'));
  });
  
  // 模拟流式数据
  const jsonChunks = [
    '{"users": [',
    '{"name": "John Doe", ',
    '"email": "john@example.com", ',
    '"age": 30},',
    '{"name": "Jane Smith", ',
    '"email": "jane@example.com", ',
    '"age": 25}',
    '],',
    '"metadata": {"count": 2, "page": 1}',
    '}'
  ];
  
  // 逐块处理
  jsonChunks.forEach(chunk => {
    parser.write(chunk);
  });
  
  // 结束解析
  parser.end();
}

// 运行测试
// testWithMockStream();
// demoWithSSE();