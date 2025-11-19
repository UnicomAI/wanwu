#!/usr/bin/env python
# -*- coding: utf-8 -*-
# import os
# from appbuilder.core.components.doc_parser.doc_parser import DocParser
# from appbuilder.core.components.doc_splitter.doc_splitter import DocSplitter
# from appbuilder.core.message import Message
#
#
# os.environ["APPBUILDER_TOKEN"] = "bce-v3/ALTAK-8D8QyXR1mMXcpfndlmQ7L/a7f251fe675d4bae08b71c08589f4157332bd691"
#
# # 先解析 [EN] # Parse first
# msg = Message("./test.pdf")
# parser = DocParser()
# parse_result = parser(msg, return_raw=True)
#
# # 基于parser的结果切分段落 [EN] # Split paragraphs based on parser results
# splitter = DocSplitter(splitter_type="split_by_chunk")
# res_paras = splitter(parse_result)
#
# # 打印结果 [EN] # print results
# print(res_paras.content)




import os
import json
from appbuilder.core.components.doc_parser.doc_parser import DocParser
from appbuilder.core.components.doc_splitter.doc_splitter import DocSplitter
from appbuilder.core.message import Message

os.environ["APPBUILDER_TOKEN"] = "bce-v3/ALTAK-8D8QyXR1mMXcpfndlmQ7L/a7f251fe675d4bae08b71c08589f4157332bd691"

# 先解析 [EN] Analyze first
pdf_path = "11-航司三字码-国内部分.txt"
msg = Message("/home/jovyan/RAG_2.0/langchain_rag_new/商飞/11-航司三字码-国内部分.txt")

parser = DocParser()
parse_result = parser(msg, return_raw=True)

# 基于parser的结果切分段落 [EN] Split paragraphs based on parser results
doc_splitter = DocSplitter(splitter_type="split_by_chunk",
                           separators=["。", "！", "？", ".", "!", "?", "……", "|"],
                           max_segment_length=800,
                           overlap=0)
res_paras = doc_splitter(parse_result)
#print("parse_result=%s" % parse_result)
print("res_paras=%s" % res_paras)


