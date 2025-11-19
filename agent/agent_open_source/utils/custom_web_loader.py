import re
from typing import Any, Optional, Union
import chardet
from bs4 import BeautifulSoup
from langchain_community.document_loaders.web_base import WebBaseLoader

class CustomWebLoader(WebBaseLoader):
    """
    自定义的 WebLoader，继承自 WebBaseLoader，
    重写 _scrape 方法，实现支持多种编码自动选择及兜底默认解码的功能。
    """
    def _scrape(
        self,
        url: str,
        parser: Union[str, None] = None,
        bs_kwargs: Optional[dict] = None,
    ) -> Any:
        # 根据 URL 后缀选择解析器 [EN] Select parser based on URL suffix
        if parser is None:
            if url.endswith(".xml"):
                parser = "xml"
            else:
                parser = self.default_parser
        self._check_parser(parser)

        html_doc = self.session.get(url, **self.requests_kwargs)
        if self.raise_for_status:
            html_doc.raise_for_status()
        content_bytes = html_doc.content
        decoded_text = None
        replacement_threshold = 0.05  # 替换字符比例阈值 [EN] Replacement character ratio threshold

        def try_decode(content: bytes, enc: Union[str, None]) -> Union[str, None]:
            """
            尝试使用指定编码解码内容，并检查解码后出现替换字符的比例是否在可接受范围内。
            """
            try:
                candidate_text = content.decode(enc, errors="replace") if enc else content.decode(errors="replace")
            except Exception:
                return None
            total_chars = len(candidate_text) or 1
            replacement_ratio = candidate_text.count("\ufffd") / total_chars
            if replacement_ratio < replacement_threshold:
                return candidate_text
            return None

        def chinese_ratio(text: str) -> float:
            """
            计算文本中中文字符所占比例
            """
            if not text:
                return 0.0
            total = len(text)
            count = sum(1 for c in text if '\u4e00' <= c <= '\u9fff')
            return count / total

        def detect_meta_charset(content: bytes) -> Optional[str]:
            """
            从 HTML 的前部分（约2000字节）中查找 meta 标签指定的 charset 信息
            """
            try:
                snippet = content[:2000].decode("ascii", errors="ignore")
                meta_match = re.search(r'<meta[^>]+charset=["\']?([\w-]+)', snippet, re.I)
                if meta_match:
                    return meta_match.group(1)
            except Exception:
                pass
            return None

        # 0. 尝试从 meta 标签中检测编码，优先使用 meta 指定的编码解码 [EN] 0. Try to detect the encoding from the meta tag, giving priority to the encoding and decoding specified by meta.
        meta_charset = detect_meta_charset(content_bytes)
        if meta_charset:
            candidate_text = try_decode(content_bytes, meta_charset)
            if candidate_text is not None:
                decoded_text = candidate_text
                html_doc.encoding = meta_charset

        # 1. 如果用户提供了 encoding 列表，则遍历所有候选，选取中文比例最高的结果 [EN] 1. If the user provides an encoding list, traverse all candidates and select the result with the highest proportion of Chinese characters.
        if decoded_text is None and self.encoding is not None:
            best_candidate = None
            best_chinese_ratio = 0.0
            best_encoding = None
            if isinstance(self.encoding, (list, tuple)) and not isinstance(self.encoding, str):
                for enc in self.encoding:
                    candidate_text = try_decode(content_bytes, enc)
                    if candidate_text is not None:
                        ratio = chinese_ratio(candidate_text)
                        # 如果候选解码的中文比例更高，则记录该编码 [EN] If the candidate decoding has a higher proportion of Chinese, record that encoding
                        if ratio > best_chinese_ratio:
                            best_chinese_ratio = ratio
                            best_candidate = candidate_text
                            best_encoding = enc
            else:
                candidate_text = try_decode(content_bytes, self.encoding)
                if candidate_text is not None:
                    best_candidate = candidate_text
                    best_encoding = self.encoding
            if best_candidate is not None:
                decoded_text = best_candidate
                html_doc.encoding = best_encoding

        # 2. 使用 chardet 自动检测编码 [EN] 2. Use chardet to automatically detect encoding
        if decoded_text is None:
            detected_encoding = chardet.detect(content_bytes)['encoding']
            if detected_encoding:
                candidate_text = try_decode(content_bytes, detected_encoding)
                if candidate_text is not None:
                    decoded_text = candidate_text
                    html_doc.encoding = detected_encoding

        # 3. 尝试默认的 utf-8 解码（不指定编码参数） [EN] 3. Try default utf-8 decoding (no encoding parameters specified)
        if decoded_text is None:
            candidate_text = try_decode(content_bytes, None)
            if candidate_text is not None:
                decoded_text = candidate_text

        # 4. 如果以上方法均未成功，根据 autoset_encoding 配置决定是否使用 apparent_encoding [EN] 4. If none of the above methods are successful, decide whether to use apparent_encoding according to the autoset_encoding configuration.
        if decoded_text is None:
            if self.autoset_encoding:
                html_doc.encoding = html_doc.apparent_encoding
                decoded_text = html_doc.text
            else:
                raise UnicodeDecodeError(
                    "无法使用提供的编码列表或默认 utf-8 对内容进行正确解码。"
                )

        text = decoded_text if decoded_text is not None else html_doc.text
        return BeautifulSoup(text, parser, **(bs_kwargs or {}))
