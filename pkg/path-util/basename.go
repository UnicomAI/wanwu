package path_util

import (
	"fmt"
	"strings"
	"unicode"
)

// BasenameMaxLen 单个文件名（路径组件）的最大长度。
// Linux 文件系统限制单个组件 255 字节；Windows MAX_PATH 组件限制同为 255。
const BasenameMaxLen = 255

// ValidateBasename 校验单个文件或目录名称（basename，不含任何路径结构）。
//
// 适用于用户可控的文件名在拼接落盘路径前的统一安全校验，覆盖三类风险：
//  1. 路径穿越：拒绝路径分隔符（/ 与 \）及 "."、".." 本身
//  2. 跨平台兼容：拒绝 Windows 保留设备名、尾随点/空格、盘符/流分隔符（:）
//     （工作区文件可能被复制到 Windows 主机，故在 Linux 服务上同样拒绝）
//  3. 注入与异常：拒绝 null 字节与控制字符，拒绝前导 "." 的隐藏条目
//
// 该函数不校验路径（含分隔符的相对路径请用 CleanRelPath），
// 也不校验路径是否越界（请配合 JoinWithinBase 使用）。
func ValidateBasename(name string) error {
	if name == "" || name == "." || name == ".." {
		return fmt.Errorf("name is required")
	}
	if len(name) > BasenameMaxLen {
		return fmt.Errorf("name too long")
	}
	if strings.ContainsAny(name, `/\:`) || strings.ContainsRune(name, '\x00') {
		return fmt.Errorf("name must be a basename")
	}
	if strings.HasSuffix(name, ".") || strings.HasSuffix(name, " ") {
		return fmt.Errorf("name has an unsafe suffix")
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return fmt.Errorf("name contains invalid control characters")
		}
	}
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("hidden entries are not allowed")
	}
	base := strings.ToUpper(strings.TrimSuffix(name, "."))
	if strings.Contains(base, ".") {
		base = strings.SplitN(base, ".", 2)[0]
	}
	switch base {
	case "CON", "PRN", "AUX", "NUL", "CLOCK$", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return fmt.Errorf("reserved device name")
	}
	return nil
}
