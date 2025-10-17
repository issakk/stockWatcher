package main

import (
	"log"
	"os"
	"os/signal"
	"stockWatcher/internal/config"
	"stockWatcher/internal/monitor"
	"stockWatcher/internal/notifier"
	"syscall"
	"time"
)

func main() {
	// 使用测试配置文件
	cfg, err := config.Load("config_test.yaml")
	if err != nil {
		log.Fatalf("加载测试配置失败: %v", err)
	}

	// 初始化通知器（使用测试URL）
	wechatNotifier, err := notifier.NewWeChatNotifier(cfg.WeChat.WebhookURL)
	if err != nil {
		log.Fatalf("初始化企微通知器失败: %v", err)
	}

	// 创建监视器
	stockMonitor := monitor.NewStockMonitor(cfg, wechatNotifier)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动监视
	log.Println("股票指数监视器（测试模式）启动...")
	log.Printf("监视指数: %s", cfg.Stock.Code)
	log.Printf("波动阈值: %.2f%%", cfg.Stock.Threshold*100)
	log.Printf("检查间隔: %v", cfg.Monitor.Interval)

	go stockMonitor.Start()

	// 运行30秒后自动退出（用于测试）
	go func() {
		time.Sleep(30 * time.Second)
		log.Println("测试时间结束，正在关闭...")
		stockMonitor.Stop()
		sigChan <- syscall.SIGTERM
	}()

	// 等待退出信号
	<-sigChan
	log.Println("程序已退出")
}