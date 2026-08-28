#!/usr/bin/env python3
"""Publish Markdown/URL articles to WeChat Official Account."""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
import textwrap
from dataclasses import dataclass
from pathlib import Path
from typing import Any
from urllib.parse import urljoin, urlparse

try:
    import markdown
    import requests
    import yaml
    from bs4 import BeautifulSoup
except Exception:
    markdown = None
    requests = None
    yaml = None
    BeautifulSoup = None

TOKEN_URL = "https://api.weixin.qq.com/cgi-bin/token"
DRAFT_ADD_URL = "https://api.weixin.qq.com/cgi-bin/draft/add"
PUBLISH_SUBMIT_URL = "https://api.weixin.qq.com/cgi-bin/freepublish/submit"
PUBLISH_GET_URL = "https://api.weixin.qq.com/cgi-bin/freepublish/get"
MATERIAL_ADD_URL = "https://api.weixin.qq.com/cgi-bin/material/add_material"


class WeChatPublishError(RuntimeError):
    """Raised when WeChat API reports an error."""


@dataclass
class Article:
    title: str
    content: str
    source_url: str
    first_image_url: str
    digest: str


def ensure_dependencies() -> None:
    if not all([markdown, requests, yaml, BeautifulSoup]):
        raise RuntimeError(
            "依赖未安装。请先运行: python scripts/publish_wechat.py --install"
        )


def install_dependencies() -> None:
    req = Path(__file__).resolve().parent / "requirements.txt"
    if not req.exists():
        raise RuntimeError(f"未找到依赖文件: {req}")
    subprocess.check_call([sys.executable, "-m", "pip", "install", "-r", str(req)])


def slugify(text: str) -> str:
    t = re.sub(r"[^\w\u4e00-\u9fff-]+", "-", text.strip().lower())
    t = re.sub(r"-{2,}", "-", t).strip("-")
    return t or "wechat-article"


def trim_utf8_bytes(text: str, max_bytes: int) -> str:
    raw = (text or "").encode("utf-8")
    if len(raw) <= max_bytes:
        return text or ""
    return raw[:max_bytes].decode("utf-8", errors="ignore")


def decode_escaped_unicode(text: str) -> str:
    """Convert literal \\uXXXX sequences to real Unicode characters."""
    if not text or "\\u" not in text:
        return text or ""
    if not re.search(r"\\u[0-9a-fA-F]{4}", text):
        return text
    try:
        return text.encode("utf-8").decode("unicode_escape")
    except Exception:
        return text


def normalize_wechat_title(text: str) -> str:
    value = decode_escaped_unicode(text or "")
    value = re.sub(r"\s+", " ", value.strip())
    value = re.sub(r"^#+\s*", "", value)
    value = value.replace("“", "\"").replace("”", "\"").replace("—", "-")
    # 标题按字符限制：64 个汉字以内（按 64 个字符处理）。
    value = value[:64].strip()
    return value or "未命名文章"


def is_url(input_value: str) -> bool:
    return bool(re.match(r"^https?://", input_value.strip(), re.I))


def load_config(path: Path) -> dict[str, Any]:
    if not path.exists():
        raise RuntimeError(f"配置文件不存在: {path}")
    data = json.loads(path.read_text(encoding="utf-8"))

    # 支持两种结构：
    # 1) {"wechat": {...}}
    # 2) {"platforms": {"wechat": {...}}}
    if isinstance(data.get("wechat"), dict):
        return data
    if isinstance(data.get("platforms", {}).get("wechat"), dict):
        return {"wechat": data["platforms"]["wechat"]}

    raise RuntimeError("配置文件缺少 wechat 配置")


def parse_frontmatter(md_text: str) -> tuple[dict[str, Any], str]:
    text = decode_escaped_unicode(md_text).lstrip("\ufeff")
    if text.startswith("---"):
        parts = text.split("---", 2)
        if len(parts) >= 3:
            try:
                frontmatter = yaml.safe_load(parts[1]) or {}
                return frontmatter, parts[2].strip()
            except Exception:
                pass
    return {}, text


def extract_title_from_md(md_text: str) -> str:
    text = decode_escaped_unicode(md_text).lstrip("\ufeff")
    match = re.search(r"^#\s+(.+)$", text, re.M)
    if match:
        return match.group(1).strip()

    for line in text.splitlines():
        s = re.sub(r"^#+\s*", "", line.strip())
        if s:
            return s[:64]
    return "未命名文章"


def first_markdown_image(md_text: str) -> str:
    match = re.search(r"!\[[^\]]*\]\(([^)]+)\)", md_text)
    return match.group(1).strip() if match else ""


def extract_from_markdown(path: Path, source_url_override: str = "") -> Article:
    md_text = decode_escaped_unicode(path.read_text(encoding="utf-8")).lstrip("\ufeff")
    frontmatter, body = parse_frontmatter(md_text)
    title = (frontmatter.get("title") or extract_title_from_md(body)).strip()[:64]
    source_url = (source_url_override or frontmatter.get("source_url") or "").strip()
    html_content = markdown_to_html(body)
    plain = html_to_plain_text(html_content)
    digest = plain[:120]
    return Article(
        title=title or "未命名文章",
        content=html_content,
        source_url=source_url,
        first_image_url=first_markdown_image(body),
        digest=digest,
    )


def extract_from_url(url: str) -> Article:
    resp = requests.get(url, timeout=30)
    resp.raise_for_status()

    soup = BeautifulSoup(resp.text, "html.parser")
    title_node = (
        soup.select_one("#cb_post_title_url")
        or soup.find("h1")
        or soup.find("title")
    )
    title = title_node.get_text(strip=True)[:64] if title_node else "未命名文章"

    body_node = (
        soup.select_one("#cnblogs_post_body")
        or soup.find("article")
        or soup.find("main")
        or soup.find("div", class_=re.compile(r"content|article|post", re.I))
        or soup.body
    )
    if not body_node:
        raise RuntimeError("无法从网页提取正文")

    for tag in body_node.select("script,style,iframe,.ad,.ads,.advertisement,.comment"):
        tag.decompose()
    for img in body_node.find_all("img"):
        src = (img.get("src") or "").strip()
        if src:
            img["src"] = urljoin(url, src)
    for a in body_node.find_all("a"):
        href = (a.get("href") or "").strip()
        if href:
            a["href"] = urljoin(url, href)

    first_image_url = ""
    first_img = body_node.find("img")
    if first_img and first_img.get("src"):
        first_image_url = first_img["src"]

    html_content = str(body_node)
    plain = html_to_plain_text(html_content)
    digest = plain[:120]

    return Article(
        title=title,
        content=html_content,
        source_url=url,
        first_image_url=first_image_url,
        digest=digest,
    )


def is_probably_html(text: str) -> bool:
    sample = (text or "").strip()[:500].lower()
    return bool(
        re.search(r"<(article|section|div|p|h[1-6]|ul|ol|pre|table|img|blockquote)\b", sample)
    )


def markdown_to_html(text: str) -> str:
    text = decode_escaped_unicode(text or "")
    if is_probably_html(text):
        return text
    return markdown.markdown(
        text,
        extensions=["extra", "tables", "nl2br", "codehilite"],
        output_format="html",
    )


def html_to_plain_text(html: str) -> str:
    soup = BeautifulSoup(html, "html.parser")
    txt = soup.get_text("\n", strip=True)
    txt = re.sub(r"\n+", "\n", txt)
    return txt


def _next_element_sibling(node):
    sibling = node.next_sibling
    while sibling is not None:
        name = getattr(sibling, "name", None)
        if name:
            return sibling
        if str(sibling).strip():
            return None
        sibling = sibling.next_sibling
    return None


def _normalize_wechat_dom(soup: BeautifulSoup) -> None:
    for list_node in soup.find_all(["ul", "ol"]):
        for child in list(list_node.contents):
            name = getattr(child, "name", None)
            if name == "li":
                continue
            text = str(child).strip() if not name else ""
            if text and text not in ("\n", "\r\n"):
                li = soup.new_tag("li")
                li.string = text
                child.replace_with(li)
            else:
                try:
                    child.extract()
                except Exception:
                    pass

    for li in soup.find_all("li"):
        txt = li.get_text(" ", strip=True)
        if not txt or txt in {"•", "-", "·"}:
            li.decompose()

    for pre in soup.find_all("pre"):
        code = pre.find("code")
        target = code if code else pre
        text = target.get_text() if target else ""
        if not text:
            continue
        text = text.replace("\r\n", "\n").replace("\r", "\n").replace("\t", "    ")
        lines = [line.replace("\u00a0", " ").replace("\u3000", " ").rstrip() for line in text.split("\n")]
        cleaned = textwrap.dedent("\n".join(lines).strip("\n"))
        target.clear()
        target.append(cleaned)


def _enhance_reference_links(soup: BeautifulSoup) -> None:
    ref_keywords = ("参考资源", "相关链接", "references", "reference")
    for heading in soup.find_all(re.compile(r"^h[1-6]$")):
        title = heading.get_text(" ", strip=True).lower()
        if not any(k in title for k in ref_keywords):
            continue

        current_level = int(heading.name[1])
        node = heading.next_sibling
        while node is not None:
            if getattr(node, "name", None) and re.match(r"^h[1-6]$", node.name):
                next_level = int(node.name[1])
                if next_level <= current_level:
                    break

            anchors = node.find_all("a") if getattr(node, "find_all", None) else []
            for a in anchors:
                href = (a.get("href") or "").strip()
                if not href:
                    continue
                parent_text = a.parent.get_text(" ", strip=True) if a.parent else ""
                if href in parent_text:
                    continue
                url_anchor = soup.new_tag("a", href=href)
                url_anchor["style"] = "color:#576b95;text-decoration:underline;word-break:break-all;"
                url_anchor["target"] = "_blank"
                url_anchor.string = href
                tail = soup.new_tag("span")
                tail["style"] = "display:block;margin-top:2px;font-size:13px;line-height:1.6;color:#7a7a7a;"
                tail.append(url_anchor)
                a.insert_after(tail)
            node = node.next_sibling


def _style_lead_paragraph(soup: BeautifulSoup, template: str) -> None:
    for p in soup.find_all("p"):
        if p.find_parent(["li", "blockquote", "td", "th"]):
            continue
        text = p.get_text(" ", strip=True)
        if not text:
            continue
        if template == "viral":
            p["style"] = (
                "margin:1em 0;padding:0.88em 0.96em;background:#fff7e6;border:1px solid #ffe1a6;"
                "border-radius:8px;line-height:1.9;font-size:16px;color:#2b2f38;text-indent:0;"
                "box-shadow:0 2px 10px rgba(251,146,60,0.12);"
            )
        else:
            p["style"] = (
                "margin:1em 0;padding:0.75em 0.9em;background:#f8fafc;border:1px solid #edf2f7;"
                "border-radius:6px;line-height:1.9;font-size:16px;color:#243044;text-indent:0;"
            )
        break


def _decorate_highlights(soup: BeautifulSoup, template: str) -> None:
    for p in soup.find_all("p"):
        text = p.get_text(" ", strip=True)
        if text not in ("看点：", "看点:"):
            continue
        if template == "viral":
            p["style"] = (
                "margin:0.95em 0 0.35em;display:inline-block;padding:0.24em 0.72em;"
                "font-size:14px;line-height:1.4;color:#b42318;font-weight:700;"
                "background:#fff1f1;border-radius:999px;border:1px solid #ffcece;"
            )
        else:
            p["style"] = (
                "margin:0.95em 0 0.35em;display:inline-block;padding:0.22em 0.65em;"
                "font-size:14px;line-height:1.4;color:#2457c5;font-weight:700;"
                "background:#eef4ff;border-radius:999px;border:1px solid #dce9ff;"
            )

        next_block = _next_element_sibling(p)
        if next_block and next_block.name in ("ul", "ol"):
            if template == "viral":
                next_block["style"] = (
                    "margin:0.45em 0 1em;padding:0.78em 0.98em 0.78em 1.6em;line-height:1.85;"
                    "color:#2d3445;background:#fff7ec;border-radius:8px;"
                    "border:1px solid #ffe3bf;box-shadow:0 2px 12px rgba(251,146,60,0.10);"
                )
            else:
                next_block["style"] = (
                    "margin:0.45em 0 1em;padding:0.75em 0.95em 0.75em 1.55em;line-height:1.82;"
                    "color:#2f3441;background:#f8fafc;border-radius:6px;border:1px solid #e8eef8;"
                )


def optimize_for_wechat_html(content_html: str, template: str = "standard") -> str:
    template = (template or "standard").strip().lower()
    soup = BeautifulSoup(content_html, "html.parser")
    for tag in soup.find_all(["script", "style"]):
        tag.decompose()

    _normalize_wechat_dom(soup)
    _enhance_reference_links(soup)
    _decorate_highlights(soup, template)

    first_h1 = soup.find("h1")
    if first_h1:
        first_h1.decompose()

    for h2 in soup.find_all("h2"):
        if template == "viral":
            h2["style"] = (
                "margin:1.45em 0 0.72em;padding:0.46em 0.72em;font-size:1.2em;line-height:1.45;"
                "color:#7a1f1f;font-weight:800;background:#fff2f2;border-left:4px solid #ef4444;border-radius:4px;"
            )
        else:
            h2["style"] = (
                "margin:1.4em 0 0.7em;padding:0.38em 0.65em;font-size:1.18em;line-height:1.45;"
                "color:#1f2937;font-weight:700;background:#f5f8ff;border-left:4px solid #3b6dd8;border-radius:2px;"
            )

    for p in soup.find_all("p"):
        if p.get_text(" ", strip=True) in ("看点：", "看点:"):
            continue
        p["style"] = (
            "margin:0.95em 0;line-height:1.92;font-size:16px;color:#2b2f38;"
            "text-align:justify;letter-spacing:0.01em;text-indent:0;"
        )

    list_style = (
        "margin:0.98em 0;padding:0.82em 1em 0.82em 1.62em;line-height:1.86;color:#2d3445;"
        "background:#fffdf8;border-radius:8px;border:1px solid #ffe8c4;"
        "box-shadow:0 2px 10px rgba(251,146,60,0.08);list-style-position:outside;"
    ) if template == "viral" else (
        "margin:0.95em 0;padding:0.75em 0.95em 0.75em 1.55em;line-height:1.8;color:#2f3441;"
        "background:#f8fafc;border-radius:6px;border:1px solid #edf2f7;list-style-position:outside;"
    )

    for ul in soup.find_all("ul"):
        ul["style"] = list_style + "list-style-type:disc;"
    for ol in soup.find_all("ol"):
        ol["style"] = list_style + "list-style-type:decimal;"
    for li in soup.find_all("li"):
        li["style"] = "margin:0.28em 0;font-size:16px;line-height:1.82;text-indent:0;"

    for a in soup.find_all("a"):
        a["style"] = (
            "color:#c2410c;text-decoration:underline;word-break:break-all;"
            if template == "viral"
            else "color:#1f57c3;text-decoration:underline;word-break:break-all;"
        )
        if a.get("href"):
            a["target"] = "_blank"

    for strong in soup.find_all("strong"):
        if template == "viral":
            strong["style"] = (
                "font-weight:800;color:#9a3412;background:#fff3d6;padding:0 2px;border-radius:3px;"
            )
        else:
            strong["style"] = "font-weight:700;color:#1f2937;"

    for blockquote in soup.find_all("blockquote"):
        if template == "viral":
            blockquote["style"] = (
                "margin:1.1em 0;padding:0.82em 1.02em;border-left:3px solid #f59e0b;"
                "background:#fff8e6;color:#5f4b20;line-height:1.85;border-radius:6px;"
            )
        else:
            blockquote["style"] = (
                "margin:1.1em 0;padding:0.8em 1em;border-left:3px solid #7aa2ff;"
                "background:#f4f7ff;color:#43506a;line-height:1.8;border-radius:4px;"
            )

    _style_lead_paragraph(soup, template)
    body_html = "".join(str(node) for node in soup.contents)
    wrapper = (
        "<section style=\"margin:0 auto;max-width:690px;padding:0;"
        "font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','PingFang SC','Hiragino Sans GB','Microsoft YaHei',sans-serif;"
        "color:#1f2937;font-size:16px;line-height:1.9;\">"
        f"{body_html}"
        "</section>"
    )
    return wrapper


def generate_cover_image(title: str, output_path: Path) -> Path | None:
    try:
        from PIL import Image, ImageDraw, ImageFont
    except Exception:
        return None

    width, height = 900, 383
    img = Image.new("RGB", (width, height))
    draw = ImageDraw.Draw(img)
    top = (60, 130, 246)
    bottom = (33, 197, 147)
    for y in range(height):
        ratio = y / max(1, height - 1)
        r = int(top[0] + (bottom[0] - top[0]) * ratio)
        g = int(top[1] + (bottom[1] - top[1]) * ratio)
        b = int(top[2] + (bottom[2] - top[2]) * ratio)
        draw.line([(0, y), (width, y)], fill=(r, g, b))

    font_candidates = [
        "C:\\Windows\\Fonts\\msyhbd.ttc",
        "C:\\Windows\\Fonts\\msyh.ttc",
        "C:\\Windows\\Fonts\\simhei.ttf",
    ]
    font = None
    for c in font_candidates:
        try:
            font = ImageFont.truetype(c, 56)
            break
        except Exception:
            continue
    if font is None:
        font = ImageFont.load_default()

    clean = re.sub(r"\s+", " ", BeautifulSoup(title, "html.parser").get_text(" ", strip=True))
    line = clean[:26] + ("..." if len(clean) > 26 else "")
    draw.text((50, 130), line, fill=(255, 255, 255), font=font)
    output_path.parent.mkdir(parents=True, exist_ok=True)
    img.save(output_path, "JPEG", quality=92)
    return output_path


class WeChatClient:
    def __init__(self, app_id: str, app_secret: str, timeout: int = 30):
        self.app_id = app_id
        self.app_secret = app_secret
        self.timeout = timeout

    def get_token(self) -> str:
        resp = requests.get(
            TOKEN_URL,
            params={
                "grant_type": "client_credential",
                "appid": self.app_id,
                "secret": self.app_secret,
            },
            timeout=self.timeout,
        )
        resp.raise_for_status()
        data = resp.json()
        if data.get("errcode", 0) != 0:
            raise WeChatPublishError(f"get_token failed: {data}")
        token = data.get("access_token", "")
        if not token:
            raise WeChatPublishError("get_token failed: missing access_token")
        return token

    def _post_json_utf8(self, url: str, params: dict[str, Any], payload: dict[str, Any]) -> dict[str, Any]:
        body = json.dumps(payload, ensure_ascii=False).encode("utf-8")
        resp = requests.post(
            url,
            params=params,
            data=body,
            headers={"Content-Type": "application/json; charset=utf-8"},
            timeout=self.timeout,
        )
        resp.raise_for_status()
        return json.loads(resp.content.decode("utf-8"))

    def upload_image_from_path(self, token: str, image_path: Path) -> str:
        with image_path.open("rb") as fh:
            files = {"media": (image_path.name, fh, "image/jpeg")}
            resp = requests.post(
                MATERIAL_ADD_URL,
                params={"access_token": token, "type": "image"},
                files=files,
                timeout=self.timeout,
            )
        resp.raise_for_status()
        data = resp.json()
        if data.get("errcode", 0) != 0:
            raise WeChatPublishError(f"upload image failed: {data}")
        media_id = data.get("media_id", "")
        if not media_id:
            raise WeChatPublishError("upload image failed: missing media_id")
        return media_id

    def upload_image_from_url(self, token: str, image_url: str) -> str:
        img = requests.get(image_url, timeout=self.timeout)
        img.raise_for_status()
        content_type = img.headers.get("Content-Type", "image/jpeg")
        files = {"media": ("thumb.jpg", img.content, content_type)}
        resp = requests.post(
            MATERIAL_ADD_URL,
            params={"access_token": token, "type": "image"},
            files=files,
            timeout=self.timeout,
        )
        resp.raise_for_status()
        data = resp.json()
        if data.get("errcode", 0) != 0:
            raise WeChatPublishError(f"upload image url failed: {data}")
        media_id = data.get("media_id", "")
        if not media_id:
            raise WeChatPublishError("upload image url failed: missing media_id")
        return media_id

    def add_draft(
        self,
        token: str,
        title: str,
        author: str,
        digest: str,
        content_html: str,
        source_url: str,
        thumb_media_id: str,
    ) -> str:
        safe_title_base = normalize_wechat_title(title)
        safe_author = trim_utf8_bytes(author, 16)
        safe_digest_base = decode_escaped_unicode(digest or "")
        safe_digest_base = re.sub(r"\s+", " ", safe_digest_base).strip()

        title_limits = [64, 48, 36, 28, 20, 12]
        digest_limits = [120, 90, 60, 40, 20]
        title_candidates: list[str] = []
        digest_candidates: list[str] = []

        for limit in title_limits:
            candidate = safe_title_base[:limit].strip()
            if candidate and candidate not in title_candidates:
                title_candidates.append(candidate)
        for limit in digest_limits:
            candidate = safe_digest_base[:limit].strip()
            if candidate and candidate not in digest_candidates:
                digest_candidates.append(candidate)
        if not digest_candidates:
            digest_candidates = ["AI 资讯速览"]

        last_error: dict[str, Any] | None = None
        for safe_title in title_candidates or ["未命名文章"]:
            should_shorten_title = False
            for safe_digest in digest_candidates:
                payload = {
                    "articles": [
                        {
                            "title": safe_title,
                            "author": safe_author,
                            "digest": safe_digest,
                            "content": content_html,
                            "content_source_url": source_url[:200],
                            "thumb_media_id": thumb_media_id,
                            "need_open_comment": 1,
                            "only_fans_can_comment": 0,
                        }
                    ]
                }
                data = self._post_json_utf8(
                    DRAFT_ADD_URL,
                    {"access_token": token},
                    payload,
                )
                if data.get("errcode", 0) == 0:
                    media_id = data.get("media_id", "")
                    if not media_id:
                        raise WeChatPublishError("draft add failed: missing media_id")
                    return media_id

                last_error = data
                errcode = data.get("errcode")
                if errcode == 45003:
                    should_shorten_title = True
                    break
                if errcode == 45004:
                    continue
                raise WeChatPublishError(f"draft add failed: {data}")

            if should_shorten_title:
                continue

        raise WeChatPublishError(f"draft add failed: {last_error}")

    def submit_publish(self, token: str, media_id: str) -> str:
        data = self._post_json_utf8(
            PUBLISH_SUBMIT_URL,
            {"access_token": token},
            {"media_id": media_id},
        )
        if data.get("errcode", 0) != 0:
            raise WeChatPublishError(f"publish submit failed: {data}")
        publish_id = data.get("publish_id", "")
        if not publish_id:
            raise WeChatPublishError("publish submit failed: missing publish_id")
        return publish_id

    def get_publish_status(self, token: str, publish_id: str) -> dict[str, Any]:
        data = self._post_json_utf8(
            PUBLISH_GET_URL,
            {"access_token": token},
            {"publish_id": publish_id},
        )
        if data.get("errcode", 0) != 0:
            raise WeChatPublishError(f"publish get failed: {data}")
        return data


def parse_args() -> argparse.Namespace:
    default_config = Path(__file__).resolve().parent.parent / "config.json"
    parser = argparse.ArgumentParser(description="Publish article to WeChat Official Account")
    parser.add_argument("input", nargs="?", help="Markdown file path or article URL")
    parser.add_argument("--config", default=str(default_config), help="Path to config.json")
    parser.add_argument("--template", choices=["standard", "viral"], default="", help="Override template")
    parser.add_argument("--author", default="", help="Override author")
    parser.add_argument("--source-url", default="", help="Override source url")
    parser.add_argument("--cover-image", default="", help="Local cover image path")
    parser.add_argument("--timeout", type=int, default=30, help="HTTP timeout seconds")
    parser.add_argument("--dry-run", action="store_true", help="Render only, no WeChat API calls")
    parser.add_argument("--publish", action="store_true", help="Submit draft for publish")
    parser.add_argument("--status", action="store_true", help="Query publish status once")
    parser.add_argument("--install", action="store_true", help="Install Python dependencies")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    if args.install:
        install_dependencies()
        print(json.dumps({"success": True, "installed": True}, ensure_ascii=False))
        return

    if not args.input:
        raise RuntimeError("缺少输入参数：请传入 Markdown 文件路径或 URL")
    ensure_dependencies()

    cfg = load_config(Path(args.config))
    wechat_cfg = cfg["wechat"]
    template = (args.template or "viral").strip().lower()
    if template not in {"standard", "viral"}:
        template = "standard"
    author = (args.author or wechat_cfg.get("author") or "").strip()
    input_value = args.input.strip()

    if is_url(input_value):
        article = extract_from_url(input_value)
        if args.source_url:
            article.source_url = args.source_url.strip()
    else:
        input_path = Path(input_value).resolve()
        if not input_path.exists():
            raise RuntimeError(f"文件不存在: {input_path}")
        article = extract_from_markdown(input_path, source_url_override=args.source_url.strip())

    article.content = optimize_for_wechat_html(article.content, template=template)

    preview_path = Path.cwd() / f"{slugify(article.title)}-wechat-preview.html"
    preview_path.write_text(article.content, encoding="utf-8")

    if args.dry_run:
        print(
            json.dumps(
                {
                    "success": True,
                    "mode": "dry-run",
                    "title": article.title,
                    "digest": article.digest,
                    "template": template,
                    "preview_html": str(preview_path),
                },
                ensure_ascii=False,
            )
        )
        return

    app_id = (wechat_cfg.get("app_id") or "").strip()
    app_secret = (wechat_cfg.get("app_secret") or "").strip()
    if not app_id or not app_secret:
        raise RuntimeError("配置缺少 wechat.app_id 或 wechat.app_secret")

    client = WeChatClient(app_id=app_id, app_secret=app_secret, timeout=args.timeout)
    token = client.get_token()

    thumb_media_id = ""
    auto_generate_cover = True
    cover_image_path = Path(args.cover_image).resolve() if args.cover_image else None
    if not thumb_media_id:
        if cover_image_path and cover_image_path.exists():
            thumb_media_id = client.upload_image_from_path(token, cover_image_path)
        else:
            generated = None
            if auto_generate_cover:
                generated = generate_cover_image(
                    article.title,
                    Path(__file__).resolve().parent.parent / "assets" / "generated_cover.jpg",
                )
            if generated and generated.exists():
                thumb_media_id = client.upload_image_from_path(token, generated)
            elif article.first_image_url and article.first_image_url.startswith(("http://", "https://")):
                thumb_media_id = client.upload_image_from_url(token, article.first_image_url)

    if not thumb_media_id:
        raise RuntimeError(
            "无法获取封面素材，请提供 --cover-image，或确保正文含可下载图片，或安装 Pillow 以自动生成封面"
        )

    draft_media_id = client.add_draft(
        token=token,
        title=article.title,
        author=author,
        digest=article.digest,
        content_html=article.content,
        source_url=article.source_url,
        thumb_media_id=thumb_media_id,
    )

    result: dict[str, Any] = {
        "success": True,
        "title": article.title,
        "template": template,
        "draft_media_id": draft_media_id,
        "preview_html": str(preview_path),
    }

    if args.publish:
        publish_id = client.submit_publish(token, draft_media_id)
        result["publish_id"] = publish_id
        if args.status:
            result["status"] = client.get_publish_status(token, publish_id)

    print(json.dumps(result, ensure_ascii=False))


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(
            json.dumps(
                {"success": False, "error": str(exc), "type": type(exc).__name__},
                ensure_ascii=False,
            )
        )
        sys.exit(1)
