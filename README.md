# RSS to Telegram Bot 使用文档

## [English](#english-version) | [简体中文](#rss-to-telegram-bot-使用文档)

## 🏷 简介

**_简称 rss2tg，用于将自定义 RSS 地址，字段，刷新时间里的相关帖子即时发送到自定义 TG 用户或频道，省去你刷帖子的时间_**

**_支持 AMD64/ARM64_**

**_镜像大小 17M，内存占用 10M_**

**——By [drfyup](https://hstz.com)**

## 📋 目录

- [简介](#🏷-简介)
- [部署方法](#1-部署方法)
- [程序使用说明](#2-程序使用说明)
- [配置详解](#配置详解)
- [Webhook 集成](#webhook-集成)
- [命令说明](#命令说明)
- [故障排查](#4-故障排查)
- [在中国大陆服务器上使用](#5-在中国大陆服务器上使用)

## 1. 部署方法

### 1.1 使用 Docker Compose（推荐）

1. 确保已安装 Docker 和 Docker Compose（方法自寻）。

2. 克隆或下载项目代码到本地。

```bash
git clone https://github.com/3377/rss2tg.git
```

3. 进入项目目录。

4. 编辑 `docker-compose.yml` 文件，修改环境变量：

-- 进入任意目录或直接当前目录，新建 docker-compose.yml 文件，填入以下内容

```yaml
version: "3"
services:
  rss2tg:
    container_name: rss2tg
    image: drfyup/rss2tg:latest
    volumes:
      - ./config/config.yaml:/app/config/config.yaml
      - ./data:/app/data
    environment:
      - TELEGRAM_BOT_TOKEN=your_bot_token_here
      - TELEGRAM_USERS=user_id_1,user_id_2
      - TELEGRAM_CHANNELS=@channel_1,@channel_2
      - TZ=Asia/Shanghai
    restart: unless-stopped
```

将`your_bot_token_here` 替换为您的 Telegram Bot Token，`user_id_1,user_id_2` 替换为您要接收消息的用户 ID，`@channel_1,@channel_2` 替换为您要发送消息的频道名称。

5. 运行以下命令启动容器：

```yaml
docker-compose up  -d
```

### 1.2 使用 Docker Run

1. 构建 Docker 镜像：

```yaml
docker pull drfyup/rss2tg:latest
```

2. 运行 Docker 容器：

```yaml
docker run -d \
--name rss2tg \
-v $(pwd)/config:/app/config \
-v $(pwd)/data:/app/data \
-e TELEGRAM_BOT_TOKEN=your_bot_token_here \
-e TELEGRAM_USERS=user_id_1,user_id_2 \
-e TELEGRAM_CHANNELS=@channel_1,@channel_2 \
-e TELEGRAM_API_URL=http://xxx.deno.dev/telegram \
-e TZ=Asia/Shanghai \
--restart unless-stopped \
drfyup/rss2tg:latest
```

请替换环境变量中的相应值。

## 2. 程序使用说明

### 2.1 配置文件

程序支持通过 YAML 配置文件或环境变量进行配置。配置文件位于 `/app/config/config.yaml`。如果该文件不存在，程序将使用环境变量进行初始配置。
环境变量读取优先级高于配置文件。

配置文件示例：

```yaml
telegram:
  bot_token: "your_bot_token_here"
  users:
    - "user_id_1"
    - "user_id_2"
  channels:
    - "@channel_1"
    - "@channel_2"
  adminuser: # 管理员用户配置（可选）
    - "admin_id_1"
    - "admin_id_2"

rss:
  - urls:
      - "https://example.com/feed1.xml"
      - "https://example.com/feed2.xml"
    interval: 300
    keywords:
      - "keyword1"
      - "keyword2"
    group: "Group1"
    allow_part_match: true # 是否允许部分关键词匹配
```

## 配置详解

### 环境变量命名规则

#### 基础配置

| 环境变量 | 必填 | 说明 | 示例 |
|----------|------|------|------|
| `TELEGRAM_BOT_TOKEN` | ✅ | Telegram Bot 的 API Token | `110201543:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw` |
| `TELEGRAM_USERS` | ✅ | 接收消息的用户 ID，多个用逗号分隔 | `123456789,987654321` |
| `TELEGRAM_CHANNELS` | ❌ | 接收消息的频道，多个用逗号分隔 | `@channel1,@channel2` |
| `TELEGRAM_ADMIN_USERS` | ❌ | 管理员用户 ID，多个用逗号分隔 | `123456789,987654321` |
| `TELEGRAM_API_URL` | ❌ | 自定义 Telegram API 服务器地址 | `http://fyapi.deno.dev/telegram` |
| `TZ` | ❌ | 时区设置 | `Asia/Shanghai` |

#### RSS 配置命名规则

**新格式（推荐）**：使用数字后缀，从 1 开始

```bash
# 第一个 RSS 源
RSS_URLS_1=https://example1.com/rss,https://example2.com/rss
RSS_KEYWORDS_1=关键词1,关键词2
RSS_GROUP_1=技术资讯
RSS_INTERVAL_1=300
RSS_ALLOW_PART_MATCH_1=true

# 第二个 RSS 源
RSS_URLS_2=https://news.example.com/rss
RSS_KEYWORDS_2=新闻,热点
RSS_GROUP_2=新闻资讯
RSS_INTERVAL_2=600
RSS_ALLOW_PART_MATCH_2=false

# 第三个 RSS 源（最多支持 10 个）
RSS_URLS_3=https://tech.example.com/feed
RSS_GROUP_3=科技动态
RSS_INTERVAL_3=900
# 不设置关键词表示推送所有文章
```

**旧格式（兼容）**：使用分号分隔多个 RSS 组

```bash
RSS_URLS=https://example1.com/rss,https://example2.com/rss;https://news.example.com/rss
RSS_KEYWORDS_0=关键词1,关键词2  # 对应第一组
RSS_KEYWORDS_1=新闻,热点       # 对应第二组
RSS_GROUP_0=技术资讯
RSS_GROUP_1=新闻资讯
```

#### Webhook 配置命名规则

**单个 Webhook（向后兼容）**：

```bash
WEBHOOK_ENABLED=true
WEBHOOK_URL=http://your-message-pusher:3000/webhook/your_webhook_id
WEBHOOK_TIMEOUT=10
WEBHOOK_RETRY_COUNT=3
```

**多个 Webhooks（推荐）**：使用数字后缀，从 1 开始

```bash
# 第一个 webhook
WEBHOOK_URL_1=http://server1:3000/webhook/webhook_id_1
WEBHOOK_NAME_1=message-pusher-1
WEBHOOK_ENABLED_1=true
WEBHOOK_TIMEOUT_1=10
WEBHOOK_RETRY_COUNT_1=3

# 第二个 webhook
WEBHOOK_URL_2=http://server2:3000/webhook/webhook_id_2
WEBHOOK_NAME_2=message-pusher-2
WEBHOOK_ENABLED_2=true
WEBHOOK_TIMEOUT_2=15
WEBHOOK_RETRY_COUNT_2=2

# 第三个 webhook（可选，最多支持 10 个）
WEBHOOK_URL_3=http://backup:3000/webhook/backup_id
WEBHOOK_NAME_3=backup-webhook
WEBHOOK_ENABLED_3=false  # 可以暂时禁用
WEBHOOK_TIMEOUT_3=5
WEBHOOK_RETRY_COUNT_3=1
```

### 配置优先级说明

1. **环境变量优先级最高**：如果设置了环境变量，将覆盖配置文件中的相应设置
2. **配置文件次之**：如果没有对应的环境变量，使用配置文件中的设置
3. **默认值最低**：如果既没有环境变量也没有配置文件设置，使用程序默认值

### 配置文件格式

详细配置请参考 `config/config.yaml.example`：

```yaml
# Telegram 配置
telegram:
  bot_token: "your_telegram_bot_token"
  users:
    - "123456789"
  channels:
    - "@your_channel"
  adminuser:
    - "123456789"

# 单个 Webhook 配置（向后兼容）
webhook:
  enabled: false
  url: "http://your-message-pusher:3000/webhook/your_webhook_id"
  timeout: 10
  retry_count: 3

# 多个 Webhooks 配置（推荐使用）
webhooks:
  - name: "message-pusher-1"
    enabled: true
    url: "http://server1:3000/webhook/webhook_id_1"
    timeout: 10
    retry_count: 3
  - name: "message-pusher-2"
    enabled: true
    url: "http://server2:3000/webhook/webhook_id_2"
    timeout: 15
    retry_count: 2
  - name: "backup-webhook"
    enabled: false  # 可以暂时禁用某个 webhook
    url: "http://backup:3000/webhook/backup_id"
    timeout: 5
    retry_count: 1

# RSS 订阅配置
rss:
  - urls:
      - "https://example.com/rss"
      - "https://example2.com/feed"
    interval: 300  # 检查间隔（秒）
    keywords:
      - "关键词1"
      - "关键词2"
    group: "技术资讯"
    allow_part_match: true  # 是否允许部分匹配
```

## Webhook 集成

现在支持通过 webhook 将消息推送到 [message-pusher](https://github.com/songquanpeng/message-pusher)，实现多平台消息推送：

- 📧 邮件推送
- 💬 企业微信推送
- 📱 钉钉推送
- 🔔 飞书推送
- 🎯 Bark 推送
- 📢 Discord 推送
- 以及更多平台...

### 新特性

✅ **多 Webhook 支持**：可以同时配置多个 webhook 地址，实现多平台并发推送  
✅ **向后兼容**：完全兼容原有的单个 webhook 配置  
✅ **优化链接预览**：改进消息格式，更好地支持链接预览功能  
✅ **独立配置**：每个 webhook 可以有独立的超时时间和重试次数  
✅ **灵活控制**：可以单独启用/禁用某个 webhook

### 快速配置

#### 1. 部署 message-pusher

参考 [message-pusher 官方文档](https://github.com/songquanpeng/message-pusher) 部署服务。

#### 2. 配置 message-pusher webhook

**详细配置请参考**：[📖 Message-Pusher 接口配置规则](docs/message-pusher-config.md)

**快速配置**：
1. 登录 message-pusher 后台
2. 进入"产品配置" -> "webhook 配置"
3. 点击"新建 webhook 通道"
4. 配置提取规则：
```json
{
  "title": "title",
  "description": "description", 
  "content": "content",
  "url": "url",
  "group": "group",
  "keywords": "keywords",
  "timestamp": "timestamp"
}
```
5. 配置构建规则（推荐）：
```json
{
  "content": "$content"
}
```
6. 复制生成的 webhook URL

#### 3. 配置 rss2tg

**环境变量方式**：
```bash
# 单个 webhook
WEBHOOK_ENABLED=true
WEBHOOK_URL=http://your-message-pusher:3000/webhook/your_webhook_id

# 多个 webhooks
WEBHOOK_URL_1=http://server1:3000/webhook/webhook_id_1
WEBHOOK_NAME_1=message-pusher-1
WEBHOOK_URL_2=http://server2:3000/webhook/webhook_id_2
WEBHOOK_NAME_2=message-pusher-2
```

**配置文件方式**：
```yaml
webhooks:
  - name: "message-pusher-1"
    enabled: true
    url: "http://server1:3000/webhook/webhook_id_1"
    timeout: 10
    retry_count: 3
  - name: "message-pusher-2"
    enabled: true
    url: "http://server2:3000/webhook/webhook_id_2"
    timeout: 15
    retry_count: 2
```

### Docker 部署示例

#### 基础部署（仅 Telegram 推送）
```bash
docker run -d \
  --name rss2tg \
  -e TELEGRAM_BOT_TOKEN=your_bot_token \
  -e TELEGRAM_USERS=123456789 \
  -e RSS_URLS_1=https://example.com/rss \
  -e RSS_KEYWORDS_1=关键词1,关键词2 \
  -e RSS_GROUP_1=技术资讯 \
  -v /path/to/data:/app/data \
  -v /path/to/config:/app/config \
  drfyup/rss2tg:latest
```

#### 单个 webhook 部署
```bash
docker run -d \
  --name rss2tg \
  -e TELEGRAM_BOT_TOKEN=your_bot_token \
  -e TELEGRAM_USERS=123456789 \
  -e RSS_URLS_1=https://example.com/rss \
  -e RSS_KEYWORDS_1=关键词1,关键词2 \
  -e RSS_GROUP_1=技术资讯 \
  -e WEBHOOK_ENABLED=true \
  -e WEBHOOK_URL=http://your-message-pusher:3000/webhook/your_webhook_id \
  -v /path/to/data:/app/data \
  -v /path/to/config:/app/config \
  drfyup/rss2tg:latest
```

#### 多个 webhook 部署
```bash
docker run -d \
  --name rss2tg \
  -e TELEGRAM_BOT_TOKEN=your_bot_token \
  -e TELEGRAM_USERS=123456789 \
  -e RSS_URLS_1=https://example.com/rss \
  -e RSS_KEYWORDS_1=关键词1,关键词2 \
  -e RSS_GROUP_1=技术资讯 \
  -e WEBHOOK_URL_1=http://server1:3000/webhook/webhook_id_1 \
  -e WEBHOOK_NAME_1=message-pusher-1 \
  -e WEBHOOK_URL_2=http://server2:3000/webhook/webhook_id_2 \
  -e WEBHOOK_NAME_2=message-pusher-2 \
  -v /path/to/data:/app/data \
  -v /path/to/config:/app/config \
  drfyup/rss2tg:latest
```

### 消息格式

#### Telegram 消息格式（保持不变）
```
📰 **文章标题**

🌐 **链接:** https://example.com/article

🔍 **关键词:** #关键词1 #关键词2

🏷️ **分组:** 技术资讯

🕒 **时间:** 2024-01-20 15:30:45
```

#### Webhook 消息格式
发送到 message-pusher 的数据格式：
```json
{
  "title": "文章标题",
  "description": "分组: 技术资讯 | 关键词: 关键词1, 关键词2 | 时间: 2024-01-20 15:30:45",
  "content": "### 📰 【技术资讯】RSS推送\n\n**标题：** 文章标题\n\nhttps://example.com/article\n\n**关键词：** #关键词1 #关键词2\n\n**时间：** 2024-01-20 15:30:45",
  "url": "https://example.com/article",
  "group": "技术资讯",
  "keywords": "关键词1, 关键词2",
  "timestamp": "2024-01-20 15:30:45"
}
```

## 命令说明

### 2.2 配置项说明

#### 2.2.1 配置文件字段说明

| 配置项                 | 类型       | 必填 | 说明                      | 示例                                           |
| ---------------------- | ---------- | ---- | ------------------------- | ---------------------------------------------- |
| telegram.bot_token     | 字符串     | 是   | Telegram Bot 的 API Token | "110201543:AAHdqTcvCH1vGWJxfSeofSAs0K5PALDsaw" |
| telegram.users         | 字符串数组 | 是   | 接收消息的用户 ID 列表    | ["123456789", "987654321"]                     |
| telegram.channels      | 字符串数组 | 否   | 接收消息的频道列表        | ["@channel1", "@channel2"]                     |
| telegram.adminuser     | 字符串数组 | 否   | 管理员用户 ID 列表        | ["123456789"]                                  |
| rss[].urls             | 字符串数组 | 是   | RSS 订阅地址列表          | ["https://example.com/feed1.xml"]              |
| rss[].interval         | 整数       | 是   | 更新间隔（秒）            | 300                                            |
| rss[].keywords         | 字符串数组 | 否   | 关键词列表                | ["vps", "优惠"]                                |
| rss[].group            | 字符串     | 否   | 分组名称                  | "科技新闻"                                     |
| rss[].allow_part_match | 布尔值     | 否   | 是否允许部分匹配          | true                                           |

#### 2.2.2 配置注意事项

1. **优先级说明**

   - 环境变量的优先级高于配置文件
   - 如果同时设置了环境变量和配置文件，将使用环境变量的值

2. **配置更新机制**

   - 系统每分钟自动检测配置文件变化
   - 配置文件变更后无需重启，自动生效

3. **关键词设置说明**

   - 如果配置了关键词，只有匹配关键词的文章才会被推送
   - 如果没有配置任何关键词，该订阅源的所有新文章都会被推送
   - 关键词匹配支持完整匹配和部分匹配两种模式
   - 可以通过 `allow_part_match` 配置是否允许部分匹配

4. **安全建议**
   - 不要在公开环境中暴露 bot_token
   - 建议设置 adminuser 限制管理权限
   - 定期更新和检查用户权限

### 2.3 权限说明

系统实现了基本的权限控制机制：

1. 管理员权限：

   - 如果未配置 `adminuser`，所有在 `users` 列表中的用户都具有管理员权限
   - 如果配置了 `adminuser`，则只有在该列表中的用户才具有管理员权限
   - 管理员可以执行所有操作，包括添加/删除用户、管理 RSS 订阅等

2. 普通用户权限：
   - 可以查看所有信息（配置、订阅列表、用户列表等）
   - 不能执行管理操作（添加/删除用户、管理 RSS 订阅）
   - 尝试执行管理操作时会收到提示："您不是系统管理员，无法操作"

### 2.4 命令说明

机器人支持以下命令：

主要命令：

- `/start` - 开始使用机器人并查看帮助信息
- `/view` - 查看类命令合集
- `/users` - 用户管理命令合集
- `/edit` - 编辑类命令合集

查看类命令（使用 `/view` 查看）：

- `/config` - 查看当前配置
- `/list` - 列出所有 RSS 订阅
- `/stats` - 查看推送统计
- `/version` - 获取当前版本信息

用户管理命令（使用 `/users` 查看）：

- `/add_user` - 添加用户（需要管理员权限）
- `/del_user` - 删除用户（需要管理员权限）
- `/list_users` - 查看用户列表

编辑类命令（使用 `/edit` 查看）：

- `/add` - 添加 RSS 订阅（需要管理员权限）
- `/edit` - 编辑 RSS 订阅（需要管理员权限）
- `/delete` - 删除 RSS 订阅（需要管理员权限）
- `/add_all` - 向所有订阅添加关键词（需要管理员权限）
- `/del_all` - 从所有订阅删除关键词（需要管理员权限）

### 2.5 用户管理

1. 添加用户（需要管理员权限）：

   - 使用 `/add_user` 命令
   - 输入要添加的用户 ID（多个 ID 用空格分隔）
   - 新添加的用户默认为普通用户权限

2. 删除用户（需要管理员权限）：

   - 使用 `/del_user` 命令
   - 查看当前用户列表
   - 输入要删除的用户编号

3. 查看用户列表：

   - 使用 `/list_users` 命令
   - 显示所有已添加的用户 ID

4. 设置管理员：
   - 在配置文件中添加 `adminuser` 字段
   - 或通过环境变量 `TELEGRAM_ADMIN_USERS` 设置
   - 多个管理员 ID 用逗号分隔

注意：如果未设置管理员，所有用户都具有管理员权限。建议在生产环境中明确设置管理员用户。

### 2.6 Bot 使用方法及命令

Bot 支持以下命令：

- `/start` - 开始使用机器人
- `/help` - 获取帮助信息
- `/config` - 查看当前配置
- `/add` - 添加 RSS 订阅
- `/edit` - 编辑 RSS 订阅
- `/delete` - 删除 RSS 订阅
- `/list` - 列出所有 RSS 订阅
- `/stats` - 查看推送统计

### 2.7 添加 RSS 订阅

#### 方式一：通过 Bot 命令

1. 发送 `/add` 命令给 Bot。
2. 按提示输入 RSS 订阅的 URL。
3. 输入更新间隔（秒）。
4. 输入关键词：
   - 输入 `1`：保持原有关键词（编辑时有效）
   - 输入 `2`：不设置关键词，该订阅源的所有新文章都会被推送
   - 直接输入关键词：输入多个关键词，用空格分隔，只有包含这些关键词的文章才会被推送
5. 输入组名。

#### 方式二：配置文件

在当前 config 目录下新建 config.yaml，填入以下内容。

```yaml
# Telegram 配置
telegram:
  bot_token: "your_telegram_bot_token"
  users:
    - "123456789"
  channels:
    - "@your_channel"
  adminuser:
    - "123456789"

# 单个 Webhook 配置（向后兼容）
webhook:
  enabled: false
  url: "http://your-message-pusher:3000/webhook/your_webhook_id"
  timeout: 10
  retry_count: 3

# 多个 Webhooks 配置（推荐使用）
webhooks:
  - name: "message-pusher-1"
    enabled: true
    url: "http://server1:3000/webhook/webhook_id_1"
    timeout: 10
    retry_count: 3
  - name: "message-pusher-2"
    enabled: true
    url: "http://server2:3000/webhook/webhook_id_2"
    timeout: 15
    retry_count: 2
  - name: "backup-webhook"
    enabled: false  # 可以暂时禁用某个 webhook
    url: "http://backup:3000/webhook/backup_id"
    timeout: 5
    retry_count: 1

# RSS 订阅配置
rss:
  - urls:
      - "https://rss.nodeseek.com"
    interval: 30
    keywords:
      - "vps"
      - "甲骨文"
      - "免费"
    group: "NS论坛"
    allow_part_match: true
  - urls:
      - "https://linux.do/latest.rss"
    interval: 30
    keywords:
      - "vps"
      - "甲骨文"
      - "免费"
      - "龟壳"
    group: "LC论坛"
    allow_part_match: true
```

**_两种方式都可以，系统会每 1 分钟自动检测，即使动态更改生效。_**

### 2.8 编辑 RSS 订阅

1. 发送 `/edit` 命令给 Bot。
2. 输入要编辑的 RSS 订阅编号。
3. 按提示修改 URL、更新间隔、关键词和组名。如果不需要修改某项，直接输入 1。

### 2.9 删除 RSS 订阅

1. 发送 `/delete` 命令给 Bot。
2. 输入要删除的 RSS 订阅编号。

### 2.10 查看订阅列表

发送 `/list` 命令给 Bot，查看当前所有 RSS 订阅。

### 2.11 查看推送统计

发送 `/stats` 命令给 Bot，查看今日和本周的推送数量。

## 3. 注意事项

- 确保 Docker 容器有足够的权限访问 `config` 和 `data` 目录。
- 如果修改了配置文件，需要重启 Docker 容器以使更改生效。
- 推送统计数据保存在 `/app/data/stats.yaml` 文件中。
- 已发送的项目记录保存在 `/app/data/sent_items.txt` 文件中。

## 4. 故障排查

### 基础问题

- 如果 Bot 无响应，请检查 Telegram Bot Token 是否正确。
- 如果无法接收消息，请确保已将您的用户 ID 添加到配置中。
- 查看 Docker 容器日志以获取更多信息：

```bash
docker logs rss2tg
```

### Webhook 相关问题

#### Webhook 推送失败
1. **检查 webhook URL 是否正确**
   ```bash
   # 检查日志中的错误信息
   docker logs rss2tg | grep -i webhook
   ```

2. **确认 message-pusher 服务是否正常运行**
   ```bash
   # 测试 message-pusher 服务连通性
   curl -I http://your-message-pusher:3000/health
   ```

3. **检查 message-pusher 配置**
   - 确保构建规则只使用了 `content` 和 `description` 字段
   - 详细配置请参考：[📖 Message-Pusher 接口配置规则](docs/message-pusher-config.md)

#### 配置不生效
1. **检查环境变量是否正确设置**
   ```bash
   # 查看容器环境变量
   docker exec rss2tg env | grep -E "(WEBHOOK|RSS|TELEGRAM)"
   ```

2. **确认配置文件格式是否正确**
   ```bash
   # 检查配置文件语法
   docker exec rss2tg cat /app/config/config.yaml
   ```

3. **重启容器使配置生效**
   ```bash
   docker restart rss2tg
   ```

#### 多 Webhook 配置问题
1. **环境变量命名错误**
   - 确保使用正确的命名格式：`WEBHOOK_URL_1`, `WEBHOOK_URL_2` 等
   - 数字从 1 开始，最多支持 10 个 webhook

2. **部分 webhook 失败**
   - 查看日志中每个 webhook 的推送结果
   - 每个 webhook 独立处理，一个失败不影响其他

#### 链接预览问题
1. **链接预览不显示**
   - 使用推荐的配置方案，将链接单独放在一行
   - 参考：[📖 Message-Pusher 接口配置规则](docs/message-pusher-config.md)

2. **格式显示异常**
   - 检查 Markdown 语法是否正确
   - 根据目标平台调整格式

如有其他问题，请参考项目的 GitHub 页面或提交 issue。

# 贴上一张效果图<br>

![image](https://github.com/user-attachments/assets/4e9ac180-5eb1-40a8-98e1-03b9fa68b691)

# English Version

# RSS to Telegram Bot usage documentation

#

[!TIP]
**_Referred to as rss2tg, it is used to instantly send related posts in custom RSS addresses, fields, and refresh times to custom TG users or channels, eliminating the time for you to swipe posts._**

**_Support AMD64/ARM64_**

**_Image size 17M, memory footprint 10M_**

**——By [drfyup](https://hstz.com)**

#

## 1. Deployment method

### 1.1 Use Docker Compose (recommended)

1. Make sure that Docker and Docker Compose are installed (the method is self-searching).

2. Clone or download the project code locally.

```bash
git clone https://github.com/3377/rss2tg.git
```

3. Enter the project directory.

4. Edit 'docker-compose.yml' file, modify environment variables：

-- Enter any directory or directly the current directory and create a new docker-compose.yml file, fill in the following content

```yaml
version: "3"
services:
  rss2tg:
    container_name: rss2tg
    image: drfyup/rss2tg:latest
    volumes:
      - ./config:/app/config
      - ./data:/app/data
    environment:
      - TELEGRAM_BOT_TOKEN=your_bot_token_here
      - TELEGRAM_USERS=user_id_1,user_id_2
      - TELEGRAM_CHANNELS=@channel_1,@channel_2
      - TZ=Asia/Shanghai
    restart: unless-stopped
```

Replace `your_bot_token_here` with your Telegram Bot Token, `user_id_1, user_id_2` with the user ID you want to receive the message, `@channel_1, @channel_2` with the channel name you want to send the message, `TELEGRAM_API_URL` with your custom Telegram API URL (useful for proxy servers in restricted regions), and `TZ=Asia/Shanghai` with your timezone settings.

5. Run the following command to start the container：

```yaml
docker-compose up  -d
```

### 1.2 Use Docker Run

1. Build a Docker image：

```yaml
docker pull drfyup/rss2tg:latest
```

2. Run the Docker container：

```yaml
docker run -d \
--name rss2tg \
-v $(pwd)/config:/app/config \
-v $(pwd)/data:/app/data \
-e TELEGRAM_BOT_TOKEN=your_bot_token_here \
-e TELEGRAM_USERS=user_id_1,user_id_2 \
-e TELEGRAM_CHANNELS=@channel_1,@channel_2 \
-e TELEGRAM_API_URL=http://fyapi.deno.dev/telegram \
-e TZ=Asia/Shanghai \
--restart unless-stopped \
drfyup/rss2tg:latest
```

Please replace the corresponding value in the environment variable.

## 2. Program instructions

### 2.1 Configuration file

The program supports configuration through YAML configuration files or environment variables.The configuration file is located in`/app/config/config.yaml`.If the file does not exist, the program will use environment variables for initial configuration.
The reading priority of environment variables is higher than that of configuration files.

Configuration file example：

```yaml
telegram:
  bot_token: "your_bot_token_here"
  users:
    - "user_id_1"
    - "user_id_2"
  channels:
    - "@channel_1"
    - "@channel_2"
  adminuser: # 管理员用户配置（可选）
    - "admin_id_1"
    - "admin_id_2"

rss:
  - urls:
      - "https://example.com/feed1.xml"
      - "https://example.com/feed2.xml"
    interval: 300
    keywords:
      - "keyword1"
      - "keyword2"
    group: "Group1"
    allow_part_match: true # 是否允许部分关键词匹配
```

### 2.2 How to use Bot and commands

The Bot supports the following commands：

-`/start`-Start using the robot -`/help`-Get help information -`/config`-View current configuration -`/add`-add RSS subscription -`/edit`-edit RSS feed -`/delete`-delete RSS feed -`/list`-list all RSS feeds -`/statistics`-View push statistics

### 2.3 Add RSS feed

#### Method One

1. Send the `/add' command to the Bot.
2. Press the prompt to enter the URL of the RSS subscription.
3. Enter the update interval (seconds).
4. Enter keywords (separated by commas, if not, you can directly enter 1).
5. Enter the group name.

#### Method Two

Create a new config in the current config directory.ymal, fill in the following.

```yaml
# Telegram 配置
telegram:
  bot_token: "your_telegram_bot_token"
  users:
    - "123456789"
  channels:
    - "@your_channel"
  adminuser:
    - "123456789"

# 单个 Webhook 配置（向后兼容）
webhook:
  enabled: false
  url: "http://your-message-pusher:3000/webhook/your_webhook_id"
  timeout: 10
  retry_count: 3

# 多个 Webhooks 配置（推荐使用）
webhooks:
  - name: "message-pusher-1"
    enabled: true
    url: "http://server1:3000/webhook/webhook_id_1"
    timeout: 10
    retry_count: 3
  - name: "message-pusher-2"
    enabled: true
    url: "http://server2:3000/webhook/webhook_id_2"
    timeout: 15
    retry_count: 2
  - name: "backup-webhook"
    enabled: false  # 可以暂时禁用某个 webhook
    url: "http://backup:3000/webhook/backup_id"
    timeout: 5
    retry_count: 1

# RSS 订阅配置
rss:
  - urls:
      - "https://rss.nodeseek.com"
    interval: 30
    keywords:
      - "vps"
      - "甲骨文"
      - "免费"
    group: "NS论坛"
    allow_part_match: true
  - urls:
      - "https://linux.do/latest.rss"
    interval: 30
    keywords:
      - "vps"
      - "甲骨文"
      - "免费"
      - "龟壳"
    group: "LC论坛"
    allow_part_match: true
```

**_两种方式都可以，系统会每 1 分钟自动检测，即使动态更改生效。_**

### 2.4 Edit RSS feed

1. Send the `/edit' command to the Bot.
2. Enter the RSS subscription number you want to edit.
3. Follow the prompts to modify the URL, update interval, keywords, and group name.If you don't need to modify an item, enter 1 directly.

### 2.5 Delete RSS feed

1. Send the `/delete' command to the Bot.
2. Enter the RSS subscription number you want to delete.

### 2.6 View subscription list

Send the `/list' command to the Bot to view all current RSS feeds.

### 2.7 View push statistics

Send the `/statistics' command to the Bot to check the number of pushes for today and this week.

## 3. Precautions

-Make sure that the Docker container has sufficient permissions to access the 'config` and'data` directories.
-If the configuration file is modified, the Docker container needs to be restarted for the changes to take effect.
-Push statistics are saved in`/app/data/statistics.In the yaml' file.
-The sent project records are saved in`/app/data/sent_items.txt` file.

## 4. Troubleshooting

-If the Bot is unresponsive, please check whether the Telegram Bot token is correct.
-If the message cannot be received, please make sure that your user ID has been added to the configuration.
-View the Docker container log for more information：

```bash
docker logs rss2tg
```

If you have other questions, please refer to the project's GitHub page or submit an issue.

# Paste a rendering <br>![image](https://github.com/user-attachments/assets/4e9ac180-5eb1-40a8-98e1-03b9fa68b691)

## 5. 在中国大陆服务器上使用

在中国大陆服务器上部署 RSS2TG 机器人时，由于网络限制，可能无法直接访问 Telegram 官方 API。为了解决这个问题，本项目支持通过自定义 API URL 使用代理服务进行通信。

### 5.1 使用代理

1. 设置环境变量 `TELEGRAM_API_URL`，指向可用的代理服务地址，例如：
   ```bash
   TELEGRAM_API_URL=http://fyapi.deno.dev/telegram
   ```

2. 在 docker-compose.yml 中配置：
   ```yaml
   environment:
     - TELEGRAM_BOT_TOKEN=your_bot_token_here
     - TELEGRAM_USERS=user_id_1,user_id_2
     - TELEGRAM_API_URL=http://fyapi.deno.dev/telegram
     - TZ=Asia/Shanghai
   ```

3. 应用重启后，机器人将通过配置的代理地址与 Telegram 进行通信。

### 5.2 注意事项

- 确保代理服务稳定可靠，否则可能导致消息发送失败
- 定期检查日志，确保通信正常
- 可能需要根据代理服务的要求进行额外设置

如果您使用 `http://fyapi.deno.dev/telegram` 作为代理，通常无需额外配置即可使用。

## 5. Using in Mainland China

When deploying RSS2TG bot on servers in mainland China, due to network restrictions, direct access to the official Telegram API might be unavailable. To solve this issue, this project supports using a proxy service through a custom API URL.

### 5.1 Using a Proxy

1. Set the environment variable `TELEGRAM_API_URL` pointing to an available proxy service address, for example:
   ```bash
   TELEGRAM_API_URL=http://fyapi.deno.dev/telegram
   ```

2. Configure in docker-compose.yml:
   ```yaml
   environment:
     - TELEGRAM_BOT_TOKEN=your_bot_token_here
     - TELEGRAM_USERS=user_id_1,user_id_2
     - TELEGRAM_API_URL=http://fyapi.deno.dev/telegram
     - TZ=Asia/Shanghai
   ```

3. After restarting the application, the bot will communicate with Telegram through the configured proxy address.

### 5.2 Important Notes

- Ensure the proxy service is stable and reliable, otherwise message delivery may fail
- Regularly check logs to ensure communication is normal
- Additional configuration may be required depending on the proxy service's requirements

If you use `http://fyapi.deno.dev/telegram` as a proxy, typically no additional configuration is needed.

## 技术特性

- ✅ **零侵入性**：完全不影响现有 Telegram 推送功能
- ✅ **异步推送**：webhook 推送失败不影响 Telegram 推送
- ✅ **自动重试**：支持配置重试次数和超时时间
- ✅ **热重载**：支持配置文件和环境变量热重载
- ✅ **统一格式**：所有平台接收相同格式的消息内容
- ✅ **多平台支持**：通过 message-pusher 支持邮件、企业微信、钉钉等多种推送方式
- ✅ **并发推送**：多个 webhook 同时推送，互不影响
- ✅ **独立配置**：每个 webhook 可以有独立的超时时间和重试次数

## 许可证

MIT License
