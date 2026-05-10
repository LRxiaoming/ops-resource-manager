# 运维资源管理系统设计

## 背景

小部门需要一个基础设施资产管理和运维工单审批系统。

## 功能需求

### 1. 用户认证
- 用户名密码登录
- 角色：系统管理员、申请人、组长、经理
- JWT token 认证

### 2. 资产管理
- 资产类型：服务器、网络设备
- 字段：IP、主机名、CPU、内存、磁盘、机房位置、责任人、状态
- 状态：在线、离线、维护中
- 支持增删改查

### 3. 工单管理
- 工单类型：故障报修、变更申请、资源申请
- 字段：类型、标题、描述、优先级、状态、申请人、处理人、关联资产
- 优先级：紧急、重要、普通
- 2级审批流程：组长审批 -> 经理审批

### 4. Dashboard
- 资产统计：按类型、按状态分组
- 工单统计：按类型、按状态、按优先级分组
- 趋势图表

## 技术栈

- 后端：Go + Gin + GORM + JWT
- 前端：Vue3 + Element Plus + ECharts
- 数据库：PostgreSQL
- 部署：Docker + Docker Compose

## 数据库设计

### users
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 主键 |
| username | varchar(50) | 用户名，唯一 |
| password | varchar(255) | 密码（bcrypt加密） |
| role | varchar(20) | admin/applicant/leader/manager |
| created_at | timestamp | 创建时间 |

### assets
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 主键 |
| type | varchar(20) | server/network |
| ip | varchar(50) | IP地址 |
| hostname | varchar(100) | 主机名 |
| cpu | varchar(50) | CPU配置 |
| memory | varchar(50) | 内存配置 |
| disk | varchar(50) | 磁盘配置 |
| location | varchar(100) | 机房位置 |
| responsible | varchar(100) | 责任人 |
| status | varchar(20) | online/offline/maintenance |
| created_at | timestamp | 创建时间 |

### tickets
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 主键 |
| type | varchar(30) | fault_change_resource |
| title | varchar(200) | 标题 |
| description | text | 描述 |
| priority | varchar(20) | urgent/important/normal |
| status | varchar(20) | pending/approved/rejected/closed |
| applicant_id | bigint FK | 申请人 |
| handler_id | bigint FK | 处理人 |
| created_at | timestamp | 创建时间 |

### ticket_assets
| 字段 | 类型 | 说明 |
|------|------|------|
| ticket_id | bigint FK | 工单ID |
| asset_id | bigint FK | 资产ID |

### approvals
| 字段 | 类型 | 说明 |
|------|------|------|
| id | bigint PK | 主键 |
| ticket_id | bigint FK | 工单ID |
| approver_id | bigint FK | 审批人 |
| level | int | 审批层级（1=组长，2=经理） |
| result | varchar(20) | approved/rejected |
| comment | text | 审批意见 |
| created_at | timestamp | 审批时间 |

## API 设计

### 认证
- POST /api/auth/login - 登录
- POST /api/auth/logout - 登出

### 用户
- GET /api/users - 用户列表（管理员）
- POST /api/users - 创建用户
- PUT /api/users/:id - 更新用户
- DELETE /api/users/:id - 删除用户

### 资产
- GET /api/assets - 资产列表
- GET /api/assets/:id - 资产详情
- POST /api/assets - 创建资产
- PUT /api/assets/:id - 更新资产
- DELETE /api/assets/:id - 删除资产

### 工单
- GET /api/tickets - 工单列表
- GET /api/tickets/:id - 工单详情
- POST /api/tickets - 创建工单
- PUT /api/tickets/:id - 更新工单
- POST /api/tickets/:id/approve - 审批工单
- POST /api/tickets/:id/close - 关闭工单

### Dashboard
- GET /api/dashboard/stats - 统计数据
- GET /api/dashboard/charts - 图表数据

## 审批流程

1. 申请人提交工单，状态为 pending
2. 组长审批（level=1），通过后状态变为 approved
3. 经理审批（level=2），通过后状态变为 approved，处理人开始处理
4. 处理完成后，状态变为 closed

## 前端页面

- 登录页
- Dashboard 首页（统计卡片 + 图表）
- 资产管理页（列表 + 表单）
- 工单管理页（列表 + 表单 + 审批流程）
- 用户管理页（管理员）

## Docker 部署

### docker-compose.yml
```yaml
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_DB: ops_manager
      POSTGRES_USER: ops
      POSTGRES_PASSWORD: ops123
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  backend:
    build: ./backend
    depends_on:
      - db
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://ops:ops123@db:5432/ops_manager

  frontend:
    build: ./frontend
    ports:
      - "3000:80"
```

### 构建流程
1. 后端：`go build -o server ./cmd/server`
2. 前端：Vue3 构建产物
3. Docker Compose 启动所有服务