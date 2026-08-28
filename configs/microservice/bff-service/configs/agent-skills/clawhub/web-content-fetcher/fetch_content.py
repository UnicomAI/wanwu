#!/usr/bin/env python3
"""
网页内容获取工具 - Web Content Fetcher
当常规爬虫被过滤时，使用替代服务获取网页内容

用法:
    python fetch_content.py <url> [method]
    python fetch_content.py https://example.com jina
"""

import sys
import subprocess
import argparse
from typing import Optional, Tuple

SERVICES = {
    "jina": {
        "url": "https://r.jina.ai/{url}",
        "desc": "Jina AI Reader - 最稳定，通用性强"
    },
    "markdown": {
        "url": "https://markdown.new/{url}",
        "desc": "Cloudflare Markdown - Cloudflare 保护网站专用"
    },
    "defuddle": {
        "url": "https://defuddle.md/{url}",
        "desc": "Defuddle - 备用方案"
    }
}

def fetch_with_service(url: str, method: str = "jina") -> Tuple[bool, str]:
    """使用指定服务获取网页内容"""
    if method not in SERVICES:
        return False, f"未知服务: {method}"
    
    service_url = SERVICES[method]["url"].format(url=url)
    
    try:
        result = subprocess.run(
            ["curl", "-s", service_url],
            capture_output=True,
            text=True,
            timeout=30
        )
        
        if result.returncode == 0 and result.stdout:
            return True, result.stdout
        else:
            return False, f"获取失败: {result.stderr}"
            
    except subprocess.TimeoutExpired:
        return False, "请求超时"
    except Exception as e:
        return False, f"错误: {str(e)}"

def fetch_with_fallback(url: str) -> str:
    """按优先级尝试所有服务"""
    errors = []
    
    for method in ["jina", "markdown", "defuddle"]:
        print(f"尝试 {method}...", file=sys.stderr)
        success, content = fetch_with_service(url, method)
        
        if success and content:
            print(f"成功使用 {method}!", file=sys.stderr)
            return content
        else:
            errors.append(f"{method}: {content}")
    
    # 所有方法都失败
    error_msg = "所有服务都失败:\n" + "\n".join(errors)
    return f"错误: {error_msg}"

def main():
    parser = argparse.ArgumentParser(description="网页内容获取工具")
    parser.add_argument("url", help="要获取的网页 URL")
    parser.add_argument("--method", "-m", choices=["jina", "markdown", "defuddle", "all"],
                        default="jina", help="获取方法 (默认: jina)")
    
    args = parser.parse_args()
    
    if args.method == "all":
        # 尝试所有方法
        for method in ["jina", "markdown", "defuddle"]:
            print(f"\n=== {method.upper()} ===", file=sys.stderr)
            success, content = fetch_with_service(args.url, method)
            if success:
                print(content)
            else:
                print(f"失败: {content}", file=sys.stderr)
    else:
        # 使用指定方法
        success, content = fetch_with_service(args.url, args.method)
        if success:
            print(content)
        else:
            print(f"错误: {content}", file=sys.stderr)
            sys.exit(1)

if __name__ == "__main__":
    main()
