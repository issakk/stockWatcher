// package main

// import (
// 	"fmt"
// 	"log"
// 	"stockWatcher/internal/monitor"
// )

// func main() {
// 	fetcher := monitor.NewStockFetcher()

// 	log.Println("测试获取上证指数数据...")

// 	data, err := fetcher.FetchStockData("sh000001")
// 	if err != nil {
// 		log.Fatalf("获取数据失败: %v", err)
// 	}

// 	fmt.Printf("股票数据获取成功:\n")
// 	fmt.Printf("代码: %s\n", data.Code)
// 	fmt.Printf("名称: %s\n", data.Name)
// 	fmt.Printf("当前: %.2f\n", data.Current)
// 	fmt.Printf("开盘: %.2f\n", data.Open)
// 	fmt.Printf("最高: %.2f\n", data.High)
// 	fmt.Printf("最低: %.2f\n", data.Low)
// 	fmt.Printf("昨收: %.2f\n", data.Previous)
// 	fmt.Printf("涨跌: %.2f%%\n", data.Change)
// 	fmt.Printf("涨跌点数: %.2f\n", data.ChangeAmt)
// 	fmt.Printf("时间: %s\n", data.Timestamp.Format("2006-01-02 15:04:05"))

// 	// 计算相对于开盘价的波动
// 	changePercent := ((data.Current - data.Open) / data.Open) * 100
// 	fmt.Printf("相对开盘价波动: %.2f%%\n", changePercent)

// 	// 测试阈值判断
// 	threshold := 0.8
// 	isSignificant := data.IsSignificantChange(threshold) || (changePercent >= threshold || changePercent <= -threshold)
// 	fmt.Printf("是否超过%.1f%%阈值: %t\n", threshold, isSignificant)
// }