package main

import (
	"flag"
	"log"
	"os"
	"time"
)

var (
	WaitFor = 10 * time.Second
	LogPath = "file2by.log"
)

var (
	waitpath string
	bypypath string

	exps  StringSlice
	stock bool
)

func init() {
	// 绑定命令行参数到变量
	flag.StringVar(&waitpath, "w", ".", "监听文件变动路径")
	flag.StringVar(&bypypath, "b", "bypy", "bypy 二进制路径")
	flag.Var(&exps, "e", "监听的文件后缀")
	flag.BoolVar(&stock, "s", false, "是否处理存量文件")

	// 解析命令行参数
	flag.Parse()
}

func main() {
	// 打开文件（不存在则创建，追加写入）
	logfile, err := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件:", err)
	}
	defer logfile.Close()

	log.SetOutput(logfile)

	watcher, err := NewWatcher(waitpath, stock, exps...)
	if err != nil {
		panic(err)
	}

	uploader := ByUploader{
		Queue:       watcher.Queue(),
		AfterDelete: true,
		BypyPath:    bypypath,
	}

	go uploader.Run()

	watcher.Wait()
	if err = watcher.Error(); err != nil {
		panic(err)
	}
}
