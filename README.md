# RSS to Telegram Bot 使用文档
#
> [!TIP]
> ***简称rss2tg，用于将自定义RSS地址，字段，刷新时间里的相关帖子即时发送到自定义TG用户或频道，省去你刷帖子的时间***
> **支持AMD64/ARM64**
> **镜像大小17M，内存占用10M**
]> **——By [drfyup](https://hstz.com)**
#

## 1. 部署方法

### 1.1 使用 Docker Compose（推荐）

1. 确保已安装 Docker 和 Docker Compose（方法自寻）。

2. 克隆或下载项目代码到本地。

```bash
git clone https://github.com/3377/rss2tg.git
```

3. 进入项目目录。

4. 编辑 `docker-compose.yml` 文件，修改环境变量：

-- 进入任意目录或直接当前目录，新建docker-compose.yml文件，填入以下内容

```yaml
version: '3'
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
#这里的tg配置优先级低于环境变量，如果填在这里，一分钟后才会读取此配置
rss:
  - url: "https://example.com/feed1.xml"
    interval: 300
    keywords:
      - "keyword1"
      - "keyword2"
    group: "Group1"
  - url: "https://example.com/feed2.xml"
    interval: 600
    keywords:
      - "keyword3"
    group: "Group2"
```

### 2.2 Bot 使用方法及命令

Bot 支持以下命令：

- `/start` - 开始使用机器人
- `/help` - 获取帮助信息
- `/config` - 查看当前配置
- `/add` - 添加 RSS 订阅
- `/edit` - 编辑 RSS 订阅
- `/delete` - 删除 RSS 订阅
- `/list` - 列出所有 RSS 订阅
- `/stats` - 查看推送统计

### 2.3 添加 RSS 订阅

#### 方式一

1. 发送 `/add` 命令给 Bot。
2. 按提示输入 RSS 订阅的 URL。
3. 输入更新间隔（秒）。
4. 输入关键词（用逗号分隔，如果没有可以直接输入 1）。
5. 输入组名。

#### 方式二

在当前config目录下新建config.ymal，填入以下内容。

```yaml
rss:
- url: https://rss.nodeseek.com
  interval: 30
  keywords:
  - vps
  - 甲骨文
  - 免费
  group: NS论坛
- url: https://linux.do/latest.rss
  interval: 30
  keywords:
  - vps
  - 甲骨文
  - 免费
  - 龟壳
  group: LC论坛
```
***两种方式都可以，系统会每1分钟自动检测，即使动态更改生效。***

### 2.4 编辑 RSS 订阅

1. 发送 `/edit` 命令给 Bot。
2. 输入要编辑的 RSS 订阅编号。
3. 按提示修改 URL、更新间隔、关键词和组名。如果不需要修改某项，直接输入 1。

### 2.5 删除 RSS 订阅

1. 发送 `/delete` 命令给 Bot。
2. 输入要删除的 RSS 订阅编号。

### 2.6 查看订阅列表

发送 `/list` 命令给 Bot，查看当前所有 RSS 订阅。

### 2.7 查看推送统计

发送 `/stats` 命令给 Bot，查看今日和本周的推送数量。

## 3. 注意事项

- 确保 Docker 容器有足够的权限访问 `config` 和 `data` 目录。
- 如果修改了配置文件，需要重启 Docker 容器以使更改生效。
- 推送统计数据保存在 `/app/data/stats.yaml` 文件中。
- 已发送的项目记录保存在 `/app/data/sent_items.txt` 文件中。

## 4. 故障排查

- 如果 Bot 无响应，请检查 Telegram Bot Token 是否正确。
- 如果无法接收消息，请确保已将您的用户 ID 添加到配置中。
- 查看 Docker 容器日志以获取更多信息：

```bash
docker logs rss2tg
```

如有其他问题，请参考项目的 GitHub 页面或提交 issue。
