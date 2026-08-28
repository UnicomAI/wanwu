#!/usr/bin/env python3
"""
微信公众号文章爬虫

将微信公号文章转换为 Markdown + 本地图片

用法:
    python main.py <文章 URL> [输出目录]
    
默认输出到工作空间下的 docs 目录
"""
import sys
import os
from datetime import datetime

from scraper import fetch_article, extract_article_content, html_to_markdown
from images import extract_images, save_images_and_update_html


# 默认输出目录：工作空间下的 docs 目录
DEFAULT_OUTPUT_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), '..', '..', 'docs')


def main():
    if len(sys.argv) < 2:
        print("❌ 用法：python main.py <文章 URL> [输出目录]")
        print(f"默认输出目录：{DEFAULT_OUTPUT_DIR}")
        sys.exit(1)
    
    article_url = sys.argv[1]
    output_dir = sys.argv[2] if len(sys.argv) > 2 else DEFAULT_OUTPUT_DIR
    
    # 创建输出目录
    os.makedirs(output_dir, exist_ok=True)
    
    print(f"📥 抓取文章：{article_url}")
    
    # 1. 抓取文章
    html = fetch_article(article_url)
    if not html:
        print("❌ 抓取失败")
        sys.exit(1)
    
    print("✅ 抓取成功")
    
    # 2. 提取内容
    article = extract_article_content(html, article_url)
    print(f"📰 标题：{article['title']}")
    if article['author']:
        print(f"✍️ 作者：{article['author']}")
    if article['publish_date']:
        print(f"📅 发布：{article['publish_date']}")
    
    # 3. 提取并下载图片
    print("🖼️  提取图片...")
    images = extract_images(article['raw_html'], article_url)
    print(f"📸 找到 {len(images)} 张图片")
    
    if images:
        image_mapping = save_images_and_update_html(images, output_dir, article_url)
    else:
        image_mapping = {}
    
    # 4. 转换为 Markdown
    print("📝 转换 Markdown...")
    markdown_content = html_to_markdown(article['content_html'], article_url, image_mapping)
    
    # 添加元数据头部
    frontmatter = []
    frontmatter.append(f"# {article['title']}")
    frontmatter.append("")
    if article['author']:
        frontmatter.append(f"**作者**: {article['author']}")
    if article['publish_date']:
        frontmatter.append(f"**发布时间**: {article['publish_date']}")
    frontmatter.append(f"**原文链接**: {article_url}")
    frontmatter.append(f"**抓取时间**: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    frontmatter.append("")
    frontmatter.append("---")
    frontmatter.append("")
    
    final_content = '\n'.join(frontmatter) + markdown_content
    
    # 5. 保存文件
    # 用标题生成文件名
    safe_title = article['title'][:50].replace('/', '_').replace('\\', '_')
    safe_title = "".join(c for c in safe_title if c.isalnum() or c in ' -_').strip()
    if not safe_title:
        safe_title = "wechat-article"
    
    md_filename = f"{safe_title}.md"
    md_path = os.path.join(output_dir, md_filename)
    
    with open(md_path, 'w', encoding='utf-8') as f:
        f.write(final_content)
    
    print(f"✅ 保存成功：{md_path}")
    print(f"📁 图片目录：{os.path.join(output_dir, 'images/')}")


if __name__ == '__main__':
    main()
