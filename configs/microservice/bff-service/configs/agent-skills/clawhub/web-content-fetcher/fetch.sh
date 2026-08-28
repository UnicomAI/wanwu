#!/bin/bash
# 网页内容获取工具 - Web Content Fetcher
# 按优先级尝试多个服务获取网页 Markdown 内容

set -e

URL="$1"
METHOD="${2:-jina}"

if [ -z "$URL" ]; then
    echo "用法: $0 <url> [method]"
    echo "方法: jina (默认), markdown, defuddle"
    exit 1
fi

# 移除 http:// 或 https:// 前缀（如果需要）
ENCODED_URL="$URL"

echo "=== 尝试获取: $URL ===" >&2
echo "使用方法: $METHOD" >&2

case "$METHOD" in
    jina)
        echo "--- 使用 r.jina.ai ---" >&2
        curl -s "https://r.jina.ai/$ENCODED_URL"
        ;;
    markdown)
        echo "--- 使用 markdown.new ---" >&2
        curl -s "https://markdown.new/$ENCODED_URL"
        ;;
    defuddle)
        echo "--- 使用 defuddle.md ---" >&2
        curl -s "https://defuddle.md/$ENCODED_URL"
        ;;
    all)
        echo "=== 尝试所有方法 ===" >&2
        echo "--- r.jina.ai ---" >&2
        curl -s "https://r.jina.ai/$ENCODED_URL"
        echo -e "\n--- markdown.new ---" >&2
        curl -s "https://markdown.new/$ENCODED_URL"
        echo -e "\n--- defuddle.md ---" >&2
        curl -s "https://defuddle.md/$ENCODED_URL"
        ;;
    *)
        echo "未知方法: $METHOD" >&2
        echo "可用方法: jina, markdown, defuddle, all" >&2
        exit 1
        ;;
esac
