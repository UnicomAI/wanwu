package shared

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// precheckCommand 在 exec 前做静态拦截。envMap key 为已注入的 skill var keys；
// 命中拒绝规则返回非 nil error，error.Error() 即作为返回给 LLM 的 stdout。
// 不打印 envMap value、不打印命令字符串中的可疑片段。
//
// 规则顺序设计原则：先做高确定性、低 FP 的硬规则（head 间接展开 → 套娃递归 → 内联/heredoc 脚本体扫描 →
// eval / 间接展开 / 命令替换枚举 → 枚举型 head → 中转变量赋值 → find/xargs 跳板 → 敏感文件 →
// 打印型 head 配合 $KEY 引用 → 写脚本文件 inline 扫描）。写脚本文件检测放最后，因为它需要全局扫描所有
// quoted 串，成本最高。head 间接展开放最前，因为一旦 head 是 $X / $(...)，后续所有基于 head 的判定都
// 失去意义（shell 展开后真实命令名 precheck 看不到），直接拒最快。
// splitCompoundCommands 按引号感知的 shell 命令分隔符拆分 cmdStr。
// 分隔符：&& / || / ; / 换行。单双引号内的分隔符视为字面量不拆。
// 管道 | 不在此处拆分——pipeShellRe 等规则已覆盖 | bash / | sh 的 dynamic exec，
// printenv | grep 等已由 cmdSubstEnumRe / secretEnumeratorCmds 覆盖。
func splitCompoundCommands(cmdStr string) []string {
	var parts []string
	var buf strings.Builder
	inSingle, inDouble := false, false
	for i := 0; i < len(cmdStr); i++ {
		c := cmdStr[i]
		if inSingle {
			buf.WriteByte(c)
			if c == '\'' {
				inSingle = false
			}
			continue
		}
		if inDouble {
			buf.WriteByte(c)
			if c == '"' && (i == 0 || cmdStr[i-1] != '\\') {
				inDouble = false
			}
			continue
		}
		if c == '\'' {
			inSingle = true
			buf.WriteByte(c)
			continue
		}
		if c == '"' {
			inDouble = true
			buf.WriteByte(c)
			continue
		}
		if c == '\\' && i+1 < len(cmdStr) {
			buf.WriteByte(c)
			buf.WriteByte(cmdStr[i+1])
			i++
			continue
		}
		// && / || / ; / 换行 作为命令分隔符
		if c == '&' && i+1 < len(cmdStr) && cmdStr[i+1] == '&' {
			flushCmdPart(&buf, &parts)
			i++ // 跳过第二个 &
			continue
		}
		if c == '|' && i+1 < len(cmdStr) && cmdStr[i+1] == '|' {
			flushCmdPart(&buf, &parts)
			i++ // 跳过第二个 |
			continue
		}
		if c == ';' {
			flushCmdPart(&buf, &parts)
			continue
		}
		buf.WriteByte(c)
	}
	flushCmdPart(&buf, &parts)
	return parts
}

func flushCmdPart(buf *strings.Builder, parts *[]string) {
	s := strings.TrimSpace(buf.String())
	if s != "" {
		*parts = append(*parts, s)
	}
	buf.Reset()
}

func precheckCommand(cmdStr string, envMap map[string]string) error {
	return precheckCommandWithDepth(cmdStr, envMap, 0)
}

// precheckCommandWithDepth 在 precheckCommand 基础上增加递归深度保护。
// depth 超过 4 时不再拆分复合命令，防止嵌套炸弹。
func precheckCommandWithDepth(cmdStr string, envMap map[string]string, depth int) error {
	if depth >= 4 {
		return nil
	}
	// 拆分复合命令：&& / || / ; / 换行 是 shell 命令分隔符，
	// 引号感知——单双引号内的分隔符不拆。
	// 拆分后逐条递归 precheck，任一命中即返回 error。
	subCmds := splitCompoundCommands(cmdStr)
	if len(subCmds) > 1 {
		for _, sub := range subCmds {
			if err := precheckCommandWithDepth(sub, envMap, depth+1); err != nil {
				return err
			}
		}
		return nil
	}

	tokens := strings.Fields(cmdStr) // 简化：按空格分词，复杂 shell quoting 不做完整解析（已知盲点）。
	if len(tokens) == 0 {
		return nil
	}

	// head normalize — 剥掉前缀赋值 (X=1) / command-exec-builtin 外壳 / (/{
	// / 反斜杠等常见绕 head 判定的语法糖，得到真实 head。
	head := normalizeHead(tokens)

	// head 是 shell 变量/命令替换展开（$X / ${X} / $(...)）—— shell 会先展开再执行，
	// precheck 静态分析看不到展开后的真实命令名，直接拒（绕过 head 判定的唯一途径）。
	if indirectCmdRe.MatchString(tokens[0]) {
		return fmt.Errorf("%s", blockedMsg("indirect-cmd", head))
	}

	// shell debug flags: bash -x / -v / -o xtrace 打印展开值
	if err := precheckShellFlags(cmdStr, head); err != nil {
		return err
	}
	// stdin-shell: 无参 bash / echo | bash 从 stdin 读命令（dynamic exec）
	if err := precheckStdinShell(cmdStr, tokens, head); err != nil {
		return err
	}
	// special-env-assign: BASH_ENV / LD_PRELOAD / PROMPT_COMMAND 等特殊 env 名赋值
	if err := precheckSpecialEnvAssign(cmdStr, head); err != nil {
		return err
	}

	// bash/sh/dash -c <arg> 套娃：把 <arg> 当新命令递归 precheck，拦下"包一层 shell 绕过"的尝试
	if head == "bash" || head == "sh" || head == "dash" {
		if inner, ok := extractDashCArg(cmdStr); ok {
			return precheckCommand(inner, envMap)
		}
	}

	// head-secret-ref: tokens[0] 含 envMap key 的 $KEY 引用（报错回显泄露）
	if err := precheckHeadSecretRef(tokens, envMap); err != nil {
		return err
	}

	// 解释器内联脚本（python/node/perl/ruby/php/awk 等）：三条件 AND（env 访问 ∧ 输出 sink ∧
	// envMap key 字面量）才拦，避免误伤合法的 "读 env 喂给 SDK" 模式（如 client=Cli(os.environ['K'])）。
	// 同时叠加 script-read-sensitive：body 内文件读 API + .env / .skill_env 字面量即拒（envMap 无关）。
	if body, ok := extractInterpreterScript(cmdStr, head, tokens); ok {
		if err := precheckScriptBody(body, envMap); err != nil {
			return err
		}
		if err := precheckScriptReadSensitive(body); err != nil {
			return err
		}
	}
	// heredoc 内联（仅在 head 是解释器或 bash/sh/dash 时触发，避免给 cat << EOF 这种普通文档误判）
	if _, isInterp := scriptInterpreterCmds[head]; isInterp || head == "bash" || head == "sh" || head == "dash" {
		if body, ok := extractHeredocBody(cmdStr); ok {
			if err := precheckHeredocBody(body, head, envMap); err != nil {
				return err
			}
			if err := precheckScriptReadSensitive(body); err != nil {
				return err
			}
		}
	}
	// eval 关键字 — 无条件拒（LLM 自动生成场景下 eval 几无合法用途）
	if evalKeywordRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("eval", head))
	}
	// 间接展开 ${!VAR} — 无条件拒（绕过 $KEY 字面匹配的唯一语法路径）
	if indirectExpandRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("indirect-expand", head))
	}
	// 命令替换嵌枚举命令 $(env)/$(printenv)/... — 等价于直接执行枚举命令，复用 enumerator 标签
	if cmdSubstEnumRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("enumerator", head))
	}
	// 进程替换 <(env) / <(printenv) 语义等价命令替换嵌枚举，同样拒
	if processSubstEnumRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("enumerator", head))
	}
	// 枚举型命令：env/printenv/set/... 无差别打印所有环境变量，无条件拒
	if _, ok := secretEnumeratorCmds[head]; ok {
		return fmt.Errorf("%s", blockedMsg("enumerator", head))
	}
	// 中转变量赋值：任何 NAME=...$KEY... 形态一律拒（赋值步骤即拒，杜绝 N 跳链）
	if err := precheckRelayAssign(cmdStr, envMap); err != nil {
		return err
	}
	// find/xargs 跳板：`find ... -exec cat` / `... | xargs cat` 到敏感文件。
	// 放在 sensitiveFileRe 之前，让 tool-jump 标签比通用 sensitive-file 更精准。
	if err := precheckToolJump(cmdStr); err != nil {
		return err
	}
	// ln/cp/mv/rsync/install + .env / .skill_env 字面量（防别名化后再读）
	if err := precheckSensitiveFileHandler(cmdStr, head); err != nil {
		return err
	}
	// 敏感文件 / /proc/<pid>/environ
	if sensitiveFileRe.MatchString(cmdStr) || procEnvironRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("sensitive-file", head))
	}
	// 写脚本文件 + inline 内容含越权代码三条件；同时覆盖 heredoc body / 变量展开 target / shell rc 文件 target。
	// 放在 ref-printer 之前，因为 `echo '...' > /tmp/x.py` 里 echo 引用 $KEY 会先命中 ref-printer；
	// 但真实语义是"写含泄露代码的脚本"，用 script-write 标签更精准。
	if err := precheckScriptWrite(cmdStr, envMap); err != nil {
		return err
	}
	// 打印型命令 + $KEY 引用，两条件必须同时满足，否则放行（避免 echo "hello" 这种被错杀）
	if _, ok := refPrinterCmds[head]; ok && referencesSecret(cmdStr, envMap) {
		return fmt.Errorf("%s", blockedMsg("ref-printer", head))
	}
	// 打印型 head + `.*` / `.?<any>` 之类宽泛 glob → 拒（防 cat .* 展开到 .env）
	if err := precheckBroadGlob(cmdStr, head); err != nil {
		return err
	}
	// base64 -d / xxd -r 解码命令：常用于 dynamic exec 后接 pipe 到 shell → 拒
	if err := precheckDecodePipe(cmdStr, head); err != nil {
		return err
	}
	// trap 'code' EXIT / SIGxxx 延迟执行：body 走 bash 三条件（防 trap 里 echo $KEY 泄露）
	if err := precheckTrapBody(cmdStr, head, envMap); err != nil {
		return err
	}
	return nil
}

// buildQuoteState 返回与 cmdStr 等长的 byte slice，标注每个字节的引号态：
//
//	0 = 引号外
//	1 = 单引号内（bash 单引号不处理转义）
//	2 = 双引号内
//
// 用途：让"按位置扫描字符串"的规则（NAME= 赋值、重定向 target、$-prefix token 判定）
// 能跳过引号内区间，避免把 echo "NAME=$KEY" / awk '$0' 里的字面内容误判为 shell 语法。
//
// 简化：不处理双引号内 \" 转义、$'...' ANSI-C quoting。最坏结果是引号态分类偏差，
// 不引入新泄露路径。
func buildQuoteState(cmdStr string) []byte {
	state := make([]byte, len(cmdStr))
	var mode byte // 0 out, 1 single, 2 double
	for i := 0; i < len(cmdStr); i++ {
		c := cmdStr[i]
		switch mode {
		case 0:
			switch c {
			case '\'':
				mode = 1
			case '"':
				mode = 2
			}
		case 1:
			if c == '\'' {
				mode = 0
			}
		case 2:
			if c == '"' {
				mode = 0
			}
		}
		state[i] = mode
	}
	return state
}

// maskQuotesToSpaces 返回一个与 cmdStr 等长的字符串，其中所有引号内的非引号、非换行字符
// 被替换为空格，引号字符本身和引号外内容保持不变。用于让"按位置匹配正则"的规则
// （如 extractRedirectTargets 找 `>` / `>>`）跳过引号内的语法字符。
//
// 位置保持：masked[i] 与 cmdStr[i] 一一对应，可直接把 masked 上的正则匹配位置映射回 cmdStr。
func maskQuotesToSpaces(cmdStr string) string {
	qs := buildQuoteState(cmdStr)
	out := []byte(cmdStr)
	for i, s := range qs {
		if s == 0 {
			continue
		}
		c := out[i]
		if c == '\'' || c == '"' || c == '\n' {
			continue
		}
		out[i] = ' '
	}
	return string(out)
}

// splitFieldsRespectQuotes 按空白分词，单/双引号包裹的整块视为一个 token
// （引号本身保留在 token 中，供下游做 $-prefix 判定）。
//
// 用途：替代 strings.Fields 供 precheckScriptFile 使用——避免 `awk 'BEGIN{print $0}'`
// 被拆散、awk 的 `$0`/`$1` 字段引用被误判为 shell 变量展开。
//
// 简化：不处理 \" / \' 转义、$'...' ANSI-C quoting。分词偏差最坏影响是"误合并/误拆分"
// 一小段 token，不引入新泄露路径。
func splitFieldsRespectQuotes(cmdStr string) []string {
	var out []string
	var cur []byte
	var mode byte // 0 out, 1 single, 2 double
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = cur[:0]
		}
	}
	for i := 0; i < len(cmdStr); i++ {
		c := cmdStr[i]
		switch mode {
		case 0:
			switch c {
			case '\'':
				mode = 1
				cur = append(cur, c)
			case '"':
				mode = 2
				cur = append(cur, c)
			case ' ', '\t', '\n', '\r':
				flush()
			default:
				cur = append(cur, c)
			}
		case 1:
			cur = append(cur, c)
			if c == '\'' {
				mode = 0
			}
		case 2:
			cur = append(cur, c)
			if c == '"' {
				mode = 0
			}
		}
	}
	flush()
	return out
}

// extractQuotedStrings 抽取 cmdStr 中所有单/双引号包裹的子串（不含引号本身）。
// 简化解析：不处理转义引号嵌套、$'...' ANSI-C quoting。
func extractQuotedStrings(cmdStr string) []string {
	var out []string
	var i int
	for i < len(cmdStr) {
		c := cmdStr[i]
		if c != '\'' && c != '"' {
			i++
			continue
		}
		// 找配对引号
		end := strings.IndexByte(cmdStr[i+1:], c)
		if end < 0 {
			break
		}
		out = append(out, cmdStr[i+1:i+1+end])
		i += 1 + end + 1
	}
	return out
}

// extractRedirectTargets 抽取 cmdStr 中所有重定向目标文件路径（> / >> / tee <flags...> <target> 后的 token）。
//
// 引号感知：先把引号内的非引号字符换成空格再跑正则，避免把 awk `'BEGIN{print $0 > "f"}'`
// 或 `echo "> path"` 里的 `>` 误识别成 shell 重定向。masked 与 cmdStr 位置一一对应，
// 匹配到的 target 直接取 masked 上的 substring；引号内的 target 因内容已被 mask，
// 天然不会作为有意义路径出现。
//
// 简化：tee 仅识别紧跟其后的首个非 flag token；不处理重定向到 /dev/fd/N 等特殊文件。
func extractRedirectTargets(cmdStr string) []string {
	masked := maskQuotesToSpaces(cmdStr)
	var out []string
	// > target / >> target
	redirRe := regexp.MustCompile(`>>?\s*([^\s;&|()<>]+)`)
	for _, m := range redirRe.FindAllStringSubmatch(masked, -1) {
		if len(m) > 1 && m[1] != "" {
			out = append(out, m[1])
		}
	}
	// tee [flags...] target
	teeRe := regexp.MustCompile(`\btee\b(?:\s+-[a-zA-Z]+)*\s+([^\s;&|()<>]+)`)
	for _, m := range teeRe.FindAllStringSubmatch(masked, -1) {
		if len(m) > 1 && m[1] != "" {
			out = append(out, m[1])
		}
	}
	return out
}

// extractHeredocBody 从 cmdStr 提取 heredoc body。
// 算法：
//  1. 用 heredocStartRe 找到 "<<TAG" 起始处，捕获 TAG 名
//  2. 从起始处往后找首个换行（heredoc body 起点）
//  3. 在 body 起点之后扫描 "\nTAG\n" / "\nTAG$" / "\nTAG;" 等结束定界（前面允许空白用于 <<- 缩进版）
//
// 不展开变量；C3 字面量仍可命中。简化：不处理多 heredoc 嵌套、heredoc 内的字符串恰好等于 TAG 的边缘情况。
func extractHeredocBody(cmdStr string) (string, bool) {
	loc := heredocStartRe.FindStringSubmatchIndex(cmdStr)
	if loc == nil {
		return "", false
	}
	tag := cmdStr[loc[2]:loc[3]]
	if tag == "" {
		return "", false
	}
	// 找起始 token 之后的首个换行
	bodyStart := strings.IndexByte(cmdStr[loc[1]:], '\n')
	if bodyStart < 0 {
		return "", false
	}
	bodyStart += loc[1] + 1
	// 在 body 起点之后扫描结束定界 "\n<空白>*TAG<空白>*(\n|$|;|&&|||)"
	rest := cmdStr[bodyStart:]
	for offset := 0; offset < len(rest); {
		idx := strings.Index(rest[offset:], "\n"+tag)
		if idx < 0 {
			// 也允许 body 单独占一行而 TAG 在文件末尾（无前导换行）
			break
		}
		absIdx := offset + idx
		// TAG 后面必须是空白/换行/分隔符/字符串结尾
		afterTag := absIdx + 1 + len(tag)
		if afterTag >= len(rest) {
			return rest[:absIdx], true
		}
		c := rest[afterTag]
		if c == '\n' || c == ' ' || c == '\t' || c == ';' || c == '&' || c == '|' || c == '\r' {
			return rest[:absIdx], true
		}
		offset = absIdx + 1
	}
	return "", false
}

// precheckHeredocBody 在 heredoc body 上跑三条件 AND（env 访问 ∧ 输出 sink ∧ envMap key 字面量）。
// bash/sh/dash 用 bashEnvAccessRe (识别 $KEY)；其它解释器沿用 scriptEnvAccessRe (os.environ / process.env / 等)。
// 命中即返回 [BLOCKED:heredoc]。
func precheckHeredocBody(body, head string, envMap map[string]string) error {
	if body == "" || len(envMap) == 0 {
		return nil
	}
	if head == "bash" || head == "sh" || head == "dash" {
		if !bashEnvAccessRe.MatchString(body) {
			return nil
		}
	} else {
		if !scriptEnvAccessRe.MatchString(body) {
			return nil
		}
	}
	if !scriptSinkRe.MatchString(body) {
		return nil
	}
	for k := range envMap {
		if k == "" {
			continue
		}
		matched, err := regexp.MatchString(`\b`+regexp.QuoteMeta(k)+`\b`, body)
		if err == nil && matched {
			return fmt.Errorf("%s", blockedMsg("heredoc", head))
		}
	}
	return nil
}

// precheckScriptWrite 拦截"把含越权代码的脚本写入文件"的命令。
// 触发：cmdStr 中存在重定向到脚本后缀文件（scriptFileExtRe）/ shell rc 文件（shellRcFileRe）
// / 变量展开 target `$VAR` 之一，AND quoted body 或 heredoc body 命中越权代码。
//
// 具体覆盖的三种 target：
//   - 脚本后缀文件（.py/.sh/.js/... 等）+ inline body 三条件命中
//   - shell rc 文件（.bashrc/.zshrc/.profile 等）+ inline body 三条件命中
//   - 变量展开 target（$VAR / ${VAR} / $(...)）+ inline body 三条件命中（无论后缀，静态无法解析路径）
//
// 其中 inline body 扫描同时覆盖 quoted string 和 heredoc body（`cat > x.py << EOF ... EOF`）。
//
// 命中返回 [BLOCKED:script-write]（第一优先）或原始规则标签。
func precheckScriptWrite(cmdStr string, envMap map[string]string) error {
	targets := extractRedirectTargets(cmdStr)
	if len(targets) == 0 {
		return nil
	}
	hasSuspiciousTarget := false
	for _, t := range targets {
		// 脚本后缀 target
		if scriptFileExtRe.MatchString(t) {
			hasSuspiciousTarget = true
			break
		}
		// shell rc 文件 target
		if shellRcFileRe.MatchString(t) {
			hasSuspiciousTarget = true
			break
		}
		// 变量展开 target
		if strings.HasPrefix(t, "$") {
			hasSuspiciousTarget = true
			break
		}
	}
	if !hasSuspiciousTarget {
		return nil
	}

	// 收集所有待扫描的 body：quoted string + heredoc body
	bodies := extractQuotedStrings(cmdStr)
	if hbody, ok := extractHeredocBody(cmdStr); ok {
		bodies = append(bodies, hbody)
	}
	for _, body := range bodies {
		if len(envMap) > 0 {
			// 解释器语义：os.environ / process.env + print/console.log + envMap key 字面量
			if err := precheckScriptBody(body, envMap); err != nil {
				return fmt.Errorf("%s", blockedMsg("script-write", "redirect"))
			}
			// bash 语义：$KEY / ${KEY} + echo/printf + envMap key 字面量
			// 覆盖 `echo 'echo $MY_TOKEN' > /tmp/x.sh` 这类"往脚本里写 shell 泄露命令"
			if err := precheckHeredocBody(body, "bash", envMap); err != nil {
				return fmt.Errorf("%s", blockedMsg("script-write", "redirect"))
			}
		}
		if err := precheckScriptReadSensitive(body); err != nil {
			return err
		}
	}
	return nil
}

// precheckRelayAssign 拦截"把 envMap key 赋值给另一个变量"的语句。
// 触发：cmdStr 中存在 NAME=<RHS>，且 RHS 中引用了 envMap 中任一 key（$KEY 或 ${KEY}）。
// 命中即返回 [BLOCKED:relay-assign]，不打印 RHS 内容。
//
// 设计：在赋值那一步就拒，杜绝中转变量的 N 跳链——A=$T 即拒，根本到不了 B=$A。
// FP 控制：
//   - 引号感知：若 NAME= 命中位置在单/双引号内（如 `echo "NAME=\$KEY"` 的字面串），跳过。
//     bash 引号内的 `NAME=...` 不构成赋值，`\$KEY` 也是转义字面 `$`（不展开、不泄露 value）。
//   - 所有中转变量在合法场景里都有 inline 等价写法（curl -H "X: $T" 替代 H=$T; curl -H "$H"），
//     拒赋值仅影响代码风格/可读性，功能完整性不受影响。
func precheckRelayAssign(cmdStr string, envMap map[string]string) error {
	if len(envMap) == 0 {
		return nil
	}
	qs := buildQuoteState(cmdStr)
	for _, m := range relayAssignRe.FindAllStringSubmatchIndex(cmdStr, -1) {
		// m 索引：0/1 整体匹配，2/3 前导边界 group，4/5 NAME group
		if len(m) < 6 || m[5] < 0 {
			continue
		}
		// NAME 起点在引号内 → 属于引号内字面串（echo/printf 输出等），跳过
		if m[4] < len(qs) && qs[m[4]] != 0 {
			continue
		}
		// "=" 在 NAME 之后；RHS 从 m[5]+1 (= 之后的位置) 开始
		rhs := extractAssignmentRHS(cmdStr, m[5]+1)
		if referencesSecret(rhs, envMap) {
			return fmt.Errorf("%s", blockedMsg("relay-assign", "assignment"))
		}
	}
	return nil
}

// extractAssignmentRHS 从 cmdStr[start:] 抽取赋值的 RHS 部分。
// 支持单/双引号包裹（引号内不切）；裸值到下个 shell 分隔符（空白 / ; / && / || / | / ) / 行尾）为止。
// 简化：不处理转义引号嵌套、$(...) 内部嵌套引号。
func extractAssignmentRHS(cmdStr string, start int) string {
	if start >= len(cmdStr) {
		return ""
	}
	// 引号包裹：取到配对引号为止
	c := cmdStr[start]
	if c == '\'' || c == '"' {
		end := strings.IndexByte(cmdStr[start+1:], c)
		if end < 0 {
			return cmdStr[start+1:] // 不配对，取剩余
		}
		return cmdStr[start+1 : start+1+end]
	}
	// 裸值：到下个 shell 分隔符为止
	i := start
	for i < len(cmdStr) {
		ch := cmdStr[i]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == ';' || ch == '|' || ch == '&' || ch == ')' {
			break
		}
		i++
	}
	return cmdStr[start:i]
}

// referencesSecret 判断 cmd 中是否引用了 envMap 中的任一 key。
// 覆盖 bash 参数展开全变体：
//   - $KEY / ${KEY}  基础形态
//   - ${KEY:-x} / ${KEY:=x} / ${KEY:?x} / ${KEY:+x}  默认值展开
//   - ${KEY#*} / ${KEY##*} / ${KEY%*} / ${KEY%%*}  前后缀移除
//   - ${KEY:0:5} / ${KEY:5}  子串
//   - ${KEY/A/B} / ${KEY//A/B}  替换
//   - ${KEY^^} / ${KEY,,} / ${KEY^} / ${KEY,}  大小写
//   - ${KEY@Q} / ${KEY@P} / ${KEY@U} / ${KEY@A}  @-operator (@P 展开成 value)
//   - ${!KEY} 间接展开也顺带命中（虽有独立 indirect-expand 规则拦，此处也算引用）
func referencesSecret(cmd string, envMap map[string]string) bool {
	for k := range envMap {
		if k == "" {
			continue
		}
		// $KEY\b：边界要求避免 $TOKEN 误命中 $TOKENS
		pattern := `\$` + regexp.QuoteMeta(k) + `\b`
		if matched, err := regexp.MatchString(pattern, cmd); err == nil && matched {
			return true
		}
		// ${KEY} 基础形态
		if strings.Contains(cmd, "${"+k+"}") {
			return true
		}
		// ${KEY<expansion-op>...} — 覆盖所有 bash 参数展开变体
		// KEY 后紧跟 bash 展开操作符或字符串结尾
		expandPattern := `\$\{[!]?` + regexp.QuoteMeta(k) + `[\s}:#%/,\^@]`
		if matched, err := regexp.MatchString(expandPattern, cmd); err == nil && matched {
			return true
		}
	}
	return false
}

// extractDashCArg 从 `bash -c <arg>` 中提取 <arg>，支持单/双引号或裸 token。
// 实现仅靠字符串切分，复杂 quoting（如转义引号嵌套）是已知盲点。
func extractDashCArg(cmd string) (string, bool) {
	_, after, found := strings.Cut(cmd, "-c")
	if !found {
		return "", false
	}
	rest := strings.TrimLeft(after, " \t")
	if rest == "" {
		return "", false
	}
	// 去掉首尾配对引号
	if (rest[0] == '\'' || rest[0] == '"') && len(rest) >= 2 && rest[len(rest)-1] == rest[0] {
		return rest[1 : len(rest)-1], true
	}
	return rest, true
}

// extractInterpreterScript 从命令字符串提取解释器内联脚本 body。
//   - 对 scriptInterpreterCmds 内的 head：在 cmdStr 中找首个匹配的 flag 紧跟空白后的下一参数（含配对引号）。
//   - 对 awk/gawk/mawk：脚本是首个位置参数（跳过 -F<sep> / -v <var=val> / -f <file> 这些 flag）。
//   - body 若被配对单/双引号包裹则剥掉一层；不命中或 body 为空返回 ("", false)。
//
// 实现：用 regexp 一次性匹配 `flag\s+('body'|"body"|bare)`；复杂 shell quoting（转义嵌套、shell 续行、
// 变量替换、--flag=value 等长选项写法）是已知盲点。
func extractInterpreterScript(cmdStr, head string, _ []string) (string, bool) {
	// awk 走位置参数语义
	if head == "awk" || head == "gawk" || head == "mawk" {
		return extractAwkProgram(cmdStr)
	}
	flags, ok := scriptInterpreterCmds[head]
	if !ok {
		return "", false
	}
	for _, f := range flags {
		// 匹配 "<空白><flag><空白>('body'|"body"|bare)"。flag 前必须有空白避免误命中 head 子串。
		// (?s) 让 . 跨行（罕见但 heredoc 替换后可能出现）。
		pattern := `\s` + regexp.QuoteMeta(f) + `\s+(?s)(?:'([^']*)'|"([^"]*)"|(\S+))`
		re := regexp.MustCompile(pattern)
		if m := re.FindStringSubmatch(" " + cmdStr); m != nil {
			for i := 1; i < len(m); i++ {
				if m[i] != "" {
					return m[i], true
				}
			}
		}
	}
	return "", false
}

// extractAwkProgram 跳过 awk 的 flag 后取首个位置参数作为 body。
// 已知盲点：仅识别独立 flag（-F,/-v X=1）形式，--field-separator= 等长格式不识别。
func extractAwkProgram(cmdStr string) (string, bool) {
	// 用正则一次性匹配：awk 头 + 零或多组（flag + 其值） + 首位置参数
	// flag 形态：-F<x>（粘连）/ -F <x> / -v <var=val> / -f <file> / 其它 -X
	// 位置参数：'body' / "body" / bare
	re := regexp.MustCompile(
		`^(?:awk|gawk|mawk)\b(?:\s+(?:-F\S*|-[vf]\s+\S+|-\S+))*\s+(?:'([^']*)'|"([^"]*)"|(\S+))`)
	m := re.FindStringSubmatch(cmdStr)
	if m == nil {
		return "", false
	}
	for i := 1; i < len(m); i++ {
		if m[i] != "" {
			return m[i], true
		}
	}
	return "", false
}

// precheckScriptBody 对解释器脚本 body 做三条件 AND 拦截，覆盖内联 (-c/-e)、
// heredoc、落盘 .py 三条 chain。三者全中返回 [BLOCKED:script-body]，任一缺失放行：
//   - 命中 scriptEnvAccessRe（脚本读取了 process env）
//   - 命中 scriptSinkRe（脚本会向 stdout/stderr 输出）
//   - body 中明文出现 envMap 任一 key 的 \b 边界匹配（确认引用了已注入的 skill key）
//
// 早期版本这里首步会跑 precheckScriptOpen —— 只要 body 出现 open( / .read( /
// fs.writeFile* 等 file I/O 内建就一律拒 —— 目的是防"open('.'+'env')" 拼装字面量
// 绕过 sensitive-file 检查。但该规则把 LLM 在 sandbox 写代码文件（heredoc /
// `cat > x.py <<EOF` / `python -c "with open(fp,'w')..."`）这一高频合法场景 100% 命中，
// 已删除；.env / .skill_env 字面量读取仍由 precheckScriptReadSensitive 兜底。
func precheckScriptBody(body string, envMap map[string]string) error {
	if body == "" || len(envMap) == 0 {
		return nil
	}
	if !scriptEnvAccessRe.MatchString(body) {
		return nil
	}
	if !scriptSinkRe.MatchString(body) {
		return nil
	}
	for k := range envMap {
		if k == "" {
			continue
		}
		matched, err := regexp.MatchString(`\b`+regexp.QuoteMeta(k)+`\b`, body)
		if err == nil && matched {
			return fmt.Errorf("%s", blockedMsg("script-body", "interpreter"))
		}
	}
	return nil
}

// precheckToolJump 检测 find / xargs 跳板到打印命令读取敏感文件的手法。
// 触发条件（任一命中）：
//   - `find ... -exec <printer>` + `-name <env pattern>`，或整条 cmdStr 命中 sensitiveFileRe
//   - `... | xargs <printer>` + pipeline 前段用 echo/printf 输出 env 路径字面量
//
// 命中即返回 [BLOCKED:tool-jump]。envMap 无关（护 .env / .skill_env 本身）。
func precheckToolJump(cmdStr string) error {
	hasFind := findExecReaderRe.MatchString(cmdStr)
	hasXargs := xargsReaderRe.MatchString(cmdStr)
	if !hasFind && !hasXargs {
		return nil
	}
	// 情况 1：find -exec printer + -name 指向 env 家族
	if hasFind && findNamePatternEnvRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("tool-jump", "find"))
	}
	// 情况 2：find -exec printer + 命令体命中 sensitiveFileRe（如 find -name .env* 已被上面命中，
	// 但如 find /workspace -path '*/.env' 走这里）
	if hasFind && sensitiveFileRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("tool-jump", "find"))
	}
	// 情况 3：pipeline 前段 echo/printf 敏感路径 → xargs cat
	if hasXargs && echoPipeEnvRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("tool-jump", "xargs"))
	}
	return nil
}

// precheckScriptReadSensitive 检测"解释器脚本读取 .env / .skill_env 文件"的路径。
// 触发条件（两条件 AND）：
//   - body 中出现文件读 API（open / .read( / fs.readFile / File.read / fopen / ...）
//   - body 中出现 .env / .skill_env 路径字面量
//
// 命中即返回 [BLOCKED:script-read-sensitive]。envMap 是否为空都拒：
// workspace/.env（含 OPENAI_API_KEY）的保护与 skill_env 独立，即使无 skill var 也不该被解释器直接读。
//
// 已知残留：字符串拼装（'.skill_e'+'nv'）/ chr / base64 编码路径静态分析穷不完，列入残留清单。
func precheckScriptReadSensitive(body string) error {
	if body == "" {
		return nil
	}
	if !scriptFileReadRe.MatchString(body) {
		return nil
	}
	if !scriptSensitivePathLiteralRe.MatchString(body) {
		return nil
	}
	return fmt.Errorf("%s", blockedMsg("script-read-sensitive", "interpreter"))
}

// blockedMsg 拼接给 LLM 的拒绝提示。固定文本不含 value，便于 LLM 模式匹配后调整策略。
func blockedMsg(rule, head string) string {
	return fmt.Sprintf(
		"[BLOCKED:%s] Command %q rejected: would print or dump skill environment variables. "+
			"Use $KEY references directly in network commands (curl, etc.) or in file writes to external systems, "+
			"but NEVER via echo / printf / cat / env / printenv / encode commands. See SKILL.md.",
		rule, head,
	)
}

// ─── 新增 helper（head 规范化 + shell/token 层规则） ───

// normalizeHead 对 tokens[0] 做规范化，剥掉常见"绕 head 判定"的语法糖，返回 (真实 head, 是否可继续 head 判定)。
// head 判定入口前置逻辑，闭合多个绕过变体：
//   - "X=fake env"（前缀赋值）→ 剥赋值前缀，返回 env
//   - "command env" / "exec env" / "builtin env"（bypass alias 内建）→ 剥外壳，返回 env
//   - "(env)" / "{ env; }"（子 shell / 命令组）→ strip 前导 (/{
//   - "\env"（backslash 绕 alias）→ strip 反斜杠
//
// 返回值：normalize 后的 head（小写、剥 /usr/bin/ 前缀）；若无法规范化返回原始 tokens[0] 处理结果。
// 递归深度限制 4 层，防止病态输入死循环。
func normalizeHead(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	idx := 0
	depth := 0
	for depth < 4 && idx < len(tokens) {
		tok := tokens[idx]
		// 剥前导 (/{
		tok = strings.TrimLeft(tok, "({")
		// 剥反斜杠
		tok = strings.TrimPrefix(tok, `\`)
		// 剥 /usr/bin/ /bin/
		tok = strings.TrimPrefix(tok, "/usr/bin/")
		tok = strings.TrimPrefix(tok, "/bin/")
		lower := strings.ToLower(tok)
		// NAME=value 前缀赋值：LHS 必须是合法 env name
		if eqIdx := strings.Index(tok, "="); eqIdx > 0 {
			lhs := tok[:eqIdx]
			if isValidEnvName(lhs) {
				idx++
				depth++
				continue
			}
		}
		// command / exec / builtin 外壳
		if lower == "command" || lower == "exec" || lower == "builtin" {
			idx++
			depth++
			continue
		}
		return lower
	}
	return ""
}

// isValidEnvName 判断 s 是否符合 POSIX env name（[A-Za-z_][A-Za-z0-9_]*）。
func isValidEnvName(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		if i == 0 {
			if c != '_' && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
				return false
			}
			continue
		}
		if c != '_' && (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') {
			return false
		}
	}
	return true
}

// precheckShellFlags 检测 bash/sh/dash 的 debug flag（-x / -v / -o xtrace / --xtrace / --verbose）。
// 这些 flag 会 stderr 打印每条命令展开值，包含 $KEY 展开结果 → 泄露。
// LLM 自动化场景无合法用途；合法 debug 用 set -x 内联已被枚举命令规则拦下。
func precheckShellFlags(cmdStr, head string) error {
	if head != "bash" && head != "sh" && head != "dash" {
		return nil
	}
	if shellDebugFlagRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("shell-flags", head))
	}
	return nil
}

// precheckStdinShell 检测无参 bash / echo | bash 等 "从 stdin 读命令" 的 dynamic exec 手法。
// head=bash/sh/dash 无位置参数（无脚本文件）+ 无 -c → 拒；
// pipeline "... | bash" 尾接 shell → 拒。
func precheckStdinShell(cmdStr string, tokens []string, head string) error {
	// pipeline 版
	if pipeShellRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("stdin-shell", head))
	}
	// 单独 head 版
	if head != "bash" && head != "sh" && head != "dash" {
		return nil
	}
	// 检查 tokens[1:] 里是否有 "-c" 或位置参数（脚本文件）
	skipNextArg := map[string]bool{"-c": true, "-o": true, "--rcfile": true, "--init-file": true}
	hasPositional := false
	hasDashC := false
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == "-c" {
			hasDashC = true
			break
		}
		if strings.HasPrefix(tok, "-") {
			if skipNextArg[tok] {
				i++
			}
			continue
		}
		hasPositional = true
		break
	}
	if !hasDashC && !hasPositional {
		return fmt.Errorf("%s", blockedMsg("stdin-shell", head))
	}
	return nil
}

// precheckSpecialEnvAssign 检测赋值 LHS 是 BASH_ENV / ENV / PROMPT_COMMAND / LD_* 等特殊 env 名。
// 这些 env 的语义是"影响 shell 启动/运行时行为"，LLM 无合法用途；
// 攻击 pattern: BASH_ENV=/tmp/rc sh -c ':' 会 source /tmp/rc。
// 一律拒（不管 RHS 引用什么）。
func precheckSpecialEnvAssign(cmdStr, head string) error {
	if specialEnvAssignRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("special-env-assign", head))
	}
	return nil
}

// precheckHeadSecretRef 检测 tokens[0] 里出现 envMap 中任一 key 的 $KEY / ${KEY} 引用。
// 攻击 pattern: nonexistent_$MY_TOKEN 会展开后 bash 报错 "nonexistent_<VALUE>: command not found"
// 到 stderr，值泄露给 LLM。合法命令名不含 secret。
//
// 排除：tokens[0] 是 NAME=... 形态的赋值（如 MY=$MY_TOKEN），这类交给 relay-assign 规则处理；
// 此处只关心"看起来像命令名"的 head token 里含 secret 引用。
func precheckHeadSecretRef(tokens []string, envMap map[string]string) error {
	if len(tokens) == 0 || len(envMap) == 0 {
		return nil
	}
	tok := tokens[0]
	// 排除赋值形态（NAME=... 由 relay-assign 处理）
	if eqIdx := strings.Index(tok, "="); eqIdx > 0 && isValidEnvName(tok[:eqIdx]) {
		return nil
	}
	if referencesSecret(tok, envMap) {
		return fmt.Errorf("%s", blockedMsg("head-secret-ref", tok))
	}
	return nil
}

// precheckTrapBody 检测 trap 命令的 body 是否含 bash 三条件 (env 访问 + sink + envMap key)。
// trap 'echo $KEY' EXIT 在 EXIT 触发时执行 body，body 展开 $KEY 并输出。
// 复用 precheckHeredocBody(head="bash") 逻辑。
func precheckTrapBody(cmdStr, head string, envMap map[string]string) error {
	if head != "trap" {
		return nil
	}
	if len(envMap) == 0 {
		return nil
	}
	for _, body := range extractQuotedStrings(cmdStr) {
		if perr := precheckHeredocBody(body, "bash", envMap); perr != nil {
			return fmt.Errorf("%s", blockedMsg("trap-body", head))
		}
	}
	return nil
}

// precheckBroadGlob 检测 refPrinter head + `.*` / `.?<any>` 之类宽泛 glob。
// cat .* 展开为所有 . 起头文件（含 .env / .skill_env）。合法用途极少，拒。
// FP 边界：`ls .*` 常见但 ls 不在 refPrinter → 天然放行。
func precheckBroadGlob(cmdStr, head string) error {
	if _, ok := refPrinterCmds[head]; !ok {
		return nil
	}
	if broadGlobRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("broad-glob", head))
	}
	return nil
}

// precheckSensitiveFileHandler 检测 ln/cp/mv/rsync/install + .env / .skill_env 字面量。
// 攻击 pattern: ln -s .skill_env.json /tmp/x; cat /tmp/x（cat /tmp/x 不含 .env 绕敏感文件正则）。
// 由文件操作命令 + 敏感文件字面量 → 拒，防止别名化绕过。
func precheckSensitiveFileHandler(cmdStr, head string) error {
	if !sensitiveFileHandlerRe.MatchString(head) {
		return nil
	}
	if sensitiveFileRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("sensitive-file", head))
	}
	return nil
}

// precheckDecodePipe 检测 base64 -d / xxd -r 解码命令。
// base64 -d <<< 'Y2F0IC5lbnY=' | sh 之类编码后管道执行手法。
// LLM 自动化场景无合法用途；无差别拒。
func precheckDecodePipe(cmdStr, head string) error {
	if base64DecodeRe.MatchString(cmdStr) || xxdRevertRe.MatchString(cmdStr) {
		return fmt.Errorf("%s", blockedMsg("decode-pipe", head))
	}
	return nil
}

// precheckScriptFile 拦截 "命令行传入 workspace 内脚本文件时脚本 body 含越权代码"。
//
// 扫所有 tokens 位置参数：任何位置参数（不论 head 是 python / bash / cat / grep / less / cp / mv 等）
// 只要指向 workspace 内的脚本后缀文件，都读进来做 body-scan。
//
// 触发规则：
//  1. 位置参数是 /dev/stdin / /dev/fd/* / /proc/self/fd/* → 拒（stdin-source 标签）
//  2. 位置参数后缀在 scriptFileExtSet 内 + 路径在 workDir 边界内 + 文件存在 + size ≤ 上限
//     → 读进来跑 precheckScriptReadSensitive + precheckScriptBody；任一命中即拒
//
// 路径信任分级：
//   - workspace/skills/** 下的脚本视为可信预置（构建期审计），免除三条件 body-scan；
//     但 precheckScriptReadSensitive（读 .env 两条件）仍生效，作为硬防线。
//   - workspace 其它路径（tmp / output）不豁免，两层扫描都走。
//
// 兜底策略：
//   - 文件不存在 / 超大 / 读失败 / 路径越出 workDir → 跳过此 arg（不阻断命令）
//   - workDir 为空 → 放行（测试或极端 corner case）
//   - envMap 空时 precheckScriptBody 不触发，但 precheckScriptReadSensitive 仍会拦（护 .env 本身）
//
// 历史注：早期版本这里有 "位置参数以 $ 开头一律拒"（indirect-target）分支，因误伤
// `python3 script.py --token "$KEY"` 这类合法 skill 调用模式已删除。头位置的变量
// 展开（$X / $(...)）仍由 precheck.go 顶层 indirectCmdRe 拦。
func precheckScriptFile(cmdStr, workDir string, envMap map[string]string) error {
	// 用引号感知 tokenizer 替代 strings.Fields：把 `awk 'BEGIN{print $0}'` 的整块 body
	// 视为一个 token，避免 awk 的字段引用 $0/$1/... 被独立看作 "$-prefix 位置参数"。
	// （历史遗留原因：早期这里的 $-prefix 判定会拒 script-file 层的位置参数，现已删除；
	// 但引号感知分词保留，因其对其它 $KEY 相关判定也更准确。）
	tokens := splitFieldsRespectQuotes(cmdStr)
	if len(tokens) < 2 {
		return nil
	}
	head := strings.TrimPrefix(tokens[0], "/usr/bin/")
	head = strings.TrimPrefix(head, "/bin/")
	head = strings.ToLower(head)

	if workDir == "" {
		return nil
	}
	absWorkDir, wderr := filepath.Abs(workDir)
	if wderr != nil {
		return nil
	}

	// "flag + 值" 组合：这些 flag 后紧跟的一个 arg 不是位置参数
	skipNextArg := map[string]bool{
		"-c": true, "-e": true, "--eval": true, "-p": true, "--print": true, "-r": true,
		"-m": true, "-E": true,
	}

	i := 1
	for i < len(tokens) {
		tok := tokens[i]
		if tok == "" || tok == "-" || tok == "--" {
			i++
			continue
		}
		if strings.HasPrefix(tok, "--") {
			i++
			continue
		}
		if strings.HasPrefix(tok, "-") {
			if skipNextArg[tok] {
				i += 2
			} else {
				i++
			}
			continue
		}
		// 位置参数：判定
		// 注：早期版本这里有 "$-prefix 位置参数一律拒" 分支（indirect-target 标签），
		// 但会误伤合法用法——例如 skill 官方调用模式 `python3 script.py --token "$KEY"`
		// 里的 flag value。由于本函数上层的 skipNextArg 无法穷举所有 skill / 用户自定义
		// flag（--token / --url / --arguments 等），"位置参数 = 脚本路径" 这一判断在
		// 混着 flag value 的现实命令上不成立。已删除该分支：
		//   - head 位置的 $X / $(...) 变量展开仍由 precheck.go 顶层 indirectCmdRe 拦；
		//   - 位置参数是变量展开时，shell 展开后要么落到 workspace 内脚本文件继续走
		//     script-file 路径检查，要么展开失败让 bash 报错——不打开新的泄露通道。
		// 剥外层配对引号
		arg := tok
		if len(arg) >= 2 {
			if (arg[0] == '"' || arg[0] == '\'') && arg[len(arg)-1] == arg[0] {
				arg = arg[1 : len(arg)-1]
			}
		}
		// /dev/stdin / /proc/self/fd/0 / /dev/fd/N 伪路径（stdin 喂脚本）→ 拒
		if _, ok := devStdinPaths[arg]; ok {
			return fmt.Errorf("%s", blockedMsg("stdin-source", head))
		}
		if strings.HasPrefix(arg, "/dev/fd/") || strings.HasPrefix(arg, "/proc/self/fd/") {
			return fmt.Errorf("%s", blockedMsg("stdin-source", head))
		}
		// 脚本后缀判定：不在集合内跳过此 arg
		ext := strings.ToLower(filepath.Ext(arg))
		if _, ok := scriptFileExtSet[ext]; !ok {
			i++
			continue
		}
		// 路径解析：绝对路径直接用；相对路径相对 workDir。
		var absPath string
		if filepath.IsAbs(arg) {
			absPath = filepath.Clean(arg)
		} else {
			absPath = filepath.Join(absWorkDir, arg)
		}
		// 边界：workDir 内
		relToWork, rerr := filepath.Rel(absWorkDir, absPath)
		if rerr != nil || relToWork == ".." || strings.HasPrefix(relToWork, ".."+string(filepath.Separator)) {
			i++
			continue
		}
		stat, serr := os.Stat(absPath)
		if serr != nil {
			if !os.IsNotExist(serr) {
				log.Printf("[Shell] script-file stat failed: %v (path=%s)", serr, absPath)
			}
			i++
			continue
		}
		if stat.IsDir() || stat.Size() > ScriptFileMaxSize {
			i++
			continue
		}
		data, rerr2 := os.ReadFile(absPath)
		if rerr2 != nil {
			log.Printf("[Shell] script-file read failed: %v (path=%s)", rerr2, absPath)
			i++
			continue
		}
		body := string(data)
		// 硬防线：任何路径下的脚本读 .env / .skill_env 都拒（含 skill 目录）
		if perr := precheckScriptReadSensitive(body); perr != nil {
			return fmt.Errorf("%s", blockedMsg("script-file", head))
		}
		// 三条件 body-scan：workspace/skills/** 下的可信预置脚本免除，其它路径（tmp / output）走。
		if len(envMap) > 0 && !isTrustedSkillScript(absPath, absWorkDir) {
			if perr := precheckScriptBody(body, envMap); perr != nil {
				return fmt.Errorf("%s", blockedMsg("script-file", head))
			}
		}
		i++
	}
	return nil
}

// isTrustedSkillScript 判断 absPath 是否位于 workspace/skills/ 下。
// 这些脚本是 sandbox 启动前静态挑选、可审计的预置文件，不是 LLM 或攻击者写入的
// 运行时产物，免除三条件 body-scan；但仍受 precheckScriptReadSensitive 两条件
// 保护（读 .env / .skill_env 不豁免）。
//
// 前提：workspace/skills/ 目录内容由构建/打包流程审计；运行时 LLM 不应有写入
// 该目录的权限（该权限边界由 sandbox 文件系统层保障，不在本函数职责内）。
func isTrustedSkillScript(absPath, absWorkDir string) bool {
	rel, err := filepath.Rel(absWorkDir, absPath)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)
	return strings.HasPrefix(rel, "skills/")
}
