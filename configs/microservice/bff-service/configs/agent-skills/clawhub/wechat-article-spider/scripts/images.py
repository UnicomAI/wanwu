"""图片下载处理模块"""
import os
import re
import hashlib
import requests
from urllib.parse import urlparse, urljoin
from typing import List, Tuple


def get_image_filename(url: str, index: int) -> str:
    """生成图片文件名，使用哈希避免重复"""
    parsed = urlparse(url)
    ext = os.path.splitext(parsed.path)[1]
    
    # 如果没有扩展名，尝试从 Content-Type 推断
    if not ext or len(ext) > 5:
        ext = '.jpg'  # 默认
    
    # 使用 URL 哈希生成唯一文件名
    url_hash = hashlib.md5(url.encode()).hexdigest()[:8]
    return f"img_{index:03d}_{url_hash}{ext}"


def download_image(url: str, save_path: str, timeout: int = 10) -> bool:
    """下载单张图片"""
    headers = {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Referer': 'https://mp.weixin.qq.com/'
    }
    
    try:
        response = requests.get(url, headers=headers, timeout=timeout, stream=True)
        response.raise_for_status()
        
        with open(save_path, 'wb') as f:
            for chunk in response.iter_content(chunk_size=8192):
                f.write(chunk)
        
        return True
    except Exception as e:
        print(f"⚠️ 图片下载失败：{url} - {e}")
        return False


def extract_images(html_content: str, base_url: str) -> List[Tuple[str, str]]:
    """从 HTML 中提取所有图片 URL"""
    from bs4 import BeautifulSoup
    
    soup = BeautifulSoup(html_content, 'lxml')
    images = []
    
    # 查找所有 img 标签
    for img in soup.find_all('img'):
        src = img.get('src') or img.get('data-src')
        if src:
            # 处理相对 URL
            full_url = urljoin(base_url, src)
            images.append((full_url, img))
    
    return images


def save_images_and_update_html(images: List[Tuple[str, any]], output_dir: str, base_url: str) -> dict:
    """
    下载所有图片并返回 URL 到本地路径的映射
    """
    images_dir = os.path.join(output_dir, 'images')
    os.makedirs(images_dir, exist_ok=True)
    
    url_mapping = {}
    downloaded_count = 0
    
    for idx, (img_url, img_tag) in enumerate(images, 1):
        if img_url in url_mapping:
            continue
        
        filename = get_image_filename(img_url, idx)
        save_path = os.path.join(images_dir, filename)
        
        if download_image(img_url, save_path):
            downloaded_count += 1
            # 相对路径用于 Markdown
            relative_path = f"images/{filename}"
            url_mapping[img_url] = relative_path
    
    print(f"✅ 下载 {downloaded_count}/{len(images)} 张图片")
    return url_mapping
