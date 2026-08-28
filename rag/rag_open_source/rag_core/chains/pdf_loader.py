import json
import PyPDF2
from PIL import Image
import fitz
import pypdf
from typing import Union, List, Optional
import time
from pdfminer.high_level import extract_pages
from pdfminer.layout import LTTextContainer, LTChar, LTRect, LTFigure, LTAnno
from langchain_core.documents import Document
from langchain_community.document_loaders import TextLoader
import pdfplumber
import re

from pathlib import Path
import uuid

import sys
import os
import logging

from utils import minio_utils
from settings import IMAGE_MINIMUM_WIDTH, IMAGE_MINIMUM_HEIGHT
from utils.constant import MIN_EMBEDDING_SIZE
logger = logging.getLogger(__name__)


def has_table(page, min_rows=2, min_cols=2):
    """
    判断给定的 pdfplumber.Page 对象中是否包含实际的表格。

    参数:
        page (pdfplumber.page.Page): 由 pdfplumber 解析得到的页面对象。
        min_rows (int): 考虑为表格的最小行数，默认为2。
        min_cols (int): 考虑为表格的最小列数，默认为2。

    返回:
        bool: 如果页面包含表格则返回 True，否则返回 False。
    """
    tables = page.find_tables(table_settings={
        "vertical_strategy": "lines_strict",
        "horizontal_strategy": "lines_strict"
    })

    for table in tables:
        # 检查表格是否至少有指定数量的行和列
        if len(table.rows) >= min_rows and len(table.cells[0]) >= min_cols:
            return True

    return False
def is_chinese_char(cp):
    """Check if a character is a Chinese character."""
    return '\u4e00' <= cp <= '\u9fff'

def is_english_word(s):
    """Check if a string is an English word (alphanumeric and underscore)."""
    return bool(re.match(r'^[A-Za-z0-9_]+$', s))

def process_hyphen(text):
    # 定义正则表达式模式以匹配带有或不带有空格的连字符
    pattern = re.compile(r'\s*-\s*')

    def replace_or_remove(match):
        before, after = match.string[:match.start()].strip()[-1:], match.string[match.end():].strip()[:1]

        # 检查连字符前后的字符是否为空白字符，并跳过这些情况
        if not before or not after:
            return match.group()

        # 判断前后字符是中文还是英文单词
        before_is_chinese = is_chinese_char(before)
        after_is_chinese = is_chinese_char(after)

        if before_is_chinese and after_is_chinese:
            # 如果前后都是中文，则去掉连字符
            return ''
        elif is_english_word(match.string[:match.start()].strip().split()[-1]) and \
             is_english_word(match.string[match.end():].strip().split()[0]):
            # 如果前后都是英文单词，则用空格替换连字符
            return ' '
        else:
            # 如果前后有中英文混合，则去掉连字符
            return ''

    # 使用正则表达式查找并替换所有符合条件的连字符
    result = pattern.sub(lambda m: replace_or_remove(m), text)

    return result
def extract_info(text, start_marker):
    # 查找 start_marker 后面的内容
    start_index = text.find(start_marker)

    if start_index == -1:
        return None  # 如果未找到 start_marker，返回 None 或者可以根据需要处理

    # 截取从 start_marker 之后的内容
    extracted_text = text[start_index + len(start_marker):]

    # 去掉所有空格和标点符号
    # cleaned_text = re.sub(r'[^\w]', '', extracted_text)
    # 去掉所有类型的空白字符（包括空格、制表符、换行符等）
    cleaned_text = re.sub(r'\s+', '', extracted_text)
    return cleaned_text


class PDFLoader(TextLoader):
    def __init__(self,
                 file_path: Union[str, Path],
                 encoding: Optional[str] = None,
                 autodetect_encoding: bool = False,
                 parser_choices: List[str] = None,
                 ocr_model_id: str = ""):
        """Initialize a PDFLoader with file path and additional chunk_type."""
        super().__init__(file_path, encoding, autodetect_encoding)  # 确保调用父类的__init__
        # 如果没有提供parser_choices，则使用默认值["text"]
        if parser_choices is None:
            parser_choices = ["text"]
        self.parser_choices = parser_choices
        self.ocr_model_id = ocr_model_id
        # self.download_link = download_link
        # self.file_name = os.path.split(file_path)[-1]
        # print(f"PDFLoader initialized with file_path: {self.file_path}, download_link: {self.download_link}")  # 添加调试语句

    def tb_text_extraction(self, text_line, height_list, title_list, page_width):
        # 从行元素中提取文本
        line_text = ""
        line_size = 0
        has_min_height = False
        # line_text = element.get_text()
        line_is_title = False
        width_full_line = False
        high_height = False
        title_level = 0
        title_coverage = False
        last_height_dict = height_list[-1]
        parent_title_list = []
        title_in_range = False
        # 探析文本的格式
        # 用文本行中出现的所有格式初始化列表
        line_formats = []
        line_meta = {"line_is_title": line_is_title, "title_level": title_level, "line_text": line_text}
        # x0, y0, x1, y1 = text_line.bbox
        # if x0 >= 40 and x1 <= 553 and y0 <= 750 and y1 >= 85:
        #     title_in_range = True
        line_size = round(text_line['bottom'] - text_line['top'])
        # 标题行不充满整行
        width_full_line = True if (text_line['x1'] - text_line['x0']) >= (page_width-10) else False
        # if text_line.y1 <= (page_height - 50):
        # 遍历文本行中的每个字符

        # current_line_dicts = []  # 用于存储当前行的字符信息
        for i, character in enumerate(text_line['chars']):
            # 判断标题条件1：单元行中字符的尺寸是否包含尺寸最小的字符信息，若包含则此行不是标题
            if isinstance(character, dict) and (83 < character['y1'] < 756):
                char_size = round(character['height'])
                if str(char_size) in last_height_dict:
                    has_min_height = True
                char_dict = {
                    "text": character['text'],
                    "bbox": (character['x0'], character['y0'], character['x1'], character['y1']),
                    "font_name": character['fontname'],
                    "size": char_size
                }

                # 尝试在之前的行中找到相同文本的字符
                for prev_char_dict in reversed(line_formats):  # 从后往前遍历，减少搜索量
                    if prev_char_dict["text"] == char_dict["text"] and prev_char_dict["font_name"] == char_dict["font_name"] and prev_char_dict["size"] == char_dict["size"]:
                        # 计算x0的差值
                        x0_diff = character['x0'] - prev_char_dict["bbox"][0]
                        if x0_diff > 1 or x0_diff < -100:  # 如果差值大于1，则添加新的char_dict
                            if i > 0 and (character['x0'] - text_line['chars'][i - 1]['x1'] > 5):
                                line_text += " " + char_dict["text"]
                            else:
                                line_text += char_dict["text"]
                            line_formats.append(char_dict)
                        break  # 不需要继续搜索，因为相同的文本在同一行中只会出现一次
                else:  # 如果在line_formats中没有找到相同文本的字符
                    # 若当前字符的x0坐标比上一个字符的x1坐标差值>5则说明有空格，需要补充空格
                    if i > 0 and (character['x0']-text_line['chars'][i-1]['x1'] > 5):
                        line_text += " " + char_dict["text"]
                    else:
                        line_text += char_dict["text"]
                    line_formats.append(char_dict)
                    # 找到行中唯一的字体大小和名称
            # line_formats.extend(current_line_dicts)
        
        # format_per_line = list(set(line_formats))
        # if line_text =='有   意   留   白':
        #     line_text = ""
        if line_text:
            title_size = str(line_size)

            # 遍历列表，找到匹配的字典项
            for index, dict_item in enumerate(height_list):

                if title_size in dict_item and (index <= len(height_list)/2) and (len(height_list) > 4):
                    # 字符尺寸在前50%
                    high_height = True
                    title_level = index + 1
                    break  # 跳出循环，因为我们已经找到了匹配的项
            # else:
            #     # 如果没有找到匹配的项，打印相应的消息
            #     print(f"没有找到标题大小为{title_size}的项")
            # 返回包含每行文本及其格式的元组
            # 标题判断：判断单元行除"."以外是否还包含其他断句标点符号
            chapter_pattern = re.compile(r'[;；!?。！？\?]', re.MULTILINE)
            #  and title_in_range
            #if (not chapter_pattern.match(line_text)) and (not has_min_height) and (not width_full_line) and high_height and (' ' not in line_text):
            if (not chapter_pattern.match(line_text)) and (not has_min_height) and (not width_full_line) and high_height:
                line_is_title = True
                if "-" in line_text:
                    line_text = process_hyphen(line_text)
                # 准备添加的新dict元素
                new_element = {'height': int(title_size), 'title': line_text, 'title_level': title_level}
                # 检查height并替换或删除元素
                for idx, item in reversed(list(enumerate(title_list))):
                    if item['height'] == new_element['height']:
                        # 替换相同的height元素
                        title_list[idx] = new_element
                        title_coverage = True
                        break  # 不需要继续检查，因为height是唯一的
                    elif item['title_level'] > new_element['title_level']:
                        # 删除title_level更大的元素
                        del title_list[idx]
                        # 如果没有找到相同的height元素，则将新元素添加到列表末尾
                if new_element not in title_list:
                    title_list.append(new_element)
                    # title_list.append({"title_level": title_level, "title": line_text, "height": int(title_size)})
            parent_title_list = [item["title"] for item in title_list if int(item["height"]) > int(title_size)]
            line_meta = {"line_is_title": line_is_title, "title_level": title_level, "line_text": line_text}
            if line_is_title is True and title_level > 0:
                parent_title_str = "" if len(parent_title_list) == 0 else " " + " ".join(parent_title_list)
                line_text = '#' * title_level + parent_title_str + ' ' + line_text
        return (line_text, line_formats, parent_title_list, line_is_title, title_coverage, line_meta)


    def text_extraction(self, element, height_list, title_list, page_width):
        # 从行元素中提取文本
        line_text = ""
        line_size = 0
        has_min_height = False
        # line_text = element.get_text()
        line_is_title = False
        width_full_line = False
        high_height = False
        title_coverage = False
        title_level = 0
        last_height_dict = height_list[-1]
        parent_title_list = []
        line_meta = {"line_is_title": line_is_title, "title_level": title_level, "line_text": line_text}
        title_in_range = False
        # 探析文本的格式
        # 用文本行中出现的所有格式初始化列表
        line_formats = []
        if len(element._objs) > 1:
            objs_with_bbox = [obj for obj in element._objs if hasattr(obj, 'bbox')]
            chars = sorted(objs_with_bbox, key=lambda char: (-char.bbox[1], char.bbox[0]))  # 排序有bbox的对象

        else:
            chars = element
        for text_line in chars:
            # line_text = line_text + self.remove_repeated_substrings(text_line.get_text())
            if isinstance(text_line, LTTextContainer):
                text_line_format = []
                # x0, y0, x1, y1 = text_line.bbox
                # if x0 >= 40 and x1 <= 553 and y0 <= 750 and y1 >= 85:
                #     title_in_range = True
                line_size = round(text_line.height)
                # 标题行不充满整行
                width_full_line = True if text_line.width >= (page_width-10) else False
                # if text_line.y1 <= (page_height - 50):
                # 遍历文本行中的每个字符
                # current_line_dicts = []  # 用于存储当前行的字符信息
                last_char_x1 = None
                for character in text_line:
                    # 判断若为LTAnno对象添加原文本中的空格
                    if isinstance(character, LTAnno):
                        line_text += character.get_text()
                    # 判断标题条件1：单元行中字符的尺寸是否包含尺寸最小的字符信息，若包含则此行不是标题

                    if isinstance(character, LTChar):
                        char_size = round(character.size)
                        if str(char_size) in last_height_dict:
                            has_min_height = True
                        char_dict = {
                            "text": character.get_text(),
                            "bbox": character.bbox,
                            "font_name": character.fontname,
                            "size": char_size
                        }

                        # 尝试在之前的行中找到相同文本的字符
                        for prev_char_dict in reversed(text_line_format):  # 从后往前遍历，减少搜索量
                            # and character.bbox[1] == prev_char_dict["bbox"][1]
                            if (prev_char_dict["text"] == char_dict["text"] and
                                    prev_char_dict["font_name"] == char_dict["font_name"]
                                    and prev_char_dict["size"] == char_dict["size"]):
                                # 计算x0的差值
                                x0_diff = character.bbox[0] - prev_char_dict["bbox"][0]

                                if x0_diff > 1 or x0_diff < -100:  # 如果差值大于1，则添加新的char_dict
                                    if last_char_x1 and (character.bbox[0] - last_char_x1 > 5):
                                        line_text += " " + char_dict["text"]
                                    else:
                                        line_text += char_dict["text"]
                                    line_formats.append(char_dict)
                                    text_line_format.append(char_dict)

                                break  # 不需要继续搜索，因为相同的文本在同一行中只会出现一次

                        else:  # 如果在line_formats中没有找到相同文本的字符
                            if last_char_x1 and (character.bbox[0] - last_char_x1 > 5):
                                line_text += " " + char_dict["text"]
                            else:
                                line_text += char_dict["text"]
                            line_formats.append(char_dict)
                            text_line_format.append(char_dict)

                            # 找到行中唯一的字体大小和名称
                        last_char_x1 = character.bbox[2]
                    # line_formats.extend(current_line_dicts)

        if line_text:
            title_size = str(line_size)
            # 遍历列表，找到标题匹配的字典项
            for index, dict_item in enumerate(height_list):

                if title_size in dict_item and (index <= len(height_list)/2) and (len(height_list) > 4):
                    # 字符尺寸在前50%
                    high_height = True
                    title_level = index + 1
                    break

            chapter_pattern = re.compile(r'[;；!?。！？\?]', re.MULTILINE)
            if (not chapter_pattern.match(line_text)) and (not has_min_height) and (not width_full_line) and high_height:
                line_is_title = True
                if "-" in line_text:
                    line_text = process_hyphen(line_text)
                # 准备添加的新dict元素
                new_element = {'height': int(title_size), 'title': line_text, 'title_level': title_level}
                # 检查height并替换或删除元素
                for idx, item in reversed(list(enumerate(title_list))):
                    if item['height'] == new_element['height']:
                        # 替换相同的height元素
                        title_list[idx] = new_element
                        title_coverage = True
                        break  # 不需要继续检查，因为height是唯一的
                    elif item['title_level'] > new_element['title_level']:
                        # 删除title_level更大的元素
                        del title_list[idx]
                        # 如果没有找到相同的height元素，则将新元素添加到列表末尾
                if new_element not in title_list:
                    title_list.append(new_element)
                    # title_list.append({"title_level": title_level, "title": line_text, "height": int(title_size)})
            parent_title_list = [item["title"] for item in title_list if int(item["height"]) > int(title_size)]
            line_text = line_text[:-1] if line_text.endswith('\n') else line_text
            line_meta = {"line_is_title": line_is_title, "title_level": title_level, "line_text": line_text}
            if line_is_title is True and title_level > 0:
                parent_title_str = "" if len(parent_title_list) == 0 else " " + " ".join(parent_title_list)
                line_text = '#' * title_level + parent_title_str + ' ' + line_text
        return (line_text, line_formats, parent_title_list, line_is_title, title_coverage, line_meta)

    def remove_repeated_twice(self, text):
        # 定义正则表达式，匹配任意连续重复两次的子字符串
        def replace_func(match):
            # 获取匹配到的重复部分
            repeated_part = match.group(0)
            # 每三个字符分为一组
            length = len(repeated_part) // 2
            # 只保留第一组字符
            return repeated_part[:length]

        pattern = re.compile(r'(?!00)(.{3,})\1{1}')

        # 使用sub方法去重
        processed_text = pattern.sub(replace_func, text)

        return processed_text

    def replace_internal_newlines(self, text):
        # 使用正则表达式替换中间的 '\n' 为 ' '
        # 其中 (?<!^) 断言确保 '\n' 不是在开头位置
        # 其中 (?!$) 断言确保 '\n' 不是在结尾位置
        return re.sub(r'(?<!^)\n(?!$)', ' ', text)

    def remove_repeated_substrings(self, text):
        # 定义替换函数，用于处理重复的子字符串
        def replace_func(match):
            # 获取匹配到的重复部分
            repeated_part = match.group(0)
            # 每三个字符分为一组
            length = len(repeated_part) // 3
            # 只保留第一组字符
            return repeated_part[:length]

        # 定义正则表达式，匹配任意连续重复三次的子字符串
        pattern = re.compile(r'((?!111)(.+?))\1{2}')
        # 使用自定义的替换函数进行替换
        text = text.replace('\n', '')
        processed_text = pattern.sub(replace_func, text)
        processed_text = self.remove_repeated_twice(processed_text)
        processed_text = self.replace_internal_newlines(processed_text)
        processed_text = self.remove_repeated_uppers(processed_text)
        return processed_text

    def remove_repeated_chars(self, text):
        # 使用正则表达式将连续出现三次及以上的中文字符替换为一个中文字符
        pattern = re.compile(r'(([a-z2-9B-Z\u4e00-\u9fa5~!@#$%^&*()_+`\-={}[\]:;"\'<>,.?/|（）℃℉～]))\1{2}')
        def process_match(match):
            # 如果匹配的是三位数且两边是空格，则直接返回该三位数
            if len(match.group()) == 3 and match.group().isdigit() and (
                    match.start() == 0 or text[match.start() - 1] == ' ') and (
                    match.end() == len(text) or text[match.end()] == ' '):
                return match.group()
            elif len(match.group()) == 3 and match.group().isdigit() and (
                    match.start() == 0 or text[match.start() - 1] == '-') and (
                    match.end() == len(text) or text[match.end()] == '-'):
                return match.group()
            else:
                # 其他情况，只保留第一个字符
                return match.group()[0]

        matches = pattern.finditer(text)
        # 构建新字符串，处理每个匹配项
        processed_text_list = []
        last_end = 0
        for match in matches:
            processed_text_list.append(text[last_end:match.start()])
            processed_text_list.append(process_match(match))
            last_end = match.end()

        # 添加剩余未匹配的部分
        processed_text_list.append(text[last_end:])
        processed_text = ''.join(processed_text_list)

        return processed_text

    def remove_repeated_uppers(self, text):
        # 使用正则表达式将连续出现三次及以上的中文字符替换为一个中文字符

        pattern = re.compile(r'([a-zA-Z1-9\u4e00-\u9fa5~!@#$%^&*()_+`\-={}[\]:;"\'<>,.?/|（）℃℉～])\1{2}')
        processed_text = pattern.sub(r'\1', text)
        processed_text = processed_text.replace('\n', '')
        return processed_text

    def remove_watermark(self, content):
        """
        删除水印信息
        """
        # 删除页数
        content = re.sub(r'- \d -', '', content)
        # 清除时间水印
        content = re.sub(r'[a-z0-9A-Z]+\s((\d{4})-(\d{2})-(\d{2})\s(\d{2}):(\d{2}):(\d{2}))', '', content)
        # print(f"清除时间水印的结果为：\n{content}")
        # 清除文件编号水印
        content = re.sub(r'([\x00-\xff]{8})', '', content)
        # 清除校对人水印
        content = re.sub(r'(\d*\x00\d*)', '', content)
        #print(f"清除文件编号及水印的结果为：\n{content}")
        return content

    def contains_chinese(self, text):
        """
        判断字符串中是否包含中文字符
        """
        # 使用正则表达式判断
        pattern = re.compile('[\u4e00-\u9fa5]+')
        match = pattern.search(text)
        return True if match else False
    # 将表格转换为适当的格式

    def contains_figure(self, page_objs):
        """
        检查页面对象列表中是否存在LTFigure类型的对象。

        :param page_objs: 页面的元素列表，通常为page._objs。
        :return: 如果列表中存在LTFigure实例，则返回True，否则返回False。
        """
        return any(isinstance(obj, LTFigure) and (83 < obj.y1 < 756) for obj in page_objs)

    def table_convert_html(self, table, last_table_header):
        embedding_content = []

        table_title_list = []

        def is_empty(cell):
            return cell is None
            # return cell is None or str(cell).strip() == ''

        def get_colspan(row, col_idx, processed_cells, row_idx):
            if (row_idx, col_idx) in processed_cells:
                return 0  # 跳过已处理的单元格

            colspan = 1
            for i in range(col_idx + 1, len(row)):
                if is_empty(row[i]) and (row_idx, i) not in processed_cells:
                    processed_cells.add((row_idx, i))  # 标记已处理
                    colspan += 1
                else:
                    break
            return colspan

        def get_rowspan(rows, row_idx, col_idx, processed_cells):
            if (row_idx, col_idx) in processed_cells:
                return 0  # 跳过已处理的单元格

            rowspan = 1
            for i in range(row_idx + 1, len(rows)):
                if is_empty(rows[i][col_idx]) and (i, col_idx) not in processed_cells:
                    processed_cells.add((i, col_idx))  # 标记已处理
                    rowspan += 1
                else:
                    break
            return rowspan

        # 开始构建 HTML 表格
        html = "<table border='1'>\n"
        processed_cells = set()  # 记录已处理的单元格位置 (row_idx, col_idx)
        for row_idx, row in enumerate(table):
            html += "<tr>"
            row_text = "<tr>"
            col_idx = 0
            while col_idx < len(row):
                if (row_idx, col_idx) in processed_cells:
                    col_idx += 1  # 跳过已处理的单元格，避免死循环
                    continue
                cell = row[col_idx]

                if is_empty(cell):
                    col_idx += 1
                    continue  # 跳过空单元格

                # 获取跨列数
                colspan = get_colspan(row, col_idx,processed_cells,row_idx)

                # 获取跨行数
                rowspan = get_rowspan(table, row_idx, col_idx,processed_cells)

                # 添加单元格
                if colspan > 1 or rowspan > 1:
                    html += f"<td colspan='{colspan}' rowspan='{rowspan}'>{cell}</td>"
                    row_text += f"<td colspan='{colspan}' rowspan='{rowspan}'>{cell}</td>"
                else:
                    html += f"<td>{cell}</td>"
                    row_text += f"<td>{cell}</td>"
                # 把除首行外的前两列的内容所谓embedding索引
                # if row_idx > 0 and col_idx < 2 and str(cell).strip() != '':
                #     # 对于表格中没有行分割线的表：通过 换行符切割索引
                #     if len(table) <= 2 and "\n" in str(cell):
                #         cell_list = str(cell).split("\n")
                #         embedding_content.extend(cell_list)
                #     else:
                #         embedding_content.append(cell)
                # 标记已经处理过的单元格
                for i in range(rowspan):
                    for j in range(colspan):
                        processed_cells.add((row_idx + i, col_idx + j))
                # 跳过已经处理过的跨列单元格
                col_idx += colspan

            html += "</tr>\n"
            row_text += "</tr>"
            # if row_idx > 0:
            #     embedding_content.append(row_text)

        html += "</table>\n"
        # print(json.dumps(embedding_content,ensure_ascii=False))
        return html, embedding_content, table_title_list

    def table_converter(self, table, last_table_header):
        """
        将表格转换为适当的格式
        """
        table_string = ''
        embedding_content = []
        fault_flag = False
        table_title_list = []
        # 处理第一行
        first_cleaned_row = [self.remove_repeated_uppers(cell) if cell else '' for cell in table[0]]
        # 如果去掉全部空格后不全是数字且包含中文描述，则说明第一行为表头
        if not first_cleaned_row[0].replace(" ", "").isdigit() and self.contains_chinese(first_cleaned_row[0]):
            for item in first_cleaned_row:
                table_title_list.append(item)
            table_string += ('|' + '|'.join(first_cleaned_row) + '|' + '\n')
            # 根据 first_cleaned_row 的长度动态生成 Markdown 表头信息
            separator_row = '| ' + ' | '.join(['---'] * len(first_cleaned_row)) + ' |'
            table_string += separator_row + '\n'
        elif len(first_cleaned_row) == len(last_table_header):
            # 无表头且列数一样，则带上一个表的表头
            table_title_list = last_table_header
            table_string += ('|' + '|'.join(table_title_list) + '|' + '\n')
            # 根据 table_title_list 的长度动态生成 Markdown 表头信息
            separator_row = '| ' + ' | '.join(['---'] * len(table_title_list)) + ' |'
            table_string += separator_row + '\n'
            # 说明第一行不是表头，则拼接embeding
            new_row_text = ""
            for i in range(len(first_cleaned_row)):
                first_cleaned_row[i] = self.remove_repeated_chars(first_cleaned_row[i].replace('\n', ' ')) if first_cleaned_row[i] else ' '
                if first_cleaned_row[i].strip() != "" and len(table_title_list) > i:
                    new_cell = "%s%s" % (table_title_list[i], first_cleaned_row[i])
                    new_row_text = new_row_text + new_cell
            if new_row_text and len(new_row_text.strip()) >= MIN_EMBEDDING_SIZE:
                embedding_content.append(new_row_text)
        else:
            # 第一行非表头且不能复用上一个表头,正常拼接行
            table_string += ('|' + '|'.join(first_cleaned_row) + '|' + '\n')

        #print(table_string)
        # 处理剩余行
        for row in table[1:]:
            new_row_text = ""
            for i in range(len(row)):
                row[i] = self.remove_repeated_chars(row[i].replace('\n', ' ')) if row[i] else ' '
                if row[i].strip() != "" and len(table_title_list) > i:
                    new_cell = "%s%s" % (table_title_list[i], row[i])
                    new_row_text = new_row_text + new_cell
            if new_row_text and len(new_row_text.strip()) >= MIN_EMBEDDING_SIZE:
                embedding_content.append(new_row_text)

            table_string += ('|' + '|'.join(row) + '|' + '\n')

        return table_string, embedding_content, table_title_list

    def is_text_garbled(self, text, threshold):
        # 判断文本是否有大量不可显示的字符或乱码
        text_garbled = False
        text = text.strip().replace('\n', '')
        if len(text) == 0:
            return True
        non_displayable_char_ratio = len(re.findall(r'[^\x20-\x7E\u4e00-\u9fff]', text)) / len(text)
        is_garbled_text_ratio = len(
            re.findall(r"[，。！？：；“”‘’（）《》【】\-_\s!\"#$%&'()01]", text)) / len(text)

        if non_displayable_char_ratio >= threshold or is_garbled_text_ratio >= threshold:
            text_garbled = True
        return text_garbled
    def is_garbled(self, text):
        if len(text) == 0:
            return True
        # 检查文本中是否包含乱码
        # 这里使用正则表达式来检测非 ASCII 字符和连续的不可打印字符
        if re.search(r'[^\x00-\x7F]', text) or re.search(r'[\x00-\x1F\x7F]{3,}', text):
            return True
        return False
    def get_chunk_type(self):
        # text = ""
        chunk_type = 1
        pdfFileObj = open(self.file_path, 'rb')
        pdf = pdfplumber.open(self.file_path)
        lines_with_height = []
        height_groups = {}
        height_list = []
        text = ""
        has_table = False
        has_image = False
        try:
            pdfReaded = pypdf.PdfReader(pdfFileObj)

            # 获取PDF总页数
            total_pages = len(pdfReaded.pages)
            logger.info(f"PDF总页数: {total_pages}")
            for pagenum, page in enumerate(extract_pages(self.file_path)):
                # 初始化从页面中提取文本所需的变量
                pageObj = pdfReaded.pages[pagenum]
                table_page = pdf.pages[pagenum]
                tables = table_page.find_tables()
                if len(tables) > 0:
                    has_table = True
                images = table_page.images
                if len(images) > 0:
                    has_image = True
                page_content = pageObj.extract_text()
                page_content = page_content.strip().replace('\u3000', '')
                if page_content:
                    text += page_content

                page_elements = []
                page_elements = [(element.y1, element) for element in page._objs]

                # 对页面中出现的所有元素进行排序
                page_elements.sort(key=lambda a: a[0], reverse=True)

                # 查找组成页面的元素
                for i, component in enumerate(page_elements):
                    # 提取PDF中元素顶部的位置
                    pos = component[0]
                    # 提取页面布局的元素
                    element = component[1]

                    # 检查该元素是否为文本元素
                    if isinstance(element, LTTextContainer):
                        for text_line in element:
                            if isinstance(text_line, LTTextContainer):
                                lines_with_height.append((round(text_line.height), len(text_line.get_text().replace("\n", ''))))
                if pagenum > 10:
                    break
            for size, count in lines_with_height:
                if size not in height_groups:
                    height_groups[size] = 0
                height_groups[size] += count
            sorted_height_dict = sorted(height_groups.items(), key=lambda x: x[1])
            data = [{str(size): count} for size, count in sorted_height_dict]
            # 按键从大到小排序，同时保持键为字符串类型
            height_list = sorted(data, key=lambda x: int(next(iter(x.keys()))), reverse=True)
            if len(height_list) >= 2 and len(text) > 0 and (has_table or has_image):
                chunk_type = 2
            # 移除 chunk_type=3 分支：text 模式提取不出文字的 PDF，返回默认 chunk_type=1
            # 不再根据 "ocr" in parser_choices 判断

        except Exception as e:
            raise RuntimeError(f"Error loading {self.file_path}") from e
        finally:
            pdfFileObj.close()
            pdf.close()
        return (chunk_type, height_list)


    # 创建一个从pdf中裁剪图像元素的函数
    def crop_image(self, element, pageObj, directory, file_name):
        # 获取从PDF中裁剪图像的坐标
        [image_left, image_top, image_right, image_bottom] = [element.x0, element.y0, element.x1, element.y1]
        # 使用坐标(left, bottom, right, top)裁剪页面
        pageObj.mediabox.lower_left = (image_left, image_bottom)
        pageObj.mediabox.upper_right = (image_right, image_top)
        # 将裁剪后的页面保存为新的PDF
        cropped_pdf_writer = PyPDF2.PdfWriter()
        cropped_pdf_writer.add_page(pageObj)
        # 将裁剪好的PDF保存到一个新文件
        image_fullname = f"{file_name}.pdf"
        # 组合成新的文件路径
        image_filepath = os.path.join(directory, image_fullname)

        with open(image_filepath, 'wb') as cropped_pdf_file:
            cropped_pdf_writer.write(cropped_pdf_file)
        return image_filepath

        # 创建一个将PDF内容转换为image的函数
    def convert_to_images(self, input_file, directory, file_name):
        dpi = 200
        with fitz.open(input_file) as doc:
            page = doc[0]
            # 使用之前裁剪的区域作为裁剪矩形
            clip_rect = fitz.Rect(page.rect)
            print(page.mediabox)
            mat = fitz.Matrix(dpi / 72, dpi / 72)
            scale_factor = 1.0  # 放大倍数
            mat = mat.prescale(scale_factor, scale_factor)  # 放大图像
            # 如果宽度或高度 > 3000像素，不放大图像
            if clip_rect.width > 3000 or clip_rect.height > 3000:
                mat = fitz.Matrix(1, 1)
            # mat = fitz.Matrix(1, 1)

            pm = page.get_pixmap(matrix=mat, clip=clip_rect, alpha=False)
            image = Image.frombytes("RGB", (pm.width, pm.height), pm.samples)
            output_file = f"{file_name}.png"
            image_filepath = os.path.join(directory, output_file)
            image.save(image_filepath, "PNG")
            # 过滤掉小图标文件
            width, height = image.size
            if width < IMAGE_MINIMUM_WIDTH or height < IMAGE_MINIMUM_HEIGHT:  # 设置你认为的小图标大小阈值
                return None  # 返回None表示此图片应被忽略
        return image_filepath

    def process_pdf_content(self, content_list, image_dict, image_labels):
        output = {'text': '', 'embedding_chunks': []}

        # 创建一个set，用于存储已经插入的图片URL，避免重复
        # inserted_urls = set()

        # 直接使用 image_dict 的键值对进行查找
        for item in content_list:
            item = item.strip()  # 去除每个元素的前后空白符

            # 如果该项是 URL，并且已存在于 image_dict 中
            if item in image_dict:
                title = image_dict[item]
                # markdown_image = f"![{title}]({item} \"{title}\")"
                markdown_image = f"![{title}]({item})"
                output['text'] += markdown_image + ' '  # 插入Markdown图片
                # 使用正则表达式匹配标题中的关键部分
                # 忽略前面的 "图" 和括号中的 "(共X张 第X张)" 或换行符
                # 通过处理 \n 和嵌套括号
                match = re.search(r'图\s*\d+\s+(.+?)(?:\s*\(\s*.*?\s*\))?$', re.sub(r'\s+', ' ', title))
                if match:
                    # 提取关键部分，例如 "空调系统-控制板"
                    title = match.group(1).strip()
                if title not in output['embedding_chunks']:
                    output['embedding_chunks'].append(title)  # 将标题加入 embedding_chunks
                    output['embedding_chunks'].append("%s图" % title)

            # 其他内容，保留原样
            else:
                if item not in image_dict.values():
                    output['text'] += item + ' '

        # 移除多余的空格和换行符
        output['text'] = output['text'].strip()
        # 追加图片中识别的文字标签
        if len(image_labels) > 0:
            output['embedding_chunks'].append("\n".join(image_labels))
            for item in image_labels:
                if len(item.strip()) >= 2:
                    output['embedding_chunks'].append(item)
        return output



    def load_and_split_doc(self, height_list) -> List[dict]:
        text = ""
        pdfFileObj = open(self.file_path, 'rb')
        pdf = pdfplumber.open(self.file_path)
        chunks = []
        page_chunks = []
        # 获取文件所在的目录路径
        directory = os.path.dirname(self.file_path)
        path_obj = Path(self.file_path)

        file_name = path_obj.stem
        try:
            logger.info('---------文字版PDF自适应解析策略按页解析切分---------')
            pdfReaded = PyPDF2.PdfReader(pdfFileObj)
            # height_list = [{'30': 17}, {'16': 244}, {'14': 760}, {'12': 1024}, {'11': 30676}, {'10': 43921}]
            title_list = []
            last_parent_title = []
            last_table_header = []
            page_title = ""
            page_title_dict = {}
            last_page_title = ""
            last_page_embed = ""
            for pagenum, page in enumerate(extract_pages(self.file_path)):
                # 初始化从页面中提取文本所需的变量
                pageObj = pdfReaded.pages[pagenum]
                page_text = ""
                page_embed_list = []
                line_format = []
                text_from_images = []
                text_from_tables = []
                chunk = {}
                page_content = []
                page_embedding_chunks = []
                # 初始化检查表的数量
                table_num = 0
                upper_side = 0
                lower_side = 0
                first_element = True
                table_extraction_flag = False
                # 打开pdf文件

                # 查找已检查的页面
                table_page = pdf.pages[pagenum]
                chunk_content = []
                content_position = []
                # page_position = []
                start_position = {}
                end_position = {}
                # 找出本页上的表格数目
                # if self.contains_figure(page._objs):
                #     continue
                try:
                    tables = table_page.find_tables()
                    table_text = ""
                    for extract_table in table_page.extract_tables():
                        for item_table in extract_table:
                            for item in item_table:
                                if item:
                                    item_content = item.strip().replace('\u3000', '')
                                    if item_content:
                                        table_text += item_content
                    # page_elements2 = [(element.y1, element) for element in page._objs]
                    ###################################################################################
                    # begin
                    if has_table(table_page) and table_text and not self.contains_figure(page._objs):
                        ts = {
                            "vertical_strategy": "lines",
                            "horizontal_strategy": "lines",
                        }
                        # Get the bounding boxes of the tables on the page.
                        # bboxes = [table.bbox for table in table_page.find_tables(table_settings=ts)]
                        bboxes = []

                        last_table_bottom = 0

                        for table in table_page.find_tables(table_settings=ts):

                            ab_parent_title_list = []
                            bt_parent_title_list = []
                            table_parent_title_list = []
                            last_line_meta = {}
                            bx0, by0, bx1, by1 = table.bbox

                            bboxes.append(table.bbox)
                            def not_within_bboxes(obj):

                                """Check if the object is in any of the table's bbox."""
                                def obj_above_bbox(_bbox):
                                    """See https://github.com/jsvine/pdfplumber/blob/stable/pdfplumber/table.py#L404"""
                                    v_mid = (obj["top"] + obj["bottom"]) / 2
                                    h_mid = (obj["x0"] + obj["x1"]) / 2
                                    x0, top, x1, bottom = _bbox
                                    return (h_mid >= x0) and (v_mid >= top)
                                    #return (h_mid >=x0) and (v_mid > x_top) and (v_mid < bottom)
                                def obj_between_bbox(_bbox):
                                    """See https://github.com/jsvine/pdfplumber/blob/stable/pdfplumber/table.py#L404"""
                                    v_mid = (obj["top"] + obj["bottom"]) / 2
                                    # h_mid = (obj["x0"] + obj["x1"]) / 2
                                    x0, top, x1, bottom = _bbox
                                    return (v_mid < last_table_bottom) or (v_mid >= top)

                                # v_mid = (obj["top"] + obj["bottom"]) / 2
                                # h_mid = (obj["x0"] + obj["x1"]) / 2
                                if table_num == 0:
                                    return not obj_above_bbox(table.bbox)
                                else:
                                    #(v_mid < by1) and (v_mid > x_top)
                                    return not obj_between_bbox(table.bbox)

                            # above_text = ""
                            # above_text = table_page.filter(not_within_bboxes).extract_text()
                            above_tb_lines = table_page.filter(not_within_bboxes).extract_text_lines()
                            for above_tb_line in above_tb_lines:
                                # print(f"Text line;{above_tb_line['text']}")
                                if above_tb_line["bottom"] > by0 or above_tb_line["bottom"] < 83:
                                    continue
                                (ab_line_text, ab_format_per_line, ab_parent_title_list, ab_is_title, ab_title_coverage, current_line_meta) = (
                                    self.tb_text_extraction(above_tb_line, height_list, title_list, page.width))
                                if ab_line_text:

                                    if ab_is_title:
                                        if last_line_meta:
                                            if last_line_meta["line_is_title"] is True and last_line_meta["title_level"] == \
                                                    current_line_meta["title_level"]:
                                                prev_content = last_line_meta["line_text"].strip()
                                                # 去掉当前行的开头 '#' 和空格
                                                current_content = current_line_meta["line_text"].lstrip('#').strip()
                                                # 合并两行的内容
                                                ab_line_text = f"{prev_content} {current_content}\n"
                                                ab_line_text = '#' * current_line_meta["title_level"] + ' ' + ab_line_text
                                                current_line_meta["line_text"] = ab_line_text
                                                if len(title_list) > 0:
                                                    if "title" in title_list[-1]:
                                                        title_list[-1]["title"] = ab_line_text.strip()
                                                if page_content:  # 确保列表不为空
                                                    page_content.pop()
                                            else:
                                                ab_line_text = ab_line_text + "\n"
                                        else:
                                            ab_line_text = ab_line_text + "\n"
                                        last_page_title = ab_line_text
                                        # current_title = ab_line_text.lstrip('#').strip()
                                        if current_line_meta["title_level"] == 2:
                                            page_title = current_line_meta["line_text"].strip()

                                            last_page_embed = current_line_meta["line_text"]
                                            page_embedding_chunks.append(current_line_meta["line_text"])
                                        elif current_line_meta["title_level"] in [3, 4]:
                                            leaf_title = "%s%s" % (page_title, current_line_meta["line_text"])
                                            page_embedding_chunks.append(leaf_title)
                                            if page_title in page_title_dict:
                                                leaf_title = "%s%s" % (page_title_dict[page_title], current_line_meta["line_text"])
                                                page_embedding_chunks.append(leaf_title)

                                            last_page_embed = leaf_title
                                        else:
                                            last_page_embed = current_line_meta["line_text"]
                                    else:
                                        if len(page_content) == 0 and last_page_title:
                                            page_content.append(last_page_title)

                                        ab_line_text = ab_line_text + "\n"
                                        if last_page_embed:
                                            if last_page_embed not in page_embedding_chunks:
                                                page_embedding_chunks.append(last_page_embed)
                                        if page_title:
                                            if page_title not in page_embedding_chunks:
                                                page_embedding_chunks.append(page_title)

                                    if ab_format_per_line and "bbox" in ab_format_per_line:
                                        if not start_position:
                                            start_position = ab_format_per_line["bbox"]
                                        end_position = ab_format_per_line["bbox"]

                                    page_content.append(ab_line_text)
                                    chunk_content.append(ab_line_text)
                                    content_position.append(ab_format_per_line)
                                    # page_position.append(ab_format_per_line)
                                    # 将最新标题层级缓存至变量中
                                    if ab_parent_title_list:
                                        last_parent_title = ab_parent_title_list
                                    last_line_meta = current_line_meta

                            t_table = table_page.extract_tables()[table_num]
                            first_row = t_table[0] if t_table else []
                            empty_cells = sum(1 for cell in first_row if cell is None or cell == '')
                            if empty_cells < 20:
                                table_string, embedding_chunks, current_table_header = self.table_convert_html(t_table, last_table_header)
                            else:
                                table_string, embedding_chunks, current_table_header = self.table_converter(t_table, last_table_header)

                            if len(page_content) == 0 and last_page_title:
                                page_content.append(last_page_title)
                            page_content.append(table_string)
                            page_embedding_chunks.extend(embedding_chunks)
                            if last_page_embed:
                                page_embedding_chunks.append(last_page_embed)
                            last_line_meta = {}
                            # 注意：表格的不写入chunk_content

                            if not start_position:
                                start_position = {"x0": bx0, "x1": bx1, "y0": by0, "y1": by1}
                            end_position = {"x0": bx0, "x1": bx1, "y0": by0, "y1": by1}

                            last_table_header = current_table_header
                            def bottom_within_bboxes(obj):
                                """Check if the object is in any of the table's bbox."""

                                def bottom_above_bbox(_bbox):
                                    """See https://github.com/jsvine/pdfplumber/blob/stable/pdfplumber/table.py#L404"""
                                    v_mid = (obj["top"] + obj["bottom"]) / 2
                                    # v_mid = (_bbox[1] + _bbox[3]) / 2
                                    h_mid = (obj["x0"] + obj["x1"]) / 2
                                    x0, top, x1, bottom = _bbox
                                    return (h_mid < x1) and (v_mid < bottom)

                                return not bottom_above_bbox(table.bbox)

                            # 表格底部文本的抽取限定在最后一个表格才抽取
                            if table_num == len(tables) - 1:

                                bottom_tb_lines = table_page.filter(bottom_within_bboxes).extract_text_lines()
                                for bottom_tb_line in bottom_tb_lines:
                                    # print(f"Text line;{bottom_tb_line['text']}")
                                    if bottom_tb_line["bottom"] > 756 or bottom_tb_line["bottom"] < 83:
                                        continue
                                    (bt_line_text, bt_format_per_line, bt_parent_title_list, bt_is_title, bt_title_coverage, current_line_meta) = (self.tb_text_extraction(bottom_tb_line, height_list, title_list, page.width))
                                    if bt_line_text:

                                        if bt_is_title:
                                            if last_line_meta:
                                                if last_line_meta["line_is_title"] is True and last_line_meta["title_level"] ==  current_line_meta["title_level"]:
                                                    prev_content = last_line_meta["line_text"].strip()
                                                    # 去掉当前行的开头 '#' 和空格
                                                    current_content = current_line_meta["line_text"].lstrip('#').strip()
                                                    # 合并两行的内容
                                                    bt_line_text = f"{prev_content} {current_content}\n"
                                                    bt_line_text = '#' * current_line_meta["title_level"] + ' ' + bt_line_text
                                                    current_line_meta["line_text"] = bt_line_text
                                                    if len(title_list) > 0:
                                                        if "title" in title_list[-1]:
                                                            title_list[-1] = bt_line_text.strip()
                                                    if page_content:  # 确保列表不为空
                                                        page_content.pop()
                                                else:
                                                    bt_line_text = bt_line_text + "\n"
                                            else:
                                                bt_line_text = bt_line_text + "\n"
                                            last_page_title = bt_line_text
                                            # current_title = bt_line_text.lstrip('#').strip()
                                            if current_line_meta["title_level"] == 2:
                                                page_title = current_line_meta["line_text"].strip()
                                                last_page_embed = current_line_meta["line_text"]
                                                page_embedding_chunks.append(current_line_meta["line_text"])
                                            elif current_line_meta["title_level"] in [3, 4]:
                                                leaf_title = "%s%s" % (page_title, current_line_meta["line_text"])
                                                page_embedding_chunks.append(leaf_title)
                                                if page_title in page_title_dict:
                                                    leaf_title = "%s%s" % (
                                                    page_title_dict[page_title], current_line_meta["line_text"])
                                                    page_embedding_chunks.append(leaf_title)
                                                last_page_embed = leaf_title
                                            else:
                                                last_page_embed = current_line_meta["line_text"]
                                        else:
                                            if len(page_content) == 0 and last_page_title:
                                                page_content.append(last_page_title)

                                            bt_line_text = bt_line_text + "\n"
                                            if last_page_embed:
                                                if last_page_embed not in page_embedding_chunks:
                                                    page_embedding_chunks.append(last_page_embed)
                                            if page_title:
                                                if page_title not in page_embedding_chunks:
                                                    page_embedding_chunks.append(page_title)

                                        page_content.append(bt_line_text)
                                        if bt_format_per_line and "bbox" in bt_format_per_line:
                                            if not start_position:
                                                start_position = bt_format_per_line["bbox"]
                                            end_position = bt_format_per_line["bbox"]
                                        chunk_content.append(bt_line_text)
                                        content_position.append(bt_format_per_line)
                                        # page_position.append(bt_format_per_line)
                                        # 将最新标题层级缓存至变量中
                                        if bt_parent_title_list:
                                            last_parent_title = bt_parent_title_list
                            if bt_parent_title_list:
                                table_parent_title_list = bt_parent_title_list
                            elif ab_parent_title_list:
                                table_parent_title_list = ab_parent_title_list
                            else:
                                table_parent_title_list = last_parent_title

                            table_num = table_num + 1
                            # last_table_bottom = by1
                            last_table_bottom = by1

                    else:

                        last_line_meta = {}
                        page_elements = [(element.y1, element) for element in page._objs]

                        page_elements.sort(key=lambda a: a[0], reverse=True)
                        image_dict = {}
                        image_labels = []
                        last_image_url = ""
                        # 查找组成页面的元素
                        for i, component in enumerate(page_elements):
                            # 提取PDF中元素顶部的位置
                            pos = component[0]
                            # 排除页眉和页脚的内容
                            if pos > 765 or pos < 83:
                                continue
                            # 提取页面布局的元素
                            element = component[1]

                            # 检查该元素是否为文本元素
                            if isinstance(element, LTTextContainer):
                                (line_text, format_per_line, parent_title_list, is_title, title_coverage, current_line_meta) = self.text_extraction(element, height_list, title_list, page.width)
                                # 将每行的文本追加到页文本
                                if line_text:

                                    if is_title:
                                        if last_line_meta:
                                            if last_line_meta["line_is_title"] is True and last_line_meta["title_level"] == \
                                                    current_line_meta["title_level"]:
                                                prev_content = last_line_meta["line_text"].strip()
                                                # 去掉当前行的开头 '#' 和空格
                                                current_content = current_line_meta["line_text"].lstrip('#').strip()
                                                # 合并两行的内容
                                                line_text = f"{prev_content} {current_content}\n"
                                                line_text = '#' * current_line_meta["title_level"] + ' ' + line_text
                                                current_line_meta["line_text"] = line_text
                                                if len(title_list) > 0:
                                                    if "title" in title_list[-1]:
                                                        title_list[-1] = line_text.strip()
                                                if page_content:  # 确保列表不为空
                                                    page_content.pop()
                                            else:
                                                line_text = line_text + "\n"
                                        else:
                                            line_text = line_text + "\n"
                                        last_page_title = line_text
                                        # current_title = line_text.lstrip('#').strip()
                                        if current_line_meta["title_level"] == 2:
                                            page_title = current_line_meta["line_text"].strip()
                                            last_page_embed = current_line_meta["line_text"]
                                            page_embedding_chunks.append(current_line_meta["line_text"])
                                        elif current_line_meta["title_level"] in [3, 4]:
                                            leaf_title = "%s%s" % (page_title, current_line_meta["line_text"])
                                            page_embedding_chunks.append(leaf_title)
                                            if page_title in page_title_dict:
                                                leaf_title = "%s%s" % (page_title_dict[page_title], current_line_meta["line_text"])
                                                page_embedding_chunks.append(leaf_title)
                                            last_page_embed = leaf_title
                                        else:
                                            last_page_embed = current_line_meta["line_text"]
                                    else:
                                        if len(page_content) == 0 and last_page_title:
                                            page_content.append(last_page_title)

                                        # line_text = line_text + "\n"
                                        if last_page_embed:
                                            if last_page_embed not in page_embedding_chunks:
                                                page_embedding_chunks.append(last_page_embed)
                                        if page_title:
                                            if page_title not in page_embedding_chunks:
                                                page_embedding_chunks.append(page_title)

                                    if len(format_per_line) > 0:
                                        start_x0, start_y0, start_x1, start_y1 = format_per_line[0]['bbox']
                                        if not start_position:
                                            start_position = {"x0": start_x0, "x1": start_x1, "y0": start_y0, "y1": start_y1}
                                        end_position = {"x0": start_x0, "x1": start_x1, "y0": start_y0, "y1": start_y1}
                                        if "图" in line_text and start_x0 >= 80 and last_image_url != "":
                                            image_dict[last_image_url] = line_text
                                            last_image_url = ""

                                    # page_content.append(line_text)
                                    chunk_content.append(line_text)
                                    content_position.append(format_per_line)
                                    # page_position.append(format_per_line)

                                    # 将最新标题层级缓存至变量中
                                    if parent_title_list:
                                        last_parent_title = parent_title_list
                            elif isinstance(element, LTFigure):
                                # 从PDF中裁剪图像
                                logger.info("-------图片解析--------")
                                unique_id = str(uuid.uuid4())
                                image_file_name = "%s_%s" % (file_name, unique_id)
                                image_file_path = self.crop_image(element, pageObj, directory, image_file_name)
                                logger.info("------>image_file_path=%s" % image_file_path)
                                # text 模式下 LTFigure 只做图片裁剪→转换→上传 MinIO，不做 OCR 提取文字

                                # 将裁剪后的pdf转换为图像
                                image_url = self.convert_to_images(image_file_path, directory, image_file_name)
                                if image_url is not None:
                                    minio_result = minio_utils.upload_local_file(image_url)
                                    if minio_result['code'] == 0:
                                        image_download_link = minio_result['download_link']
                                        last_image_url = image_download_link
                                        image_line_text = image_download_link
                                        # page_content.append(image_line_text)
                                        chunk_content.append(f"![image]({image_line_text})")
                                        image_line_formats = []
                                        image_format = {
                                            "text": image_download_link,
                                            "bbox": (element.x0, element.y0, element.x1, element.y1),
                                            "font_name": '',
                                            "size": 0
                                        }
                                        image_line_formats.append(image_format)
                                        content_position.append(image_line_formats)
                                        # page_position.append(image_line_formats)
                                        if not start_position:
                                            start_position = {"x0": element.x0, "x1": element.x1, "y0": element.y0, "y1": element.y1}
                                        end_position = {"x0": element.x0, "x1": element.x1, "y0": element.y0, "y1": element.y1}

                        if len(chunk_content) == 0:
                            chunk_content = table_page.extract_text()
                        if len(chunk_content) > 0:

                            join_text = " ".join(chunk_content)
                            # if "https" in join_text and "图" in join_text:
                            #     chunk = self.process_pdf_content(chunk_content, image_dict, image_labels)
                            # else:
                            #     chunk = {"text": join_text, "embedding_chunks": []}
                            chunk = {"text": join_text, "embedding_chunks": []}
                            page_content.append(chunk["text"])
                            if chunk["embedding_chunks"]:
                                page_embedding_chunks.extend(chunk["embedding_chunks"])

                    current_page_content = "".join(page_content)
                    current_page_content_len = len(current_page_content)
                    if current_page_content:
                        page_chunk = {}
                        page_chunk["text"] = current_page_content
                        page_chunk["page_num"] = [pagenum + 1]
                        page_chunk["file_path"] = self.file_path
                        page_chunk["type"] = "text"
                        page_chunk["embedding_chunks"] = list(dict.fromkeys(page_embedding_chunks))
                        page_chunk["start_position"] = start_position
                        page_chunk["end_position"] = end_position
                        page_chunks.append(page_chunk)
                        text += current_page_content + "\n"
                    # text 模式下页面为空则不追加，不做 OCR 回退
                    logger.info("------>page_num=%s,content_len=%s" % (pagenum + 1, current_page_content_len))
                except Exception as error:
                    import traceback
                    logger.error("------> page_num:%s, error: %s" % (pagenum + 1, error))
                    logger.error(traceback.format_exc())
                    continue

        except Exception as e:
            import traceback
            logger.error("------> pdf load_and_split error %s" % e)
            logger.error(traceback.format_exc())
            # raise RuntimeError(f"Error loading {self.file_path}") from e
        finally:
            pdfFileObj.close()
            pdf.close()

        return page_chunks

    def load(self) -> List[Document]:

        text = ""
        chunks = []

        pdf = pdfplumber.open(self.file_path)

        try:
            logger.info('---------文字版PDF默认解析逻辑load()---------')

            last_parent_title = []
            last_table_header = []
            page_empty_count = 0
            for pagenum, page in enumerate(extract_pages(self.file_path)):
                chunk = {}
                page_content = []
                # 初始化检查表的数量
                table_num = 0

                plumber_page = pdf.pages[pagenum]
                chunk_content = []
                try:
                    # 找出本页上的表格数目
                    tables = plumber_page.find_tables()
                    table_text = ""
                    for extract_table in plumber_page.extract_tables():
                        for item_table in extract_table:
                            for item in item_table:
                                if item:
                                    item_content = item.strip().replace('\u3000', '')
                                    if item_content:
                                        table_text += item_content
                    if tables and table_text:
                        ts = {
                            "vertical_strategy": "lines",
                            "horizontal_strategy": "lines",
                        }
                        # Get the bounding boxes of the tables on the page.
                        # bboxes = [table.bbox for table in table_page.find_tables(table_settings=ts)]
                        bboxes = []
                        for table in plumber_page.find_tables(table_settings=ts):

                            bx0, by0, bx1, by1 = table.bbox

                            bboxes.append(table.bbox)
                            t_table = plumber_page.extract_tables()[table_num]
                            if t_table:
                                def not_within_bboxes(obj):
                                    """Check if the object is in any of the table's bbox."""

                                    def obj_above_bbox(_bbox):
                                        """See https://github.com/jsvine/pdfplumber/blob/stable/pdfplumber/table.py#L404"""
                                        v_mid = (obj["top"] + obj["bottom"]) / 2
                                        h_mid = (obj["x0"] + obj["x1"]) / 2
                                        x0, top, x1, bottom = _bbox
                                        return (h_mid >= x0) and (v_mid >= top)

                                    return not obj_above_bbox(table.bbox)


                                above_text = plumber_page.filter(not_within_bboxes).extract_text()
                                above_lines = above_text.split('\n')
                                above_line = ""
                                for line in above_lines:
                                    above_line += self.remove_repeated_substrings(line) + "\n"
                                page_content.append(above_line)

                                table_string, embedding_chunks, current_table_header = self.table_converter(t_table, last_table_header)
                                page_content.append(table_string)

                                last_table_header = current_table_header
                                table_num = table_num + 1

                                def bottom_within_bboxes(obj):
                                    """Check if the object is in any of the table's bbox."""

                                    def bottom_above_bbox(_bbox):
                                        """See https://github.com/jsvine/pdfplumber/blob/stable/pdfplumber/table.py#L404"""
                                        v_mid = (obj["top"] + obj["bottom"]) / 2
                                        h_mid = (obj["x0"] + obj["x1"]) / 2
                                        x0, top, x1, bottom = _bbox
                                        return (h_mid < x1) and (v_mid < bottom)

                                    return not bottom_above_bbox(table.bbox)

                                # print(table_page.filter(bottom_within_bboxes).extract_text())
                                # page_content.append(table_page.filter(bottom_within_bboxes).extract_text())
                                bottom_text = plumber_page.filter(bottom_within_bboxes).extract_text()
                                bottom_lines = bottom_text.split('\n')
                                bottom_line = ""
                                for line in bottom_lines:
                                    bottom_line += self.remove_repeated_substrings(line) + "\n"
                                page_content.append(bottom_line)
                    else:
                        clean_text = plumber_page.extract_text()
                        # logger.info("---->clean_text=%s" % clean_text)
                        # print("---->clean_text=%s" % clean_text)
                        page_content.append(clean_text)

                    current_page_content = "".join(page_content)
                    current_page_content_len = len(current_page_content)
                    if current_page_content:
                        text += current_page_content + "\n"
                    else:
                        page_empty_count += 1
                    logger.info("------>page_num=%s,content_len=%s" % (pagenum + 1, current_page_content_len))
                except Exception as error:
                    import traceback
                    logger.error("------> page_num:%s, error: %s" % (pagenum + 1, error))
                    logger.error(traceback.format_exc())
                    continue

        except Exception as e:
            import traceback
            logger.error("------> pdf load() error %s" % e)
            logger.error(traceback.format_exc())
        finally:
            pdf.close()

        # text 模式下提取不出文字就是空结果，不再回退到 OCR
        metadata = {"source": self.file_path}
        return [Document(page_content=text, metadata=metadata)]



if __name__ == "__main__":

    filepath = "your_file.pdf"
    loader = PDFLoader(filepath)

    height_list = [{'50': 26}, {'39': 63}, {'28': 9}]
    # docs = loader.load()
    docs = loader.load_and_split_doc(height_list)

    for doc in docs:
        # print(doc.page_content)
        print(json.dumps(doc,ensure_ascii=False))
    #         processed_file.write(doc.page_content)
