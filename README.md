# 面试吧 - AI智能面试平台 (Interview Bar)

基于字节跳动 Hertz 框架和 Eino 大语言模型框架开发的智能面试代理系统，助力求职者轻松应对各类面试挑战。

## 项目概述

「面试吧」是一个智能面试系统，利用 Eino 框架提供的大语言模型能力，实现简历分析、面试问题生成、答案评估等功能。系统采用 Hertz 作为 Web 框架，提供 RESTful API 接口，帮助用户提升面试技能，轻松拿到心仪的offer。

## 技术栈

- **后端框架**: [Hertz](https://github.com/cloudwego/hertz) - 字节跳动高性能 Web 框架
- **AI 框架**: [Eino](https://github.com/cloudwego/eino) - 大语言模型应用框架
- **数据库**: MySQL + GORM
- **缓存**: Redis
- **认证**: JWT

## 功能特性

- **用户管理**: 用户注册、登录、微信登录、用户信息管理
- **简历管理**: 简历上传、解析、智能分析、简历押题
- **面试系统**: 
  - 综合面试（校招/社招）
  - 专项面试（Java、Go、MySQL、Redis、消息队列等）
  - AI驱动的面试问题生成
  - 智能答案评估与反馈
  - 面试记录管理
- **AI智能体**: 
  - 简历分析智能体
  - 面试问题生成智能体
  - 答案评估智能体
  - 预测推荐智能体
- **向量检索**: 基于Milvus的文档向量检索
- **消息队列**: Redis消息队列处理异步任务

## 项目结构

```
├── README.md            # 项目说明文档
├── backend/             # 后端代码目录
│   ├── main.go          # 后端应用程序入口
│   ├── config.yaml      # 后端配置文件
│   ├── config.example.yaml # 后端配置示例文件
│   ├── go.mod           # Go 模块文件
│   ├── go.sum           # Go 依赖校验文件
│   ├── db_schema.sql    # 数据库模式定义
│   ├── api/             # API层
│   │   ├── handler/     # API处理器
│   │   ├── model/       # API数据模型
│   │   ├── response/    # 响应封装
│   │   └── router/      # 路由配置
│   ├── chatApp/         # AI聊天应用核心
│   │   ├── agent/       # AI智能体
│   │   ├── agent_service/ # AI智能体服务
│   │   ├── chat/        # 聊天功能
│   │   ├── config/      # 聊天应用配置
│   │   └── tool/        # 聊天应用工具
│   ├── idl/             # 接口定义语言文件
│   ├── internal/        # 内部包，不对外暴露
│   │   ├── config/      # 配置管理
│   │   ├── eino/        # Eino框架集成
│   │   ├── errors/      # 错误处理
│   │   ├── middleware/  # 中间件
│   │   ├── model/       # 数据模型定义
│   │   ├── mq/          # 消息队列
│   │   ├── repository/  # 数据访问层
│   │   └── service/     # 业务逻辑层
│   ├── mcp-moduel/      # MCP协议模块
│   ├── modelmgr/        # 模型管理器
│   └── pkg/             # 可重用的公共包
├── doc/                 # 项目文档目录
│   ├── 需求文档/        # 需求文档子目录
│   ├── AI编程开发规范.md # AI编程开发规范
│   ├── 后端架构设计.md    # 后端架构设计文档
│   ├── 技术实现方案.md    # 技术实现方案文档
│   └── 项目优化建议_*.md  # 各类优化建议
├── frontend/            # 前端代码目录 (Next.js)
│   ├── src/             # 前端源代码
│   ├── public/          # 静态资源
│   └── 配置文件         # 前端相关配置
└── mcpserver/           # MCP服务器
    ├── internal/        # MCP服务器内部实现
    └── config.yaml      # MCP服务器配置
```

### 目录功能说明

**根目录**: 项目的根目录，包含项目的整体说明文档和各主要组件目录
- `README.md`: 项目的主要说明文档，包含项目概述、安装说明等

**backend/**: 后端代码目录，包含所有服务器端代码
- `main.go`: 后端应用程序的主入口，负责初始化和启动服务
- `config.yaml`: 后端配置文件，包含数据库、Redis、API等配置
- `config.example.yaml`: 配置示例文件，提供配置模板
- `go.mod`: Go模块定义文件，管理项目依赖
- `go.sum`: 依赖版本锁定文件，确保依赖一致性
- `db_schema.sql`: 数据库表结构定义脚本
- `internal/`: 内部包目录，根据Go语言约定，这些包不会被外部项目导入
  - `config/`: 配置管理，负责加载和解析配置文件
  - `model/`: 数据模型定义，定义数据库表对应的Go结构体
  - `repository/`: 数据访问层，封装数据库和缓存操作
  - `middleware/`: 中间件，包含认证、日志等横切关注点
  - `service/`: 业务逻辑层，实现核心业务功能
  - `handler/`: API处理器，处理HTTP请求和响应
  - `utils/`: 工具函数，提供通用辅助功能
- `pkg/`: 可重用的公共包，可以被其他项目导入使用
  - `eino/`: Eino框架集成，封装AI大模型相关功能
  - `hertz/`: Hertz框架配置，提供Web框架相关的配置和工具
- `docs/`: API文档目录，包含自动生成的Swagger文档
- `chatApp/`: 聊天应用相关代码，实现与AI模型的对话功能
- `cmd/`: 命令行入口目录，包含可执行程序的入口点

**doc/**: 项目文档目录，包含项目相关的各种文档
- 包含架构设计、技术方案、需求文档等项目说明文件
- 包含前后端开发规范和计划文档

**frontend/**: 前端代码目录，包含所有客户端代码
- 存放前端应用代码，负责用户界面实现
- 包含React组件、页面、服务和工具函数等

## 配置说明

配置文件 `config.yaml` 包含以下主要配置项：

- **服务配置**: 端口、主机、日志配置
- **数据库配置**: MySQL连接信息、连接池配置
- **Redis配置**: Redis连接信息、连接池配置
- **Hertz框架配置**: 读写超时、日志级别
- **面试系统配置**: 面试时长、问题数量限制
- **安全性配置**: JWT密钥、CORS配置
- **AI服务配置**: 
  - OpenAI API配置
  - Embedding服务配置
  - 文档分割器配置
- **向量数据库**: Milvus向量数据库配置（可选）
- **微信配置**: 微信小程序登录配置
- **Google搜索**: Google搜索API配置（可选）

## AI架构与智能体

### 智能体架构
系统采用多智能体架构，每个智能体负责特定的业务功能：

- **简历分析智能体**: 分析用户简历，提取关键技能和知识点
- **面试问题生成智能体**: 基于简历和岗位要求生成个性化面试问题
- **答案评估智能体**: 评估用户答案的准确性、完整性和深度
- **预测推荐智能体**: 基于用户表现预测面试结果并推荐改进方向

### 向量检索
- 支持文档向量化存储和检索
- 基于Milvus向量数据库实现知识库检索
- 支持多种距离度量方式（余弦相似度、内积、L2距离）

### 消息队列
- 基于Redis的消息队列系统
- 支持异步任务处理
- 解耦AI处理流程，提升系统响应速度

## 安装和运行

### 前提条件

- Go 1.20+
- MySQL 数据库
- Redis 服务
- 有效的 OpenAI API Key（或其他支持的模型 API Key）
- （可选）Milvus向量数据库
- （可选）Google搜索API Key

### 安装步骤

1. 克隆项目
   ```bash
   git clone git@codeup.aliyun.com:60fadd729187b7df39056384/training_camp/go-eino-interview-agent.git
   cd go-eino-interview-agent
   ```

2. 配置环境
   进入 backend 目录并编辑 `config.yaml` 文件，填写相关配置：
   ```bash
   cd backend
   ```
   - 数据库连接信息
   - Redis 连接信息
   - API Key 等
   - 把 db_schema.sql 中的数据库 schema 导入到 MySQL 数据库中

3. 安装依赖
   ```bash
   go mod download
   ```

4. 运行服务
   ```bash
   go run main.go
   ```

## API 接口

### 用户相关
- `POST /api/v1/user/register` - 用户注册
- `POST /api/v1/user/login` - 用户登录
- `GET /api/v1/user/wechat/login` - 微信登录
- `GET /api/v1/user/wechat/callback` - 微信登录回调
- `GET /api/v1/user/profile` - 获取用户信息
- `PUT /api/v1/user/profile` - 更新用户信息

### 简历相关
- `POST /api/v1/resume` - 创建简历
- `GET /api/v1/resume` - 获取用户所有简历
- `GET /api/v1/resume/:id` - 获取指定简历
- `PUT /api/v1/resume/:id` - 更新简历
- `DELETE /api/v1/resume/:id` - 删除简历
- `POST /api/v1/resume/:id/analyze` - 分析简历
- `POST /api/v1/resume/:id/prediction` - 简历押题

### 面试相关
- `POST /api/v1/interview` - 创建面试
- `GET /api/v1/interview` - 获取用户所有面试
- `GET /api/v1/interview/:id` - 获取指定面试
- `POST /api/v1/interview/:id/start` - 开始面试
- `POST /api/v1/interview/:id/end` - 结束面试
- `POST /api/v1/interview/:id/answer` - 提交答案
- `GET /api/v1/interview/:id/questions` - 获取面试问题
- `GET /api/v1/interview/:id/result` - 获取面试结果
- `GET /api/v1/interview/:id/evaluation` - 获取面试评估

### 专项面试
- `POST /api/v1/special-interview/go` - 创建Go专项面试
- `POST /api/v1/special-interview/java` - 创建Java专项面试
- `POST /api/v1/special-interview/mysql` - 创建MySQL专项面试
- `POST /api/v1/special-interview/redis` - 创建Redis专项面试
- `POST /api/v1/special-interview/mq` - 创建消息队列专项面试

## 健康检查

- `GET /health` - 服务健康检查

## 开发说明

### 前端开发
- 基于Next.js 14+和TypeScript
- 使用Tailwind CSS进行样式设计
- 支持响应式设计，适配移动端和桌面端
- 集成Axios进行API调用
- 使用Zustand进行状态管理

### 中间件

- JWT 认证中间件保护需要授权的接口
- CORS 中间件处理跨域请求
- Recovery中间件捕获panic，防止服务崩溃
- 自定义错误处理中间件返回统一格式的错误响应

### 错误处理

系统采用统一的错误响应格式：
```json
{
  "error": "错误信息",
  "code": "错误代码",
  "details": "详细错误信息"
}
```

### 日志记录

- 使用结构化日志记录关键操作和错误信息
- 支持日志级别配置（debug、info、warn、error）
- 日志文件自动轮转和归档

### 代码规范
- 遵循Go官方代码规范
- 使用golangci-lint进行代码质量检查
- 单元测试覆盖率要求达到80%以上
- API文档使用Swagger自动生成

## 容器化部署

### Docker支持
- 提供完整的Docker容器化方案
- 支持docker-compose一键部署
- 包含开发环境、测试环境和生产环境配置

### 服务组件
- **后端服务**: Go应用容器
- **数据库**: MySQL容器
- **缓存**: Redis容器
- **向量数据库**: Milvus容器（可选）
- **前端服务**: Next.js应用容器
- **反向代理**: Nginx容器

### 部署文件
- `docker-compose.yml`: 基础部署配置
- `docker-compose-dev.yml`: 开发环境配置
- `docker-compose-prod.yml`: 生产环境配置
- `nginx.conf`: Nginx反向代理配置

## 注意事项

- 确保配置文件中的 API Key 和数据库连接信息安全存储
- 生产环境建议使用环境变量或密钥管理服务
- 定期更新依赖包以获取安全补丁
- 向量数据库和Google搜索为可选功能，可根据需要启用
- 建议配置适当的资源限制，防止AI服务消耗过多资源

## 许可证

MIT