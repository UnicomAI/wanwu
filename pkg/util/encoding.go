package util

import (
	"bytes"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// ZipEFSFlag 是 zip 规范的 Language encoding flag (EFS, bit 11)。
// 置位表示条目名按 UTF-8 编码，未置位则按本地编码（中文环境通常为 GBK/GB18030）。
const ZipEFSFlag = 0x800

// replacementRune 是 x/text 解码器遇到非法字节时插入的 Unicode 替换字符 U+FFFD。
const replacementRune = "�"

// DecodeGBKToUTF8 把可能是 GBK/GB18030 编码的字符串解码为 UTF-8。
//
// 判定与回退策略：
//  1. 若 s 本身是合法 UTF-8（utf8.ValidString），原样返回，不做任何解码。
//  2. 否则尝试 GB18030 解码（GB18030 是 GBK 超集，覆盖面最广，含繁体、生僻字、4 字节区）。
//  3. 仅当解码产物是合法 UTF-8 且不含替换字符 U+FFFD 时才采用；否则原样返回 s。
//  4. 不兜底 Big5：Big5 与 GB18030 双字节空间高度重叠，几乎任意 Big5 字节串都会被
//     GB18030 “误成功”解码成错码（utf8.ValidString 无法识别），先试谁就吞掉谁，
//     无上下文编码检测无法可靠区分。本项目以简体 GBK/GB18030 为主，故只处理 GB18030；
//     Big5 字节场景维持乱码但不比修复前更差。
//
// 步骤 1 的 utf8.ValidString 前置过滤是防止“合法 UTF-8 被误当 GBK 解出错码”的关键防线，
// 不可省略。例如 UTF-8 的 “中文”（E4 B8 AD E6 96 87）在 GB18030 下也能“解码成功”得到
// 另一个错误字符串，必须靠前置 ValidString 规避。
// 步骤 3 的 U+FFFD 检测用于拦截非法字节：GB18030 解码器遇到非法字节会插入 U+FFFD 而非报错，
// 若不拦截会把 "\xff\xfe\xfd" 这类非法串“洗”成替换字符序列，违反“不比修复前更差”原则。
func DecodeGBKToUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	if d, ok := tryDecode(s, simplifiedchinese.GB18030.NewDecoder()); ok && utf8.ValidString(d) && !strings.Contains(d, replacementRune) {
		return d
	}
	return s
}

// EncodeUTF8ToGBK 把 UTF-8 字符串编码为 GB18030 字节序列。
// 用于把前端 UTF-8 相对路径反向编码，以在磁盘上定位存量 GBK 文件名。
// 编码失败（含不可表示字符）时原样返回 s。
func EncodeUTF8ToGBK(s string) string {
	var buf bytes.Buffer
	enc := transform.NewWriter(&buf, simplifiedchinese.GB18030.NewEncoder())
	if _, err := io.WriteString(enc, s); err != nil {
		_ = enc.Close()
		return s
	}
	if err := enc.Close(); err != nil {
		return s
	}
	return buf.String()
}

// tryDecode 用指定 decoder 解码字节串，返回解码后的字符串与是否成功。
// 不吞错误，区别于历史内联代码的 `content, _ := io.ReadAll(decoder)`。
func tryDecode(s string, dec transform.Transformer) (string, bool) {
	r := transform.NewReader(bytes.NewReader([]byte(s)), dec)
	content, err := io.ReadAll(r)
	if err != nil {
		return "", false
	}
	return string(content), true
}
