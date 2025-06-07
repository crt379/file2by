package main

import (
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Vary struct {
	timer *time.Timer
	op    fsnotify.Op
}

type ByUploader struct {
	Queue       <-chan File
	AfterDelete bool
	BypyPath    string
}

func (u *ByUploader) Run() {
	for t := range u.Queue {
		if _, err := os.Stat(t.Path); err != nil {
			log.Printf("文件不存在: %s", t.Path)
			continue
		}

		log.Printf("🚀 上传任务: %+v", t)
		u.Upload(t.Path, t.Path[len(t.Root):])
	}
}

func (u *ByUploader) Upload(filepath string, targetpath string) {
	// 上传文件
	log.Printf("⬆️ 开始上传: %s", filepath)

	cmd := exec.Command(
		u.BypyPath,
		"-s",
		"100M",
		"upload",
		filepath,
		targetpath,
	)

	log.Printf("⬆️ 执行命令: %s", cmd.String())

	output, err := cmd.CombinedOutput()

	// 检查上传结果
	if err == nil {
		log.Printf("🎉 上传成功: %s, 输出: %s", filepath, output)

		// 如果设置删除，安全删除文件
		if u.AfterDelete {
			u.SafeDelete(filepath)
		}
	} else {
		log.Printf("❌ 上传失败: %s, 输出: %s", filepath, output)
	}
}

func (u *ByUploader) SafeDelete(filepath string) {
	log.Printf("🗑️ 开始安全删除: %s", filepath)
	// 二次确认文件存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		log.Printf("⚠️ 文件已不存在，跳过删除: %s", filepath)
		return
	}

	// 永久删除
	if err := os.Remove(filepath); err != nil {
		log.Printf("❌ 永久删除失败: %s - %v", filepath, err)
		return
	}
	log.Printf("🗑️ 文件已删除: %s", filepath)
}
