#!/usr/bin/env python3
"""
Script to translate Chinese characters to English in mcp_config.yaml
This script will process the file and translate Chinese text to English.
"""

import re
import sys

# Translation dictionary for common Chinese terms
TRANSLATIONS = {
    # Common field names
    "名称": "name",
    "描述": "description", 
    "特性": "feature",
    "手册": "manual",
    "场景": "scenario",
    "摘要": "summary",
    "工具": "tool",
    "类型": "type",
    "字段": "field",
    "必需": "required",
    "可选": "optional",
    
    # Common terms
    "支持": "Supports",
    "提供": "Provides",
    "基于": "Based on",
    "服务": "service",
    "功能": "function",
    "配置": "configuration",
    "安装": "Install",
    "启动": "Start",
    "运行": "Run",
    "依赖": "dependencies",
    "默认": "default",
    "文件": "file",
    "路径": "path",
    "数据": "data",
    "查询": "query",
    "结果": "result",
    "格式": "format",
    "输出": "output",
    "输入": "input",
    "索引": "index",
    "执行": "Execute",
    "导出": "Export",
    "获取": "Get",
    "检查": "Check",
    "分析": "analysis",
    "处理": "processing",
    "识别": "recognition",
    "转写": "transcription",
    "音频": "audio",
    "文本": "text",
    "增强": "enhancement",
    "模型": "model",
    "信号": "signal",
    "医疗": "medical",
    "医学": "medical",
    "医药": "medical",
    "专业": "professional",
    "语音": "voice",
    "会议": "meeting",
    "记录": "record",
    "管理": "management",
    "系统": "system",
    "应用": "application",
    "适用于": "Applicable to",
    "包括": "including",
    "自动": "automatic",
    "智能": "intelligent",
    "实时": "real-time",
}

def has_chinese(text):
    """Check if text contains Chinese characters"""
    return bool(re.search(r'[\u4e00-\u9fff]', text))

def count_chinese_lines(filepath):
    """Count lines containing Chinese characters"""
    count = 0
    with open(filepath, 'r', encoding='utf-8') as f:
        for line in f:
            if has_chinese(line):
                count += 1
    return count

def main():
    filepath = '/Users/mohankumarv/Desktop/SAFVR/wanwu/configs/microservice/mcp-service/configs/mcp_config.yaml'
    
    count = count_chinese_lines(filepath)
    print(f"Total lines with Chinese characters: {count}")
    
    # Show sample lines with Chinese
    print("\nSample lines with Chinese (first 20):")
    with open(filepath, 'r', encoding='utf-8') as f:
        chinese_lines = []
        for i, line in enumerate(f, 1):
            if has_chinese(line):
                chinese_lines.append((i, line.rstrip()))
                if len(chinese_lines) >= 20:
                    break
        
        for line_num, line in chinese_lines:
            print(f"{line_num}: {line[:100]}...")

if __name__ == '__main__':
    main()
