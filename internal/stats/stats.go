package stats

import (
    "encoding/json"
    "io/ioutil"
    "log"
    "os"
    "sync"
    "time"
)

type Stats struct {
    DailyCount  int       `json:"daily_count"`
    WeeklyCount int       `json:"weekly_count"`
    TotalCount  int       `json:"total_count"`
    LastReset   time.Time `json:"last_reset"`
    filePath    string
    mu          sync.Mutex
}

func NewStats(filePath string) (*Stats, error) {
    s := &Stats{filePath: filePath}
    err := s.load()
    if err != nil {
        return nil, err
    }
    go s.resetCounters()
    return s, nil
}

func (s *Stats) load() error {
    data, err := ioutil.ReadFile(s.filePath)
    if os.IsNotExist(err) {
        s.LastReset = time.Now()
        return s.save()
    }
    if err != nil {
        return err
    }
    return json.Unmarshal(data, s)
}

func (s *Stats) save() error {
    data, err := json.Marshal(s)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(s.filePath, data, 0644)
}

func (s *Stats) IncrementMessageCount() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.DailyCount++
    s.WeeklyCount++
    s.TotalCount++
    if err := s.save(); err != nil {
        log.Printf("保存统计信息失败: %v", err)
    }
}

func (s *Stats) GetMessageCounts() (int, int, int) {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.DailyCount, s.WeeklyCount, s.TotalCount
}

func (s *Stats) resetCounters() {
    for {
        now := time.Now()
        nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
        time.Sleep(nextMidnight.Sub(now))

        s.mu.Lock()
        s.DailyCount = 0
        if now.Sub(s.LastReset) >= 7*24*time.Hour {
            s.WeeklyCount = 0
            s.LastReset = now
        }
        if err := s.save(); err != nil {
            log.Printf("重置后保存统计信息失败: %v", err)
        }
        s.mu.Unlock()
    }
}
