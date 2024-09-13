[English](#English) | [简体中文](#简体中文文档)
# 简体中文文档
# RSS to Telegram Bot 使用文档

#
>
> [!TIP]
> ***简称rss2tg，用于将自定义RSS地址，字段，刷新时间里的相关帖子即时发送到自定义TG用户或频道，省去你刷帖子的时间*** <br>
> ***支持AMD64/ARM64*** <br>
> ***镜像大小17M，内存占用10M*** <br>
> **——By [drfyup](https://hstz.com)**
>
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

# 贴上一张效果图<br>

![image](https://github.com/user-attachments/assets/4e9ac180-5eb1-40a8-98e1-03b9fa68b691)

# English
# RSS to Telegram Bot usage documentation
#
> [!TIP]
>*** Referred to as rss2tg, it is used to instantly send related posts in custom RSS addresses, fields, and refresh times to custom TG users or channels, eliminating the time for you to swipe posts.*** <br>
>*** Support AMD64/ARM64*** <br>
>*** Image size 17M, memory footprint 10M*** <br>
>  **——By [drfyup](https://hstz.com)**
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

Replace `your_bot_token_here` with your Telegram Bot Token, `user_id_1, user_id_2` with the user ID you want to receive the message, and `@channel_1, @channel_2` with the channel name you want to send the message, `TZ=Asia/Shanghai` with your timezone settings.

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
#The priority of the tg configuration here is lower than that of the environment variable. If you fill in here, this configuration will not be read until one minute later.
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

### 2.2 How to use Bot and commands

The Bot supports the following commands：

-`/start`-Start using the robot
-`/help`-Get help information
-`/config`-View current configuration
-`/add`-add RSS subscription
-`/edit`-edit RSS feed
-`/delete`-delete RSS feed
-`/list`-list all RSS feeds
-`/statistics`-View push statistics

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
rss:
- url: https://rss.nodeseek.com
  interval: 30
  keywords:
  - vps
  -Oracle
  -Free
  group: NS Forum
- url: https://linux.do/latest.rss
  interval: 30
  keywords:
  - vps
  -Oracle
  -Free
  -Turtle shell
  group: LC Forum
```
***Both methods are possible, the system will automatically detect every 1 minute, even if the dynamic changes take effect.***

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

# Paste a rendering <br>

![image](https://github.com/user-attachments/assets/4e9ac180-5eb1-40a8-98e1-03b9fa68b691)
