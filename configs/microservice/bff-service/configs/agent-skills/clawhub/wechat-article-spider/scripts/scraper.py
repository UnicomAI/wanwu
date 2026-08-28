"""微信文章抓取和解析模块"""
import re
from typing import Optional, Dict
from bs4 import BeautifulSoup
import requests


def fetch_article(url: str) -> Optional[str]:
    """抓取文章 HTML 内容"""
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8',
        'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8',
    }
    
    try:
        response = requests.get(url, headers=headers, timeout=15)
        response.raise_for_status()
        response.encoding = response.apparent_encoding  # 自动检测编码
        return response.text
    except Exception as e:
        print(f"❌ 抓取失败：{e}")
        return None


def extract_article_content(html: str, url: str) -> Dict:
    """提取文章标题、正文等核心内容"""
    soup = BeautifulSoup(html, 'lxml')
    
    # 提取标题
    title = ""
    title_tag = soup.find('h1', class_='rich_media_title') or soup.find('h2', class_='rich_media_title')
    if title_tag:
        title = title_tag.get_text(strip=True)
    else:
        # 备用：从 meta 标签获取
        meta_title = soup.find('meta', property='og:title')
        if meta_title:
            title = meta_title.get('content', '')
    
    # 如果没有标题，用 URL 生成
    if not title:
        title = f"微信文章 - {url}"
    
    # 提取正文
    content_html = ""
    content_div = soup.find('div', id='js_content') or soup.find('div', class_='rich_media_content')
    
    if content_div:
        content_html = str(content_div)
    else:
        # 备用：查找包含主要内容的 div
        for div in soup.find_all('div', class_='rich_media_area_primary'):
            content_div = div.find('section')
            if content_div:
                content_html = str(content_div)
                break
    
    # 提取作者
    author = ""
    author_tag = soup.find('span', class_='rich_media_meta_nickname')
    if author_tag:
        author = author_tag.get_text(strip=True)
    
    # 提取发布时间
    publish_date = ""
    date_tag = soup.find('em', class_='rich_media_meta_text')
    if date_tag:
        publish_date = date_tag.get_text(strip=True)
    
    return {
        'title': title,
        'author': author,
        'publish_date': publish_date,
        'content_html': content_html,
        'raw_html': html
    }


def html_to_markdown(content_html: str, url: str, image_mapping: dict) -> str:
    """将 HTML 内容转换为 Markdown，替换图片链接"""
    if not content_html:
        return ""
    
    soup = BeautifulSoup(content_html, 'lxml')
    markdown_parts = []
    
    # 递归处理元素
    def process_element(element):
        if isinstance(element, str):
            text = element.strip()
            if text:
                return text
            return None
        
        if element.name == 'img':
            src = element.get('src') or element.get('data-src')
            if src:
                from urllib.parse import urljoin
                full_url = urljoin(url, src)
                if full_url in image_mapping:
                    alt = element.get('alt', '')
                    return f"![{alt}]({image_mapping[full_url]})"
            return None
        
        if element.name in ['p', 'section']:
            parts = []
            for child in element.children:
                result = process_element(child)
                if result:
                    parts.append(result)
            if parts:
                return ' '.join(parts)
            return None
        
        if element.name in ['h1', 'h2', 'h3', 'h4', 'h5', 'h6']:
            level = int(element.name[1])
            text = element.get_text(strip=True)
            if text:
                return f"{'#' * level} {text}"
            return None
        
        if element.name in ['ul', 'ol']:
            return process_list(element)
        
        if element.name == 'blockquote':
            text = element.get_text(strip=True)
            if text:
                return f"> {text}"
            return None
        
        if element.name == 'br':
            return "\n"
        
        # 默认：处理所有子元素
        parts = []
        for child in element.children:
            result = process_element(child)
            if result:
                parts.append(result)
        
        if parts:
            return ' '.join(parts)
        return None
    
    # 处理所有顶级元素
    for element in soup.children:
        result = process_element(element)
        if result:
            markdown_parts.append(result)
    
    # 合并结果
    result = '\n\n'.join(markdown_parts)
    
    # 清理多余空白
    result = re.sub(r'\n{3,}', '\n\n', result)
    result = result.strip()
    
    return result


def process_list(list_tag) -> str:
    """处理列表元素"""
    items = []
    is_ordered = list_tag.name == 'ol'
    
    for idx, li in enumerate(list_tag.find_all('li', recursive=False), 1):
        text = li.get_text(strip=True)
        if is_ordered:
            items.append(f"{idx}. {text}")
        else:
            items.append(f"- {text}")
    
    return '\n'.join(items) if items else ""
