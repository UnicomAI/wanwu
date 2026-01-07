package tools

import (
	"log"
	"runtime"
	"strings"
)

var (
	panicLogLen = 2048
)

// PrintPanicStack recover并打印堆栈
// 用法：defer tools.PrintPanicStack()，注意 defer func() { tools.PrintPanicStack() } 是无效的
func PrintPanicStack() {
	if r := recover(); r != nil {
		buf := make([]byte, panicLogLen)
		l := runtime.Stack(buf, false)
		str := strings.ReplaceAll(string(buf[:l]), "\n", " ")
		log.Printf("[PANIC] %v: %s", r, str)
	}
}
