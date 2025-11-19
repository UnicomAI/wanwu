import asyncio
import re
import datetime
import logging
import functools
from typing import List, Dict, Optional

# To use BeautifulSoup, please keep
import bs4

# To use langchain's RecursiveCharacterTextSplitter, please keep
from langchain_text_splitters import RecursiveCharacterTextSplitter

# Custom references: If these classes/functions do exist, keep the correct import
# If it is not in the same directory, you need to modify the import path
from utils.custom_web_loader import CustomWebLoader
from utils.timing import advanced_timing_decorator

from utils.uni_id import generate_unique_id

logger = logging.getLogger(__name__)


QUERY_SHORT = 15

def clean_text(text: str) -> str:
    """
    清除文本中的特殊字符、多余空白以及部分 HTML 标签。
    """
    patterns = [
        r'\xa0+',      # Clear non-breaking whitespace characters
        r'\u3000',     # Clear Chinese full-width spaces
        r'\t+',        # Clear tabs
        r'\r+',        # Clear carriage returns
        r'\n+',        # Clear consecutive newlines
        r'<[/]?b>',    # Clear <b> and </b> tags
        r'&gt;',       # Clear HTML entity characters >
        r'&lt;'        # Clear HTML entity characters <
    ]
    for pattern in patterns:
        text = re.sub(pattern, '', text)
    return text.strip()

###########################################################
#          Main function: async_crawl_and_parse_webpage
###########################################################
# @advanced_timing_decorator()
async def async_crawl_and_parse_webpage(
    bing_single_item: Dict,
    query: str = "",
    sentence_size: int = 600,
    overlap_size: Optional[int] = 20,
    separators: Optional[List[str]] = None,
    time_out: Optional[float] = None
) -> List[Dict[str, str]]:
    """
    异步根据给定的 URL 爬取网页内容，并使用 RecursiveCharacterTextSplitter 拆分文本，返回拆分后的文档列表。
    如果超过 time_out（秒）依然未完成，则返回空列表。
    """
    if separators is None:
        separators = [
            "\n\n",
            "\n",
            " ",
            ",",
            "\u200b",  # zero width space
            "\uff0c",  # Full-width comma
            "\u3001",  # comma
            "\uff0e",  # full-width period
            "\u3002",  # period
            ".",
            "",
        ]

    url = bing_single_item["link"]
    loader = CustomWebLoader(
        web_path=url,
        requests_kwargs={"headers": {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"}},
        requests_per_second=5000,
        raise_for_status=False
    )

    async def gather_docs() -> List:
        """从异步生成器中把文档全部收集到列表里。"""
        _docs = []
        async for doc in loader.alazy_load():
            _docs.append(doc)
        return _docs

    start_time = datetime.datetime.now()

    # Use asyncio.wait_for to wrap the web crawling process, and asyncio.TimeoutError will be thrown once it times out.
    try:
        docs = await asyncio.wait_for(gather_docs(), timeout=time_out)
    except asyncio.TimeoutError:
        # Timeout logging
        logger.error(f"解析【超时】:{query[:QUERY_SHORT]} ---> 超过{time_out}秒未完成，返回空列表 ---> {url}")
        return []

    # Initialize RecursiveCharacterTextSplitter
    text_splitter = RecursiveCharacterTextSplitter(
        chunk_size=sentence_size,
        chunk_overlap=overlap_size,
        length_function=len,
        is_separator_regex=False,
        separators=separators,
    )

    split_docs = text_splitter.split_documents(docs)

    def convert_documents(docs) -> List[Dict[str, str]]:
        results = []
        for doc in docs:
            title = bing_single_item.get("title", "")
            snippet = clean_text(doc.page_content)
            link = doc.metadata.get("source", "")
            results.append({
                "type": "SE",
                "id": generate_unique_id(),
                "title": title,
                "snippet": snippet,
                "link": link,
                "datePublished": "",
                "dateLastCrawled": "",
            })
        return results

    results = convert_documents(split_docs)
    elapsed_time = (datetime.datetime.now() - start_time).total_seconds()
    print(f"解析【正常】:{query[:QUERY_SHORT]} ---> 耗时: {elapsed_time}s ---> 个数：{len(results)} ---> {url}")
    logger.info(f"解析【正常】:{query[:QUERY_SHORT]} ---> 耗时: {elapsed_time}s ---> 个数：{len(results)} ---> {url}")

    return results

###########################################################
#                 Asynchronous example call (main)
###########################################################
if __name__ == "__main__":
    async def main():
        # Several example URLs will actually only take effect with the last assignment.
        url = "https://xueqiu.com/S/00700?md5__1038=n4%2BxRD9DuiDQKxx0x0HwbDyADgYkbDclpr0hoD"
        # url = "https://www.weather.com.cn/weathern/101010100.shtml"
        # url = "https://news.qq.com/rain/a/20250206A09CH400"
        # url = "https://forecast.weather.com.cn/town/weather1dn/101010100.shtml"
        # url = "https://news.qq.com/rain/a/20250209A01SDJ00"
        # url = " http://www.baidu.com/link?url=k3loy5W-scFez9wYcMrV2nsuCe81Jaf6XvtEhQAU-lErDV7Us3TdJPt0t7FIXQRx"
        # url = "http://www.nmc.cn/publish/forecast/ABJ/beijing.html"
        # url = "https://xueqiu.com/S/00700?xueqiu_status_id=309683094&xueqiu_status_from_source=utl&xueqiu_status_source=statusdetail&xueqiu_private_from_source=0105"
        # url =  "https://news.gmw.cn/2025-02/17/content_37853418.htm"


        item = {"link": url, "title": "123"}
        query = "今天北京天气"

        # Set timeout time_out=2 seconds
        docs_list = await async_crawl_and_parse_webpage(
            bing_single_item=item,
            query=query,
            sentence_size=1000,
            overlap_size=0,
            time_out=10  # 2 seconds timeout
        )
        if not docs_list:
            print("处理失败：未能获取到网页内容或处理超时，返回空列表。")
        else:
            print("处理成功：")
            for idx, doc in enumerate(docs_list, start=1):
                print(f"文档块 {idx}:")
                print(doc)
                print("=" * 40)

    # Run asynchronous main
    asyncio.run(main())
