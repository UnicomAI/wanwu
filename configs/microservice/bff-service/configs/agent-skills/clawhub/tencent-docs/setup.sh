#!/bin/bash
#
# Setup script for 腾讯文档 MCP Skill (内部 OpenClaw 版本) 一体化配置与授权脚本
#
# 功能：
#   1. 检查 mcporter 是否已配置 tencent-docs（含 Authorization 可用）
#   2. 未配置或 Token 失效时，展示授权链接并等待用户主动确认已完成授权
#   3. 用户确认后主动查询一次 Token 并写入 mcporter 配置
#   4. 对过期、错误等场景给出友好提示
#
# 用法（供 AI Agent 调用）：
#   第一步：检查状态（立即返回，不阻塞）
#     bash ./setup.sh tdoc_check_and_start_auth
#     输出：
#       READY                  → 服务已就绪，直接执行用户任务，无需后续步骤
#       AUTH_REQUIRED:<url>    → 向用户展示授权链接，等待用户确认已完成授权后执行第二步
#       ERROR:*                → 告知用户对应错误
#
#   第二步：用户确认授权后，主动查询 Token（立即返回）
#     bash ./setup.sh tdoc_fetch_token
#     输出：
#       TOKEN_READY            → 授权成功，继续执行用户任务
#       ERROR:not_authorized   → 用户尚未完成授权，请稍后重试
#       ERROR:expired          → 授权码已过期，请重新发起请求
#       ERROR:token_invalid    → Token 已失效，请重新授权
#       ERROR:*                → 告知用户对应错误
#
#   可选：直接带 Token 设置服务（跳过 OAuth 流程，适合已有 Token 的场景）
#     bash ./setup.sh tdoc_set_token <token>
#     输出：
#       TOKEN_READY            → Token 写入成功，可直接执行用户任务
#       ERROR:missing_token    → 未提供 token 参数
#       ERROR:*                → 告知用户对应错误
#
# 直接执行（排查问题）：
#   bash ./setup.sh
#

# ── 全局配置 ──────────────────────────────────────────────────────────────────
_TDOC_API_BASE="${TDOC_API_BASE_URL:-https://docs.qq.com}"
_TDOC_AUTH_BASE="${TDOC_AUTH_BASE_URL:-https://docs.qq.com/scenario/open-claw.html}"
_TDOC_MCP_URL="https://docs.qq.com/openapi/mcp"
_TDOC_SERVICE_NAME="tencent-docs"

# 临时文件
_TDOC_CODE_FILE="${TMPDIR:-/tmp}/.tdoc_auth_code"
_TDOC_URL_FILE="${TMPDIR:-/tmp}/.tdoc_auth_url"

# ── 清理函数 ──────────────────────────────────────────────────────────────────
_tdoc_cleanup() {
    rm -f "$_TDOC_CODE_FILE" "$_TDOC_URL_FILE"
}

# ── 检查 mcporter 是否已安装 ──────────────────────────────────────────────────
_tdoc_check_mcporter() {
    if ! command -v mcporter &> /dev/null; then
        echo "⚠️  未找到 mcporter，正在安装..."
        if command -v npm &>/dev/null; then
            npm install -g mcporter@0.8.1 2>&1 | tail -3
            echo "✅ mcporter 安装完成"
        else
            echo "ERROR:no_npm"
            return 1
        fi
    fi
    return 0
}

# 从 mcporter config get 读取当前 Authorization Token
# 输出：token 字符串（空则表示服务未注册或 Token 未配置）
_tdoc_get_token() {
    local output
    output=$(mcporter config get "$_TDOC_SERVICE_NAME" 2>/dev/null) || return 1

    # 从输出中提取 Authorization 头的值
    local token
    token=$(echo "$output" | grep -i '^\s*Authorization:' | sed 's/.*Authorization:[[:space:]]*//' | tr -d '[:space:]')
    echo "$token"
}

# ── 将 Token 写入 mcporter 配置 ───────────────────────────────────────────────
# 用法：_tdoc_save_token <token>
_tdoc_save_token() {
    # 添加 MCP 配置
    echo "🔧 配置 mcporter..."

    local token="$1"
    [[ -z "$token" ]] && return 1

    # 使用传入的 token 写入 mcporter 配置（tencent-docs）
    mcporter config add "$_TDOC_SERVICE_NAME" "$_TDOC_MCP_URL" \
        --header "Authorization=$token" \
        --transport http \
        --scope home

    echo ""
    echo "✅ 配置完成！"
    echo ""

    echo "🧪 验证配置..."
    if mcporter list 2>&1 | grep -q "$_TDOC_SERVICE_NAME"; then
        echo "✅ tencent-docs 配置验证成功！"
        echo ""
        mcporter list | grep -A 1 "$_TDOC_SERVICE_NAME" || true
    else
        echo "⚠️  tencent-docs 配置验证失败，请检查网络或 Token 是否有效"
    fi

    echo ""
    echo "如有问题，请访问 ${_TDOC_API_BASE}/scenario/open-claw.html?nlc=1 获取 Token"

    echo ""
    echo "─────────────────────────────────────"
    echo "🎉 设置完成！"
    echo ""
    echo "📖 使用方法："
    echo "   mcporter call ${_TDOC_SERVICE_NAME}.create_smartcanvas_by_mdx"
    echo ""
    echo "🏠 腾讯文档主页：${_TDOC_API_BASE}/home"
    echo ""
    echo "📖 更多信息请查看 SKILL.md"
    echo ""
    return 0
}

# ── 检查 tencent-docs 服务状态 ────────────────────────────────────────────────
# 返回值：
#   0 = 服务正常可用（有 Token）
#   1 = 服务未注册（mcporter config get 失败）
#   2 = Token 为空或未配置
_tdoc_check_service() {
    if ! mcporter list 2>/dev/null | grep -q "$_TDOC_SERVICE_NAME"; then
        return 1
    fi

    local token
    token=$(_tdoc_get_token)
    local rc=$?

    # mcporter config get 返回非 0 表示服务未注册
    if [[ $rc -ne 0 ]]; then
        return 1
    fi

    # Token 为空表示服务已注册但未配置 Authorization
    if [[ -z "$token" ]]; then
        return 2
    fi

    return 0
}

# ── 生成授权链接 ──────────────────────────────────────────────────────────────
# 输出：auth_url 字符串，同时将 code 写入 $_TDOC_CODE_FILE
_tdoc_generate_auth_url() {
    local code
    code=$(openssl rand -hex 8 2>/dev/null || \
           cat /dev/urandom | LC_ALL=C tr -dc 'a-zA-Z0-9' 2>/dev/null | head -c 16 || \
           date +%s%N 2>/dev/null | sha256sum 2>/dev/null | head -c 16 || \
           echo "$(date +%s)$$")

    echo "$code" > "$_TDOC_CODE_FILE"
    echo "${_TDOC_AUTH_BASE}?nlc=1&authType=1&code=${code}&mcp_source=desktop"
}

# ── 主入口函数 A：检查状态 / 生成授权链接（立即返回，不阻塞）────────────────
#
# AI Agent 第一步调用此函数，命令执行完毕后立即拿到输出：
#   READY                  服务已就绪，直接执行用户任务，无需后续步骤
#   AUTH_REQUIRED:<url>    需要授权：向用户展示链接，等用户确认后执行第二步
#   ERROR:*                错误信息
#
tdoc_check_and_start_auth() {
    _tdoc_check_mcporter || {
        echo "ERROR:mcporter_not_found - 请先安装 Node.js 和 npm 后重试"
        return 1
    }

    _tdoc_check_service
    local status=$?

    case $status in
        0)
            echo "READY"
            return 0
            ;;
        1|2)
            _tdoc_cleanup

            # 生成授权链接（同时写入 code 文件）
            local auth_url
            auth_url=$(_tdoc_generate_auth_url)

            # 将 URL 写入文件，供后续阶段读取
            echo "$auth_url" > "$_TDOC_URL_FILE"

            echo "AUTH_REQUIRED:$auth_url"
            return 0
            ;;
    esac
}

# ── 主入口函数 B：用户确认授权后，主动查询 Token 并写入配置（立即返回）────────
#
# AI Agent 在用户确认已完成授权后调用此函数，主动查询一次 Token：
#   TOKEN_READY            授权成功，Token 已写入配置，直接执行用户任务
#   ERROR:not_authorized   用户尚未完成授权，请稍后重试或重新发起请求
#   ERROR:expired          授权码已过期，告知用户重新发起请求
#   ERROR:token_invalid    Token 鉴权失败，告知用户重新授权
#   ERROR:*                错误信息
#
tdoc_fetch_token() {
    # 读取 code 文件
    if [[ ! -f "$_TDOC_CODE_FILE" ]]; then
        echo "ERROR:no_code - 未找到授权码，请先执行 tdoc_check_and_start_auth"
        return 1
    fi

    local code
    code=$(cat "$_TDOC_CODE_FILE")
    if [[ -z "$code" ]]; then
        echo "ERROR:empty_code - 授权码为空，请重新发起请求"
        return 1
    fi

    local url="${_TDOC_API_BASE}/oauth/v2/mcp/token/get?code=${code}"

    local response
    response=$(curl -s -f -L "$url" 2>/dev/null)
    if [[ $? -ne 0 || -z "$response" ]]; then
        echo "ERROR:network - 网络请求失败，请检查网络连接后重试"
        return 1
    fi

    # 提取 token
    local token
    token=$(echo "$response" | jq -r '.data.token // empty' 2>/dev/null || echo "")
    if [[ -n "$token" ]]; then
        if _tdoc_save_token "$token"; then
            _tdoc_cleanup
            echo "TOKEN_READY"
            return 0
        else
            _tdoc_cleanup
            echo "ERROR:save_token_failed"
            return 1
        fi
    fi

    # 提取错误码
    local ret
    ret=$(echo "$response" | jq -r '.ret // empty' 2>/dev/null || echo "")

    case "$ret" in
        "11510")
            # 用户还未完成授权
            echo "ERROR:not_authorized - 您尚未完成授权，请在浏览器中完成授权后重试"
            return 1
            ;;
        "400006")
            # Token 鉴权失败
            _tdoc_cleanup
            echo "ERROR:token_invalid - Token 鉴权失败，请重新授权"
            return 1
            ;;
        "400007")
            # VIP 权限不足
            echo "ERROR:vip_required - 当前操作需要腾讯文档 VIP 权限，请升级 VIP：https://docs.qq.com/vip?immediate_buy=1?part_aid=persnlspace_mcp"
            return 1
            ;;
        *)
            local expired
            expired=$(echo "$response" | jq -r '.data.expired // empty' 2>/dev/null || echo "")
            if [[ "$expired" == "true" ]]; then
                _tdoc_cleanup
                echo "ERROR:expired - Token 已过期"
                return 1
            fi
            echo "ERROR:unknown(ret=${ret}, response=${response}) - 授权失败，请尝试手动设置 Token"
            return 1
            ;;
    esac
}

# ── 主入口函数 C：直接带 token 参数设置 mcporter 服务 ────────────────────────
#
# AI Agent 在已知 token 的情况下可直接调用此函数，跳过 OAuth 授权流程：
#   TOKEN_READY            Token 写入成功，可直接执行用户任务
#   ERROR:missing_token    未提供 token 参数
#   ERROR:save_token_failed  写入配置失败
#
# 用法：
#   bash ./setup.sh tdoc_set_token <token>
#
tdoc_set_token() {
    local token="$1"
    if [[ -z "$token" ]]; then
        echo "ERROR:missing_token - 请提供 token 参数，用法：bash ./setup.sh tdoc_set_token <token>"
        return 1
    fi

    _tdoc_check_mcporter || {
        echo "ERROR:mcporter_not_found - 请先安装 Node.js 和 npm 后重试"
        return 1
    }

    if _tdoc_save_token "$token"; then
        echo "TOKEN_READY"
        return 0
    else
        echo "ERROR:save_token_failed - Token 写入配置失败"
        return 1
    fi
}

# ── 直接执行时的交互式安装流程 ───────────────────────────────────────────────
_tdoc_interactive_setup() {
    echo ""
    echo "╔══════════════════════════════════════════════╗"
    echo "║     腾讯文档 MCP Skill 配置向导              ║"
    echo "╚══════════════════════════════════════════════╝"
    echo ""

    # 检查 mcporter
    echo "🔍 检查 mcporter..."
    if ! _tdoc_check_mcporter; then
        echo "❌ mcporter 安装失败，请先安装 Node.js (https://nodejs.org) 后重试"
        exit 1
    fi
    echo "✅ mcporter 已就绪"
    echo ""

    # 检查服务状态
    echo "🔍 检查 tencent-docs 服务配置..."
    _tdoc_check_service
    local status=$?

    case $status in
        0)
            echo "✅ tencent-docs 服务已配置且运行正常！"
            echo ""
            echo "🎉 无需重新配置，您可以直接使用腾讯文档功能。"
            echo ""
            echo "📖 使用示例："
            echo "   mcporter call tencent-docs manage.recent_online_file --args '{\"num\":10}'"
            return 0
            ;;
        1|2)
            echo "⚠️  Token 未配置，需要授权..."
            ;;
    esac

    echo ""
    echo "🔐 需要完成腾讯文档授权"
    echo ""

    # 清理旧状态
    _tdoc_cleanup

    # 生成授权链接（同时写入 code 文件）
    local auth_url
    auth_url=$(_tdoc_generate_auth_url)

    echo "┌─────────────────────────────────────────────────────────┐"
    echo "│  请在浏览器中打开以下链接完成授权：                      │"
    echo "│                                                         │"
    printf "│  %s\n" "$auth_url"
    echo "│                                                         │"
    echo "│  ⚠️  请使用 QQ 或微信 扫码 / 登录授权                   │"
    echo "└─────────────────────────────────────────────────────────┘"
    echo ""
    echo "完成授权后，请按回车键继续..."
    read -r

    # 用户确认后主动查询 Token
    echo "⏳ 正在查询授权结果..."
    local result
    result=$(tdoc_fetch_token)

    case "$result" in
        TOKEN_READY)
            echo ""
            echo "🎉 配置完成！现在可以直接使用腾讯文档功能了。"
            echo ""
            echo "📖 使用示例："
            echo "   mcporter call ${_TDOC_SERVICE_NAME} manage.recent_online_file --args '{\"num\":10}'"
            echo ""
            echo "🏠 腾讯文档主页：${_TDOC_API_BASE}/home"
            ;;
        ERROR:not_authorized*)
            echo ""
            echo "⚠️  您似乎尚未完成授权，请在浏览器中完成授权后重新运行：bash ./setup.sh"
            exit 1
            ;;
        ERROR:expired*)
            echo ""
            echo "❌ Token 已过期，请访问 https://docs.qq.com/scenario/open-claw.html 重新获取 Token，然后重新授权"
            exit 1
            ;;
        ERROR:token_invalid*)
            echo ""
            echo "❌ Token 鉴权失败，请重新运行：bash ./setup.sh"
            exit 1
            ;;
        ERROR:*)
            echo ""
            echo "❌ 授权失败：$result"
            echo "   如问题持续，请联系腾讯文档客服：${_TDOC_API_BASE}/home/feedback"
            exit 1
            ;;
    esac

    return 0
}

# ── 脚本入口 ──────────────────────────────────────────────────────────────────
# 直接执行时：
#   bash ./setup.sh tdoc_check_and_start_auth  → 第一步：检查状态 / 生成授权链接
#   bash ./setup.sh tdoc_fetch_token           → 第二步：用户确认后主动查询 Token
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    if [[ -n "$1" ]]; then
        # 参数分发：将第一个参数作为函数名执行
        case "$1" in
            tdoc_check_and_start_auth|tdoc_fetch_token)
                "$1"
                exit $?
                ;;
            tdoc_set_token)
                tdoc_set_token "$2"
                exit $?
                ;;
            setup)
                echo "🚀 腾讯文档 MCP Skill 人工配置向导"
                echo ""
                _tdoc_interactive_setup
                ;;
            *)
                echo "ERROR:unknown_command - 未知命令: $1"
                echo "可用命令: tdoc_check_and_start_auth, tdoc_fetch_token, tdoc_set_token, setup"
                exit 1
                ;;
        esac
    else
        echo "用法："
        echo "  bash ./setup.sh tdoc_check_and_start_auth      # 第一步：检查状态 / 生成授权链接"
        echo "  bash ./setup.sh tdoc_fetch_token               # 第二步：用户确认后主动查询 Token"
        echo "  bash ./setup.sh tdoc_set_token <token>         # 直接设置 Token（跳过 OAuth 流程）"
    fi
fi
