# File Share Tool

一个轻量的局域网文件分享工具。启动后会把指定目录以网页形式分享给同一局域网内的设备，支持浏览目录、下载文件、二维码分享和复制下载链接。

## 功能

- 局域网内通过浏览器访问，无需安装客户端
- 文件夹优先展示，支持返回上级和面包屑导航
- 文件下载链接可生成二维码，也可一键复制
- 支持 Windows、Linux、macOS 作为分享端
- 后端使用相对路径访问共享目录，阻止路径穿越和符号链接逃逸
- 前端使用 React + MUI，支持桌面和移动端布局

## 快速使用

分享当前目录：

```bash
go run .
```

分享指定目录：

```bash
go run . -t "D:\Downloads"
```

分享用户主目录：

```bash
go run . -t home
```

启动后终端会输出访问地址，例如：

```text
http://192.168.1.3:8000
```

同一局域网内的手机、平板或电脑打开这个地址即可浏览和下载文件。

## 命令参数

```text
-t       要分享的目录，默认为当前目录；也可以传 home 表示用户主目录
-p       服务端口，默认 8000
-host    监听地址，默认 0.0.0.0
-open    是否自动打开浏览器，默认 true
```

示例：

```bash
go run . -t home -p 9000 -host 0.0.0.0 -open=false
```

如果只想本机访问，可以使用：

```bash
go run . -host 127.0.0.1
```

## 构建

需要安装：

- Go 1.24+
- Node.js
- npm
- make（可选，用于 Makefile）

构建前端：

```bash
make build-frontend
```

构建三平台可执行文件：

```bash
make build-all
```

产物会输出到 `build/`：

```text
file-share-windows-amd64.exe
file-share-linux-amd64
file-share-darwin-arm64
```

默认构建会把 `frontend/dist` 嵌入到 Go 二进制中，因此构建前需要先生成前端产物。开发时如果想直接使用 `frontend/public` 或本地 `frontend/dist` 目录，可以加 `-tags dev_frontend`。

没有 make 时，可以手动执行：

```bash
cd frontend
npm ci
npm run build
cd ..
go build -o build/file-share
```

## 开发检查

后端测试（需要已生成 `frontend/dist`）：

```bash
go test ./...
```

前端检查：

```bash
cd frontend
npm run lint
npm run build
```

开发模式后端测试（不需要提前生成 `frontend/dist`）：

```bash
go test -tags dev_frontend ./...
```

## API

列出目录：

```http
GET /api/files?path=<relative-path>
```

下载文件：

```http
GET /api/download?path=<relative-path>
```

说明：

- `path` 是共享根目录内的相对路径
- 根目录可以省略 `path` 或传空值
- 旧版 `fname` 参数仍兼容下载接口

## 安全说明

这个工具默认保持免登录，适合在可信局域网内临时分享文件。同一网络内知道地址的人可以访问共享目录中的文件。

建议：

- 只分享需要传输的目录，不要直接分享整个磁盘
- 公共 Wi-Fi 或不可信网络中不要使用默认局域网分享
- 需要仅本机访问时使用 `-host 127.0.0.1`

后端会阻止 `..` 路径穿越、绝对路径逃逸和指向共享目录外部的符号链接访问，但它不是权限系统，也不提供账号认证。
