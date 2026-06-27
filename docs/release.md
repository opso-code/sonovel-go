# 发布流程

## 自动构建发布

### 工作流程

GitHub Actions 会自动在以下情况触发构建和发布：

1. **代码提交到 main 分支** → 自动创建 release
2. **Pull Request** → 自动测试和构建
3. **标签推送** → 创建指定版本的 release

### 构建产物

- `sonovel-linux-amd64` - Linux 64 位二进制
- `sonovel-windows-amd64.exe` - Windows 64 位二进制
- `sonovel-darwin-amd64` - macOS 64 位二进制

### 版本管理

```bash
# 创建版本标签
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 或创建预发布版本
git tag -a v1.0.0-alpha -m "Alpha v1.0.0"
git push origin v1.0.0-alpha
```

## 手动发布

### 1. 准备发布

```bash
# 更新版本号
# 编辑 cmd/sonovel/main.go 中的版本信息

# 构建
go build -o ./bin/sonovel ./cmd/sonovel

# 测试
go test ./...
./bin/sonovel -help
```

### 2. 创建 Release

```bash
# 打标签
git tag -a v1.0.0 -m "Release v1.0.0"

# 推送到远程
git push origin v1.0.0

# 在 GitHub Web 界面创建 release，上传构建产物
```

### 3. 上传构建产物

在 GitHub Release 页面上传：
- `bin/sonovel` (Linux/macOS)
- `bin/sonovel.exe` (Windows)

## 分发方式

### 1. GitHub Releases

直接下载 Release 资产

### 2. Homebrew (macOS/Linux)

```bash
# 添加到 Homebrew 公式
brew create https://github.com/opso-code/sonovel-go.git

# 用户安装
brew install opso/sonovel-go/sonovel
```

### 3. Scoop (Windows)

```bash
# 添加到 Scoop bucket
scoop bucket add opso https://github.com/opso-code/sonovel-go

# 用户安装
scoop install opso/sonovel
```

### 4. 直接下载

- GitHub Releases 页面
- 项目 `Releases` 标签

## 发布清单

- [ ] 更新版本号
- [ ] 验证所有功能正常
- [ ] 运行 `go test ./...`
- [ ] 运行 `go build ./...`
- [ ] 创建 git tag
- [ ] 推送到远程
- [ ] 在 GitHub Web 创建 release
- [ ] 上传构建产物
- [ ] 编写 release notes
