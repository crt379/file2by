package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	waitpath string
	bypypath string

	exps  StringSlice
	stock bool

	logPath string

	waitfor time.Duration
)

func init() {
	// 绑定命令行参数到变量
	flag.StringVar(&waitpath, "wf", ".", "监听文件变动路径")
	flag.StringVar(&bypypath, "bp", "bypy", "bypy 二进制路径")
	flag.Var(&exps, "exps", "监听的文件后缀")
	flag.BoolVar(&stock, "stock", false, "是否处理存量文件")
	flag.StringVar(&logPath, "log", "file2by.log", "日志文件路径")
	flag.DurationVar(&waitfor, "waitfor", 10*time.Second, "等待时间")

	// 解析命令行参数
	flag.Parse()
}

func main() {
	// 打开文件（不存在则创建，追加写入）
	logfile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件:", err)
	}
	defer logfile.Close()

	log.SetOutput(logfile)

	watcher, err := NewWatcher(waitpath, stock, waitfor, exps...)
	if err != nil {
		panic(err)
	}

	uploader := ByUploader{
		AfterDelete: true,
		BypyPath:    bypypath,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 优雅关闭通道：监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	uploadch := make(chan UploadInfo)

	wg.Add(1)
	go func() {
		defer wg.Done()
		uploader.Run(ctx, uploadch)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		watcher.Wait(ctx, func(f File) {
			uf := UFile(f)
			uploadch <- &uf
		})
		if err = watcher.Error(); err != nil {
			log.Printf("Watcher error: %v", err)
			os.Exit(1)
		}
	}()

	<-sigChan

	log.Println("Shutting down...")
	cancel()

	wg.Wait()
	log.Println("Bye!")
}
