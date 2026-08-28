#!/usr/bin/env bash
# setup.sh - 从 TOOLS.md 读取微信公众号环境变量
# Usage: source ./setup.sh

TOOLS_MD="$HOME/.openclaw/workspace/TOOLS.md"

# 检查 TOOLS.md 是否存在
if [ ! -f "$TOOLS_MD" ]; then
    echo "❌ 找不到 TOOLS.md 文件: $TOOLS_MD"
    echo ""
    echo "请在 TOOLS.md 中添加微信公众号凭证："
    echo ""
    echo "## 🔐 WeChat Official Account (微信公众号)"
    echo ""
    echo "export WECHAT_APP_ID=your_app_id"
    echo "export WECHAT_APP_SECRET=your_app_secret"
    exit 1
fi

# 从 TOOLS.md 提取凭证
WECHAT_APP_ID=$(grep "export WECHAT_APP_ID=" "$TOOLS_MD" | head -1 | sed 's/.*export WECHAT_APP_ID=//' | tr -d ' ')
WECHAT_APP_SECRET=$(grep "export WECHAT_APP_SECRET=" "$TOOLS_MD" | head -1 | sed 's/.*export WECHAT_APP_SECRET=//' | tr -d ' ')

# 检查是否成功提取
if [ -z "$WECHAT_APP_ID" ] || [ -z "$WECHAT_APP_SECRET" ]; then
    echo "❌ 无法从 TOOLS.md 读取凭证！"
    echo ""
    echo "请确保 TOOLS.md 包含以下格式："
    echo ""
    echo "export WECHAT_APP_ID=your_app_id"
    echo "export WECHAT_APP_SECRET=your_app_secret"
    exit 1
fi

# 设置环境变量
export WECHAT_APP_ID
export WECHAT_APP_SECRET

echo "✅ 微信公众号环境变量已从 TOOLS.md 加载！"
echo ""
echo "  WECHAT_APP_ID=${WECHAT_APP_ID:0:10}..."
echo "  WECHAT_APP_SECRET=****** (已隐藏)"
echo ""
echo "💡 提示：这些变量仅在当前 shell 会话有效"
