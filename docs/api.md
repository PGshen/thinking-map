### 接口
#### 6.3.1 认证相关接口
```yaml
# 用户注册
POST /api/v1/auth/register
Content-Type: application/json

Request:
{
  "username": "string",
  "email": "string", 
  "password": "string",
  "fullName": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "userID": "uuid",
    "username": "string",
    "email": "string",
    "fullName": "string",
    "accessToken": "string",
    "refreshToken": "string",
    "expiresIn": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

Response 400 Bad Request:
{
  "code": 400,
  "message": "invalid request parameters",
  "data": {
    "field": "username",
    "error": "username already exists"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 用户登录
POST /api/v1/auth/login
Content-Type: application/json

Request:
{
  "email": "string",
  "password": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "userID": "uuid",
    "username": "string",
    "email": "string",
    "fullName": "string",
    "accessToken": "string",
    "refreshToken": "string",
    "expiresIn": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

Response 401 Unauthorized:
{
  "code": 401,
  "message": "invalid credentials",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 刷新Token
POST /api/v1/auth/refresh
Authorization: Bearer <refresh_token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "accessToken": "string",
    "refreshToken": "string",
    "expiresIn": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

Response 401 Unauthorized:
{
  "code": 401,
  "message": "invalid refresh token",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 登出
POST /api/v1/auth/logout
Authorization: Bearer <access_token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}
```

#### 6.3.2 思考图接口
```yaml
# 创建思维导图
POST /api/v1/maps
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "problem": "string",         // 必填，问题描述，最大1000字符
  "problemType": "string",    // 可选，问题类型，最大50字符
  "target": "string",         // 可选，目标，最大1000字符
  "keyPoints": [ ... ],         // 可选，关键点（数组/对象，结构见下）
  "constraints": [ ... ]        // 可选，约束条件（数组/对象，结构见下）
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "string",             // 导图ID（UUID）
    "status": 1,                 // 状态码
    "problem": "string",        // 问题描述
    "problemType": "string",    // 问题类型
    "target": "string",         // 目标
    "keyPoints": [ ... ],         // 关键点
    "constraints": [ ... ],       // 约束条件
    "conclusion": "string",      // 结论
    "metadata": {},              // 元数据（对象）
    "createdAt": "string",      // 创建时间（ISO8601）
    "updatedAt": "string"       // 更新时间（ISO8601）
  },
  "timestamp": "string",        // 响应时间（ISO8601）
  "requestID": "string"         // 请求ID（UUID）
}

// 说明：
// keyPoints、constraints 字段的具体结构请参考后端 model.KeyPoints、model.Constraints 的定义。

# 获取用户的思维导图列表
GET /api/v1/maps?page=1&limit=20&status=1
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "limit": 20,
    "items": [
      {
        "id": "uuid",
        "status": 1,
        "title": "string",
        "problem": "string",
        "problemType": "string",
        "target": "string",
        "keyPoints": ["string"],
        "constraints": ["string"],
        "conclusion": "string",
        "progress": 0.5,
        "metadata": {},
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 获取思维导图详情
GET /api/v1/maps/{mapID}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "status": 1,
    "title": "string",
    "problem": "string",
    "problemType": "string",
    "target": "string",
    "keyPoints": ["string"],
    "constraints": ["string"],
    "conclusion": "string",
    "progress": 0.5,
    "metadata": {},
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 更新思维导图
PUT /api/v1/maps/{mapID}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "status": 1,
  "title": "string",
  "problem": "string",
  "problemType": "string",
  "target": "string",
  "keyPoints": ["string"],
  "constraints": ["string"],
  "conclusion": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "status": 1,
    "title": "string",
    "problem": "string",
    "problemType": "string",
    "target": "string",
    "keyPoints": [],
    "constraints": [],
    "conclusion": "string",
    "progress": 0.5,
    "metadata": {},
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 删除思维导图
DELETE /api/v1/maps/{mapID}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}
```

#### 6.3.3 节点管理接口
```yaml
# 获取思维导图的所有节点
GET /api/v1/maps/{mapID}/nodes
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "nodes": [
      {
        "id": "uuid",
        "mapID": "uuid",
        "parentID": "uuid",
        "nodeType": "analysis",
        "question": "string",
        "target": "string",
        "context": {
          "ancestor": [
            {
              "question": "string",
              "target": "string",
              "conclusion": "string",
              "abstract": "string",
              "status": "string"
            }
          ],
          "prevSibling": [
            {
              "question": "string",
              "target": "string",
              "conclusion": "string",
              "abstract": "string",
              "status": "string"
            }
          ],
          "children": [
            {
              "question": "string",
              "target": "string",
              "conclusion": "string",
              "abstract": "string",
              "status": "string"
            }
          ]
        },
        "conclusion": "string",
        "status": 0,
        "position": {
          "x": 100,
          "y": 200,
          "width": 0,
          "height": 0
        },
        "metadata": {},
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 创建节点
POST /api/v1/maps/{mapID}/nodes
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "mapID": "uuid",
  "parentID": "uuid",
  "nodeType": "analysis",
  "question": "string",
  "target": "string",
  "position": {
    "x": 100,
    "y": 200,
    "width": 0,
    "height": 0
  }
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "mapID": "uuid",
    "parentID": "uuid",
    "nodeType": "analysis",
    "question": "string",
    "target": "string",
    "context": {
      "ancestor": [],
      "prevSibling": [],
      "children": []
    },
    "conclusion": "string",
    "status": 0,
    "position": {
      "x": 100,
      "y": 200,
      "width": 0,
      "height": 0
    },
    "metadata": {},
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 更新节点
PUT /api/v1/{mapID}/nodes/{nodeID}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "question": "string",
  "target": "string",
  "position": {
    "x": 100,
    "y": 200,
    "width": 0,
    "height": 0
  }
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "mapID": "uuid",
    "parentID": "uuid",
    "nodeType": "analysis",
    "question": "string",
    "target": "string",
    "context": {
      "ancestor": [],
      "prevSibling": [],
      "children": []
    },
    "conclusion": "string",
    "status": 0,
    "position": {
      "x": 100,
      "y": 200,
      "width": 0,
      "height": 0
    },
    "metadata": {},
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 更新节点上下文
PUT /api/v1/{mapID}/nodes/{nodeID}/context
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "context": {
    "ancestor": [
      {
        "question": "string",
        "target": "string",
        "conclusion": "string",
        "abstract": "string",
        "status": "string"
      }
    ],
    "prevSibling": [
      {
        "question": "string",
        "target": "string",
        "conclusion": "string",
        "abstract": "string",
        "status": "string"
      }
    ],
    "children": [
      {
        "question": "string",
        "target": "string",
        "conclusion": "string",
        "abstract": "string"
      }
    ]
  }
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "mapID": "uuid",
    "parentID": "uuid",
    "nodeType": "analysis",
    "question": "string",
    "target": "string",
    "context": {
      "ancestor": [],
      "prevSibling": [],
      "children": []
    },
    "conclusion": "string",
    "status": 0,
    "position": {
      "x": 100,
      "y": 200,
      "width": 0,
      "height": 0
    },
    "metadata": {},
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 重置上下文
PUT /api/v1/{mapID}/nodes/{nodeID}/context/reset
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "mapID": "uuid",
    "parentID": "uuid",
    "nodeType": "string",
    "question": "string",
    "target": "string",
    "position": {
      "x": 0,
      "y": 0
    },
    "context": {
      "ancestor": [],
      "prevSibling": [],
      "children": []
    },
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 删除节点
DELETE /api/v1/{mapID}/nodes/{nodeID}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 获取下一个可执行节点
GET /api/v1/maps/{mapID}/executable-nodes
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "nodeIDs": ["uuid"]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

#### 6.3.5 思考相关接口
```yaml


# SSE连接
GET /api/v1/sse/connect?map_id={map_id}
Authorization: Bearer <token>

Response 200 OK:
event: connection_established
data: {
  "sessionID": "uuid",
  "clientID": "uuid",
  "message": "SSE连接已建立"
}

# SSE事件格式
event: nodeCreated
data: {
  "nodeID": "uuid",
  "parentID": "uuid",
  "nodeType": "analysis",
  "question": "string",
  "target": "string",
  "position": {"x": 100, "y": 200},
  "dependencies": ["uuid"]
}

event: nodeUpdated
data: {
  "nodeID": "uuid",
  "mode": "replace",
  "updates": {
    "status": "completed",
    "conclusion": "string"
  }
}

event: thinkingProgress
data: {
  "nodeID": "uuid",
  "stage": "analyzing|reasoning|synthesizing",
  "progress": 50,
  "message": "string"
}


# 问题分析
POST /api/v1/thinking/analyze
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "question": "string",
  "questionType": "research|creative|analysis|planning",
  "userID": "uuid"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "understanding": {
      "coreQuestion": "string",
      "target": "string",
      "keyPoints": ["string"],
      "constraints": ["string"],
      "context": "string",
      "complexity": "high|medium|low"
    },
    "suggestions": ["string"],
    "clarificationQuestions": ["string"]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 问题澄清
POST /api/v1/thinking/clarify
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "sessionID": "uuid",
  "clarifications": {
    "answers": ["string"],
    "additionalInfo": "string",
    "modifications": {
      "target": "string",
      "constraints": ["string"]
    }
  }
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "understanding": {
      "coreQuestion": "string",
      "target": "string",
      "keyPoints": ["string"],
      "constraints": ["string"],
      "context": "string",
      "complexity": "high|medium|low"
    },
    "suggestions": ["string"],
    "clarificationQuestions": ["string"]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 问题确认
POST /api/v1/thinking/confirm
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "sessionID": "uuid",
  "finalUnderstanding": {
    "problem": "string",
    "target": "string",
    "keyPoints": ["string"],
    "constraints": ["string"]
  }
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "mapID": "uuid",
    "rootNode": {
      "nodeID": "uuid",
      "question": "string",
      "target": "string",
      "status": "pending"
    }
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 节点执行
POST /api/v1/thinking/execute
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeID": "uuid",
  "action": "start"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "action": "decompose|conclude",
    "reason": "string",
    "nextTab": "decompose|conclusion"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 问题拆解
POST /api/v1/thinking/decompose
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeID": "uuid",
  "question": "string",
  "context": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "subProblems": [
      {
        "question": "string",
        "target": "string",
        "priority": 1,
        "dependencies": ["string"]
      }
    ],
    "strategy": "string"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 问题拆解反馈
POST /api/v1/thinking/decompose/feedback
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeID": "节点ID",
  "feedback": {
    "rating": 3,
    "comments": "拆解不够细致，缺少技术实现层面",
    "issues": [
      {
        "subProblemID": "子问题ID",
        "issue": "问题描述过于宽泛"
      }
    ],
    "suggestions": "希望增加技术选型相关的子问题"
  },
  "action": "regenerate|adjust|confirm"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "feedbackID": "反馈ID",
    "status": "received",
    "nextAction": "regenerate"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}


# 结论生成
POST /api/v1/thinking/conclude
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeID": "uuid",
  "question": "string",
  "context": "string",
  "subConclusions": ["string"]
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "conclusion": "string",
    "confidence": 0.85,
    "evidence": ["string"],
    "limitations": ["string"],
    "recommendations": ["string"]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

# 用户反馈
POST /api/v1/thinking/conclude/feedback
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeID": "uuid",
  "feedback": "string",
  "action": "adjust|confirm"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "conclusion": "string",
    "confidence": 0.85,
    "evidence": ["string"],
    "limitations": ["string"],
    "recommendations": ["string"]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestID": "uuid"
}

```

