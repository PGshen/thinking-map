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
    "userId": "uuid",
    "username": "string",
    "email": "string",
    "fullName": "string",
    "accessToken": "string",
    "refreshToken": "string",
    "expiresIn": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
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
  "requestId": "uuid"
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
    "userId": "uuid",
    "username": "string",
    "email": "string",
    "fullName": "string",
    "accessToken": "string",
    "refreshToken": "string",
    "expiresIn": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

Response 401 Unauthorized:
{
  "code": 401,
  "message": "invalid credentials",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
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
  "requestId": "uuid"
}

Response 401 Unauthorized:
{
  "code": 401,
  "message": "invalid refresh token",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
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
  "requestId": "uuid"
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
  "title": "string",
  "description": "string",
  "rootQuestion": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "title": "string",
    "description": "string",
    "rootQuestion": "string",
    "rootNodeId": "uuid",
    "status": 1,
    "metadata": {},
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

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
        "title": "string",
        "description": "string",
        "rootQuestion": "string",
        "status": 1,
        "nodeCount": 10,
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 获取思维导图详情
GET /api/v1/maps/{mapId}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "title": "string",
    "description": "string",
    "rootQuestion": "string",
    "rootNodeId": "uuid",
    "status": 1,
    "metadata": {},
    "nodeCount": 10,
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 更新思维导图
PUT /api/v1/maps/{mapId}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "title": "string",
  "description": "string",
  "status": 1
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "title": "string",
    "description": "string",
    "status": 1,
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 删除思维导图
DELETE /api/v1/maps/{mapId}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}
```

#### 6.3.3 节点管理接口
```yaml
# 获取思维导图的所有节点
GET /api/v1/maps/{mapId}/nodes
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "nodes": [
      {
        "id": "uuid",
        "parentId": "uuid",
        "nodeType": "analysis",
        "question": "string",
        "target": "string",
        "status": 0,
        "position": {
          "x": 100,
          "y": 200
        }
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 创建节点
POST /api/v1/maps/{mapId}/nodes
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "parentId": "uuid",
  "nodeType": "analysis",
  "question": "string",
  "target": "string",
  "context": "string",
  "position": {"x": 100, "y": 200}
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "mapId": "uuid",
    "parentId": "uuid",
    "nodeType": "analysis",
    "question": "string",
    "target": "string",
    "context": "string",
    "status": 0,
    "position": {
      "x": 100,
      "y": 200
    },
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 更新节点
PUT /api/v1/nodes/{nodeId}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "question": "string",
  "target": "string",
  "context": "string",
  "position": {"x": 100, "y": 200}
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "question": "string",
    "target": "string",
    "context": "string",
    "position": {
      "x": 100,
      "y": 200
    },
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 删除节点
DELETE /api/v1/nodes/{nodeId}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 检查节点依赖
GET /api/v1/nodes/{nodeId}/dependencies
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "dependencies": [
      {
        "nodeId": "uuid",
        "dependencyType": "prerequisite",
        "required": true,
        "status": 2
      }
    ],
    "dependentNodes": [
      {
        "nodeId": "uuid",
        "dependencyType": "dependent",
        "required": true,
        "status": 0
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 添加节点依赖
POST /api/v1/nodes/{nodeId}/dependencies
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "dependencyNodeId": "uuid", // 依赖的节点ID
  "dependencyType": "prerequisite", // prerequisite | dependent
  "required": true
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "nodeId": "uuid",
    "dependencyNodeId": "uuid",
    "dependencyType": "prerequisite",
    "required": true,
    "status": 0
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 删除节点依赖
DELETE /api/v1/nodes/{nodeId}/dependencies/{dependencyNodeId}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}
```

#### 6.3.4 节点详情接口
```yaml
# 获取节点详情
GET /api/v1/nodes/{nodeId}/details
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "details": [
      {
        "id": "uuid",
        "nodeId": "uuid",
        "detailType": "string",
        "content": {
          "context": [
            {
              "nodeId": "uuid",
              "type": "string",
              "question": "string",
              "target": "string",
              "conclusion": "string"
            }
          ],
          "question": "string",
          "target": "string",
          "message": [[1,2],[3,4]],
          "decomposeResult": [
            {
              "question": "string",
              "target": "string"
            }
          ],
          "conclusion": "string"
        },
        "status": 1,
        "metadata": {},
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 创建节点详情
POST /api/v1/nodes/{nodeId}/details
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "detailType": "string",
  "content": {
    "context": [
      {
        "nodeId": "uuid",
        "type": "string",
        "question": "string",
        "target": "string",
        "conclusion": "string"
      }
    ],
    "question": "string",
    "target": "string",
    "message": [[1,2],[3,4]],
    "decomposeResult": [
      {
        "question": "string",
        "target": "string"
      }
    ],
    "conclusion": "string"
  },
  "status": 1,
  "metadata": {}
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "nodeId": "uuid",
    "detailType": "string",
    "content": { ... },
    "status": 1,
    "metadata": {},
    "createdAt": "2024-01-01T00:00:00Z",
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 更新节点详情
PUT /api/v1/node-details/{detailId}
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "content": { ... },
  "status": 1,
  "metadata": {}
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "nodeId": "uuid",
    "detailType": "string",
    "content": { ... },
    "status": 1,
    "metadata": {},
    "updatedAt": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 删除节点详情
DELETE /api/v1/node-details/{detailId}
Authorization: Bearer <token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}
```

#### 6.3.5 思考相关接口
```yaml


# SSE连接
GET /api/v1/sse/connect?map_id={map_id}&user_id={user_id}
Authorization: Bearer <token>

Response 200 OK:
event: connection_established
data: {
  "mapId": "uuid",
  "currentStatus": "ready",
  "activeNodes": ["uuid"]
}

# SSE事件格式
event: node_created
data: {
  "nodeId": "uuid",
  "parentId": "uuid",
  "nodeType": "analysis",
  "question": "string",
  "target": "string",
  "position": {"x": 100, "y": 200},
  "dependencies": ["uuid"]
}

event: node_updated
data: {
  "nodeId": "uuid",
  "updates": {
    "status": "completed",
    "conclusion": "string"
  }
}

event: thinking_progress
data: {
  "nodeId": "uuid",
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
  "userId": "uuid"
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
  "requestId": "uuid"
}

# 问题澄清
POST /api/v1/thinking/clarify
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "sessionId": "uuid",
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
  "requestId": "uuid"
}

# 问题确认
POST /api/v1/thinking/confirm
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "sessionId": "uuid",
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
    "mapId": "uuid",
    "rootNode": {
      "nodeId": "uuid",
      "question": "string",
      "target": "string",
      "status": "pending"
    }
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}

# 节点执行
POST /api/v1/thinking/execute
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeId": "uuid",
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
  "requestId": "uuid"
}

# 问题拆解
POST /api/v1/thinking/decompose
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeId": "uuid",
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
  "requestId": "uuid"
}

# 问题拆解反馈
POST /api/v1/thinking/decompose/feedback
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeId": "节点ID",
  "feedback": {
    "rating": 3,
    "comments": "拆解不够细致，缺少技术实现层面",
    "issues": [
      {
        "subProblemId": "子问题ID",
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
    "feedbackId": "反馈ID",
    "status": "received",
    "nextAction": "regenerate"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "requestId": "uuid"
}


# 结论生成
POST /api/v1/thinking/conclude
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeId": "uuid",
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
  "requestId": "uuid"
}

# 用户反馈
POST /api/v1/thinking/conclude/feedback
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "nodeId": "uuid",
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
  "requestId": "uuid"
}

```

