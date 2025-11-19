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
import nltk
current_file_path = os.path.abspath(__file__)
# Get the directory where the current file is located
current_dir = os.path.dirname(current_file_path)
# Splice the path to the nltk_data folder
nltk_data_path = os.path.join(current_dir, 'nltk_data')
nltk.data.path.append(nltk_data_path)

from utils import minio_utils
from utils import ocr_utils
from logging_config import setup_logging
#
logger_name='rag_pdf_loader'
app_name = os.getenv("LOG_FILE")
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))




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
        # Checks whether the table has at least the specified number of rows and columns
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
    # Define regular expression pattern to match hyphens with or without spaces
    pattern = re.compile(r'\s*-\s*')

    def replace_or_remove(match):
        before, after = match.string[:match.start()].strip()[-1:], match.string[match.end():].strip()[:1]

        # Check if the characters before and after the hyphen are whitespace characters and skip these cases
        if not before or not after:
            return match.group()

        # Determine whether the preceding and following characters are Chinese or English words
        before_is_chinese = is_chinese_char(before)
        after_is_chinese = is_chinese_char(after)

        if before_is_chinese and after_is_chinese:
            # If both characters are in Chinese, remove the hyphen.
            return ''
        elif is_english_word(match.string[:match.start()].strip().split()[-1]) and \
             is_english_word(match.string[match.end():].strip().split()[0]):
            # If there are English words before and after, replace the hyphen with a space
            return ' '
        else:
            # If there is a mixture of Chinese and English before and after, remove the hyphen.
            return ''

    # Find and replace all matching hyphens using regular expressions
    result = pattern.sub(lambda m: replace_or_remove(m), text)

    return result
def extract_info(text, start_marker):
    # Find the content after start_marker
    start_index = text.find(start_marker)

    if start_index == -1:
        return None  # If start_marker is not found, returns None or can be processed as needed

    # Intercept content after start_marker
    extracted_text = text[start_index + len(start_marker):]

    # Remove all spaces and punctuation
    # cleaned_text = re.sub(r'[^\w]', '', extracted_text)
    # Remove all types of whitespace characters (including spaces, tabs, newlines, etc.)
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
        super().__init__(file_path, encoding, autodetect_encoding)  # Make sure to call the parent class's __init__
        # If parser_choices is not provided, the default value ["text"] is used
        if parser_choices is None:
            parser_choices = ["text"]
        self.parser_choices = parser_choices
        self.ocr_model_id = ocr_model_id
        # self.download_link = download_link
        # self.file_name = os.path.split(file_path)[-1]
        # print(f"PDFLoader initialized with file_path: {self.file_path}, download_link: {self.download_link}") #Add debugging statements

    def tb_text_extraction(self, text_line, height_list, title_list, page_width):
        # Extract text from row element
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
        # Exploring the format of text
        # Initialize a list with all formats present in a text line
        line_formats = []
        line_meta = {"line_is_title": line_is_title, "title_level": title_level, "line_text": line_text}
        # x0, y0, x1, y1 = text_line.bbox
        # if x0 >= 40 and x1 <= 553 and y0 <= 750 and y1 >= 85:
        #     title_in_range = True
        line_size = round(text_line['bottom'] - text_line['top'])
        # Title row does not fill the entire row
        width_full_line = True if (text_line['x1'] - text_line['x0']) >= (page_width-10) else False
        # if text_line.y1 <= (page_height - 50):
        # Iterate over each character in a line of text

        # current_line_dicts = [] # Used to store character information of the current line
        for i, character in enumerate(text_line['chars']):
            # Determine title condition 1: Whether the size of the characters in the unit row contains the smallest character information. If it does, the row is not a title.
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

                # Try to find the characters of the same text in the previous line
                for prev_char_dict in reversed(line_formats):  # Traverse from back to front to reduce search volume
                    if prev_char_dict["text"] == char_dict["text"] and prev_char_dict["font_name"] == char_dict["font_name"] and prev_char_dict["size"] == char_dict["size"]:
                        # Calculate the difference of x0
                        x0_diff = character['x0'] - prev_char_dict["bbox"][0]
                        if x0_diff > 1 or x0_diff < -100:  # If the difference is greater than 1, add a new char_dict
                            if i > 0 and (character['x0'] - text_line['chars'][i - 1]['x1'] > 5):
                                line_text += " " + char_dict["text"]
                            else:
                                line_text += char_dict["text"]
                            line_formats.append(char_dict)
                        break  # No need to continue searching because the same text will only appear once in the same line
                else:  # If no characters for the same text are found in line_formats
                    # If the difference between the x0 coordinate of the current character and the x1 coordinate of the previous character is >5, it means there are spaces and spaces need to be added.
                    if i > 0 and (character['x0']-text_line['chars'][i-1]['x1'] > 5):
                        line_text += " " + char_dict["text"]
                    else:
                        line_text += char_dict["text"]
                    line_formats.append(char_dict)
                    # Find the unique font size and name in a row
            # line_formats.extend(current_line_dicts)
        
        # format_per_line = list(set(line_formats))
        # if line_text =='Intentionally left blank':
        #     line_text = ""
        if line_text:
            title_size = str(line_size)

            # Iterate through the list and find matching dictionary items
            for index, dict_item in enumerate(height_list):

                if title_size in dict_item and (index <= len(height_list)/2) and (len(height_list) > 4):
                    # Character size is in the top 50%
                    high_height = True
                    title_level = index + 1
                    break  # Break out of the loop because we have found a match
            # else:
            #     # If no matching item is found, print the appropriate message
            #     print(f"No item with title size {title_size} found")
            # Returns a tuple containing each line of text and its format
            # Title judgment: Determine whether the unit line contains other sentence-breaking punctuation marks besides "."
            chapter_pattern = re.compile(r'[;；!?。！？\?]', re.MULTILINE)
            #  and title_in_range
            #if (not chapter_pattern.match(line_text)) and (not has_min_height) and (not width_full_line) and high_height and (' ' not in line_text):
            if (not chapter_pattern.match(line_text)) and (not has_min_height) and (not width_full_line) and high_height:
                line_is_title = True
                if "-" in line_text:
                    line_text = process_hyphen(line_text)
                # New dict elements ready to be added
                new_element = {'height': int(title_size), 'title': line_text, 'title_level': title_level}
                # Check height and replace or remove elements
                for idx, item in reversed(list(enumerate(title_list))):
                    if item['height'] == new_element['height']:
                        # Replace the same height element
                        title_list[idx] = new_element
                        title_coverage = True
                        break  # No need to continue checking because height is unique
                    elif item['title_level'] > new_element['title_level']:
                        # Remove elements with larger title_level
                        del title_list[idx]
                        # If the same height element is not found, add the new element to the end of the list
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
        # Extract text from row element
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
        # Exploring the format of text
        # Initialize a list with all formats present in a text line
        line_formats = []
        if len(element._objs) > 1:
            objs_with_bbox = [obj for obj in element._objs if hasattr(obj, 'bbox')]
            chars = sorted(objs_with_bbox, key=lambda char: (-char.bbox[1], char.bbox[0]))  # Sort objects with bbox

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
                # Title row does not fill the entire row
                width_full_line = True if text_line.width >= (page_width-10) else False
                # if text_line.y1 <= (page_height - 50):
                # Iterate over each character in a line of text
                # current_line_dicts = [] # Used to store character information of the current line
                last_char_x1 = None
                for character in text_line:
                    # Determine if spaces in the original text are added to the LTAno object
                    if isinstance(character, LTAnno):
                        line_text += character.get_text()
                    # Determine title condition 1: Whether the size of the characters in the unit row contains the smallest character information. If it does, the row is not a title.

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

                        # Try to find the characters of the same text in the previous line
                        for prev_char_dict in reversed(text_line_format):  # Traverse from back to front to reduce search volume
                            # and character.bbox[1] == prev_char_dict["bbox"][1]
                            if (prev_char_dict["text"] == char_dict["text"] and
                                    prev_char_dict["font_name"] == char_dict["font_name"]
                                    and prev_char_dict["size"] == char_dict["size"]):
                                # Calculate the difference of x0
                                x0_diff = character.bbox[0] - prev_char_dict["bbox"][0]

                                if x0_diff > 1 or x0_diff < -100:  # If the difference is greater than 1, add a new char_dict
                                    if last_char_x1 and (character.bbox[0] - last_char_x1 > 5):
                                        line_text += " " + char_dict["text"]
                                    else:
                                        line_text += char_dict["text"]
                                    line_formats.append(char_dict)
                                    text_line_format.append(char_dict)

                                break  # No need to continue searching because the same text will only appear once in the same line

                        else:  # If no characters for the same text are found in line_formats
                            if last_char_x1 and (character.bbox[0] - last_char_x1 > 5):
                                line_text += " " + char_dict["text"]
                            else:
                                line_text += char_dict["text"]
                            line_formats.append(char_dict)
                            text_line_format.append(char_dict)

                            # Find the unique font size and name in a row
                        last_char_x1 = character.bbox[2]
                    # line_formats.extend(current_line_dicts)

        if line_text:
            title_size = str(line_size)
            # Iterate through the list and find the dictionary item whose title matches
            for index, dict_item in enumerate(height_list):

                if title_size in dict_item and (index <= len(height_list)/2) and (len(height_list) > 4):
                    # Character size is in the top 50%
                    high_height = True
                    title_level = index + 1
                    break

            chapter_pattern = re.compile(r'[;；!?。！？\?]', re.MULTILINE)
            if (not chapter_pattern.match(line_text)) and (not has_min_height) and (not width_full_line) and high_height:
                line_is_title = True
                if "-" in line_text:
                    line_text = process_hyphen(line_text)
                # New dict elements ready to be added
                new_element = {'height': int(title_size), 'title': line_text, 'title_level': title_level}
                # Check height and replace or remove elements
                for idx, item in reversed(list(enumerate(title_list))):
                    if item['height'] == new_element['height']:
                        # Replace the same height element
                        title_list[idx] = new_element
                        title_coverage = True
                        break  # No need to continue checking because height is unique
                    elif item['title_level'] > new_element['title_level']:
                        # Remove elements with larger title_level
                        del title_list[idx]
                        # If the same height element is not found, add the new element to the end of the list
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
        # Define a regular expression to match any substring repeated twice in a row
        def replace_func(match):
            # Get the matched duplicates
            repeated_part = match.group(0)
            # Group every three characters
            length = len(repeated_part) // 2
            # Keep only the first set of characters
            return repeated_part[:length]

        pattern = re.compile(r'(?!00)(.{3,})\1{1}')

        # Use sub method to remove duplicates
        processed_text = pattern.sub(replace_func, text)

        return processed_text

    def replace_internal_newlines(self, text):
        # Use a regular expression to replace the '\n' in the middle with ' '
        # where the (?<!^) assertion ensures that '\n' is not at the beginning
        # The (?!$) assertion ensures that '\n' is not at the end
        return re.sub(r'(?<!^)\n(?!$)', ' ', text)

    def remove_repeated_substrings(self, text):
        # Define a replacement function to handle repeated substrings
        def replace_func(match):
            # Get the matched duplicates
            repeated_part = match.group(0)
            # Group every three characters
            length = len(repeated_part) // 3
            # Keep only the first set of characters
            return repeated_part[:length]

        # Define a regular expression to match any substring repeated three times in a row
        pattern = re.compile(r'((?!111)(.+?))\1{2}')
        # Use custom replacement function to replace
        text = text.replace('\n', '')
        processed_text = pattern.sub(replace_func, text)
        processed_text = self.remove_repeated_twice(processed_text)
        processed_text = self.replace_internal_newlines(processed_text)
        processed_text = self.remove_repeated_uppers(processed_text)
        return processed_text

    def remove_repeated_chars(self, text):
        # Use regular expressions to replace Chinese characters that appear three times or more in a row with one Chinese character
        pattern = re.compile(r'(([a-z2-9B-Z\u4e00-\u9fa5~!@#$%^&*()_+`\-={}[\]:;"\'<>,.?/|（）℃℉～]))\1{2}')
        def process_match(match):
            # If the match is a three-digit number and there are spaces on both sides, the three-digit number is returned directly.
            if len(match.group()) == 3 and match.group().isdigit() and (
                    match.start() == 0 or text[match.start() - 1] == ' ') and (
                    match.end() == len(text) or text[match.end()] == ' '):
                return match.group()
            elif len(match.group()) == 3 and match.group().isdigit() and (
                    match.start() == 0 or text[match.start() - 1] == '-') and (
                    match.end() == len(text) or text[match.end()] == '-'):
                return match.group()
            else:
                # Otherwise, only the first character is retained
                return match.group()[0]

        matches = pattern.finditer(text)
        # Build a new string, processing each match
        processed_text_list = []
        last_end = 0
        for match in matches:
            processed_text_list.append(text[last_end:match.start()])
            processed_text_list.append(process_match(match))
            last_end = match.end()

        # Add remaining unmatched parts
        processed_text_list.append(text[last_end:])
        processed_text = ''.join(processed_text_list)

        return processed_text

    def remove_repeated_uppers(self, text):
        # Use regular expressions to replace Chinese characters that appear three times or more in a row with one Chinese character

        pattern = re.compile(r'([a-zA-Z1-9\u4e00-\u9fa5~!@#$%^&*()_+`\-={}[\]:;"\'<>,.?/|（）℃℉～])\1{2}')
        processed_text = pattern.sub(r'\1', text)
        processed_text = processed_text.replace('\n', '')
        return processed_text

    def remove_watermark(self, content):
        """
        删除水印信息
        """
        # delete pages
        content = re.sub(r'- \d -', '', content)
        # Clear time watermark
        content = re.sub(r'[a-z0-9A-Z]+\s((\d{4})-(\d{2})-(\d{2})\s(\d{2}):(\d{2}):(\d{2}))', '', content)
        # print(f"The result of clearing the time watermark is:\n{content}")
        # Clear file number watermark
        content = re.sub(r'([\x00-\xff]{8})', '', content)
        # Clear proofreader watermark
        content = re.sub(r'(\d*\x00\d*)', '', content)
        #print(f"The result of clearing the file number and watermark is:\n{content}")
        return content

    def contains_chinese(self, text):
        """
        判断字符串中是否包含中文字符
        """
        # Use regular expressions to determine
        pattern = re.compile('[\u4e00-\u9fa5]+')
        match = pattern.search(text)
        return True if match else False
    # Convert table to appropriate format

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
                return 0  # Skip processed cells

            colspan = 1
            for i in range(col_idx + 1, len(row)):
                if is_empty(row[i]) and (row_idx, i) not in processed_cells:
                    processed_cells.add((row_idx, i))  # Flag processed
                    colspan += 1
                else:
                    break
            return colspan

        def get_rowspan(rows, row_idx, col_idx, processed_cells):
            if (row_idx, col_idx) in processed_cells:
                return 0  # Skip processed cells

            rowspan = 1
            for i in range(row_idx + 1, len(rows)):
                if is_empty(rows[i][col_idx]) and (i, col_idx) not in processed_cells:
                    processed_cells.add((i, col_idx))  # Flag processed
                    rowspan += 1
                else:
                    break
            return rowspan

        # Start building an HTML table
        html = "<table border='1'>\n"
        processed_cells = set()  # Record the processed cell position (row_idx, col_idx)
        for row_idx, row in enumerate(table):
            html += "<tr>"
            row_text = "<tr>"
            col_idx = 0
            while col_idx < len(row):
                if (row_idx, col_idx) in processed_cells:
                    col_idx += 1  # Skip processed cells to avoid infinite loops
                    continue
                cell = row[col_idx]

                if is_empty(cell):
                    col_idx += 1
                    continue  # Skip empty cells

                # Get the number of spanned columns
                colspan = get_colspan(row, col_idx,processed_cells,row_idx)

                # Get the number of spanned rows
                rowspan = get_rowspan(table, row_idx, col_idx,processed_cells)

                # Add cell
                if colspan > 1 or rowspan > 1:
                    html += f"<td colspan='{colspan}' rowspan='{rowspan}'>{cell}</td>"
                    row_text += f"<td colspan='{colspan}' rowspan='{rowspan}'>{cell}</td>"
                else:
                    html += f"<td>{cell}</td>"
                    row_text += f"<td>{cell}</td>"
                # The contents of the first two columns except the first row are called embedding indexes
                if row_idx > 0 and col_idx < 2 and str(cell).strip() != '':
                    # For tables without row separators: cut index by newline character
                    if len(table) <= 2 and "\n" in str(cell):
                        cell_list = str(cell).split("\n")
                        embedding_content.extend(cell_list)
                    else:
                        embedding_content.append(cell)
                # Mark cells that have been processed
                for i in range(rowspan):
                    for j in range(colspan):
                        processed_cells.add((row_idx + i, col_idx + j))
                # Skip already processed cross-column cells
                col_idx += colspan

            html += "</tr>\n"
            row_text += "</tr>"
            if row_idx > 0:
                embedding_content.append(row_text)

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
        # Process the first line
        first_cleaned_row = [self.remove_repeated_uppers(cell) if cell else '' for cell in table[0]]
        # If all the spaces are removed and they are not all numbers and contain Chinese descriptions, then the first line is the header.
        if not first_cleaned_row[0].replace(" ", "").isdigit() and self.contains_chinese(first_cleaned_row[0]):
            for item in first_cleaned_row:
                table_title_list.append(item)
            table_string += ('|' + '|'.join(first_cleaned_row) + '|' + '\n')
            # Dynamically generate Markdown header information based on the length of first_cleaned_row
            separator_row = '| ' + ' | '.join(['---'] * len(first_cleaned_row)) + ' |'
            table_string += separator_row + '\n'
        elif len(first_cleaned_row) == len(last_table_header):
            # If there is no header and the number of columns is the same, bring the header of a table.
            table_title_list = last_table_header
            table_string += ('|' + '|'.join(table_title_list) + '|' + '\n')
            # Dynamically generate Markdown header information based on the length of table_title_list
            separator_row = '| ' + ' | '.join(['---'] * len(table_title_list)) + ' |'
            table_string += separator_row + '\n'
            # If the first line is not the header, then embedding will be spliced.
            new_row_text = ""
            for i in range(len(first_cleaned_row)):
                first_cleaned_row[i] = self.remove_repeated_chars(first_cleaned_row[i].replace('\n', ' ')) if first_cleaned_row[i] else ' '
                if first_cleaned_row[i].strip() != "" and len(table_title_list) > i:
                    new_cell = "%s%s" % (table_title_list[i], first_cleaned_row[i])
                    new_row_text = new_row_text + new_cell
            if new_row_text:
                embedding_content.append(new_row_text)
        else:
            # The first row is not a header and the previous header cannot be reused. Rows are spliced ​​normally.
            table_string += ('|' + '|'.join(first_cleaned_row) + '|' + '\n')

        #print(table_string)
        # Process remaining rows
        for row in table[1:]:
            new_row_text = ""
            for i in range(len(row)):
                row[i] = self.remove_repeated_chars(row[i].replace('\n', ' ')) if row[i] else ' '
                if row[i].strip() != "" and len(table_title_list) > i:
                    new_cell = "%s%s" % (table_title_list[i], row[i])
                    new_row_text = new_row_text + new_cell
            if new_row_text:
                embedding_content.append(new_row_text)

            table_string += ('|' + '|'.join(row) + '|' + '\n')

        return table_string, embedding_content, table_title_list

    def is_text_garbled(self, text, threshold):
        # Determine whether the text contains a large number of undisplayable characters or garbled characters
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
        # Check whether the text contains garbled characters
        # Regular expressions are used here to detect non-ASCII characters and consecutive unprintable characters
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

            # Get the total number of PDF pages
            total_pages = len(pdfReaded.pages)
            logger.info(f"PDF总页数: {total_pages}")
            for pagenum, page in enumerate(extract_pages(self.file_path)):
                # Initialize the variables needed to extract text from the page
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

                # Sort all elements that appear on the page
                page_elements.sort(key=lambda a: a[0], reverse=True)

                # Find the elements that make up a page
                for i, component in enumerate(page_elements):
                    # Extract the position of the top of an element in a PDF
                    pos = component[0]
                    # Extract elements of page layout
                    element = component[1]

                    # Check if the element is a text element
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
            # Sort keys from largest to smallest while keeping keys as string types
            height_list = sorted(data, key=lambda x: int(next(iter(x.keys()))), reverse=True)
            if len(height_list) >= 2 and len(text) > 0 and (has_table or has_image):
                chunk_type = 2
            elif (text == '') and ('ocr' in self.parser_choices):
                chunk_type = 3

        except Exception as e:
            raise RuntimeError(f"Error loading {self.file_path}") from e
        finally:
            pdfFileObj.close()
            pdf.close()
        return (chunk_type, height_list)


    # Create a function that crops image elements from pdf
    def crop_image(self, element, pageObj, directory, file_name):
        # Get coordinates of cropped image from PDF
        [image_left, image_top, image_right, image_bottom] = [element.x0, element.y0, element.x1, element.y1]
        # Crop the page using coordinates (left, bottom, right, top)
        pageObj.mediabox.lower_left = (image_left, image_bottom)
        pageObj.mediabox.upper_right = (image_right, image_top)
        # Save cropped pages as new PDF
        cropped_pdf_writer = PyPDF2.PdfWriter()
        cropped_pdf_writer.add_page(pageObj)
        # Save the cropped PDF to a new file
        image_fullname = f"{file_name}.pdf"
        # Combined into a new file path
        image_filepath = os.path.join(directory, image_fullname)

        with open(image_filepath, 'wb') as cropped_pdf_file:
            cropped_pdf_writer.write(cropped_pdf_file)
        return image_filepath

        # Create a function that converts PDF content to image
    def convert_to_images(self, input_file, directory, file_name):
        dpi = 200
        with fitz.open(input_file) as doc:
            page = doc[0]
            # Use the previously cropped area as the cropping rectangle
            clip_rect = fitz.Rect(page.rect)
            print(page.mediabox)
            mat = fitz.Matrix(dpi / 72, dpi / 72)
            scale_factor = 1.0  # magnification
            mat = mat.prescale(scale_factor, scale_factor)  # Enlarge image
            # Don't enlarge image if width or height > 3000 pixels
            if clip_rect.width > 3000 or clip_rect.height > 3000:
                mat = fitz.Matrix(1, 1)
            # mat = fitz.Matrix(1, 1)

            pm = page.get_pixmap(matrix=mat, clip=clip_rect, alpha=False)
            image = Image.frombytes("RGB", (pm.width, pm.height), pm.samples)
            output_file = f"{file_name}.png"
            image_filepath = os.path.join(directory, output_file)
            image.save(image_filepath, "PNG")
        return image_filepath

    def process_pdf_content(self, content_list, image_dict, image_labels):
        output = {'text': '', 'embedding_chunks': []}

        # Create a set to store the inserted image URL to avoid duplication
        # inserted_urls = set()

        # Directly use the key-value pairs of image_dict to search
        for item in content_list:
            item = item.strip()  # Remove leading and trailing whitespace from each element

            # If the item is a URL and already exists in image_dict
            if item in image_dict:
                title = image_dict[item]
                # markdown_image = f"![{title}]({item} \"{title}\")"
                markdown_image = f"![{title}]({item})"
                output['text'] += markdown_image + ' '  # Insert Markdown picture
                # Use regular expressions to match key parts in titles
                # Ignore the preceding "picture" and the "(X pictures in total)" or line breaks in brackets
                # By handling \n and nested brackets
                match = re.search(r'图\s*\d+\s+(.+?)(?:\s*\(\s*.*?\s*\))?$', re.sub(r'\s+', ' ', title))
                if match:
                    # Extract key sections, such as "Air conditioning system - control panel"
                    title = match.group(1).strip()
                if title not in output['embedding_chunks']:
                    output['embedding_chunks'].append(title)  # Add title to embedding_chunks
                    output['embedding_chunks'].append("%s图" % title)

            # Leave other content as is
            else:
                if item not in image_dict.values():
                    output['text'] += item + ' '

        # Remove extra spaces and newlines
        output['text'] = output['text'].strip()
        # Append text tags recognized in images
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
        # Get the directory path where the file is located
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
                # Initialize the variables needed to extract text from the page
                pageObj = pdfReaded.pages[pagenum]
                page_text = ""
                page_embed_list = []
                line_format = []
                text_from_images = []
                text_from_tables = []
                chunk = {}
                page_content = []
                page_embedding_chunks = []
                # Number of initialization checklists
                table_num = 0
                upper_side = 0
                lower_side = 0
                first_element = True
                table_extraction_flag = False
                # open pdf file

                # Find checked pages
                table_page = pdf.pages[pagenum]
                chunk_content = []
                content_position = []
                # page_position = []
                start_position = {}
                end_position = {}
                # Find the number of tables on this page
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
                                                # Remove the '#' and spaces at the beginning of the current line
                                                current_content = current_line_meta["line_text"].lstrip('#').strip()
                                                # Merge the contents of two lines
                                                ab_line_text = f"{prev_content} {current_content}\n"
                                                ab_line_text = '#' * current_line_meta["title_level"] + ' ' + ab_line_text
                                                current_line_meta["line_text"] = ab_line_text
                                                if len(title_list) > 0:
                                                    if "title" in title_list[-1]:
                                                        title_list[-1]["title"] = ab_line_text.strip()
                                                if page_content:  # Make sure the list is not empty
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
                                    # Cache the latest title level into a variable
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
                            # Note: table does not write chunk_content

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

                            # The extraction of text at the bottom of the table is limited to the last table.
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
                                                    # Remove the '#' and spaces at the beginning of the current line
                                                    current_content = current_line_meta["line_text"].lstrip('#').strip()
                                                    # Merge the contents of two lines
                                                    bt_line_text = f"{prev_content} {current_content}\n"
                                                    bt_line_text = '#' * current_line_meta["title_level"] + ' ' + bt_line_text
                                                    current_line_meta["line_text"] = bt_line_text
                                                    if len(title_list) > 0:
                                                        if "title" in title_list[-1]:
                                                            title_list[-1] = bt_line_text.strip()
                                                    if page_content:  # Make sure the list is not empty
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
                                        # Cache the latest title level into a variable
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
                        # Find the elements that make up a page
                        for i, component in enumerate(page_elements):
                            # Extract the position of the top of an element in a PDF
                            pos = component[0]
                            # Exclude header and footer content
                            if pos > 765 or pos < 83:
                                continue
                            # Extract elements of page layout
                            element = component[1]

                            # Check if the element is a text element
                            if isinstance(element, LTTextContainer):
                                (line_text, format_per_line, parent_title_list, is_title, title_coverage, current_line_meta) = self.text_extraction(element, height_list, title_list, page.width)
                                # Append the text of each line to the page text
                                if line_text:

                                    if is_title:
                                        if last_line_meta:
                                            if last_line_meta["line_is_title"] is True and last_line_meta["title_level"] == \
                                                    current_line_meta["title_level"]:
                                                prev_content = last_line_meta["line_text"].strip()
                                                # Remove the '#' and spaces at the beginning of the current line
                                                current_content = current_line_meta["line_text"].lstrip('#').strip()
                                                # Merge the contents of two lines
                                                line_text = f"{prev_content} {current_content}\n"
                                                line_text = '#' * current_line_meta["title_level"] + ' ' + line_text
                                                current_line_meta["line_text"] = line_text
                                                if len(title_list) > 0:
                                                    if "title" in title_list[-1]:
                                                        title_list[-1] = line_text.strip()
                                                if page_content:  # Make sure the list is not empty
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

                                    # Cache the latest title level into a variable
                                    if parent_title_list:
                                        last_parent_title = parent_title_list
                            elif isinstance(element, LTFigure):
                                # Crop image from PDF
                                logger.info("-------图片解析--------")
                                unique_id = str(uuid.uuid4())
                                image_file_name = "%s_%s" % (file_name, unique_id)
                                image_file_path = self.crop_image(element, pageObj, directory, image_file_name)
                                logger.info("------>image_file_path=%s" % image_file_path)
                                if "ocr" in self.parser_choices:
                                    ocr_parser_data = ocr_utils.ocr_parser_native(image_file_path, self.ocr_model_id)
                                    logger.info(json.dumps(ocr_parser_data, ensure_ascii=False))
                                    if 'data' in ocr_parser_data:
                                        for item in ocr_parser_data["data"]:
                                            if item["type"] == "figure":
                                                image_extract_text = item["text"]
                                                if image_extract_text:
                                                    image_labels.extend(image_extract_text.split("\n"))

                                # Convert cropped pdf to image
                                image_url = self.convert_to_images(image_file_path, directory, image_file_name)
                                minio_result = minio_utils.upload_local_file(image_url)
                                if minio_result['code'] == 0:
                                    image_download_link = minio_result['download_link']
                                    last_image_url = image_download_link
                                    image_line_text = image_download_link
                                    # page_content.append(image_line_text)
                                    chunk_content.append(image_line_text)
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
                            # if "https" in join_text and "graph" in join_text:
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
                    elif "ocr" in self.parser_choices:
                        page_data, page_num = ocr_utils.get_page_data(pagenum, self.file_path, self.ocr_model_id)
                        if page_data is not None:
                            for item in page_data:
                                if "text" not in item:
                                    continue
                                if item["type"] not in ['page-header', 'page-footer']:
                                    current_page_content_len = len(item["text"])
                                    page_chunk = {}
                                    page_chunk["text"] = item["text"]
                                    page_chunk["page_num"] = [pagenum + 1]
                                    page_chunk["file_path"] = self.file_path
                                    page_chunk["type"] = "text"
                                    page_chunk["embedding_chunks"] = []
                                    page_chunk["start_position"] = {}
                                    page_chunk["end_position"] = {}
                                    page_chunks.append(page_chunk)
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
                # Number of initialization checklists
                table_num = 0

                plumber_page = pdf.pages[pagenum]
                chunk_content = []
                try:
                    # Find the number of tables on this page
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
                        logger.info("---->clean_text=%s" % clean_text)
                        print("---->clean_text=%s" % clean_text)
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

        # doc_list = []
        metadata = {"source": self.file_path}
        if (text == '' or len(text) <= 150 or page_empty_count > 2) and ("ocr" in self.parser_choices):
            try:
                chunks = ocr_utils.ocr_parser(self.file_path, self.ocr_model_id)
                for chunk in chunks:
                    text += chunk["text"] + "\n"
            except Exception as err:
                import traceback
                logger.error("------> ocr error %s" % err)
                logger.error(traceback.format_exc())

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