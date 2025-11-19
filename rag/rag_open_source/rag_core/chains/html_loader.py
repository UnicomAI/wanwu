from typing import List, Optional
from langchain_core.documents import Document
from langchain_community.document_loaders import TextLoader
# import readability
import html_text
import chardet

from bs4 import BeautifulSoup, Tag


def remove_styles_from_table(table):
    attrs_to_remove = ['x:num','x:str','bgcolor', 'bordercolor', 'width', 'height','align','nowrap','valign','style','class','href','_href', 'cellspacing','border', 'data-sort','cellpadding']
    # 首先处理 <table> 标签本身的样式属性 [EN] First deal with the style attributes of the <table> tag itself
    for attr in attrs_to_remove:
        if attr in table.attrs:
            del table[attr]
    """移除表格及其子元素中的样式属性"""
    for tag in table.find_all(True):  # 查找所有标签 [EN] Find all tags
        if isinstance(tag, Tag):
            if tag.name not in ['table', 'tr', 'td']:
                tag.unwrap()  # 移除非必要的标签，但保留其内容 [EN] Remove non-essential tags but keep their content
                continue

            # 移除其他样式相关属性（根据需要添加） [EN] Remove other style related properties (add as needed)
            for attr in attrs_to_remove:
                if attr in tag.attrs:
                    del tag[attr]
    return table

def get_encoding(file):
    with open(file,'rb') as f:
        tmp = chardet.detect(f.read())
        return tmp['encoding']
def extract_text_with_tables(html_content):
    # 解析 HTML 内容 [EN] Parse HTML content
    soup = BeautifulSoup(html_content, 'html.parser')

    # 存储表格及其位置 [EN] Store tables and their locations
    tables = []
    for idx, table in enumerate(soup.find_all('table')):
        # 创建一个唯一的占位符 [EN] Create a unique placeholder
        placeholder = f"[TABLE_{idx}]"
        tables.append((placeholder, str(remove_styles_from_table(table))))
        # tables.append((placeholder, str(table)))
        table.replace_with(BeautifulSoup(placeholder, 'html.parser'))

    # 提取带格式的文本 [EN] Extract formatted text
    formatted_text = html_text.extract_text(str(soup))

    # 将占位符替换回原始表格 HTML [EN] Replace placeholders back to original table HTML
    for placeholder, table_html in tables:
        formatted_text = formatted_text.replace(placeholder, table_html)

    return formatted_text
class HTMLLoader(TextLoader):
    def load(self) -> List[Document]:
        txt = ""
        try:
            with open(self.file_path, "r", encoding=self.encoding) as f:
                content = f.read()
            txt = extract_text_with_tables(content)
            # html_doc = readability.Document(txt)
            # title = html_doc.title()
            # print("title=%s" % title)
            # content = html_text.extract_text(html_doc.summary(html_partial=True))
            # txt = f'{title}\n{content}'
            # sections = txt.split("\n")

        except Exception as e:
            import traceback
            print("====> error %s" % e)
            print(traceback.format_exc())
            raise RuntimeError(f"Error loading {self.file_path}") from e

        metadata = {"source": self.file_path}
        return [Document(page_content=txt, metadata=metadata)]


if __name__ == "__main__":

    filepath = "your_file.html"
    encoding = get_encoding(filepath)
    print("encoding=%s" % encoding)
    loader = HTMLLoader(filepath, encoding=encoding, autodetect_encoding=True)
    docs = loader.load()
    for doc in docs:
        print(doc)