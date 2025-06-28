# ThinkingMap 交互时序图

## 1. 问题理解阶段时序图

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant B as 后端
    participant IA as Intent Agent
    participant AA as Analysis Agent

    Note over U,AA: 问题输入与初步分析
    U->>F: 输入问题文本和类型
    F->>F: 验证问题格式
    F->>B: POST /api/v1/thinking/analyze
    Note right of F: {question, question_type, user_id}
    
    B->>IA: 调用用户意图识别
    IA-->>B: 返回意图分析结果
    B->>AA: 调用问题初步分析
    AA-->>B: 返回问题理解结果
    
    B-->>F: 返回问题理解
    Note right of B: {understanding, suggestions, clarification_questions}
    F-->>U: 显示问题理解结果

    Note over U,AA: 问题澄清与确认
    U->>F: 回答澄清问题/提供补充信息
    F->>B: POST /api/v1/thinking/clarify
    Note right of F: {session_id, clarifications}
    
    B->>AA: 更新问题理解
    AA-->>B: 返回更新后的理解
    B-->>F: 返回更新结果
    F-->>U: 显示更新后的理解
    
    Note over U,AA: 重复澄清过程直到用户确认
    
    U->>F: 确认最终问题理解
    F->>B: POST /api/v1/thinking/confirm
    Note right of F: {session_id, final_understanding}
    
    B->>B: 创建thinking_map和根节点
    B-->>F: 返回map_id和根节点信息
    F-->>U: 进入问题解答阶段
```

## 2. 问题解答阶段时序图

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant B as 后端
    participant SSE as SSE服务
    participant DA as Decompose Agent
    participant RA as Reasoning Agent
    participant SA as Synthesis Agent

    Note over U,SA: 工作区初始化
    F->>B: GET /api/v1/sse/connect\n{map_id, user_id}
    B->>SSE: 建立SSE连接
    SSE-->>F: 发送连接建立事件
    F->>B: GET /api/v1/thinking/maps/{map_id}\nGET /api/v1/thinking/nodes?map_id={map_id}
    B-->>F: 返回思考图和所有节点数据
    F-->>U: 显示工作区界面

    Note over U,SA: 节点执行流程
    F->>B: GET /api/v1/nodes/{node_id}
    B-->>F: 返回节点详细信息
    F-->>U: 显示节点信息面板

    U->>F: 点击"开始执行"
    F->>B: POST /api/v1/thinking/execute\n{node_id, action: "start"}
    B->>B: 判断执行动作
    B-->>F: 返回执行决策\n{action: "decompose|conclude", reason, next_tab}

    alt 选择拆解
        Note over U,SA: 问题拆解流程
        F-->>U: 显示拆解确认对话框
        U->>F: 确认拆解
        F->>B: POST /api/v1/thinking/decompose
        B->>DA: 调用问题拆解服务\n{node_id, question, context}
        DA-->>B: 返回拆解结果\n{sub_problems, strategy}
        B->>B: 创建子节点
        B->>SSE: 发送节点创建事件
        SSE-->>F: event: node_created
        F-->>U: 实时更新可视化工作区

        U->>F: 调整拆解结果（可选）
        F->>B: POST /api/v1/thinking/adjust-decomposition
        B-->>F: 确认调整结果

    else 选择直接解答
        Note over U,SA: 问题解答流程
        F-->>U: 切换到结论生成Tab

        B->>RA: 调用信息检索
        B->>SSE: 发送思考进度事件
        SSE-->>F: event: thinking_progress
        F-->>U: 显示思考进度

        B->>RA: 调用逻辑推理
        B->>SSE: 发送推理进度
        SSE-->>F: event: thinking_progress
        F-->>U: 更新进度显示

        B->>SA: 调用结论生成
        SA-->>B: 返回初步结论
        B-->>F: 返回结论结果
        F-->>U: 显示生成的结论

        U->>F: 提供反馈或确认
        F->>B: POST /api/v1/thinking/feedback\n{node_id, feedback, action}

        alt 需要调整
            B->>SA: 根据反馈调整结论
            SA-->>B: 返回调整后的结论
            B-->>F: 返回新结论
            F-->>U: 显示调整后的结论
        else 确认结论
            B->>B: 标记节点为完成
            B->>SSE: 发送节点完成事件
            SSE-->>F: event: node_updated
            F-->>U: 更新节点状态
        end
    end

    Note over U,SA: 下一个节点选择
    B->>B: 检查依赖关系
    B->>B: 选择下一个执行节点
    B->>SSE: 发送节点切换事件
    SSE-->>F: event: node_activated
    F-->>U: 高亮新节点并展开信息面板
```

## 3. 异常处理时序图

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant B as 后端
    participant SSE as SSE服务
    participant A as Agent服务

    Note over U,A: 网络异常处理
    F->>SSE: 检测连接状态
    alt SSE连接断开
        F->>F: 启动自动重连
        F->>SSE: 重新建立连接
        SSE-->>F: 连接恢复
        F->>B: 同步当前状态
        B-->>F: 返回状态信息
        F-->>U: 显示连接恢复提示
    end

    Note over U,A: API请求失败
    F->>B: API请求
    alt 请求失败
        B-->>F: 返回错误响应
        F->>F: 保存本地状态
        F-->>U: 显示错误提示和重试选项
        U->>F: 选择重试
        F->>B: 重新发送请求
        B-->>F: 成功响应
        F-->>U: 显示成功结果
    end

    Note over U,A: Agent服务异常
    B->>A: 调用Agent服务
    alt 服务不可用
        A-->>B: 服务错误响应
        B-->>F: 返回服务状态提示
        F-->>U: 显示服务不可用提示
        F->>F: 保存当前进度
        U->>F: 选择重试或等待
        F->>B: 重试请求
        B->>A: 重新调用服务
        A-->>B: 服务恢复响应
        B-->>F: 返回正常结果
        F-->>U: 继续执行流程
    end

    Note over U,A: 执行超时处理
    B->>A: 长时间执行请求
    alt 执行超时
        A-->>B: 超时响应
        B-->>F: 返回超时提示
        F-->>U: 显示超时选项
        U->>F: 选择继续等待或取消
        alt 继续等待
            F->>B: 继续等待请求
            B->>A: 继续执行
            A-->>B: 最终完成
            B-->>F: 返回完成结果
            F-->>U: 显示完成结果
        else 取消执行
            F->>B: 取消执行请求
            B->>B: 记录断点信息
            B-->>F: 确认取消
            F-->>U: 显示取消确认
        end
    end
```

## 4. 完整交互流程概览

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端
    participant B as 后端
    participant A as Agent服务

    Note over U,A: 问题理解阶段
    U->>F: 输入问题
    F->>B: 问题分析请求
    B->>A: 意图识别和问题分析
    A-->>B: 返回理解结果
    B-->>F: 返回问题理解
    F-->>U: 显示理解结果
    
    loop 问题澄清
        U->>F: 提供澄清信息
        F->>B: 更新理解请求
        B->>A: 更新问题理解
        A-->>B: 返回更新结果
        B-->>F: 返回更新后的理解
        F-->>U: 显示更新结果
    end
    
    U->>F: 确认问题理解
    F->>B: 确认请求
    B-->>F: 返回map_id和根节点
    F-->>U: 进入工作区

    Note over U,A: 问题解答阶段
    loop 节点执行循环
        F->>B: 获取节点信息
        B-->>F: 返回节点详情
        F-->>U: 显示节点信息
        
        U->>F: 开始执行节点
        F->>B: 执行请求
        B->>B: 判断执行动作
        
        alt 需要拆解
            B->>A: 问题拆解
            A-->>B: 返回拆解结果
            B-->>F: 返回子节点
            F-->>U: 显示拆解结果
        else 直接解答
            B->>A: 问题解答
            A-->>B: 返回解答结果
            B-->>F: 返回结论
            F-->>U: 显示结论
            
            U->>F: 反馈或确认
            F->>B: 反馈请求
            B->>A: 调整结论（如需要）
            A-->>B: 返回调整结果
            B-->>F: 返回最终结论
            F-->>U: 确认完成
        end
        
        B->>B: 选择下一个节点
        B-->>F: 返回下一个节点
        F-->>U: 切换到新节点
    end
    
    Note over U,A: 完成
    B-->>F: 所有节点完成
    F-->>U: 显示最终结果
```

## 5. 关键状态转换图

```mermaid
stateDiagram-v2
    [*] --> 问题输入
    问题输入 --> 问题分析: 提交问题
    问题分析 --> 问题澄清: 需要澄清
    问题澄清 --> 问题分析: 提供澄清信息
    问题分析 --> 问题确认: 理解完成
    问题确认 --> 工作区初始化: 确认问题
    
    工作区初始化 --> 节点选择: 初始化完成
    节点选择 --> 节点信息展示: 选择节点
    节点信息展示 --> 执行决策: 查看节点信息
    执行决策 --> 问题拆解: 选择拆解
    执行决策 --> 问题解答: 选择直接解答
    
    问题拆解 --> 子节点创建: 拆解完成
    子节点创建 --> 节点选择: 创建完成
    
    问题解答 --> 结论生成: 开始解答
    结论生成 --> 用户反馈: 生成结论
    用户反馈 --> 结论生成: 需要调整
    用户反馈 --> 节点完成: 确认结论
    
    节点完成 --> 节点选择: 还有未完成节点
    节点完成 --> 流程结束: 所有节点完成
    
    流程结束 --> [*]
    
    note right of 问题拆解
        调用Decompose Agent
        生成子问题列表
        确定依赖关系
    end note
    
    note right of 问题解答
        调用Analysis Agent
        调用Reasoning Agent
        调用Synthesis Agent
    end note
```

这些时序图完整展示了ThinkingMap的交互流程，包括：

1. **问题理解阶段**：从问题输入到最终确认的完整流程
2. **问题解答阶段**：节点执行、拆解、解答的详细过程
3. **异常处理**：网络异常、服务异常、超时等情况的处理
4. **完整流程概览**：整个系统的宏观交互流程
5. **状态转换**：各个阶段之间的状态转换关系

每个时序图都基于交互流程文档中的具体实现细节，确保了技术实现的准确性和完整性。 