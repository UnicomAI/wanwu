#!/usr/bin/env python3
"""
Scan source files for Chinese comments and append English translations inline.
Translations are added in the format "[EN] <translation>" so the script can
skip already-processed comments on subsequent runs.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import sys
import time
import urllib.parse
import urllib.request
from pathlib import Path
from typing import Dict, List, Sequence, Tuple

HAN_RE = re.compile(r"[\u4e00-\u9fff]")

TRANSLATE_URL = "https://translate.googleapis.com/translate_a/single"
USER_AGENT = "wanwu-translate-script/1.0"

SKIP_DIRS = {
    ".git",
    ".idea",
    ".vscode",
    ".venv",
    "vendor",
    "node_modules",
    "dist",
    "build",
    "__pycache__",
    "docs",
    ".nuxt",
    ".output",
}

DOUBLE_SLASH_SUFFIXES = {
    ".go",
    ".proto",
    ".ts",
    ".tsx",
    ".js",
    ".jsx",
    ".vue",
    ".java",
    ".c",
    ".h",
    ".cc",
    ".cpp",
    ".mjs",
    ".cjs",
    ".rs",
}

HASH_SUFFIXES = {
    ".sh",
    ".bash",
    ".zsh",
    ".py",
    ".yaml",
    ".yml",
    ".toml",
    ".ini",
    ".cfg",
    ".conf",
    ".mk",
    ".env",
    ".dockerfile",
}

HASH_FILENAMES = {
    "Dockerfile",
    "Makefile",
    "Makefile.develop",
}

CODE_EXTENSIONS = DOUBLE_SLASH_SUFFIXES | HASH_SUFFIXES | {
    ".json",
    ".scss",
    ".css",
    ".less",
    ".sass",
    ".html",
    ".mdx",
}


class TranslateClient:
    def __init__(self, source_lang: str = "zh-CN", target_lang: str = "en") -> None:
        self.source_lang = source_lang
        self.target_lang = target_lang

    def translate(self, text: str) -> str:
        params = {
            "client": "gtx",
            "sl": self.source_lang,
            "tl": self.target_lang,
            "dt": "t",
            "q": text,
        }
        url = TRANSLATE_URL + "?" + urllib.parse.urlencode(params)
        req = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
        with urllib.request.urlopen(req, timeout=10) as response:
            payload = response.read()
        data = json.loads(payload.decode("utf-8"))
        if not data or not data[0]:
            raise RuntimeError(f"Unexpected translation response: {data}")
        return "".join(segment[0] for segment in data[0] if segment and segment[0])


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Translate Chinese comments to English.")
    parser.add_argument("--root", type=Path, default=Path("."), help="Repository root to scan.")
    parser.add_argument("--dry-run", action="store_true", help="Scan and report without writing files.")
    parser.add_argument(
        "--limit",
        type=int,
        default=None,
        help="Optional limit for number of comment translations (useful for testing).",
    )
    return parser.parse_args()


def path_has_skip_part(path: Path, root: Path) -> bool:
    try:
        rel = path.relative_to(root)
    except ValueError:
        return True
    for part in rel.parts:
        if part in SKIP_DIRS:
            return True
    return False


def should_process_file(path: Path) -> bool:
    suffix = path.suffix.lower()
    if suffix in CODE_EXTENSIONS:
        return True
    if path.name in HASH_FILENAMES or path.name.startswith("Dockerfile."):
        return True
    return False


def markers_for_file(path: Path) -> Sequence[str]:
    markers: List[str] = []
    suffix = path.suffix.lower()
    if suffix in DOUBLE_SLASH_SUFFIXES or suffix in {".json", ".scss", ".css", ".less", ".sass", ".html", ".mdx"}:
        markers.append("//")
    if suffix in HASH_SUFFIXES or path.name in HASH_FILENAMES or path.name.startswith("Dockerfile"):
        markers.append("#")
    return markers


def find_comment(line: str, markers: Sequence[str]) -> Tuple[int, str] | Tuple[int, None]:
    in_single = False
    in_double = False
    in_backtick = False
    escape = False
    i = 0
    while i < len(line):
        if not in_single and not in_double and not in_backtick:
            if "//" in markers and line.startswith("//", i):
                if i > 0 and line[i - 1] == ":":
                    pass
                else:
                    return i, "//"
            if "#" in markers and line[i] == "#":
                return i, "#"
        ch = line[i]
        if escape:
            escape = False
        else:
            if ch == "\\" and (in_single or in_double):
                escape = True
            elif ch == '"' and not in_single and not in_backtick:
                in_double = not in_double
            elif ch == "'" and not in_double and not in_backtick:
                in_single = not in_single
            elif ch == "`" and not in_single and not in_double:
                in_backtick = not in_backtick
        i += 1
    return -1, None


def translate_comment(text: str, client: TranslateClient, cache: Dict[str, str]) -> str:
    key = text.strip()
    if not key:
        return ""
    cached = cache.get(key)
    if cached:
        return cached

    attempts = 0
    last_exc: Exception | None = None
    while attempts < 3:
        attempts += 1
        try:
            translated = client.translate(key).strip()
            cache[key] = translated
            return translated
        except Exception as exc:
            last_exc = exc
            time.sleep(1.5 * attempts)
    raise RuntimeError(f"Failed to translate comment: {key}") from last_exc


def process_line(line: str, markers: Sequence[str], client: TranslateClient, cache: Dict[str, str]) -> Tuple[str, bool]:
    if not markers:
        return line, False
    idx, marker = find_comment(line, markers)
    if idx == -1 or marker is None:
        return line, False

    comment_body = line[idx + len(marker) :]
    stripped = comment_body.strip()
    if not stripped or "[EN]" in stripped:
        return line, False
    if not HAN_RE.search(stripped):
        return line, False

    translation = translate_comment(stripped, client, cache)
    no_newline = comment_body.rstrip("\n")
    base = no_newline.rstrip()
    trailing = no_newline[len(base) :]
    prefix = line[: idx + len(marker)]
    if base:
        new_comment = f"{base} [EN] {translation}{trailing}"
    else:
        new_comment = f" [EN] {translation}{trailing}"
    suffix = "\n" if line.endswith("\n") else ""
    return f"{prefix}{new_comment}{suffix}", True


def process_file(path: Path, client: TranslateClient, dry_run: bool, cache: Dict[str, str], limit: List[int]) -> int:
    markers = markers_for_file(path)
    if not markers:
        return 0

    try:
        lines = path.read_text(encoding="utf-8").splitlines(keepends=True)
    except UnicodeDecodeError:
        return 0

    changed = False
    translated_count = 0
    new_lines: List[str] = []
    for line in lines:
        if limit[0] is not None and limit[1] >= limit[0]:
            new_lines.append(line)
            continue
        new_line, updated = process_line(line, markers, client, cache)
        if updated:
            limit[1] += 1
            translated_count += 1
            changed = True
        new_lines.append(new_line)

    if changed and not dry_run:
        path.write_text("".join(new_lines), encoding="utf-8")
    return translated_count


def collect_files(root: Path) -> List[Path]:
    files: List[Path] = []
    for dirpath, dirnames, filenames in os.walk(root):
        current = Path(dirpath)
        if path_has_skip_part(current, root):
            dirnames[:] = []
            continue
        keep_dirs = []
        for dirname in dirnames:
            candidate = current / dirname
            if not path_has_skip_part(candidate, root):
                keep_dirs.append(dirname)
        dirnames[:] = keep_dirs

        for filename in filenames:
            candidate = current / filename
            if path_has_skip_part(candidate, root):
                continue
            if should_process_file(candidate):
                files.append(candidate)
    files.sort()
    return files


def main() -> int:
    args = parse_args()
    root = args.root.resolve()
    client = TranslateClient()
    cache: Dict[str, str] = {}
    files = collect_files(root)
    total_translated = 0
    limit = [args.limit, 0]

    for path in files:
        translated = process_file(path, client, args.dry_run, cache, limit)
        total_translated += translated

    print(f"Processed {len(files)} files, translated {total_translated} comments.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
