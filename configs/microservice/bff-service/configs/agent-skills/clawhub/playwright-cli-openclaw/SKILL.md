---
name: playwright-cli
description: 官方Microsoft Playwright CLI网页自动化工具，支持所有主流浏览器的无头/有头自动化操作，包括页面导航、元素交互、截图、录制、测试等功能。当用户提到网页自动化、浏览器操作、爬虫、截图、录制用户操作、E2E测试时触发。
---

# Playwright CLI 技能

Playwright CLI是微软官方的浏览器自动化工具，支持Chromium、Firefox、WebKit三大主流浏览器，提供强大的网页自动化能力。

## 安装步骤

### 1. 安装Playwright CLI
```bash
npm install -g @playwright/test
# 或者
pip install playwright
```

### 2. 安装浏览器
```bash
playwright install
# 仅安装特定浏览器
playwright install chromium firefox
# 安装系统依赖（Linux环境）
playwright install-deps
```

## 核心常用命令

### 页面操作
```bash
# 打开指定网页
playwright open https://example.com

# 无头模式打开网页并截图
playwright screenshot https://example.com example.png
# 全屏截图
playwright screenshot https://example.com --full-page full.png
# 指定视口大小
playwright screenshot https://example.com --viewport-size=1920,1080 desktop.png

# 生成PDF
playwright pdf https://example.com example.pdf
# 自定义PDF格式
playwright pdf https://example.com --format=A4 --landscape report.pdf
```

### 录制用户操作
```bash
# 录制用户操作并生成代码
playwright codegen https://example.com
# 保存录制结果到文件
playwright codegen https://example.com --output script.py
# 生成指定语言的代码（python, javascript, java, csharp）
playwright codegen https://example.com --target python
```

### 测试相关
```bash
# 运行测试
playwright test
# 运行特定测试文件
playwright test tests/example.spec.js
# 有头模式运行测试（显示浏览器界面）
playwright test --headed
# 调试模式运行
playwright test --debug
# 生成测试报告
playwright show-report
```

### 浏览器管理
```bash
# 列出已安装的浏览器
playwright list-browsers
# 更新浏览器到最新版本
playwright install --force
# 卸载浏览器
playwright uninstall chromium
```

## 使用示例

### 示例1：批量截图多个网页
```bash
# 对多个网页进行全屏截图
for url in "https://google.com" "https://github.com" "https://stackoverflow.com"; do
  name=$(echo $url | sed 's/https\?:\/\///' | sed 's/\//_/g')
  playwright screenshot $url --full-page "${name}.png"
done
```

### 示例2：自动填写表单
```python
# 录制生成的自动登录脚本示例
from playwright.sync_api import Playwright, sync_playwright, expect

def run(playwright: Playwright) -> None:
    browser = playwright.chromium.launch(headless=False)
    context = browser.new_context()
    page = context.new_page()
    page.goto("https://example.com/login")
    page.get_by_label("用户名").fill("your_username")
    page.get_by_label("密码").fill("your_password")
    page.get_by_role("button", name="登录").click()
    # 截图登录后的页面
    page.screenshot(path="logged_in.png")
    context.close()
    browser.close()

with sync_playwright() as playwright:
    run(playwright)
```

### 示例3：模拟移动端访问
```bash
# 模拟iPhone 14访问网页并截图
playwright screenshot https://example.com --device="iPhone 14" mobile.png
# 模拟安卓设备
playwright screenshot https://example.com --device="Galaxy S23" android.png
```

## 最佳实践
1. 优先使用无头模式执行自动化任务，性能更高
2. 复杂操作优先使用`codegen`录制生成代码，再手动调整
3. 执行长时间任务时添加`--slowmo=1000`参数减慢操作速度，避免被反爬
4. 保存登录状态使用`--save-storage=auth.json`，下次可以直接`--load-storage=auth.json`跳过登录
5. 网络不稳定时添加`--timeout=60000`延长超时时间
