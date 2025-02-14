package bot

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strconv"
    "strings"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "rss2tg/internal/config"
    "rss2tg/internal/storage"
    "rss2tg/internal/stats"
)

type MessageHandler func(title, url, group string, pubDate time.Time, matchedKeywords []string) error

type Bot struct {
    api              *tgbotapi.BotAPI
    users            []int64
    channels         []string
    db               *storage.Storage
    config           *config.Config
    configFile       string
    stats            *stats.Stats
    userState        map[int64]string
    messageHandler   MessageHandler
    updateRSSHandler func()
}

func NewBot(token string, users []string, channels []string, db *storage.Storage, config *config.Config, configFile string, stats *stats.Stats) (*Bot, error) {
    api, err := tgbotapi.NewBotAPI(token)
    if err != nil {
        return nil, err
    }

    userIDs := make([]int64, len(users))
    for i, user := range users {
        userID, err := strconv.ParseInt(user, 10, 64)
        if err != nil {
            return nil, fmt.Errorf("无效的用户ID: %s", user)
        }
        userIDs[i] = userID
    }

    return &Bot{
        api:              api,
        users:            userIDs,
        channels:         channels,
        db:               db,
        config:           config,
        configFile:       configFile,
        stats:            stats,
        userState:        make(map[int64]string),
        updateRSSHandler: func() {}, // 初始化为空函数
    }, nil
}

func (b *Bot) SetMessageHandler(handler MessageHandler) {
    b.messageHandler = handler
}

func (b *Bot) SetUpdateRSSHandler(handler func()) {
    b.updateRSSHandler = handler
}

func (b *Bot) Start() {
    log.Println("机器人已启动")
    
    commands := []tgbotapi.BotCommand{
        {Command: "start", Description: "开始/帮助"},
        {Command: "view", Description: "查看类命令"},
        {Command: "edit", Description: "编辑类命令"},
        {Command: "stats", Description: "推送统计"},
    }
    
    setMyCommandsConfig := tgbotapi.NewSetMyCommands(commands...)
    _, err := b.api.Request(setMyCommandsConfig)
    if err != nil {
        log.Printf("设置命令失败: %v", err)
    }

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := b.api.GetUpdatesChan(u)

    for update := range updates {
        if update.CallbackQuery != nil {
            // 处理按钮点击
            chatID := update.CallbackQuery.Message.Chat.ID
            userID := update.CallbackQuery.From.ID
            
            switch update.CallbackQuery.Data {
            case "config":
                b.handleConfig(chatID)
            case "list":
                b.handleList(chatID)
            case "stats":
                b.handleStats(chatID)
            case "version":
                b.handleVersion(chatID)
            case "add":
                b.handleAdd(chatID, userID)
            case "edit":
                b.handleEdit(chatID, userID)
            case "delete":
                b.handleDelete(chatID, userID)
            case "add_all":
                b.handleAddAll(chatID, userID)
            case "del_all":
                b.handleDelAll(chatID, userID)
            }
            
            // 回应按钮点击
            callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
            if _, err := b.api.Request(callback); err != nil {
                log.Printf("回应按钮点击失败: %v", err)
            }
            
            continue
        }

        if update.Message == nil {
            continue
        }

        userID := update.Message.From.ID
        chatID := update.Message.Chat.ID

        if update.Message.IsCommand() {
            switch update.Message.Command() {
            case "start":
                b.handleStart(chatID)
            case "stats":
                b.handleStats(chatID)
            case "view":
                b.handleView(chatID, userID)
            case "edit":
                b.handleEditCommand(chatID, userID)
            case "config":
                b.handleConfig(chatID)
            case "list":
                b.handleList(chatID)
            case "version":
                b.handleVersion(chatID)
            case "add":
                b.handleAdd(chatID, userID)
            case "delete":
                b.handleDelete(chatID, userID)
            default:
                b.sendMessage(chatID, "未知命令，请使用 /start 查看可用命令。")
            }
        } else {
            b.handleUserInput(update.Message)
        }
    }
}

// escapeMarkdownV2Text 转义普通文本中的特殊字符
func escapeMarkdownV2Text(text string) string {
    // 按照 Telegram MarkdownV2 格式要求转义特殊字符
    // 参考: https://core.telegram.org/bots/api#markdownv2-style
    text = strings.ReplaceAll(text, "\\", "\\\\") // 必须首先转义反斜杠
    text = strings.ReplaceAll(text, "_", "\\_")
    text = strings.ReplaceAll(text, "*", "\\*")
    text = strings.ReplaceAll(text, "[", "\\[")
    text = strings.ReplaceAll(text, "]", "\\]")
    text = strings.ReplaceAll(text, "(", "\\(")
    text = strings.ReplaceAll(text, ")", "\\)")
    text = strings.ReplaceAll(text, "~", "\\~")
    text = strings.ReplaceAll(text, "`", "\\`")
    text = strings.ReplaceAll(text, ">", "\\>")
    text = strings.ReplaceAll(text, "#", "\\#")
    text = strings.ReplaceAll(text, "+", "\\+")
    text = strings.ReplaceAll(text, "-", "\\-")
    text = strings.ReplaceAll(text, "=", "\\=")
    text = strings.ReplaceAll(text, "|", "\\|")
    text = strings.ReplaceAll(text, "{", "\\{")
    text = strings.ReplaceAll(text, "}", "\\}")
    text = strings.ReplaceAll(text, ".", "\\.")
    text = strings.ReplaceAll(text, "!", "\\!")
    return text
}

// formatBoldText 格式化加粗文本
func formatBoldText(text string) string {
    if text == "" {
        return "*无*"
    }
    return "*" + escapeMarkdownV2Text(text) + "*"
}

func (b *Bot) SendMessage(title, url, group string, pubDate time.Time, matchedKeywords []string) error {
    chinaLoc, _ := time.LoadLocation("Asia/Shanghai")
    pubDateChina := pubDate.In(chinaLoc)
    
    // 处理标题（加粗）
    formattedTitle := formatBoldText(title)
    
    // 处理URL（转义所有特殊字符）
    formattedURL := url
    formattedURL = strings.ReplaceAll(formattedURL, "\\", "\\\\")
    formattedURL = strings.ReplaceAll(formattedURL, ".", "\\.")
    formattedURL = strings.ReplaceAll(formattedURL, "(", "\\(")
    formattedURL = strings.ReplaceAll(formattedURL, ")", "\\)")
    formattedURL = strings.ReplaceAll(formattedURL, "!", "\\!")
    formattedURL = strings.ReplaceAll(formattedURL, "-", "\\-")
    formattedURL = strings.ReplaceAll(formattedURL, "_", "\\_")
    
    // 处理关键词（加粗并添加#）
    formattedKeywords := make([]string, len(matchedKeywords))
    for i, keyword := range matchedKeywords {
        formattedKeywords[i] = "\\#" + formatBoldText(keyword)
    }
    
    // 处理分组（加粗）
    formattedGroup := formatBoldText(group)
    
    // 处理时间（加粗，需要转义所有特殊字符）
    timeStr := pubDateChina.Format("2006-01-02 15:04:05")
    timeStr = strings.ReplaceAll(timeStr, "-", "\\-")
    timeStr = strings.ReplaceAll(timeStr, ":", "\\:")
    timeStr = strings.ReplaceAll(timeStr, ".", "\\.")
    formattedTime := "*" + timeStr + "*"
    
    // 构建消息文本（标签文本也需要加粗）
    text := fmt.Sprintf("%s\n\n🌐 *链接:* %s\n\n🔍 *关键词:* %s\n\n🏷️ *分组:* %s\n\n🕒 *时间:* %s", 
        formattedTitle,
        formattedURL,
        strings.Join(formattedKeywords, " "),
        formattedGroup,
        formattedTime)
    
    log.Printf("发送消息: %s", text)

    // 发送消息
    for _, userID := range b.users {
        msg := tgbotapi.NewMessage(userID, text)
        msg.ParseMode = "MarkdownV2"
        if _, err := b.api.Send(msg); err != nil {
            log.Printf("发送消息给用户 %d 失败: %v", userID, err)
        } else {
            log.Printf("成功发送消息给用户 %d", userID)
            b.stats.IncrementMessageCount()
        }
    }

    for _, channel := range b.channels {
        msg := tgbotapi.NewMessageToChannel(channel, text)
        msg.ParseMode = "MarkdownV2"
        if _, err := b.api.Send(msg); err != nil {
            log.Printf("发送消息到频道 %s 失败: %v", channel, err)
        } else {
            log.Printf("成功发送消息到频道 %s", channel)
            b.stats.IncrementMessageCount()
        }
    }

    return nil
}

func (b *Bot) reloadConfig() error {
    newConfig, err := config.Load(b.configFile)
    if err != nil {
        return err
    }
    b.config = newConfig
    return nil
}

func (b *Bot) handleStart(chatID int64) {
    helpText := "欢迎使用RSS订阅机器人！\n\n" +
        "主要命令：\n" +
        "/start \\- 开始使用机器人并查看帮助信息\n" +
        "/stats \\- 查看推送统计\n" +
        "/view \\- 查看类命令合集\n" +
        "/edit \\- 编辑类命令合集\n\n" +
        "查看类命令（使用 /view 查看）：\n" +
        "/config \\- 查看当前配置\n" +
        "/list \\- 列出所有RSS订阅\n" +
        "/stats \\- 查看推送统计\n" +
        "/version \\- 获取当前版本信息\n\n" +
        "编辑类命令（使用 /edit 查看）：\n" +
        "/add \\- 添加RSS订阅\n" +
        "/edit \\- 编辑RSS订阅\n" +
        "/delete \\- 删除RSS订阅\n" +
        "/add_all \\- 向所有订阅添加关键词\n" +
        "/del_all \\- 从所有订阅删除关键词"
    
    // 转义特殊字符，但保持命令格式
    helpText = strings.ReplaceAll(helpText, "!", "\\!")
    helpText = strings.ReplaceAll(helpText, "(", "\\(")
    helpText = strings.ReplaceAll(helpText, ")", "\\)")
    
    msg := tgbotapi.NewMessage(chatID, helpText)
    msg.ParseMode = "MarkdownV2"
    if _, err := b.api.Send(msg); err != nil {
        log.Printf("发送消息失败: %v", err)
    }
}

func (b *Bot) handleView(chatID int64, userID int64) {
    text := "查看类命令列表："
    
    // 创建按钮列表
    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("📋 查看当前配置", "config"),
            tgbotapi.NewInlineKeyboardButtonData("📜 列出RSS订阅", "list"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("📊 查看推送统计", "stats"),
            tgbotapi.NewInlineKeyboardButtonData("ℹ️ 获取当前版本", "version"),
        ),
    )

    msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2Text(text))
    msg.ParseMode = "MarkdownV2"
    msg.ReplyMarkup = keyboard
    if _, err := b.api.Send(msg); err != nil {
        log.Printf("发送消息失败: %v", err)
    }
}

func (b *Bot) handleEditCommand(chatID int64, userID int64) {
    text := "编辑类命令列表："
    
    // 创建按钮列表
    keyboard := tgbotapi.NewInlineKeyboardMarkup(
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("➕ 添加RSS订阅", "add"),
            tgbotapi.NewInlineKeyboardButtonData("✏️ 编辑RSS订阅", "edit"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("❌ 删除RSS订阅", "delete"),
            tgbotapi.NewInlineKeyboardButtonData("📝 添加全局关键词", "add_all"),
        ),
        tgbotapi.NewInlineKeyboardRow(
            tgbotapi.NewInlineKeyboardButtonData("🗑️ 删除全局关键词", "del_all"),
        ),
    )

    msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2Text(text))
    msg.ParseMode = "MarkdownV2"
    msg.ReplyMarkup = keyboard
    if _, err := b.api.Send(msg); err != nil {
        log.Printf("发送消息失败: %v", err)
    }
}

func (b *Bot) handleConfig(chatID int64) {
    log.Printf("正在处理查看配置请求，chatID: %d", chatID)
    if err := b.reloadConfig(); err != nil {
        log.Printf("加载配置失败: %v", err)
        b.sendMessage(chatID, fmt.Sprintf("加载配置时出错：%v\n请检查配置文件格式是否正确。", err))
        return
    }
    
    config := b.getConfig()
    if config == "" {
        b.sendMessage(chatID, "当前没有配置信息或配置为空")
        return
    }
    
    b.sendMessage(chatID, config)
    log.Printf("成功发送配置信息到chatID: %d", chatID)
}

func (b *Bot) handleAdd(chatID int64, userID int64) {
    b.userState[userID] = "add_url"
    message := b.listSubscriptions()
    message += "\n请输入要添加的RSS订阅URL（如需添加多个URL，请用英文逗号分隔）："
    
    msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2Text(message))
    msg.ParseMode = "MarkdownV2"
    if _, err := b.api.Send(msg); err != nil {
        log.Printf("发送消息失败: %v", err)
    }
}

func (b *Bot) handleEdit(chatID int64, userID int64) {
    b.userState[userID] = "edit_index"
    message := b.listSubscriptions()
    message += "\n请输入要编辑的RSS订阅编号："
    
    msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2Text(message))
    msg.ParseMode = "MarkdownV2"
    if _, err := b.api.Send(msg); err != nil {
        log.Printf("发送消息失败: %v", err)
    }
}

func (b *Bot) handleDelete(chatID int64, userID int64) {
    b.userState[userID] = "delete"
    message := b.listSubscriptions()
    message += "\n请输入要删除的RSS订阅编号："
    
    msg := tgbotapi.NewMessage(chatID, escapeMarkdownV2Text(message))
    msg.ParseMode = "MarkdownV2"
    if _, err := b.api.Send(msg); err != nil {
        log.Printf("发送消息失败: %v", err)
    }
}

func (b *Bot) handleList(chatID int64) {
    log.Printf("正在处理列表请求，chatID: %d", chatID)
    if err := b.reloadConfig(); err != nil {
        log.Printf("加载配置失败: %v", err)
        b.sendMessage(chatID, fmt.Sprintf("加载配置时出错：%v\n请检查配置文件格式是否正确。", err))
        return
    }
    
    list := b.listSubscriptions()
    if list == "" {
        b.sendMessage(chatID, "当前没有RSS订阅")
        return
    }
    
    b.sendMessage(chatID, list)
    log.Printf("成功发送订阅列表到chatID: %d", chatID)
}

func (b *Bot) handleStats(chatID int64) {
    b.sendMessage(chatID, b.getStats())
}

func (b *Bot) handleUserInput(message *tgbotapi.Message) {
    userID := message.From.ID
    chatID := message.Chat.ID
    text := message.Text

    switch b.userState[userID] {
    case "view_command":
        switch text {
        case "1":
            b.handleConfig(chatID)
        case "2":
            b.handleStats(chatID)
        case "3":
            b.handleList(chatID)
        case "4":
            b.handleVersion(chatID)
        default:
            b.sendMessage(chatID, "无效的命令编号，请使用 /view 重新选择。")
        }
        delete(b.userState, userID)
    case "edit_command":
        switch text {
        case "1":
            b.handleAdd(chatID, userID)
        case "2":
            b.handleEdit(chatID, userID)
        case "3":
            b.handleDelete(chatID, userID)
        case "4":
            b.handleAddAll(chatID, userID)
        case "5":
            b.handleDelAll(chatID, userID)
        default:
            b.sendMessage(chatID, "无效的命令编号，请使用 /edit 重新选择。")
            delete(b.userState, userID)
            return
        }
    case "add_url":
        b.userState[userID] = "add_interval"
        urls := strings.Split(text, ",")
        // 清理URL列表
        cleanURLs := make([]string, 0)
        for _, url := range urls {
            url = strings.TrimSpace(url)
            if url != "" {
                cleanURLs = append(cleanURLs, url)
            }
        }
        // 创建新的RSSEntry并添加到配置中
        b.config.RSS = append(b.config.RSS, config.RSSEntry{
            URLs: cleanURLs,
        })
        b.sendMessage(chatID, "请输入订阅的更新间隔（秒）：")
    case "add_interval":
        interval, err := strconv.Atoi(text)
        if err != nil {
            b.sendMessage(chatID, "无效的间隔时间，请输入一个整数。")
            return
        }
        lastIndex := len(b.config.RSS) - 1
        if lastIndex >= 0 {
            b.config.RSS[lastIndex].Interval = interval
            b.userState[userID] = "add_keywords"
            b.sendMessage(chatID, "请输入关键词（用空格分隔，如果没有可以直接输入1）：")
        } else {
            b.sendMessage(chatID, "添加订阅失败：找不到要编辑的订阅")
            delete(b.userState, userID)
        }
    case "add_keywords":
        lastIndex := len(b.config.RSS) - 1
        if lastIndex >= 0 {
            if text != "1" {
                keywords := strings.Fields(text) // 使用 Fields 替代 Split，自动按空格分割
                b.config.RSS[lastIndex].Keywords = keywords
            }
            b.userState[userID] = "add_group"
            b.sendMessage(chatID, "请输入组名：")
        } else {
            b.sendMessage(chatID, "添加订阅失败：找不到要编辑的订阅")
            delete(b.userState, userID)
        }
    case "add_group":
        lastIndex := len(b.config.RSS) - 1
        if lastIndex >= 0 {
            b.config.RSS[lastIndex].Group = text
            delete(b.userState, userID)
            if err := b.config.Save(b.configFile); err != nil {
                b.sendMessage(chatID, "添加订阅成功，但保存配置失败。")
            } else {
                b.sendMessage(chatID, "成功添加RSS订阅。")
                b.updateRSSHandler()
            }
        } else {
            b.sendMessage(chatID, "添加订阅失败：找不到要编辑的订阅")
            delete(b.userState, userID)
        }
    case "edit_index":
        index, err := strconv.Atoi(text)
        if err != nil || index < 1 || index > len(b.config.RSS) {
            b.sendMessage(chatID, "无效的编号。请使用 /edit 重新开始。")
            delete(b.userState, userID)
            return
        }
        b.userState[userID] = fmt.Sprintf("edit_url_%d", index-1)
        b.sendMessage(chatID, fmt.Sprintf("当前URL列表为：\n%s\n请输入新的URL列表（多个URL用英文逗号分隔，如不修改请输入1）：", 
            strings.Join(b.config.RSS[index-1].URLs, "\n")))
    case "delete":
        index, err := strconv.Atoi(text)
        if err != nil || index < 1 || index > len(b.config.RSS) {
            b.sendMessage(chatID, "无效的编号。请使用 /delete 重新开始。")
            delete(b.userState, userID)
            return
        }
        deletedRSS := b.config.RSS[index-1]
        b.config.RSS = append(b.config.RSS[:index-1], b.config.RSS[index:]...)
        if err := b.config.Save(b.configFile); err != nil {
            b.sendMessage(chatID, "删除订阅成功，但保存配置失败。")
        } else {
            b.sendMessage(chatID, fmt.Sprintf("成功删除订阅: %v", deletedRSS.URLs))
            b.updateRSSHandler()
        }
        delete(b.userState, userID)
    case "add_all_keywords":
        keywords := strings.Fields(text)
        if len(keywords) == 0 {
            b.sendMessage(chatID, "请输入至少一个关键词。")
            return
        }
        
        // 向所有订阅添加关键词
        for i := range b.config.RSS {
            existingKeywords := make(map[string]bool)
            for _, k := range b.config.RSS[i].Keywords {
                existingKeywords[strings.ToLower(k)] = true
            }
            
            // 添加新关键词（避免重复）
            for _, newKeyword := range keywords {
                if !existingKeywords[strings.ToLower(newKeyword)] {
                    b.config.RSS[i].Keywords = append(b.config.RSS[i].Keywords, newKeyword)
                }
            }
        }
        
        if err := b.config.Save(b.configFile); err != nil {
            b.sendMessage(chatID, "添加关键词成功，但保存配置失败。")
        } else {
            b.sendMessage(chatID, fmt.Sprintf("成功向所有订阅添加关键词：%v", keywords))
            b.updateRSSHandler()
        }
        delete(b.userState, userID)
        
    case "del_all_keywords":
        keywords := strings.Fields(text)
        if len(keywords) == 0 {
            b.sendMessage(chatID, "请输入至少一个关键词。")
            return
        }
        
        // 从所有订阅中删除关键词
        keywordsToRemove := make(map[string]bool)
        for _, k := range keywords {
            keywordsToRemove[strings.ToLower(k)] = true
        }
        
        for i := range b.config.RSS {
            newKeywords := make([]string, 0)
            for _, k := range b.config.RSS[i].Keywords {
                if !keywordsToRemove[strings.ToLower(k)] {
                    newKeywords = append(newKeywords, k)
                }
            }
            b.config.RSS[i].Keywords = newKeywords
        }
        
        if err := b.config.Save(b.configFile); err != nil {
            b.sendMessage(chatID, "删除关键词成功，但保存配置失败。")
        } else {
            b.sendMessage(chatID, fmt.Sprintf("成功从所有订阅中删除关键词：%v", keywords))
            b.updateRSSHandler()
        }
        delete(b.userState, userID)
    default:
        if strings.HasPrefix(b.userState[userID], "edit_url_") {
            index, _ := strconv.Atoi(strings.TrimPrefix(b.userState[userID], "edit_url_"))
            if text != "1" {
                urls := strings.Split(text, ",")
                // 清理URL列表
                cleanURLs := make([]string, 0)
                for _, url := range urls {
                    url = strings.TrimSpace(url)
                    if url != "" {
                        cleanURLs = append(cleanURLs, url)
                    }
                }
                b.config.RSS[index].URLs = cleanURLs
            }
            b.userState[userID] = fmt.Sprintf("edit_interval_%d", index)
            b.sendMessage(chatID, fmt.Sprintf("当前间隔为：%d秒\n请输入新的间隔时间（秒）如不修改请输入1）：", b.config.RSS[index].Interval))
        } else if strings.HasPrefix(b.userState[userID], "edit_interval_") {
            index, _ := strconv.Atoi(strings.TrimPrefix(b.userState[userID], "edit_interval_"))
            if text != "1" {
                interval, err := strconv.Atoi(text)
                if err != nil {
                    b.sendMessage(chatID, "无效的间隔时间，请输入一个整数。不修改请输入1。")
                    return
                }
                b.config.RSS[index].Interval = interval
            }
            b.userState[userID] = fmt.Sprintf("edit_keywords_%d", index)
            b.sendMessage(chatID, fmt.Sprintf("当前关键词为：%v\n请输入新的关键词（用空格分隔，如不修改请输入1）：", b.config.RSS[index].Keywords))
        } else if strings.HasPrefix(b.userState[userID], "edit_keywords_") {
            index, _ := strconv.Atoi(strings.TrimPrefix(b.userState[userID], "edit_keywords_"))
            if text != "1" {
                keywords := strings.Fields(text) // 使用 Fields 替代 Split，自动按空格分割
                b.config.RSS[index].Keywords = keywords
            }
            b.userState[userID] = fmt.Sprintf("edit_group_%d", index)
            b.sendMessage(chatID, fmt.Sprintf("当前组名为：%s\n请输入新的组名（如不修改请输入1）：", b.config.RSS[index].Group))
        } else if strings.HasPrefix(b.userState[userID], "edit_group_") {
            index, _ := strconv.Atoi(strings.TrimPrefix(b.userState[userID], "edit_group_"))
            if text != "1" {
                b.config.RSS[index].Group = text
            }
            delete(b.userState, userID)
            if err := b.config.Save(b.configFile); err != nil {
                b.sendMessage(chatID, "编辑订阅成功，但保存配置失败。")
            } else {
                b.sendMessage(chatID, "成功编辑RSS订阅。")
                b.updateRSSHandler()
            }
        }
    }
}

func (b *Bot) getConfig() string {
    config := "当前配置信息：\n"
    config += fmt.Sprintf("用户: %v\n", b.users)
    config += fmt.Sprintf("频道: %v\n", b.channels)
    config += "RSS订阅:\n"
    for i, rss := range b.config.RSS {
        config += fmt.Sprintf("%d. 📡 URLs:\n", i+1)
        for j, url := range rss.URLs {
            config += fmt.Sprintf("   %d) %s\n", j+1, url)  // 直接显示URL，不进行转义
        }
        keywords := strings.Join(rss.Keywords, ", ")
        config += fmt.Sprintf("   ⏱️ 间隔: %d秒\n   🔑 关键词: %s\n   🏷️ 组名: %s\n", 
            rss.Interval, 
            formatBoldText(keywords),
            formatBoldText(rss.Group))
    }
    return config
}

func (b *Bot) listSubscriptions() string {
    list := "当前RSS订阅列表:\n"
    for i, rss := range b.config.RSS {
        list += fmt.Sprintf("%d. 📡 URLs:\n", i+1)
        for j, url := range rss.URLs {
            list += fmt.Sprintf("   %d) %s\n", j+1, url)  // 直接显示URL，不进行转义
        }
        keywords := strings.Join(rss.Keywords, ", ")
        list += fmt.Sprintf("   ⏱️ 间隔: %d秒\n   🔑 关键词: %s\n   🏷️ 组名: %s\n", 
            rss.Interval, 
            formatBoldText(keywords),
            formatBoldText(rss.Group))
    }
    return list
}

func (b *Bot) getStats() string {
    dailyCount, weeklyCount, totalCount := b.stats.GetMessageCounts()
    return fmt.Sprintf("推送统计:\n📊 今日推送: %s\n📈 本周推送: %s\n📋 总计推送: %s", 
        formatBoldText(strconv.Itoa(dailyCount)),
        formatBoldText(strconv.Itoa(weeklyCount)),
        formatBoldText(strconv.Itoa(totalCount)))
}

func (b *Bot) UpdateConfig(cfg *config.Config) {
    b.config = cfg
}

func (b *Bot) handleVersion(chatID int64) {
    // 获取当前版本
    currentVersion, err := b.getCurrentVersion()
    if err != nil {
        b.sendMessage(chatID, fmt.Sprintf("获取当前版本失败：%v", err))
        return
    }

    // 获取最新版本
    latestVersion, err := b.getLatestVersion()
    if err != nil {
        b.sendMessage(chatID, fmt.Sprintf("获取最新版本失败：%v", err))
        return
    }

    // 发送版本信息
    message := fmt.Sprintf("当前版本：%s\n最新版本：%s", currentVersion, latestVersion)
    b.sendMessage(chatID, message)
}

func (b *Bot) getCurrentVersion() (string, error) {
    versionFile := "/app/config/version"
    content, err := os.ReadFile(versionFile)
    if err != nil {
        return "", fmt.Errorf("读取版本文件失败: %v", err)
    }
    return strings.TrimSpace(string(content)), nil
}

func (b *Bot) getLatestVersion() (string, error) {
    // 直接从远程获取最新版本
    resp, err := http.Get("https://raw.githubusercontent.com/3377/rss2tg/refs/heads/main/version")
    if err != nil {
        return "", fmt.Errorf("无法获取最新版本: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("获取最新版本失败，状态码: %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("读取最新版本内容失败: %v", err)
    }

    return strings.TrimSpace(string(body)), nil
}

func (b *Bot) handleAddAll(chatID int64, userID int64) {
    b.userState[userID] = "add_all_keywords"
    b.sendMessage(chatID, "请输入要添加到所有订阅的关键词（用空格分隔）：")
}

func (b *Bot) handleDelAll(chatID int64, userID int64) {
    b.userState[userID] = "del_all_keywords"
    b.sendMessage(chatID, "请输入要从所有订阅中删除的关键词（用空格分隔）：")
}

// sendMessage 发送普通消息
func (b *Bot) sendMessage(chatID int64, text string) {
    // 转义特殊字符
    text = escapeMarkdownV2Text(text)
    
    msg := tgbotapi.NewMessage(chatID, text)
    msg.ParseMode = "MarkdownV2"
    if _, err := b.api.Send(msg); err != nil {
        log.Printf("发送消息失败: %v", err)
    }
}
