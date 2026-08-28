---
name: wechat-article
description: 抓取微信公众号文章，提取标题和内容并输出为 Markdown 格式。支持多种提取方式。
metadata: { "openclaw": { "emoji": "📰", "requires": { "bins": ["python3"], "env": [] } } }
---

# 微信公众号文章抓取

通过公众号链接获取文章的标题和内容，输出为 Markdown 格式。

## Usage

```bash
python3 skills/wechat-article/scripts/wechat_article.py '<公众号文章链接>'
```

## 输出示例

```markdown
# 文章标题

**发布时间**: 2026-03-04
**公众号**: 某某公众号

---

文章正文内容...
```

## 提取策略

按顺序尝试以下方式：

1. **r.jina.ai** - 第三方内容提取服务（推荐）
2. **web_fetch** - 直接抓取（可能被拦截）
3. **agent-browser** - 浏览器自动化（需要处理验证码）

## 注意事项

- 微信公众号有反爬虫机制，部分文章可能无法获取
- 需要验证码的文章无法自动处理
- 建议优先使用 r.jina.ai 方式

## 示例

```bash
# 基本使用
python3 skills/wechat-article/scripts/wechat_article.py "https://mp.weixin.qq.com/s/xxx"

# 输出到文件
python3 skills/wechat-article/scripts/wechat_article.py "https://mp.weixin.qq.com/s/xxx" > article.md
```
