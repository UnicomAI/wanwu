# wenyan-cli 主题列表

wenyan-cli 支持多种内置主题，也支持自定义主题。

## 内置主题

查看所有内置主题：
```bash
wenyan theme -l
```

**常用主题：**

1. **default** - 默认主题
   - 简洁、通用
   - 适合大部分文章

2. **lapis** - 青金石（推荐）
   - 优雅的蓝色调
   - 适合技术文章

3. **phycat** - 物理猫
   - 轻量、现代
   - 适合科技类内容

**完整主题列表：** https://github.com/caol64/wenyan-core/tree/main/src/assets/themes

## 代码高亮主题

### 亮色主题
- `atom-one-light` - Atom 编辑器亮色
- `github` - GitHub 风格
- `solarized-light` - Solarized 亮色（推荐）
- `xcode` - Xcode 默认

### 暗色主题
- `atom-one-dark` - Atom 编辑器暗色
- `dracula` - Dracula 主题
- `github-dark` - GitHub 暗色
- `monokai` - Monokai 经典
- `solarized-dark` - Solarized 暗色

## 自定义主题

### 临时使用
```bash
wenyan publish -f article.md -c /path/to/theme.css
```

### 永久安装
```bash
# 从本地文件
wenyan theme --add --name my-theme --path /path/to/theme.css

# 从网络
wenyan theme --add --name my-theme --path https://example.com/theme.css
```

### 使用已安装主题
```bash
wenyan publish -f article.md -t my-theme
```

### 删除主题
```bash
wenyan theme --rm my-theme
```

## 主题定制

如果你想创建自己的主题，可以参考：

1. **查看现有主题源码：** https://github.com/caol64/wenyan-core/tree/main/src/assets/themes
2. **CSS 变量参考：** wenyan 使用 CSS 变量定制样式
3. **测试主题：** 使用 `wenyan render` 命令仅渲染不发布

**示例：**
```bash
# 渲染测试（不发布）
wenyan render -f article.md -t my-theme -h github
```

## 推荐组合

### 技术文章
```bash
wenyan publish -f article.md -t lapis -h solarized-light
```

### 深色风格
```bash
wenyan publish -f article.md -t phycat -h dracula
```

### 简洁风格
```bash
wenyan publish -f article.md -t default -h github
```

## 更多选项

### 关闭 Mac 风格代码块
```bash
wenyan publish -f article.md -t lapis --no-mac-style
```

### 关闭链接转脚注
```bash
wenyan publish -f article.md -t lapis --no-footnote
```

### 组合所有选项
```bash
wenyan publish -f article.md \
  -t lapis \
  -h solarized-light \
  --no-mac-style \
  --no-footnote
```
