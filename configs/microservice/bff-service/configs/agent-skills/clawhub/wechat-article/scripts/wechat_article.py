#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
微信公众号文章抓取工具
通过公众号链接获取文章的标题和内容，输出为 Markdown 格式
"""

import sys
import json
import requests
import re
from urllib.parse import quote
from datetime import datetime


def fetch_via_jina(url: str) -> dict:
    """
    使用 r.jina.ai 提取文章内容
    这是最可靠的方式，支持微信公众号
    """
    jina_url = f"https://r.jina.ai/{url}"
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        "X-Return-Format": "markdown",
        "X-With-Links-Summary": "true",
    }
    
    try:
        response = requests.get(jina_url, headers=headers, timeout=30)
        response.raise_for_status()
        content = response.text.strip()
        
        if content and "环境异常" not in content and "验证码" not in content:
            return {
                "success": True,
                "method": "jina.ai",
                "content": content
            }
        return {"success": False, "method": "jina.ai", "error": "内容被拦截或包含验证码"}
    except requests.exceptions.Timeout:
        return {"success": False, "method": "jina.ai", "error": "请求超时"}
    except requests.exceptions.RequestException as e:
        return {"success": False, "method": "jina.ai", "error": str(e)}
    except Exception as e:
        return {"success": False, "method": "jina.ai", "error": str(e)}


def fetch_via_web_fetch(url: str) -> dict:
    """
    直接抓取网页内容（可能被微信拦截）
    """
    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
        "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
    }
    
    try:
        response = requests.get(url, headers=headers, timeout=30)
        response.raise_for_status()
        html = response.text
        
        # 检查是否被拦截
        if "环境异常" in html or "验证码" in html:
            return {"success": False, "method": "web_fetch", "error": "需要验证码"}
        
        # 提取标题
        title_match = re.search(r'<title>(.*?)</title>', html, re.IGNORECASE)
        title = title_match.group(1) if title_match else "未知标题"
        
        # 提取正文（简单提取）
        content_match = re.search(r'<div[^>]*id="js_content"[^>]*>(.*?)</div>', html, re.DOTALL)
        if content_match:
            content_html = content_match.group(1)
            # 去除 HTML 标签
            content = re.sub(r'<[^>]+>', '', content_html)
            content = content.strip()
        else:
            content = html
        
        return {
            "success": True,
            "method": "web_fetch",
            "title": title,
            "content": content
        }
    except requests.exceptions.Timeout:
        return {"success": False, "method": "web_fetch", "error": "请求超时"}
    except requests.exceptions.RequestException as e:
        return {"success": False, "method": "web_fetch", "error": str(e)}
    except Exception as e:
        return {"success": False, "method": "web_fetch", "error": str(e)}


def extract_title_from_url(url: str) -> str:
    """
    尝试从 URL 或其他来源提取标题
    """
    # 可以尝试通过搜索获取标题
    return ""


def format_markdown(title: str, content: str, url: str, publish_time: str = "", account: str = "") -> str:
    """
    格式化为 Markdown 输出
    """
    md = []
    md.append(f"# {title}")
    md.append("")
    
    if publish_time:
        md.append(f"**发布时间**: {publish_time}")
    if account:
        md.append(f"**公众号**: {account}")
    if url:
        md.append(f"**原文链接**: {url}")
    
    md.append("")
    md.append("---")
    md.append("")
    md.append(content)
    
    return "\n".join(md)


def parse_jina_content(content: str) -> dict:
    """
    解析 jina.ai 返回的内容，提取标题和正文
    """
    lines = content.split("\n")
    title = ""
    body_lines = []
    in_body = False
    
    for i, line in enumerate(lines):
        line = line.strip()
        
        # 第一行非空通常是标题
        if not title and line and not line.startswith("http"):
            title = line.lstrip("# ").strip()
            continue
        
        # 跳过 URL 行
        if line.startswith("http"):
            continue
        
        # 跳过空行标记
        if line == "---":
            continue
        
        body_lines.append(line)
    
    # 清理正文
    body = "\n".join(body_lines).strip()
    
    # 尝试提取发布时间和公众号
    publish_time = ""
    account = ""
    
    return {
        "title": title,
        "content": body,
        "publish_time": publish_time,
        "account": account
    }


def wechat_article(url: str) -> dict:
    """
    主函数：抓取微信公众号文章
    """
    # 验证 URL
    if not url or "mp.weixin.qq.com" not in url:
        return {
            "success": False,
            "error": "无效的微信公众号文章链接"
        }
    
    # 策略 1: 使用 jina.ai（最可靠）
    print(f"📡 尝试使用 jina.ai 提取...", file=sys.stderr)
    result = fetch_via_jina(url)
    if result["success"]:
        print(f"✅ 使用 {result['method']} 成功提取", file=sys.stderr)
        parsed = parse_jina_content(result["content"])
        return {
            "success": True,
            "method": result["method"],
            "title": parsed["title"],
            "content": parsed["content"],
            "publish_time": parsed["publish_time"],
            "account": parsed["account"],
            "url": url
        }
    
    # 策略 2: 直接抓取
    print(f"📡 尝试直接抓取...", file=sys.stderr)
    result = fetch_via_web_fetch(url)
    if result["success"]:
        print(f"✅ 使用 {result['method']} 成功提取", file=sys.stderr)
        return {
            "success": True,
            "method": result["method"],
            "title": result.get("title", "未知标题"),
            "content": result.get("content", ""),
            "url": url
        }
    
    return {
        "success": False,
        "error": "所有提取方式都失败了，文章可能需要验证码或无法访问"
    }


def main():
    if len(sys.argv) < 2:
        print("用法：python3 wechat_article.py '<公众号文章链接>'")
        print("示例：python3 wechat_article.py 'https://mp.weixin.qq.com/s/xxx'")
        sys.exit(1)
    
    url = sys.argv[1]
    
    # 抓取文章
    result = wechat_article(url)
    
    if result["success"]:
        # 输出 Markdown 格式
        md_output = format_markdown(
            title=result.get("title", "未知标题"),
            content=result.get("content", ""),
            url=result.get("url", url),
            publish_time=result.get("publish_time", ""),
            account=result.get("account", "")
        )
        print(md_output)
    else:
        print(f"❌ 提取失败：{result.get('error', '未知错误')}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
