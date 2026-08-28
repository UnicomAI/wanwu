package shared

import "regexp"

// --- Skill 环境变量保护：命令前置拦截 ---
//
// 目的：防止 LLM 在恶意 prompt 引导下，通过 bash 输出回流 skill 已注入的敏感 env 值。
// 策略：在 cmd.Run 之前对命令字符串做静态 precheck，命中规则即拒绝执行；不做输出 redact,
// 因为字节匹配式 redact 会误伤 value 撞合法输出的场景（污染整个 sandbox 输出）。
//
// 指导思想：承认 shell + 内联解释器攻击面不可穷举；precheck 目标是抬高攻击者构造成本，
// 不是证明安全。skill 内脚本自带的安全隐患也尽力发现拦截，责任归属不做区分。
// 架构层根治方案（网络白名单 / 不落盘 / token broker）由部署层实现，不在本层。
//
// 拦截规则按攻击向量分类：
//
//   [枚举型]
//     - 独立命令：env / printenv / set / declare / export / compgen / typeset 无差别拒
//     - 命令替换：$(env) / `printenv` / <(env) / <(printenv) 等嵌套语义等价，同样拒
//
//   [引用型]（打印命令/子命令引用 $KEY / ${KEY} / bash 参数展开变体）
//     - 打印类命令 (echo/printf/cat/tee/less/more/head/tail/sed/awk/tr/rev/cut/paste/sort/uniq/grep/
//       nl/fold/fmt/column/strings/base64/xxd/od/hexdump/hd) 命令行含 $KEY 引用 → 拒
//     - NAME=...$KEY... 中转变量赋值（赋值步骤即拒，杜绝 N 跳链）
//     - tokens[0] 里含 $KEY 引用（如 nonexistent_$KEY 报错回显泄露）
//     - trap 'echo $KEY' EXIT 之类延迟执行
//     - bash 参数展开变体覆盖：${K} / ${K:-x} / ${K:=x} / ${K:?x} / ${K:+x} / ${K#*} / ${K##*} /
//       ${K%*} / ${K%%*} / ${K:0:5} / ${K:5} / ${K/A/B} / ${K//A/B} / ${K^^} / ${K,,} /
//       ${K^} / ${K,} / ${K@Q} / ${K@P} / ${K@U} / ${K@A} / ${!K}
//
//   [文件读取]
//     - 命令体含 .env / .skill_env 系列路径（含 glob 通配符 * ? [）
//     - 命令体含 /proc/*/env* 路径（含 shell 变量展开）
//     - ln / cp / mv / rsync / install + .env / .skill_env 字面量（防别名化后再读）
//     - 打印命令 + `.*` / `.?<any>` 之类宽泛 glob（防 cat .* 展开到 .env）
//     - find / xargs 跳板到打印命令读敏感文件
//
//   [解释器 body]
//     - 内联脚本 body 同时命中 (env 访问) ∧ (stdout sink) ∧ (envMap key 字面量) 三条件
//     - heredoc body 同上（bash 家族用 $KEY 语法，其余用 os.environ / process.env 等）
//     - body 含文件读 API + .env / .skill_env 字面量（护 workspace/.env 里 OPENAI_API_KEY）
//
//   [dynamic exec / shell 语法]
//     - bash / sh / dash -c "<arg>" 递归 precheck，拦下"包一层 shell 绕过"
//     - eval 关键字无条件拒
//     - ${!VAR} 间接展开无条件拒
//     - head 是 $X / ${X} / $(...) 变量或命令替换展开 → 无条件拒
//     - bash -x / -v / -o xtrace debug flag（每条命令展开值输出到 stderr）
//     - 无参 bash / echo | bash / pipeline 尾接 shell（从 stdin 读命令的 dynamic exec）
//     - BASH_ENV / ENV / PROMPT_COMMAND / LD_PRELOAD / LD_LIBRARY_PATH / LD_AUDIT / SHELLOPTS /
//       BASH_XTRACEFD / PS4 等特殊 env 赋值（会触发 shell 启动 source 或运行时行为改变）
//     - 重定向 target / 脚本路径以 $ 开头（变量展开无法静态解析）
//     - base64 -d / --decode / xxd -r / --revert 解码命令（防编码后 pipe 到 shell）
//     - head normalize：剥掉 NAME=value 前缀赋值、command/exec/builtin 外壳、 (/{ 子 shell 前导、
//       反斜杠等常见绕 head 判定的语法糖后，再做后续所有 head 判定
//
//   [写盘]
//     - 重定向到 .py / .sh / .js / .rb / .pl / .php / .awk 等脚本后缀文件 + inline body（quoted
//       string 或 heredoc body）含越权代码或文件读 API + .env 字面量
//     - 重定向 target 是 shell rc 文件 (.bashrc / .zshrc / .profile 等) → 同样走 body 扫描
//     - 重定向 target 是 $VAR 变量展开 → 无论后缀强制走 body 扫描
//
// 另外文件形式脚本 body-scan (script-file) 规则在 ShellOnlyBackend.Execute 层做 IO check
// （precheck 通过后追加），扫所有 tokens 位置参数指向的 workspace 内脚本文件（不限 head，
// 因为无论 head 是 python / bash / cat / grep / less / cp / mv 都能达成"让越权代码被读"的效果）。
// 独立于 precheckCommand 分层：precheckCommand 是 pure-static（无 IO），script-file 需要读盘。

var (
	// secretEnumeratorCmds 无差别打印所有 env 的命令，envMap 是否为空都拒（给 LLM 一致 UX）。
	secretEnumeratorCmds = map[string]struct{}{
		"env": {}, "printenv": {}, "set": {}, "declare": {},
		"compgen": {}, "typeset": {}, "export": {},
	}
	// refPrinterCmds 必须叠加 "命令体引用注册 key" 条件才拒，避免 echo "hello" 这种合法用法被错杀。
	// 覆盖到 sed/awk/tr/rev/cut/paste/sort/uniq/grep/nl/fold/fmt/column 等所有会向 stdout 输出的
	// 文本处理命令。命令行无 $KEY 引用不受影响，FP 理论为 0；有 $KEY 引用即视为泄露。
	refPrinterCmds = map[string]struct{}{
		"echo": {}, "printf": {}, "cat": {}, "tee": {},
		"less": {}, "more": {}, "head": {}, "tail": {},
		"strings": {}, "base64": {}, "xxd": {}, "od": {},
		"hexdump": {}, "hd": {},
		"sed": {}, "awk": {}, "tr": {}, "rev": {}, "cut": {}, "paste": {},
		"sort": {}, "uniq": {}, "grep": {}, "nl": {}, "fold": {}, "fmt": {}, "column": {},
	}
	// sensitiveFileRe 匹配 .skill_env* / .env* 文件路径，需有 word/路径边界，避免误伤 readme.envsetup 这种。
	// 末尾边界含 glob 元字符 * ? [，覆盖 `cat .env*` / `cat .env?` / `cat .env[abc]` 这类通配符探查。
	// 前置边界保持 `^|[\s/]` 不含通配符：`readme.envsetup` 仍不会被误命中（`me` 不是路径分隔符）。
	sensitiveFileRe = regexp.MustCompile(`(^|[\s/])(\.env|\.skill_env)(\.[\w.]+)?($|[\s/*?\[])`)
	// procEnvironRe 放宽 pid 段与 environ 段均为"任意非空白/路径分隔符 + env 前缀"，
	// 覆盖 self / 数字 / $$ / $BASHPID / 通配符（如 se* / envi* / env*）。
	// 副作用：`/proc/<x>/env<any>` 之类不存在的路径也会命中，但那类命令不会成功执行，
	// precheck 提前拒无实质影响。
	procEnvironRe = regexp.MustCompile(`/proc/[^\s/]+/env[^\s/]*`)

	// scriptInterpreterCmds 解释器内联脚本拦截覆盖的命令集合及其"内联脚本"flag。
	// 匹配 head 后，extractInterpreterScript 在 cmdStr 中找任一 flag，取 flag 后的下一参数作为脚本 body。
	// awk/gawk/mawk 走单独的位置参数分支，不在本 map（脚本是首个非 flag 的位置参数）。
	scriptInterpreterCmds = map[string][]string{
		"python":  {"-c"},
		"python3": {"-c"},
		"pypy":    {"-c"},
		"pypy3":   {"-c"},
		"node":    {"-e", "--eval", "-p", "--print"},
		"nodejs":  {"-e", "--eval", "-p", "--print"},
		"perl":    {"-e", "-E"},
		"ruby":    {"-e"},
		"php":     {"-r"},
	}

	// scriptEnvAccessRe 命中即视为"脚本读取了 process env"。任一语言的常见 env 读法即可。
	scriptEnvAccessRe = regexp.MustCompile(
		`os\.environ|os\.getenv\s*\(|process\.env\b|\$ENV\{|\bENV\[|ENVIRON\[|\bgetenv\s*\(`)

	// scriptSinkRe 命中即视为"脚本会向 stdout/stderr 输出"。
	// 关键：\bprint\b 同时覆盖 Python 的 print(...) 与 Perl/awk 的裸 print（后者无括号）。
	// 不会误命中 printf / pprint —— 两者都不存在以 \b 切分的 "print" 边界。
	// \becho\b 主要覆盖 PHP / awk 内联用法；bash 的 echo 走的是打印型命令分支（head 在 refPrinterCmds），不会到这里。
	scriptSinkRe = regexp.MustCompile(
		`\bprint\b|console\.(log|error|warn|info|dir)\s*\(|` +
			`sys\.std(out|err)\.write\s*\(|process\.std(out|err)\.write\s*\(|` +
			`\bputs\b|\bsay\b|\bprintf\b|\becho\b`)

	// scriptFileExtRe 写入文件的目标若为可执行脚本后缀，需扫描 inline 内容做三条件检查。
	scriptFileExtRe = regexp.MustCompile(`\.(py|pyw|sh|bash|js|mjs|cjs|pl|rb|php|awk)\b`)

	// heredocStartRe 识别 `<head> ... <<-?['"]?TAG['"]?` 的 heredoc 起始 token，捕获 TAG 名。
	// body 与结束 TAG 由 extractHeredocBody 用 strings.Index 扫描提取（RE2 不支持反向引用）。
	heredocStartRe = regexp.MustCompile(`<<-?\s*['"]?([A-Za-z_][A-Za-z0-9_]*)['"]?`)

	// bashEnvAccessRe 仅在 bash/sh/dash heredoc body 内使用：$KEY 与 ${KEY} 即视为 env 访问。
	// Python/Node 脚本内的 $KEY 是字面量，不在此处用。
	bashEnvAccessRe = regexp.MustCompile(`\$[A-Za-z_][A-Za-z0-9_]*|\$\{[A-Za-z_][A-Za-z0-9_]*\}`)

	// evalKeywordRe head 形态或作为独立 token 出现的 eval 关键字。
	// 边界要求是空白/分隔符，避免误命中 eval_expr_func 这种函数名子串。
	evalKeywordRe = regexp.MustCompile(`(^|[\s;&|])eval(\s|$)`)

	// indirectExpandRe bash 的间接展开语法 ${!VAR}。LLM 自动生成 bash 几乎不用，可无条件拒。
	indirectExpandRe = regexp.MustCompile(`\$\{!\s*[A-Za-z_][A-Za-z0-9_]*\s*\}`)

	// cmdSubstEnumRe $( env|printenv|set|declare|compgen|typeset|export ... ) 或反引号形态。
	// 等价于直接执行枚举命令，复用 enumerator 标签。
	cmdSubstEnumRe = regexp.MustCompile(
		"\\$\\(\\s*(env|printenv|set|declare|compgen|typeset|export)\\b" +
			"|`\\s*(env|printenv|set|declare|compgen|typeset|export)\\b")

	// relayAssignRe `<NAME>=` 赋值语句首部。RHS 若引用 envMap 中任一 key 即拒。
	// 前导边界 (^|[\s;&|()])避免误命中 X==Y 比较 / 字符串中间的 VAR=value 子串。
	relayAssignRe = regexp.MustCompile(`(^|[\s;&|()])([A-Za-z_][A-Za-z0-9_]*)=`)

	// scriptFileReadRe 覆盖 python / node / perl / ruby / php / awk 的常见文件读 API。
	// 命中即视为"脚本读文件"。是 precheckScriptReadSensitive 两条件硬防线的第一条
	// （另一条件是 .env / .skill_env 字面量），单独命中不拒。
	// 匹配示例：open( / .read( / .readlines( / fs.readFile / require('fs') / File.read / IO.read /
	// file_get_contents( / fopen( / readfile( / getline < （awk）
	scriptFileReadRe = regexp.MustCompile(
		`\bopen\s*\(|\.read\s*\(|\.readlines\s*\(|` +
			`\bfs\.readFile|\bfs\.readFileSync|require\s*\(\s*['"]fs['"]|` +
			`\bFile\.(read|open)|\bIO\.(read|readlines)|` +
			`\bfile_get_contents\s*\(|\bfopen\s*\(|\breadfile\s*\(|` +
			`\bgetline\s*<`)

	// scriptSensitivePathLiteralRe 匹配脚本 body 中出现的 .env / .skill_env 字面量。
	// 三条件的第二条：脚本参数含敏感路径字面量（不做扁平路径展开，绕过手法列入残留）。
	scriptSensitivePathLiteralRe = regexp.MustCompile(`\.(env|skill_env)(\.[\w.]+)?`)

	// findExecReaderRe 匹配 `find ... -exec <printer>` / `find ... -exec <printer> ;` 形态。
	// 覆盖 find 跳板到打印命令的所有主流 pattern（sh 会拆解 -exec 后的命令行）。
	findExecReaderRe = regexp.MustCompile(
		`\bfind\b[^|;&]*?-exec\s+(cat|less|more|head|tail|xxd|hexdump|hd|od|strings|base64|tee|printf|echo)\b`)

	// xargsReaderRe 匹配 `<anything> | xargs <printer>` 形态。
	// xargs 可能带 -n / -I 等 flag，一起吞掉。
	xargsReaderRe = regexp.MustCompile(
		`\|\s*xargs\b(?:\s+-[a-zA-Z]+)*\s+(cat|less|more|head|tail|xxd|hexdump|hd|od|strings|base64)\b`)

	// findNamePatternEnvRe 检测 `find ... -name` 参数指向 env 文件族。
	// pattern 允许含 glob 元字符（* ? [ ]）和 . 后缀，覆盖 .env* / .skill_env* / .e / .s 等模糊探查。
	findNamePatternEnvRe = regexp.MustCompile(`-name\s+['"]?\.(env|skill_env|e|s)[\w*.?\[\]]*`)

	// echoPipeEnvRe 检测 pipeline 前段用 echo/printf 输出 env 路径字面量再通过 xargs 类跳板消费。
	// 例：echo .skill_env.json | xargs cat
	echoPipeEnvRe = regexp.MustCompile(`(echo|printf)[^|;&]*\.(env|skill_env)`)

	// indirectCmdRe head token 是 shell 变量/命令替换展开（$X / ${X} / $(...)）。
	// shell 会先展开再执行，precheck 静态分析看不到展开后的真实命令名——绕过 head 判定的唯一途径。
	// LLM 自动化任务里几乎无合法用途（合法用 printenv 就直接写 printenv），无条件拒。
	indirectCmdRe = regexp.MustCompile(`^\$[A-Za-z_{(]`)

	// shellDebugFlagRe bash -x / -v / -o xtrace 会把每条命令的展开值输出到 stderr（含 $KEY 展开值）。
	// LLM 自动化场景几乎无合法 debug 需求；bash 家族 + 这些 debug flag 无条件拒。
	shellDebugFlagRe = regexp.MustCompile(`\s(-[a-zA-Z]*[xv]|--xtrace|--verbose|-o\s+xtrace)\b`)

	// pipeShellRe pipeline 尾接 shell（... | bash / ... | sh / ... | dash）。
	// 无参 shell 从 stdin 读命令等价于 dynamic exec bash -c "<attacker code>"；无条件拒。
	pipeShellRe = regexp.MustCompile(`\|\s*(bash|sh|dash)\s*(\||;|&&|\|\||$)`)

	// specialEnvAssignRe 影响 shell 启动/运行时行为的特殊 env 名。
	// 攻击 pattern: BASH_ENV=/tmp/rc sh -c ':' 会在 bash 启动前自动 source /tmp/rc。
	// LHS 命中即拒（不管 RHS 引用什么，因为这些 env 无合法业务用途）。
	specialEnvAssignRe = regexp.MustCompile(
		`(^|[\s;&|()])(BASH_ENV|ENV|PROMPT_COMMAND|LD_PRELOAD|LD_LIBRARY_PATH|LD_AUDIT|SHELLOPTS|BASH_XTRACEFD|PS4)=`)

	// sensitiveFileHandlerRe 与 sensitiveFileRe 叠加判定（head 是文件操作 + 命令体含 .env / .skill_env 字面量）
	// 防止 ln -s .skill_env.json /tmp/x 之类的别名化绕过。
	sensitiveFileHandlerRe = regexp.MustCompile(`^(ln|cp|mv|rsync|install)$`)

	// broadGlobRe cat .* / less .* / head .? 之类宽泛 glob 展开到 . 起头文件。
	// shell 会展开为 .bashrc / .env / .skill_env.json 等；打印命令读所有匹配文件。
	// FP 低：合法用户几乎不 cat .* （一般用完整名或 ls .*）。
	// 匹配：空白 + .* / .? (后跟非路径分隔符或行尾) / .??* 等
	broadGlobRe = regexp.MustCompile(`\s\.(\*|\?($|[^/\s]))`)

	// base64DecodeRe / xxdRevertRe 解码命令：常与管道到 shell 组合成 dynamic exec
	// （base64 -d <<< 'Y2F0IC5lbnY=' | sh）。LLM 自动化场景无合法用途，无差别拒。
	base64DecodeRe = regexp.MustCompile(`(^|[\s;&|()])base64\s+(-D|-d|--decode)\b`)
	xxdRevertRe    = regexp.MustCompile(`(^|[\s;&|()])xxd\s+(-r|--revert)\b`)

	// shellRcFileRe 写入 shell rc 文件（.bashrc / .profile 等）会在下次 shell 启动时自动 source。
	// 补 script-write 覆盖面：脚本后缀正则只匹配 .sh / .py / ... 后缀，rc 文件后缀不同会漏掉。
	shellRcFileRe = regexp.MustCompile(`\.(bashrc|bash_profile|zshrc|zprofile|zshenv|profile|envrc)($|\s)`)

	// processSubstEnumRe 进程替换 <(env) / <(printenv) 等，语义等价命令替换嵌枚举命令。
	// 与 cmdSubstEnumRe 独立判定，标签同 enumerator。
	processSubstEnumRe = regexp.MustCompile(`<\(\s*(env|printenv|set|declare|compgen|typeset|export)\b`)
)

// --- 通用安全校验 ---
//
// 设计原则：
//   1. wga-sandbox 本身已是隔离容器，本层不再承担"防止一切危险字符串"的职责，
//      只防"破坏沙箱自身 + 泄露宿主敏感信息"。
//   2. 区分读/写：cat /proc/cpuinfo、cat /etc/hosts 这类只读操作不拦，只在
//      写入或删除时拦截敏感路径。
//   3. 高危且无合法替代的动作（mkfs、shutdown、写块设备、反向 shell 通道、
//      rm 系统目录）无条件拦截。

var dangerousPatterns = []*regexp.Regexp{
	// 无可逆/无合法用途的系统破坏命令
	regexp.MustCompile(`(?i)\bmkfs\b`),
	regexp.MustCompile(`(?i)\b(shutdown|reboot|halt|poweroff|init\s+[06])\b`),

	// 直接向块设备写入
	regexp.MustCompile(`(?i)\bdd\b[^|;&]*\bof=/dev/(sd|nvme|hd|mmcblk|loop|mapper)`),
	regexp.MustCompile(`(?i)>\s*/dev/(sd|nvme|hd|mmcblk)`),

	// bash 反向 shell 通道
	regexp.MustCompile(`/dev/(tcp|udp)/`),

	// 删除根目录或系统目录（工作目录内 rm -rf 放行）。
	// 注意：故意不列 /home —— 工作区前缀就是 /home/root/workspace/...，列了会误杀。
	// /root\b 是兜底；走到 sensitiveWritePatterns 的 /root/ + writeAction 也会再拦一次。
	regexp.MustCompile(`(?i)\brm\s+-[a-zA-Z]*[rRf][a-zA-Z]*\s+(/\s|/$|/\*|\$HOME\b|~/?\*?\s|~/?\*?$|/etc\b|/var\b|/usr\b|/bin\b|/sbin\b|/lib\b|/lib64\b|/boot\b|/root\b)`),

	// 明显的代码注入
	regexp.MustCompile(`(?i)\b__import__\s*\(\s*['"]os['"]\s*\)`),
}

// sensitiveReadPatterns：无论读写，命中即拦。
// 真正存放凭证/秘密的位置才进来。
var sensitiveReadPatterns = []*regexp.Regexp{
	regexp.MustCompile(`/etc/(shadow|sudoers)\b`),
	regexp.MustCompile(`\.ssh/(id_rsa|id_ed25519|id_ecdsa|id_dsa)\b`),
	regexp.MustCompile(`\.aws/credentials\b`),
	regexp.MustCompile(`/var/lib/(mysql|postgresql)\b`),
	regexp.MustCompile(`/proc/\d+/(mem|maps|environ)\b`),
	regexp.MustCompile(`/proc/sys/kernel/`),
}

// sensitiveWritePatterns：仅当紧邻写动作（>, >>, tee, rm, mv, cp, chmod, chown）时拦截。
var sensitiveWritePatterns = []*regexp.Regexp{
	regexp.MustCompile(`/etc/`),
	regexp.MustCompile(`/proc/`),
	regexp.MustCompile(`/sys/`),
	regexp.MustCompile(`/var/lib/`),
	regexp.MustCompile(`/boot/`),
	regexp.MustCompile(`/root/`),
}

// writeActionPattern：命中点之前若以这些动作结尾，则视为"对该路径写"。
// 用 $ 锚点配合"取前缀子串"的方式判定。
var writeActionPattern = regexp.MustCompile(`(?i)(>\s*|>>\s*|\|\s*tee\s+(-[a-zA-Z]+\s+)*|\btee\s+(-[a-zA-Z]+\s+)*|\brm\s+(-[a-zA-Z]+\s+)*|\bmv\s+\S+\s+|\bcp\s+(-[a-zA-Z]+\s+)*\S+\s+|\bchmod\s+\S+\s+|\bchown\s+\S+\s+)$`)

var (
	pathTraversalPattern = regexp.MustCompile(`\.\./`)
	symlinkPattern       = regexp.MustCompile(`(?i)\bln\s+-[a-z]*s`)
)

// scriptFileExtSet 与 precheck.go 里 scriptFileExtRe 保持同源。
// script-file 扫 tokens 位置参数时用它做后缀快速匹配。
var scriptFileExtSet = map[string]struct{}{
	".py": {}, ".pyw": {}, ".sh": {}, ".bash": {},
	".js": {}, ".mjs": {}, ".cjs": {},
	".pl": {}, ".rb": {}, ".php": {}, ".awk": {},
}

// devStdinPaths 解释器 "从 stdin/fd 读脚本" 的伪路径集合。命中即拒（stdin-source 标签）。
var devStdinPaths = map[string]struct{}{
	"/dev/stdin": {}, "/proc/self/fd/0": {},
}

// ScriptFileMaxSize 读取脚本文件做 body-scan 的大小上限。
// 越权脚本通常几 KB - 几十 KB；超过 1 MB 一般是数据文件或大型工具，
// 静态扫描收益低、成本高，直接放行让 shell 自己去跑。
const ScriptFileMaxSize = 1 << 20 // 1 MB
