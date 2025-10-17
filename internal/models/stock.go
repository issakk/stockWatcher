package models

import (
	"time"
)

// StockData 表示股票指数数据
type StockData struct {
	Code      string    `json:"code"`
	Name      string    `json:"name"`
	Current   float64   `json:"current"`   // 当前点数
	Open      float64   `json:"open"`      // 开盘点位
	High      float64   `json:"high"`      // 最高点位
	Low       float64   `json:"low"`       // 最低点位
	Previous  float64   `json:"previous"`  // 昨收点位
	Change    float64   `json:"change"`    // 涨跌幅(百分比)
	ChangeAmt float64   `json:"changeAmt"` // 涨跌点数
	Timestamp time.Time `json:"timestamp"` // 数据时间
}

// ChangePercent 计算涨跌百分比
func (s *StockData) ChangePercent() float64 {
	if s.Previous == 0 {
		return 0
	}
	return ((s.Current - s.Previous) / s.Previous) * 100
}

// IsSignificantChange 判断是否显著变化
func (s *StockData) IsSignificantChange(threshold float64) bool {
	return s.ChangePercent() >= threshold || s.ChangePercent() <= -threshold
}