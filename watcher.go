package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Vary struct {
	timer *time.Timer
	op    fsnotify.Op
}

type Watcher struct {
	// 监控对象
	waitfor     time.Duration
	watcher     *fsnotify.Watcher
	waitpath    string
	abswaitpath string

	err error

	// 处理的文件扩展名
	exps []string

	// 是否处理存量文件
	stock bool

	// 监听文件 事件
	mu    sync.Mutex
	varys map[string]*Vary
}

func NewWatcher(path string, stock bool, waitfor time.Duration, exps ...string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	var abswaitpath string
	if abswaitpath, err = filepath.Abs(path); err != nil {
		abswaitpath = path
	}

	w := &Watcher{
		waitfor:     waitfor,
		watcher:     watcher,
		waitpath:    path,
		abswaitpath: abswaitpath,
		exps:        exps,
		stock:       stock,
		varys:       make(map[string]*Vary),
	}

	return w, nil
}

func (w *Watcher) ishandle(path string) bool {
	for _, exp := range w.exps {
		if strings.HasSuffix(path, exp) {
			return true
		}
	}
	return false
}

func (w *Watcher) varyOnf(path string, f func(vary *Vary)) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if vary, ok := w.varys[path]; ok {
		if f != nil {
			f(vary)
		}

		return true
	}

	return false
}

func (w *Watcher) varyAdd(k string, v *Vary) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.varys[k] = v
}

func (w *Watcher) varyDel(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.varys, path)
}

func (w *Watcher) Wait(ctx context.Context, fun func(f File)) {
	// 递归添加监控目录
	err := filepath.Walk(w.waitpath, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return w.watcher.Add(path)
		}

		return nil
	})

	if err != nil {
		w.err = err
		return
	}

	// 处理存量文件
	if w.stock {
		w.handleStock(fun)
	}

	for {
		select {
		case <-ctx.Done():
			w.watcher.Close()
			return
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			fileinfog, err := os.Stat(event.Name)
			if err == nil && fileinfog.IsDir() {
				switch {
				case event.Has(fsnotify.Create):
					w.watcher.Add(event.Name)
				case event.Has(fsnotify.Remove), event.Has(fsnotify.Rename):
					w.watcher.Remove(event.Name)
				}

				continue
			}

			if !w.ishandle(event.Name) {
				continue
			}

			switch {
			case event.Has(fsnotify.Create), event.Has(fsnotify.Write):
				if w.varyOnf(event.Name, func(vary *Vary) {
					if vary.op != event.Op {
						log.Printf("%s %s -> %s", vary.op.String(), event.Name, event.Op.String())
					}

					vary.op = event.Op
					vary.timer.Reset(w.waitfor)
				}) {
					continue
				}

				log.Printf("%s %s", event.Op.String(), event.Name)
				w.fileTiming(event.Name, event.Op, fun)
			case event.Has(fsnotify.Rename), event.Has(fsnotify.Remove):
				log.Printf("%s %s", event.Op.String(), event.Name)
				w.varyOnf(event.Name, func(vary *Vary) {
					vary.op = event.Op
					vary.timer.Stop()
					delete(w.varys, event.Name)
				})
			}
		case err = <-w.watcher.Errors:
			w.err = err
			log.Println("⚠️ 监控错误:", err)
		}
	}
}

func (w *Watcher) fileTiming(path string, op fsnotify.Op, fun func(f File)) {
	timer := time.NewTimer(w.waitfor)
	w.varyAdd(path, &Vary{timer: timer, op: op})

	go func() {
		<-timer.C

		if !w.varyOnf(path, nil) {
			return
		}

		abspath, err := filepath.Abs(path)
		if err != nil {
			abspath = path
		}

		fun(File{
			Path: abspath,
			Root: w.abswaitpath,
		})

		w.varyDel(path)
	}()
}

func (w *Watcher) handleStock(fun func(f File)) {
	// 使用切片存储所有文件路径
	var files []string

	// 使用 filepath.Walk 遍历目录
	err := filepath.Walk(w.waitpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("访问路径出错 %q: %v\n", path, err)
			return nil
		}

		// 如果不是目录，则添加到文件列表
		if !info.IsDir() && w.ishandle(path) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		log.Printf("遍历目录时出错: %v\n", err)
		return
	}

	// 打印找到的所有文件
	log.Printf("找到 %d 个存量文件:\n", len(files))
	for _, file := range files {
		log.Println(file)
		w.fileTiming(file, 0, fun)
	}
}

func (w *Watcher) Error() error {
	return w.err
}
