package util

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
)

var (
	panicLogLen = 2048
)

// PrintPanicStack recover并打印堆栈 [EN] PrintPanicStack recover and print the stack
// 用法：defer util.PrintPanicStack()，注意 defer func() { util.PrintPanicStack() } 是无效的 [EN] Usage: defer util.PrintPanicStack(), note that defer func() { util.PrintPanicStack() } is invalid
func PrintPanicStack() {
	if r := recover(); r != nil {
		buf := make([]byte, panicLogLen)
		l := runtime.Stack(buf, false)
		str := strings.ReplaceAll(string(buf[:l]), "\n", " ")
		log.Errorf("%v: %s", r, str)
	}
}

func PrintPanicStackWithCall(postProcessor func(panicOccur bool, recoverError error)) {
	var panicOccur = false
	var err error
	if r := recover(); r != nil {
		buf := make([]byte, panicLogLen)
		l := runtime.Stack(buf, false)
		str := strings.ReplaceAll(string(buf[:l]), "\n", " ")
		log.Errorf("%v: %s", r, str)
		panicOccur = true
		err = fmt.Errorf("panic %v", r)
	}
	if postProcessor != nil {
		postProcessor(panicOccur, err)
	}
}
