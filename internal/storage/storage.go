package storage

import (
    "bufio"
    "log"
    "os"
    "strings"
    "sync"
)

type Storage struct {
    sentItems map[string]bool
    filePath  string
    mu        sync.Mutex
}

func NewStorage(filePath string) *Storage {
    s := &Storage{
        sentItems: make(map[string]bool),
        filePath:  filePath,
    }
    s.loadSentItems()
    return s
}

func (s *Storage) loadSentItems() {
    file, err := os.Open(s.filePath)
    if err != nil {
        if !os.IsNotExist(err) {
            log.Printf("打开已发送项目文件时出错: %v", err)
        }
        return
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        s.sentItems[strings.TrimSpace(scanner.Text())] = true
    }

    if err := scanner.Err(); err != nil {
        log.Printf("读取已发送项目文件时出错: %v", err)
    }
}

func (s *Storage) WasSent(url string) bool {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.sentItems[url]
}

func (s *Storage) MarkAsSent(url string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.sentItems[url] = true

    file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    if _, err := file.WriteString(url + "\n"); err != nil {
        return err
    }

    return nil
}
