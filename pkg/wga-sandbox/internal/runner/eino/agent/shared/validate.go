package shared

import (
	"fmt"
	"log"
	"strings"
)

// hasWriteActionBefore 判断 command[:idx] 末尾是否以"写动作"结束。
// lookback 限制回看窗口避免误判跨命令的远距离 token。
func hasWriteActionBefore(command string, idx int) bool {
	const lookback = 64
	start := max(idx-lookback, 0)
	prefix := command[start:idx]
	return writeActionPattern.MatchString(prefix)
}

func validateCommand(command string) error {
	// 1) 高危动作：无论上下文都拦
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(command) {
			matched := pattern.FindString(command)
			log.Printf("[安全拦截] 检测到危险操作: %s", matched)
			return fmt.Errorf("安全拦截：检测到高危命令片段 [%s]，已拒绝执行", matched)
		}
	}

	// 2) 读敏感路径：命中即拦
	for _, pattern := range sensitiveReadPatterns {
		if matched := pattern.FindString(command); matched != "" {
			log.Printf("[安全拦截] 检测到对敏感路径读取尝试: %s", matched)
			return fmt.Errorf("安全拦截：检测到对敏感路径 [%s] 的访问尝试，已拒绝执行", matched)
		}
	}

	// 3) 写敏感路径：仅当命中位置前紧邻写动作时拦截
	for _, pattern := range sensitiveWritePatterns {
		for _, loc := range pattern.FindAllStringIndex(command, -1) {
			if hasWriteActionBefore(command, loc[0]) {
				matched := command[loc[0]:loc[1]]
				log.Printf("[安全拦截] 检测到对敏感路径写入尝试: %s", matched)
				return fmt.Errorf("安全拦截：检测到对敏感路径 [%s] 的写入尝试，已拒绝执行", matched)
			}
		}
	}

	// 4) 路径穿越
	if strings.Contains(command, "..") && pathTraversalPattern.MatchString(command) {
		return fmt.Errorf("安全拦截：检测到路径穿越尝试（../），已拒绝执行")
	}

	// 5) 指向敏感路径的符号链接
	if symlinkPattern.MatchString(command) {
		for _, pattern := range sensitiveReadPatterns {
			if matched := pattern.FindString(command); matched != "" {
				return fmt.Errorf("安全拦截：禁止创建指向敏感路径 [%s] 的符号链接", matched)
			}
		}
		for _, pattern := range sensitiveWritePatterns {
			if matched := pattern.FindString(command); matched != "" {
				return fmt.Errorf("安全拦截：禁止创建指向敏感路径 [%s] 的符号链接", matched)
			}
		}
	}

	return nil
}
