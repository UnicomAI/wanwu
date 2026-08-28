package request

import (
	"regexp"
)

// 预编译恶意文件名检测正则表达式（不区分大小写）
var (
	// 匹配双管道后跟常见解释器名称（如 python, perl, sh 等）
	rePipeWithInterpreter = regexp.MustCompile(`(?i)\|\|.*(?:python|perl|ruby|sh|bash|cmd|powershell)`)
	// 匹配 python -c 直接执行代码
	rePythonC = regexp.MustCompile(`(?i)python\d?\s+-c\s+`)
	// 匹配 exec( 函数调用
	reExecCall = regexp.MustCompile(`(?i)exec\s*\(`)
	// 匹配命令替换：$(...) 或反引号 `
	reCmdSubst = regexp.MustCompile(`\$\x60|` + "`")
	// 匹配分号、与号、管道（单独使用或组合）后跟 shell 命令特征（可选，增强检测）
	reShellMeta = regexp.MustCompile(`(?i)[;&|]\s*(?:python|perl|ruby|sh|bash|wget|curl|nc|powershell)`)
)

type CommonCheck struct {
}

func (c *CommonCheck) Check() error {
	return nil
}

type PageSearch struct {
	PageSize int `json:"pageSize" form:"pageSize" validate:"required"`
	PageNo   int `json:"pageNo" form:"pageNo"`
}

type LoginEmailCheck struct {
	Email string `json:"email" validate:"required"` // 邮箱
	Code  string `json:"code" validate:"required"`  // 邮箱验证码
}

func (l *LoginEmailCheck) Check() error {
	return nil
}

type ChangeUserPasswordByEmail struct {
	OldCipher string `json:"oldCipher" validate:"required"` // 旧密码RSA加密后的Base64字符串，包含{password, challenge}
	NewCipher string `json:"newCipher" validate:"required"` // 新密码RSA加密后的Base64字符串，包含{password, challenge}
	Email     string `json:"email" validate:"required"`     // 邮箱
	Code      string `json:"code" validate:"required"`      // 邮箱验证码
	KeyID     string `json:"keyId" validate:"required"`     // RSA公钥ID
}

func (c *ChangeUserPasswordByEmail) Check() error {
	return nil
}

// IsMaliciousFilename 检查文件名是否包含潜在的恶意命令执行模式
// 返回 true 表示文件名危险，应拒绝上传；false 表示相对安全
func IsMaliciousFilename(filename string) bool {
	// 可选：仅检查文件名的基础部分，避免路径干扰
	// base := path.Base(filename)
	// 此处直接使用完整字符串，攻击者可能直接传入文件名
	if rePipeWithInterpreter.MatchString(filename) {
		return true
	}
	if rePythonC.MatchString(filename) {
		return true
	}
	if reExecCall.MatchString(filename) {
		return true
	}
	if reCmdSubst.MatchString(filename) {
		return true
	}
	if reShellMeta.MatchString(filename) {
		return true
	}
	return false
}
