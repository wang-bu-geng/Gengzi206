# 更子206

一个用于整理短剧制作流程的本地项目，包含剧本、角色、场景、分镜、图片和视频生成等功能。后端使用 Go，前端使用 Vue 3。部分生成能力需要自行配置对应的服务。

这是我此前短剧项目的后续整理版本，当前仓库展示的是现阶段的代码与使用说明。

## 当前包含的功能

- 创建和管理短剧项目、剧集及相关资料
- 处理剧本内容，并整理角色、场景和分镜
- 在工作流中生成和管理图片、视频素材
- 配置 AI 服务，查看生成任务的进度与结果

这些功能依赖所配置的服务和本地运行环境；具体可用范围以实际配置为准。

## 本地运行

需要 Go 1.23 或更高版本、Node.js 和 npm。视频处理功能还需要 FFmpeg。

```bash
cp configs/config.example.yaml configs/config.yaml
go mod download
cd web && npm install && cd ..
```

分别启动后端与前端：

```bash
go run main.go
```

```bash
cd web
npm run dev
```

默认情况下，前端在 `http://localhost:3012`，后端 API 在 `http://localhost:5678/api/v1`。前端开发服务器会代理 API 和静态资源请求。

`configs/config.yaml` 是本地配置文件，不应提交真实密钥。数据库、上传素材和编译产物也应保留在本地。详细数据迁移说明见 `docs/DATA_MIGRATION.md`。

## 项目结构

- `api/`：HTTP 接口与路由
- `application/`：业务流程
- `domain/`：数据模型
- `infrastructure/`：存储及外部服务接入
- `web/`：Vue 前端
- `configs/`：配置示例

