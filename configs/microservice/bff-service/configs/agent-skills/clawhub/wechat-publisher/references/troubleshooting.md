# 故障排查指南

wechat-publisher skill 的常见问题和解决方案。

## 1. ❌ IP 不在白名单

**错误信息：**
```
Error: ip not in whitelist
```

**原因：** 你的 IP 地址未添加到微信公众号后台白名单。

**解决方法：**

1. **获取你的公网 IP：**
   ```bash
   curl ifconfig.me
   ```

2. **登录微信公众号后台：** https://mp.weixin.qq.com/

3. **添加 IP 白名单：**
   - 开发 → 基本配置
   - IP 白名单 → 添加你的 IP

4. **重试发布**

**详细说明：** https://yuzhi.tech/docs/wenyan/upload

---

## 2. ❌ wenyan-cli 未安装

**错误信息：**
```
wenyan: command not found
```

**解决方法：**
```bash
npm install -g @wenyan-md/cli
```

**验证安装：**
```bash
wenyan --help
```

---

## 3. ❌ 环境变量未设置

**错误信息：**
```
Error: WECHAT_APP_ID is required
```

**解决方法：**

**方式 1: 使用 setup.sh**
```bash
cd /Users/leebot/.openclaw/workspace/wechat-publisher
source ./scripts/setup.sh
```

**方式 2: 手动设置（临时）**
```bash
export WECHAT_APP_ID=your_wechat_app_id
export WECHAT_APP_SECRET=your_wechat_app_secret
```

**方式 3: 永久设置**

编辑 `~/.zshrc` 或 `~/.bashrc`，添加：
```bash
export WECHAT_APP_ID=your_wechat_app_id
export WECHAT_APP_SECRET=your_wechat_app_secret
```

然后：
```bash
source ~/.zshrc  # 或 source ~/.bashrc
```

---

## 4. ❌ Frontmatter 缺失（最常见！）

**错误信息：**
```
Error: 未能找到文章封面
```

**原因：** Markdown 文件缺少必需的 frontmatter（特别是 `title` 字段）。

**⚠️ 重要：** wenyan-cli **强制要求** frontmatter，必须在 Markdown 文件顶部添加！

**解决方法：**

**方案 1：有封面图**
```markdown
---
title: 你的文章标题
cover: /path/to/cover.jpg
---

# 正文开始
...
```

**方案 2：无封面图（推荐，正文有图片即可）**
```markdown
---
title: 你的文章标题
---

# 正文

![配图](https://example.com/image.jpg)  # 正文图片会自动上传

内容...
```

**⚠️ 关键点：**
- `title` 字段是**必须的**，缺少会报错 "未能找到文章封面"
- `cover` 字段可选：如果正文中有图片，wenyan 会自动使用第一张图作为封面
- frontmatter 必须在文件最顶部，前面不能有空行或其他内容
- frontmatter 使用三个短横线 `---` 包围

**错误示例（会报错）：**
```markdown
# 我的文章

没有 frontmatter，wenyan 会报错！
```

---

## 5. ❌ 图片上传失败

**错误信息：**
```
Error: Failed to upload image
```

**可能原因：**

1. **图片路径错误** - 检查本地路径是否正确
2. **图片格式不支持** - 微信支持 jpg/png/gif
3. **图片过大** - 微信限制单张图片 < 10MB
4. **网络图片无法访问** - 检查 URL 是否可访问

**解决方法：**

1. **检查图片路径：**
   ```bash
   ls -lh /path/to/image.jpg
   ```

2. **检查图片格式：**
   ```bash
   file /path/to/image.jpg
   ```

3. **压缩图片（如果过大）：**
   ```bash
   # 使用 ImageMagick
   convert large.jpg -quality 80 -resize 1200x compressed.jpg
   ```

4. **测试网络图片：**
   ```bash
   curl -I https://example.com/image.jpg
   ```

---

## 6. ❌ API 凭证错误

**错误信息：**
```
Error: invalid credential
```

**原因：** AppID 或 AppSecret 错误。

**解决方法：**

1. **检查 TOOLS.md 中的凭证是否正确**

2. **重新获取凭证：**
   - 登录：https://mp.weixin.qq.com/
   - 开发 → 基本配置
   - 查看 AppID 和 AppSecret

3. **更新环境变量**

4. **重试发布**

---

## 7. ❌ Node.js 版本过低

**错误信息：**
```
Error: Requires Node.js >= 14.0.0
```

**解决方法：**

1. **检查当前版本：**
   ```bash
   node --version
   ```

2. **升级 Node.js：**
   ```bash
   # 使用 Homebrew (macOS)
   brew upgrade node
   
   # 或使用 nvm
   nvm install stable
   nvm use stable
   ```

---

## 8. ❌ 网络连接问题

**错误信息：**
```
Error: connect ETIMEDOUT
```

**可能原因：**

1. **网络不稳定** - 检查网络连接
2. **防火墙阻止** - 检查防火墙设置
3. **微信 API 服务异常** - 稍后重试

**解决方法：**

1. **测试网络连接：**
   ```bash
   curl -I https://api.weixin.qq.com
   ```

2. **使用代理（如需要）：**
   ```bash
   export HTTP_PROXY=http://proxy:port
   export HTTPS_PROXY=http://proxy:port
   ```

3. **重试发布**

---

## 9. 🐛 调试模式

如果以上方法都不行，启用详细日志：

```bash
# 设置详细日志
export DEBUG=wenyan:*

# 运行发布
wenyan publish -f article.md -t lapis -h solarized-light
```

查看完整错误信息，然后：
- 检查 wenyan-cli GitHub Issues: https://github.com/caol64/wenyan-cli/issues
- 或提交新 Issue

---

## 10. 📞 获取帮助

**wenyan-cli 帮助：**
```bash
wenyan --help
wenyan publish --help
wenyan theme --help
```

**wechat-publisher 帮助：**
```bash
cd /Users/leebot/.openclaw/workspace/wechat-publisher
./scripts/publish.sh --help
```

**参考资料：**
- wenyan-cli GitHub: https://github.com/caol64/wenyan-cli
- wenyan 官网: https://wenyan.yuzhi.tech
- 微信公众号开发文档: https://developers.weixin.qq.com/doc/offiaccount/

---

## 💡 最佳实践

**1. 始终添加 frontmatter**
```markdown
---
title: 文章标题（必填！）
---
```

**2. 正文添加至少一张图片**
- wenyan 会自动使用第一张图作为封面
- 图片会自动上传到微信图床

**3. 测试流程**
```bash
# 1. 先用 render 测试（不发布）
wenyan render -f article.md

# 2. 确认无误后再 publish
wenyan publish -f article.md -t lapis -h solarized-light
```

**4. 检查列表**
- ✅ frontmatter 中有 title
- ✅ 正文有至少一张图片（或 frontmatter 有 cover）
- ✅ 环境变量已设置
- ✅ IP 在白名单中

---

**如果问题仍未解决，请联系 Bruce 或提交 Issue。**
