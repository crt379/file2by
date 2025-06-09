# file2by

file2by 是一个通过 bypy 批量上传文件到百度云的工具。支持监听指定目录的变动，自动上传新增的文件。

运行前请先安装 bypy。

运行方法：
go build -o file2by main.go
./file2by -wf=/path/to/watch -bp=/path/to/bypy -exps=.txt,.log -stock=true -log=/path/to/log/file2by.log -waitfor=30s

参数说明：
-wf: 监听的目录。
-bp: bypy 的目录。
-exps:  监听的文件后缀，如 .txt, .log 等，多个后缀用逗号分隔。
-stock: 是否处理存量文件，布尔值，默认为 false。
-log: 日志文件。
-waitfor: 文件变动后等待的时间，如果期间有文件有变动，则重新计算等待时间，比如a.txt变动后在waitfor内有变动，则重置等待时间来防止文件未写入完成就上传。
