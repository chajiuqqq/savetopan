# 小红书收集器

## 项目概述

这是一个基于Golang和Gin框架开发的资源处理工具，主要用于从特定来源（如小红书）下载资源并上传到网盘（如AList）进行保存。工具提供了Web界面和API接口两种使用方式，方便用户操作。

[Alist](https://alistgo.com/zh/) 支持对接多种网盘，比如夸克网盘、百度网盘等，用户可以挂载目录到不同的网盘进行资源上传。

## 功能特性

- **提取URL**：自动从输入内容中提取URL
- **资源下载**：调用下载服务获取媒体资源，目前仅支持小红书
- **网盘上传**：将资源上传到 Alist 支持的网盘
- **Web界面**：提供用户友好的Web界面，自动读取剪贴板内容
- **配置管理**：支持通过配置文件自定义服务参数
- **系统服务**：提供systemd服务配置，便于后台稳定运行

## 技术栈

- **后端**：Golang 1.24.2 + Gin框架
- **HTTP客户端**：Resty
- **配置管理**：ini配置文件
- **前端**：HTML5 + CSS3 + JavaScript

## 安装部署

### 1. 准备工作

- Go 1.24.2或更高版本
- [XHS-Downloader小红书资源下载器](https://github.com/JoeanAmier/XHS-Downloader)
- [Alist](https://alistgo.com/zh/)

### 2. 编译程序

```bash
# 克隆仓库（如果有）
# git clone <repository-url>

# 进入项目目录
cd /path/to/res_saver

# 获取依赖
go mod tidy

# 编译程序
go build -o res_saver
```

### 3. 配置文件

编辑`config.ini`文件，根据实际情况修改配置：

```ini
[server]
port = 9092

[xhs_downloader]
url = http://xxx/xhs/detail
download_dir = /root/res_saver/xhs_downloader_volume/Download

[alist]
url = http://xxx/api/fs/put
# 在Alist的设置-其他，最底下的token
token = xxx
# 上传路径，直接用文件夹名称，支持子文件夹
upload_path = /夸克
as_task = true

[cors]
allow_origin = *
allow_methods = POST, OPTIONS
allow_headers = Origin, Content-Type, Accept
```

### 4. 启动服务

#### 直接运行

```bash
./res_saver
```

#### 作为系统服务运行

1. 将`res_saver.service`复制到系统服务目录：
   ```bash
   sudo cp res_saver.service /etc/systemd/system/
   ```

2. 重载systemd配置：
   ```bash
   sudo systemctl daemon-reload
   ```

3. 启动服务：
   ```bash
   sudo systemctl start res_saver.service
   ```

4. 设置开机自启：
   ```bash
   sudo systemctl enable res_saver.service
   ```

5. 查看服务状态：
   ```bash
   sudo systemctl status res_saver.service
   ```

## 使用方法

### Web界面

1. 启动服务后，在浏览器中访问：`http://服务器IP:9092`
2. 页面会自动尝试读取剪贴板内容（需浏览器权限）
3. 确认URL无误后，点击「提交」按钮
4. 等待处理完成，查看结果信息

### API接口

可以直接调用`/api/process`接口：

```bash
curl -X POST http://服务器IP:9092/api/process \
  -H "Content-Type: application/json" \
  -d '{"message":{"text":"https://www.xiaohongshu.com/explore/作品ID"}}'
```

## URL格式支持

系统支持以下URL格式：
- `https://www.xiaohongshu.com/explore/作品ID?xsec_token=XXX`
- `https://www.xiaohongshu.com/discovery/item/作品ID?xsec_token=XXX`
- `https://www.xiaohongshu.com/user/profile/作者ID/作品ID?xsec_token=XXX`
- `https://xhslink.com/分享码`

## 错误排查

1. **服务无法启动**：检查配置文件是否正确，端口是否被占用
2. **API请求失败**：查看服务日志，确认URL格式是否正确
3. **上传失败**：检查AList配置和网络连接

## 日志查看

- 直接运行时，日志输出到控制台
- 作为systemd服务运行时，使用以下命令查看日志：
  ```bash
  journalctl -u res_saver.service
  ```

## 注意事项

1. 确保下载目录和相关文件有正确的读写权限
2. 定期清理下载目录中的临时文件，避免磁盘空间占用过大
3. 生产环境中应修改默认配置，特别是安全相关的token等信息
4. 对于大量资源的处理，可能需要调整内存限制