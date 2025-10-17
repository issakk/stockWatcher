package main

import (
	"log"
	"os"
	"os/signal"
	"stockWatcher/internal/config"
	"stockWatcher/internal/monitor"
	"stockWatcher/internal/notifier"
	"syscall"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化通知器
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
	log.Println("股票指数监视器启动...")
	log.Printf("监视指数: %s", cfg.Stock.Code)
	log.Printf("波动阈值: %.2f%%", cfg.Stock.Threshold)
	log.Printf("检查间隔: %v", cfg.Monitor.Interval)

	go stockMonitor.Start()

	// 等待退出信号
	<-sigChan
	log.Println("接收到退出信号，正在关闭...")
	stockMonitor.Stop()
	log.Println("程序已退出")
}