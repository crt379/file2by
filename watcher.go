package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	watcher     *fsnotify.Watcher
	waitpath    string
	abswaitpath string
	queue       chan File
	err         error
	handleexps  []string
	stock       bool

	// 监听文件 事件
	mu    sync.Mutex
	varys map[string]*Vary
}

func NewWatcher(path string, stock bool, handleexps ...string) (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	var abswaitpath string
	if abswaitpath, err = filepath.Abs(path); err != nil {
		abswaitpath = path
	}

	w := &Watcher{
		watcher:     watcher,
		waitpath:    path,
		abswaitpath: abswaitpath,
		queue:       make(chan File),
		handleexps:  handleexps,
		stock:       stock,
		varys:       make(map[string]*Vary),
	}

	return w, nil
}

func (w *Watcher) Queue() <-chan File {
	return w.queue
}

func (w *Watcher) ishandle(path string) bool {
	for _, exp := range w.handleexps {
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

func (w *Watcher) Wait() {
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

	w.handleStock()

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			fileinfog, err := os.Stat(event.Name)
			if err == nil && fileinfog.IsDir() {
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
					vary.timer.Reset(WaitFor)
				}) {
					continue
				}

				log.Printf("%s %s", event.Op.String(), event.Name)
				w.fileTiming(event.Name, event.Op)
			case event.Has(fsnotify.Rename), event.Has(fsnotify.Remove):
				log.Printf("%s %s", event.Op.String(), event.Name)
				if w.varyOnf(event.Name, func(vary *Vary) {
					vary.op = event.Op
					vary.timer.Stop()
					delete(w.varys, event.Name)
				}) {
					return
				}
			}
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}

			w.err = err
			log.Println("⚠️ 监控错误:", err)
		}
	}
}

func (w *Watcher) fileTiming(path string, op fsnotify.Op) {
	timer := time.NewTimer(WaitFor)
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

		w.queue <- File{
			Path: abspath,
			Root: w.abswaitpath,
		}

		w.varyDel(path)
	}()
}

func (w *Watcher) handleStock() {
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
		w.fileTiming(file, 0)
	}
}

func (w *Watcher) Error() error {
	return w.err
}
