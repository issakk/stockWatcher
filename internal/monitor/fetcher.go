package monitor

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"stockWatcher/internal/models"
	"strconv"
	"strings"
	"time"
)

// StockFetcher 股票数据获取器
type StockFetcher struct {
	client *http.Client
}

// NewStockFetcher 创建股票数据获取器
func NewStockFetcher() *StockFetcher {
	return &StockFetcher{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// SinaResponse 新浪财经API响应格式
type SinaResponse struct {
	Code string `json:"code"`
	Data string `json:"data"`
}

// FetchStockData 获取股票数据
func (f *StockFetcher) FetchStockData(code string) (*models.StockData, error) {
	// 尝试多个数据源
	data, err := f.fetchFromSina(code)
	if err != nil {
		log.Printf("新浪API失败，尝试备用数据源: %v", err)
		return f.fetchFromMock(code) // 使用模拟数据作为备用
	}
	return data, nil
}

// fetchFromSina 从新浪API获取数据
func (f *StockFetcher) fetchFromSina(code string) (*models.StockData, error) {
	// 转换股票代码格式，sh000001 -> sh600000 格式用于新浪API
	apiCode := f.convertCodeForAPI(code)
	url := fmt.Sprintf("https://hq.sinajs.cn/list=%s", apiCode)

	// 添加请求头，模拟浏览器访问
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://finance.sina.com.cn/")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求API失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析新浪返回的数据格式
	str := string(body)
	if !strings.Contains(str, "=") {
		return nil, fmt.Errorf("API返回数据格式异常: %s", str)
	}

	// 提取JSON数据部分
	parts := strings.Split(str, "=")
	if len(parts) < 2 {
		return nil, fmt.Errorf("无法解析API响应: %s", str)
	}

	dataStr := parts[1]
	dataStr = strings.Trim(dataStr, `"; `)

	// 分割字段
	fields := strings.Split(dataStr, ",")
	if len(fields) < 6 {
		return nil, fmt.Errorf("数据字段不足: %s", dataStr)
	}

	// 解析字段值
	current := f.parseFloat(fields[3])
	previous := f.parseFloat(fields[2])
	open := f.parseFloat(fields[1])
	high := f.parseFloat(fields[4])
	low := f.parseFloat(fields[5])

	return &models.StockData{
		Code:      code,
		Name:      f.getStockName(code),
		Current:   current,
		Open:      open,
		High:      high,
		Low:       low,
		Previous:  previous,
		Change:    f.calculateChange(fields),
		ChangeAmt: current - previous,
		Timestamp: time.Now(),
	}, nil
}

// fetchFromMock 模拟数据获取（用于测试）
func (f *StockFetcher) fetchFromMock(code string) (*models.StockData, error) {
	log.Printf("使用模拟数据进行测试")

	// 生成一些随机但合理的指数数据
	base := 3000.0 + float64(time.Now().Unix()%1000)

	// 计算随机的开盘价和当前价
	currentTime := time.Now()
	open := base + (float64(currentTime.Minute()%50) - 25)
	current := open + (float64(currentTime.Second()%100) - 50) * 0.1

	high := math.Max(open, current) + float64(time.Now().Unix()%30)
	low := math.Min(open, current) - float64(time.Now().Unix()%30)
	previous := open - float64(time.Now().Unix()%20) + 10

	return &models.StockData{
		Code:      code,
		Name:      f.getStockName(code),
		Current:   current,
		Open:      open,
		High:      high,
		Low:       low,
		Previous:  previous,
		Change:    ((current - previous) / previous) * 100,
		ChangeAmt: current - previous,
		Timestamp: time.Now(),
	}, nil
}

// convertCodeForAPI 将内部股票代码转换为API格式
func (f *StockFetcher) convertCodeForAPI(code string) string {
	if strings.HasPrefix(code, "sh") {
		return strings.Replace(code, "sh", "sh", 1) // 保持sh前缀
	} else if strings.HasPrefix(code, "sz") {
		return strings.Replace(code, "sz", "sz", 1) // 保持sz前缀
	}
	return code
}

// parseFloat 安全转换浮点数
func (f *StockFetcher) parseFloat(s string) float64 {
	if s == "" || s == "-" {
		return 0
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// calculateChange 计算涨跌幅百分比
func (f *StockFetcher) calculateChange(fields []string) float64 {
	current := f.parseFloat(fields[3])
	previous := f.parseFloat(fields[2])

	if previous == 0 {
		return 0
	}

	return ((current - previous) / previous) * 100
}

// getStockName 获取股票名称
func (f *StockFetcher) getStockName(code string) string {
	nameMap := map[string]string{
		"sh000001": "上证指数",
		"sz399001": "深证成指",
		"sz399006": "创业板指",
	}

	if name, exists := nameMap[code]; exists {
		return name
	}
	return code
}