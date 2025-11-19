import asyncio
import re
import datetime
import logging
import functools
from typing import List, Dict, Optional

# 如需使用 BeautifulSoup，请保留 [EN] To use BeautifulSoup, please keep
import bs4

# 如需使用 langchain 的 RecursiveCharacterTextSplitter，请保留 [EN] To use langchain's RecursiveCharacterTextSplitter, please keep
from langchain_text_splitters import RecursiveCharacterTextSplitter

# 自定义引用：如果这些类/函数确实存在，请保留正确的 import [EN] Custom references: If these classes/functions do exist, keep the correct import
# 如果不在同一目录，需要修改导入路径 [EN] If it is not in the same directory, you need to modify the import path
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
        r'\xa0+',      # 清除不间断空白字符 [EN] Clear non-breaking whitespace characters
        r'\u3000',     # 清除中文全角空格 [EN] Clear Chinese full-width spaces
        r'\t+',        # 清除制表符 [EN] Clear tabs
        r'\r+',        # 清除回车符 [EN] Clear carriage returns
        r'\n+',        # 清除连续换行符 [EN] Clear consecutive newlines
        r'<[/]?b>',    # 清除 <b> 和 </b> 标签 [EN] Clear <b> and </b> tags
        r'&gt;',       # 清除 HTML 实体字符 > [EN] Clear HTML entity characters >
        r'&lt;'        # 清除 HTML 实体字符 < [EN] Clear HTML entity characters <
    ]
    for pattern in patterns:
        text = re.sub(pattern, '', text)
    return text.strip()

###########################################################
#          主函数：async_crawl_and_parse_webpage [EN] Main function: async_crawl_and_parse_webpage
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
            "\u200b",  # 零宽空格 [EN] zero width space
            "\uff0c",  # 全角逗号 [EN] Full-width comma
            "\u3001",  # 顿号 [EN] comma
            "\uff0e",  # 全角句号 [EN] full-width period
            "\u3002",  # 句号 [EN] period
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

    # 使用 asyncio.wait_for 包裹网络爬取过程，一旦超时会抛出 asyncio.TimeoutError [EN] Use asyncio.wait_for to wrap the web crawling process, and asyncio.TimeoutError will be thrown once it times out.
    try:
        docs = await asyncio.wait_for(gather_docs(), timeout=time_out)
    except asyncio.TimeoutError:
        # 超时日志记录 [EN] Timeout logging
        logger.error(f"解析【超时】:{query[:QUERY_SHORT]} ---> 超过{time_out}秒未完成，返回空列表 ---> {url}")
        return []

    # 初始化 RecursiveCharacterTextSplitter [EN] Initialize RecursiveCharacterTextSplitter
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
#                 异步示例调用 (main) [EN] Asynchronous example call (main)
###########################################################
if __name__ == "__main__":
    async def main():
        # 几个示例网址，实际只会用最后一次赋值生效 [EN] Several example URLs will actually only take effect with the last assignment.
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

        # 设置超时时间 time_out=2 秒 [EN] Set timeout time_out=2 seconds
        docs_list = await async_crawl_and_parse_webpage(
            bing_single_item=item,
            query=query,
            sentence_size=1000,
            overlap_size=0,
            time_out=10  # 2 秒超时 [EN] 2 seconds timeout
        )
        if not docs_list:
            print("处理失败：未能获取到网页内容或处理超时，返回空列表。")
        else:
            print("处理成功：")
            for idx, doc in enumerate(docs_list, start=1):
                print(f"文档块 {idx}:")
                print(doc)
                print("=" * 40)

    # 运行异步 main [EN] Run asynchronous main
    asyncio.run(main())
