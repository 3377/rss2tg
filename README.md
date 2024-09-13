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

4. Edit`docker-compose.yml' file, modify environment variables：

--Enter any directory or directly the current directory and create a new docker-compose.yml file, fill in the following content

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

Replace `your_bot_token_here` with your Telegram Bot Token, `user_id_1, user_id_2` with the user ID you want to receive the message, and `@channel_1, @channel_2` with the channel name you want to send the message.

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

###2.7 View push statistics

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
