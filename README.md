# Minimax 语音合成工作台 (Minimax Voice Workbench)

这是一个基于 Minimax API 开发的语音合成工作台，集成了音色复刻、语音合成、音色管理等功能。项目采用前后端一体化设计，最终可编译为单个可执行文件，方便部署与使用。

## 核心功能

- **密钥管理**：集中管理 Minimax API 密钥（API Key & Group ID）。
- **音色库管理**：
    - 查询、删除可用音色。
    - **音色复刻**：上传音频样本，快速复刻指定声音。
    - **音色设计**：通过参数微调生成自定义音色。
    - 音色试听与详细信息查看。
- **语音合成**：
    - 支持同步语音合成。
    - 丰富的参数自定义（语速、音量、情感强度等）。
    - 合成历史记录管理，支持本地保存与下载。
- **国际化**：支持中英文界面切换。

## 技术栈

- **后端**：Go 1.24 + [Gin](https://github.com/gin-gonic/gin) 框架
- **数据库**：SQLite ([GORM](https://gorm.io/))
- **前端**：Vue 3 + Vite + [Lucide Icons](https://lucide.dev/)
- **打包**：Go `embed` 实现前端静态资源嵌入

## 快速开始

### 1. 开发环境运行

#### 后端 (Go)
```bash
go run main.go
```
服务默认运行在 `http://localhost:8080`。

#### 前端 (Vue)
```bash
cd web
npm install
npm run dev
```
前端开发服务器通常运行在 `http://localhost:5173`。

> **注意**：前端开发时，API 请求会通过 Vite 代理转发到后端 `8080` 端口（详见 `web/vite.config.js`）。

### 2. 编译打包

如需生成单个可执行文件，请按照以下步骤操作：

1. **构建前端**：
   ```bash
   cd web
   npm run build
   ```
   这将在 `web/dist` 目录下生成静态文件。

2. **构建后端**：
   在项目根目录下执行：
   ```bash
   go build -o minimax-workbench.exe main.go
   ```
   Go 的 `embed` 指令会自动将 `web/dist` 的内容打包进二进制文件中。

### 3. 使用方法

直接运行生成的二进制文件：
```bash
./minimax-workbench.exe
```
程序启动后会尝试自动打开浏览器。如果未自动打开，请手动访问 `http://localhost:8080`。

#### 命令行参数
- `--data-dir`: (待实现) 指定数据库和上传文件的存储目录，默认为当前目录。

## 目录结构

```text
├── cmd/                # 命令行工具
├── internal/           # 内部业务逻辑
│   ├── api/            # 接口定义与路由
│   ├── database/       # 数据库模型与初始化
│   ├── model/          # 数据模型
│   └── service/        # 核心服务逻辑
├── pkg/                # 封装的第三方 SDK 或工具类
│   └── minimax/        # Minimax API 封装
├── web/                # Vue 前端项目
│   ├── src/
│   │   ├── views/      # 页面组件
│   │   ├── locales/    # 国际化语言包
│   │   └── api/        # 前端 API 请求封装
│   └── dist/           # 前端构建产物 (Git 忽略)
├── main.go             # 程序入口
├── 需求.md             # 项目需求文档
└── README.md           # 本说明文件
```

## 注意事项

- 请确保在使用前在“设置”页面配置正确的 Minimax API Key。
- 本项目产生的音频文件会默认保存至 `uploads/` 目录。
- 数据库文件 `minimax.db` 会在首次运行时于根目录自动创建。

## 许可证

[MIT License](LICENSE)
