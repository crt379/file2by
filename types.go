package main

import "strings"

type File struct {
	Root string
	Path string
}

// StringSlice 自定义字符串切片类型
type StringSlice []string

// 实现 flag.Value 接口的 Set 方法
func (s *StringSlice) Set(value string) error {
	for v := range strings.SplitSeq(value, ",") {
		*s = append(*s, strings.TrimSpace(v))
	}
	return nil
}

// 实现 flag.Value 接口的 String 方法
func (s *StringSlice) String() string {
	return strings.Join(*s, ",")
}
