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
  "full_name": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "username": "string",
    "email": "string",
    "full_name": "string",
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
  "request_id": "uuid"
}

# 用户登录
POST /api/v1/auth/login
Content-Type: application/json

Request:
{
  "username": "string",
  "password": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "uuid",
    "username": "string",
    "email": "string",
    "full_name": "string",
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

Response 401 Unauthorized:
{
  "code": 401,
  "message": "invalid credentials",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 刷新Token
POST /api/v1/auth/refresh
Authorization: Bearer <refresh_token>

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "access_token": "string",
    "refresh_token": "string",
    "expires_in": 900
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

Response 401 Unauthorized:
{
  "code": 401,
  "message": "invalid refresh token",
  "data": null,
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
  "request_id": "uuid"
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
  "root_question": "string"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "title": "string",
    "description": "string",
    "root_question": "string",
    "root_node_id": "uuid",
    "status": 1,
    "metadata": {},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
        "root_question": "string",
        "status": 1,
        "node_count": 10,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
    "root_question": "string",
    "root_node_id": "uuid",
    "status": 1,
    "metadata": {},
    "node_count": 10,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
  "request_id": "uuid"
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
        "parent_id": "uuid",
        "node_type": "analysis",
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
  "request_id": "uuid"
}

# 创建节点
POST /api/v1/maps/{mapId}/nodes
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "parent_id": "uuid",
  "node_type": "analysis",
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
    "map_id": "uuid",
    "parent_id": "uuid",
    "node_type": "analysis",
    "question": "string",
    "target": "string",
    "context": "string",
    "status": 0,
    "position": {
      "x": 100,
      "y": 200
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
  "request_id": "uuid"
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
        "node_id": "uuid",
        "dependency_type": "prerequisite",
        "required": true,
        "status": 2
      }
    ],
    "dependent_nodes": [
      {
        "node_id": "uuid",
        "dependency_type": "dependent",
        "required": true,
        "status": 0
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
        "node_id": "uuid",
        "detail_type": "string",
        "content": {
          "context": [
            {
              "node_id": "uuid",
              "type": "string",
              "question": "string",
              "target": "string",
              "conclusion": "string"
            }
          ],
          "question": "string",
          "target": "string",
          "message": [[1,2],[3,4]],
          "decompose_result": [
            {
              "question": "string",
              "target": "string"
            }
          ],
          "conclusion": "string"
        },
        "status": 1,
        "metadata": {},
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 创建节点详情
POST /api/v1/nodes/{nodeId}/details
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "detail_type": "string",
  "content": {
    "context": [
      {
        "node_id": "uuid",
        "type": "string",
        "question": "string",
        "target": "string",
        "conclusion": "string"
      }
    ],
    "question": "string",
    "target": "string",
    "message": [[1,2],[3,4]],
    "decompose_result": [
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
    "node_id": "uuid",
    "detail_type": "string",
    "content": { ... },
    "status": 1,
    "metadata": {},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
    "node_id": "uuid",
    "detail_type": "string",
    "content": { ... },
    "status": 1,
    "metadata": {},
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
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
  "request_id": "uuid"
}
```

#### 6.3.5 AI思考接口
```yaml
# 开始问题分析
POST /api/v1/thinking/analyze
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "node_id": "uuid",
  "context": "string",
  "options": {
    "model": "gpt-4",
    "temperature": 0.7
  }
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "uuid",
    "node_id": "uuid",
    "status": "processing",
    "estimated_time": 30
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 开始问题拆解
POST /api/v1/thinking/decompose
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "node_id": "uuid",
  "decompose_strategy": "breadth_first",
  "max_depth": 3
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "uuid",
    "node_id": "uuid",
    "status": "processing",
    "estimated_time": 60
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 生成结论
POST /api/v1/thinking/conclude
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "node_id": "uuid",
  "evidence": ["string"],
  "reasoning_type": "deductive"
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "task_id": "uuid",
    "node_id": "uuid",
    "status": "processing",
    "estimated_time": 45
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}

# 对话交互
POST /api/v1/thinking/chat
Authorization: Bearer <token>
Content-Type: application/json

Request:
{
  "node_id": "uuid",
  "message": "string",
  "context": "decompose" // decompose | conclude
}

Response 200 OK:
{
  "code": 200,
  "message": "success",
  "data": {
    "message_id": "uuid",
    "content": "string",
    "role": "assistant",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "timestamp": "2024-01-01T00:00:00Z",
  "request_id": "uuid"
}
```

#### 6.3.6 SSE接口
```yaml
# 建立SSE连接
GET /api/v1/sse/connect/{mapId}
Authorization: Bearer <token>
Accept: text/event-stream
Cache-Control: no-cache
Connection: keep-alive

# 连接成功响应
event: connected
data: {
  "connection_id": "uuid",
  "map_id": "uuid",
  "timestamp": "2024-01-01T00:00:00Z"
}

# 连接关闭响应
event: disconnected
data: {
  "connection_id": "uuid",
  "reason": "user_disconnected",
  "timestamp": "2024-01-01T00:00:00Z"
}
```
