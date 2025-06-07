package main

// 判断是否为临时文件
// func isTempFile(filepath string) bool {
// 	filename := strings.ToLower(filepath)
// 	return strings.HasPrefix(filepath, ".") ||
// 		strings.HasSuffix(filename, ".tmp") ||
// 		strings.HasSuffix(filename, ".swp") ||
// 		strings.HasSuffix(filename, ".part") ||
// 		strings.Contains(filename, "temp") ||
// 		strings.Contains(filename, "cache")
// }

// 捕获命令输出
// func captureOutput(stdout, stderr io.ReadCloser) string {
// 	var builder strings.Builder

// 	{
// 		scanner := bufio.NewScanner(stdout)
// 		for scanner.Scan() {
// 			line := scanner.Text()
// 			builder.WriteString("OUT: " + line + "\n")
// 		}
// 	}

// 	{
// 		scanner := bufio.NewScanner(stderr)
// 		for scanner.Scan() {
// 			line := scanner.Text()
// 			builder.WriteString("ERR: " + line + "\n")
// 		}
// 	}

// 	return builder.String()
// }
