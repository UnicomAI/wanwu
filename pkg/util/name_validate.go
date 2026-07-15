package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	NameMinLen = 2
	NameMaxLen = 50
	DescMaxLen = 200
	// FileNameMaxLen 上传文档名的最大长度，按 Unicode 字符数计。
	// 文档会以文档名落盘，受 Linux 文件名 255 字节上限约束；按中文每字 3 字节估算，80 字约 240 字节，留有余量。
	FileNameMaxLen = 80
)

// 各模块名称/描述校验时用作错误文案前缀的 subject 常量。
const (
	SubjectAssistant            = "智能体"
	SubjectRag                  = "知识问答"
	SubjectWorkflow             = "工作流"
	SubjectChatflow             = "对话流"
	SubjectModel                = "模型"
	SubjectMCP                  = "MCP"
	SubjectMCPServer            = "MCP服务"
	SubjectKnowledge            = "知识库"
	SubjectKnowledgeExternalAPI = "外部知识库API"
	SubjectCustomTool           = "工具"
	SubjectSensitiveWordTable   = "敏感词表"
	SubjectPrompt               = "提示词"
)

var (
	nameCharsetRegex = regexp.MustCompile(`^[A-Za-z0-9\x{4e00}-\x{9fa5}_.\-]+$`)
	nameHasValidChar = regexp.MustCompile(`[A-Za-z0-9\x{4e00}-\x{9fa5}]`)
)

// ValidateName 规范化并校验名称：
// 1) 先对 *name 做 TrimSpace 原地写回，保证后续入库的也是规范化后的值
// 2) 长度 2-50（按 Unicode 字符数）
// 3) 不能以下划线开头
// 4) 仅支持中英文、数字、下划线、中划线、小数点
// 5) 不能仅由下划线/中划线/小数点组成
// 按顺序短路，命中第一条即返回。
func ValidateName(name *string, subject string) error {
	*name = strings.TrimSpace(*name)
	n := utf8.RuneCountInString(*name)
	if n < NameMinLen || n > NameMaxLen {
		return fmt.Errorf("%s名称长度需为 %d-%d 字符", subject, NameMinLen, NameMaxLen)
	}
	if strings.HasPrefix(*name, "_") {
		return fmt.Errorf("%s名称不能以下划线开头", subject)
	}
	if !nameCharsetRegex.MatchString(*name) {
		return fmt.Errorf("%s名称仅支持中英文、数字、下划线、中划线、小数点", subject)
	}
	if !nameHasValidChar.MatchString(*name) {
		return fmt.Errorf("%s名称不能仅由下划线、中划线、小数点组成", subject)
	}
	return nil
}

// ValidateFileNameLength 仅按 Unicode 字符数限制上传文档名长度（原始名含扩展名/空格等，不套 ValidateName 白名单）。
func ValidateFileNameLength(name string) bool {
	return utf8.RuneCountInString(name) <= FileNameMaxLen
}

// ValidateDesc 规范化并校验描述： 检查 ≤ 200 字符（按 Unicode 字符数）
func ValidateDesc(desc *string, subject string) error {
	*desc = strings.TrimSpace(*desc)
	if utf8.RuneCountInString(*desc) > DescMaxLen {
		return fmt.Errorf("%s描述长度不能超过 %d 字符", subject, DescMaxLen)
	}
	return nil
}

// ValidateNameUpdate 用于更新名字场景： 如果用户没有改过名字，则跳过格式校验
func ValidateNameUpdate(newName *string, oldName, subject string) error {
	*newName = strings.TrimSpace(*newName)
	if *newName == strings.TrimSpace(oldName) {
		return nil
	}
	return ValidateName(newName, subject)
}

// ValidateDescUpdate 用于更新描述场景：未变更则跳过校验
func ValidateDescUpdate(newDesc *string, oldDesc, subject string) error {
	*newDesc = strings.TrimSpace(*newDesc)
	if *newDesc == strings.TrimSpace(oldDesc) {
		return nil
	}
	return ValidateDesc(newDesc, subject)
}

// ValidateBriefCreate 新建时校验名称+描述。name / desc 都会被原地 trim。
func ValidateBriefCreate(name, desc *string, subject string) error {
	if err := ValidateName(name, subject); err != nil {
		return err
	}
	return ValidateDesc(desc, subject)
}

// ValidateBriefUpdate 编辑名称+描述
func ValidateBriefUpdate(newName *string, oldName string, newDesc *string, oldDesc, subject string) error {
	if err := ValidateNameUpdate(newName, oldName, subject); err != nil {
		return err
	}
	return ValidateDescUpdate(newDesc, oldDesc, subject)
}

// 副本序号后缀保留长度："_" + 最多 4 位数字，即副本编号上限 9999
const copySuffixReservedLen = 5

// CopyNameBase 把原名规范化成合规的副本名主体：trim、非法字符换 _、剥前导 _、空兜底 "copy"、
// 按字符截到最多 45（留 5 字符给 "_序号"，覆盖 1-9999 的编号）。
func CopyNameBase(original string) string {
	name := replaceInvalidChars(strings.TrimSpace(original))
	name = strings.TrimLeft(name, "_")
	if name == "" {
		name = "copy"
	}
	runes := []rune(name)
	if len(runes)+copySuffixReservedLen > NameMaxLen {
		name = string(runes[:NameMaxLen-copySuffixReservedLen])
	}
	return name
}

// GenCopyName 按 "规范化原名_序号" 生成合规的副本名
func GenCopyName(original string, n int) string {
	return CopyNameBase(original) + "_" + strconv.Itoa(n)
}

// replaceInvalidChars 将不在 ValidateName 白名单内的字符替换为 _
func replaceInvalidChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z',
			r >= 'a' && r <= 'z',
			r >= '0' && r <= '9',
			r >= 0x4e00 && r <= 0x9fa5,
			r == '_', r == '-', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}
