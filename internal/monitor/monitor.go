package monitor

import (
	"fmt"
	"log"
	"math"
	"stockWatcher/internal/models"
	"stockWatcher/internal/notifier"
	"sync"
	"time"
)

// StockMonitor 股票监视器
type StockMonitor struct {
	config        *models.Config
	fetcher       *StockFetcher
	notifier      notifier.Notifier
	isRunning     bool
	stopChan      chan struct{}
	mu            sync.RWMutex
	lastData      *models.StockData // 上次的数据
	lastAlertTime time.Time         // 上次发送警报的时间
	dayStart      float64           // 当日开盘价
	maxChange     float64           // 当日最大波动
	minChange     float64           // 当日最小波动
}

// NewStockMonitor 创建股票监视器
func NewStockMonitor(config *models.Config, notifier notifier.Notifier) *StockMonitor {
	return &StockMonitor{
		config:   config,
		fetcher:  NewStockFetcher(),
		notifier: notifier,
		stopChan: make(chan struct{}),
	}
}

// Start 开始监视
func (m *StockMonitor) Start() {
	m.mu.Lock()
	m.isRunning = true
	m.mu.Unlock()

	ticker := time.NewTicker(m.config.Monitor.Interval)
	defer ticker.Stop()

	// 立即执行一次检查
	m.checkStock()

	for {
		select {
		case <-ticker.C:
			m.checkStock()
		case <-m.stopChan:
			return
		}
	}
}

// Stop 停止监视
func (m *StockMonitor) Stop() {
	m.mu.Lock()
	m.isRunning = false
	close(m.stopChan)
	m.mu.Unlock()
}

// checkStock 检查股票数据
func (m *StockMonitor) checkStock() {
	data, err := m.fetcher.FetchStockData(m.config.Stock.Code)
	if err != nil {
		log.Printf("获取股票数据失败: %v", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 第一次获取数据，初始化基准值
	if m.lastData == nil {
		m.dayStart = data.Open
		m.lastData = data
		m.maxChange = 0
		m.minChange = 0

		log.Printf("初始数据: %s 当前: %.2f 开盘: %.2f",
			data.Name, data.Current, data.Open)
		return
	}

	// 直接使用 StockData.Change 字段（相对于昨收的涨跌百分比）
	currentChange := data.Change

	// 计算相对于开盘价的波动用于统计
	openChangePercent := m.calculateChangePercent(m.dayStart, data.Current)

	// 更新当日最大最小波动（相对于开盘价）
	m.maxChange = math.Max(m.maxChange, math.Abs(openChangePercent))
	m.minChange = math.Min(m.minChange, math.Abs(openChangePercent))

	// 使用 Change 字段检查是否超过阈值
	if math.Abs(currentChange) >= m.config.Stock.Threshold {
		m.sendAlert(data, currentChange)
	}

	// 记录当前状态
	log.Printf("%s: 当前 %.2f 昨收 %.2f 开盘 %.2f 涨跌 %.2f%% (相对昨收) 波动 %.2f%% (相对开盘)",
		data.Name, data.Current, data.Previous, data.Open, currentChange, openChangePercent)

	m.lastData = data
}

// calculateChangePercent 计算变化百分比
func (m *StockMonitor) calculateChangePercent(base, current float64) float64 {
	if base == 0 {
		return 0
	}
	return ((current - base) / base) * 100
}

// isWithinAlertWindow 检查当前时间是否在警报时间窗口内 (周一至周五, 14:30-17:00)
func (m *StockMonitor) isWithinAlertWindow() bool {
	now := time.Now()
	weekday := now.Weekday()
	hour := now.Hour()
	minute := now.Minute()

	// 周一至周五
	if weekday < time.Monday || weekday > time.Friday {
		return false
	}

	// 14:30 到 17:00 (包含17:00整)
	if (hour == 14 && minute >= 30) || (hour > 14 && hour < 15) || (hour == 15 && minute == 0) {
		return true
	}

	return false
}

// sendAlert 发送警报
func (m *StockMonitor) sendAlert(data *models.StockData, changePercent float64) {
	// 检查是否在通知时间段内
	if !m.isWithinAlertWindow() {
		return // 不在时间窗口内，静默返回
	}

	// 避免重复发送相同警报
	if !m.shouldSendAlert(data, changePercent) {
		return
	}

	var direction string
	var emoji string
	if changePercent > 0 {
		direction = "上涨"
		emoji = "📈"
	} else {
		direction = "下跌"
		emoji = "📉"
	}

	message := m.formatAlertMessage(data, changePercent, direction, emoji)

	if err := m.notifier.Send(message); err != nil {
		log.Printf("发送通知失败: %v", err)
	} else {
		log.Printf("已发送%.2f%%波动警报通知", math.Abs(changePercent))
		// 更新最后发送警报时间
		m.lastAlertTime = data.Timestamp
	}
}

// shouldSendAlert 判断是否应该发送警报（避免重复）
func (m *StockMonitor) shouldSendAlert(data *models.StockData, changePercent float64) bool {
	// 如果从未发送过警报，允许发送
	if m.lastAlertTime.IsZero() {
		return true
	}

	// 距离上次警报至少间隔5分钟，避免频繁通知
	timeDiff := data.Timestamp.Sub(m.lastAlertTime)

	if timeDiff < 5*time.Minute {
		// 但是如果涨跌幅度继续显著增加，仍然发送
		lastChange := math.Abs(m.lastData.Change)
		currentChange := math.Abs(changePercent)
		changeDiff := currentChange - lastChange

		// 如果当前涨跌比上次涨跌增加超过0.2%，仍然发送
		if changeDiff < 0.2 {
			return false
		}
	}

	return true
}

// formatAlertMessage 格式化警报消息
func (m *StockMonitor) formatAlertMessage(data *models.StockData, changePercent float64, direction string, emoji string) string {
	template := `
%s 股票市场警报 %s

指数: %s
当前点位: %.2f
涨跌: %s %.2f%% (%.2f 点)
开盘: %.2f
最高: %.2f
最低: %.2f
时间: %s
阈值: %.2f%%

统计数据 (与昨收对比):
- 当前涨跌: %.2f%%
- 日内波动: %.2f%%
- 日内最大波动: %.2f%%

请注意风险！
`

	// 计算开盘后波动
	openChangePercent := m.calculateChangePercent(m.dayStart, data.Current)

	return fmt.Sprintf(template,
		emoji, emoji,
		data.Name,
		data.Current,
		direction,
		math.Abs(changePercent),
		data.ChangeAmt,
		data.Open,
		data.High,
		data.Low,
		data.Timestamp.Format("15:04:05"),
		m.config.Stock.Threshold,
		math.Abs(changePercent),
		math.Abs(openChangePercent),
		m.maxChange,
	)
}
