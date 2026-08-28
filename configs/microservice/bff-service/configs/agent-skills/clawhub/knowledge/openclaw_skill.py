# E:/knowledge-base/openclaw_skill.py
"""
OpenClaw Skill - 本地知识库集成
提供知识库检索、投喂等功能
"""

import sys
from pathlib import Path

# 添加知识库路径
sys.path.insert(0, str(Path(__file__).parent / "scripts"))

from openclaw_integration import handle_request


def skill_handler(message: str, context: dict = None) -> str:
    """
    Skill 处理函数

    Args:
        message: 用户消息
        context: 上下文信息

    Returns:
        处理结果
    """
    # 解析用户请求
    message_lower = message.lower().strip()

    # ==================== 模式切换 ====================
    if any(keyword in message_lower for keyword in ['切换到本地', '用本地知识库', 'local_kb', '本地检索']):
        result = handle_request('set_mode', {'mode': 'local_kb'})
        return f"✓ 已切换到本地知识库检索模式\n\n{result.get('message', '')}"

    if any(keyword in message_lower for keyword in ['切换到anything', '用anythingllm', 'anything_llm', '对话模式', '聊天模式']):
        result = handle_request('set_mode', {'mode': 'anything_llm'})
        return f"✓ 已切换到 AnythingLLM 对话模式\n\n{result.get('message', '')}"

    if any(keyword in message_lower for keyword in ['切换模式', '换一个', '换系统']):
        status = handle_request('get_status', {})
        current = status.get('current_mode', 'local_kb')
        next_mode = 'anything_llm' if current == 'local_kb' else 'local_kb'
        result = handle_request('switch_mode', {'mode': next_mode})
        return result.get('message', result.get('error', '切换失败'))

    # ==================== 状态查询 ====================
    if any(keyword in message_lower for keyword in ['系统状态', '模式状态', '什么模式', '当前模式', 'status']):
        status = handle_request('get_status', {})

        mode_name = {
            'local_kb': '本地知识库检索',
            'anything_llm': 'AnythingLLM 对话'
        }.get(status.get('current_mode', ''), status.get('current_mode', ''))

        response = f"📊 系统状态\n\n"
        response += f"当前模式: {mode_name}\n"
        response += f"本地知识库: {'✓ 可用' if status.get('local_kb_available') else '✗ 不可用'}\n"
        response += f"AnythingLLM: {'✓ 可用' if status.get('anythingllm_available') else '✗ 不可用'}\n"
        return response

    # ==================== 检索请求 ====================
    if any(keyword in message_lower for keyword in ['查', '搜索', '查找', '问问', '咨询', '了解']):
        # 提取查询词
        query = extract_query(message)
        if query:
            # 检查是否指定模式
            mode = None
            if '本地' in message:
                mode = 'local_kb'
            elif '对话' in message or '聊天' in message:
                mode = 'anything_llm'

            # 使用双轨查询
            params = {'query': query, 'mode': mode, 'auto': True}
            result = handle_request('query', params)

            if result['success']:
                response = f"{result['answer']}"

                # 添加来源说明
                source = result.get('source', '')
                mode = result.get('mode', '')
                if source == 'anythingllm':
                    response += f"\n\n🤖 [AnythingLLM 对话模式]"
                elif source == 'local_kb':
                    confidence = result.get('confidence', '')
                    if confidence == 'high':
                        response += f"\n\n📚 [本地知识库 - 高置信度]"
                    elif confidence == 'medium':
                        response += f"\n\n📚 [本地知识库 - 中等置信度]"
                    else:
                        response += f"\n\n📚 [本地知识库]"

                # 添加自动选择说明
                if result.get('auto_selected'):
                    response += f"\n{result.get('reason', '')}"

                return response
            else:
                return f"检索失败: {result.get('error', '未知错误')}"

    # ==================== 统计请求 ====================
    if any(keyword in message_lower for keyword in ['统计', '有多少', '多少个', '总数', '信息']):
        result = handle_request('get_stats', {})

        if result['success']:
            stats = result['stats']
            return f"知识库统计信息：\n- 总文档数: {stats['total_documents']}\n- 集合名称: {stats['collection_name']}"
        else:
            return f"获取统计失败: {result.get('error', '未知错误')}"

    # ==================== 文档列表请求 ====================
    if any(keyword in message_lower for keyword in ['列表', '文档', '有哪些', '列出']):
        params = {'limit': 10}
        result = handle_request('list_documents', params)

        if result['success']:
            docs = result['documents']
            if docs:
                response = f"知识库中有 {len(docs)} 个文档:\n"
                for i, doc in enumerate(docs, 1):
                    preview = doc['preview'][:100] + "..." if len(doc['preview']) > 100 else doc['preview']
                    response += f"\n{i}. {doc['id']}\n   预览: {preview}\n"
                return response
            else:
                return "知识库中暂无文档"
        else:
            return f"获取文档列表失败: {result.get('error', '未知错误')}"

    # ==================== 帮助信息 ====================
    return get_help_text()


def extract_query(message: str) -> str:
    """
    从消息中提取查询词
    
    Args:
        message: 原始消息
    
    Returns:
        查询词
    """
    # 移除常见的查询前缀
    prefixes = ['查', '搜索', '查找', '问问', '咨询', '了解', '关于', '有关']
    
    clean_message = message.strip()
    
    for prefix in prefixes:
        if clean_message.startswith(prefix):
            # 移除前缀
            clean_message = clean_message[len(prefix):].strip()
            break
    
    # 如果消息仍包含疑问词，也移除
    question_words = ['什么', '怎么', '如何', '为什么', '哪里', '何时', '谁', '哪']
    for word in question_words:
        if word in clean_message:
            # 从疑问词之后的部分作为查询
            idx = clean_message.find(word)
            clean_message = clean_message[idx + len(word):].strip()
            break
    
    return clean_message if clean_message else message.strip()


def get_help_text() -> str:
    """
    获取帮助文本

    Returns:
        帮助信息
    """
    return """📚 本地知识库功能说明（双轨模式）

🔀 模式切换：
- "切换到本地知识库" - 使用本地检索系统
- "切换到 AnythingLLM" - 使用对话模式
- "切换模式" - 在两个系统间切换
- "当前模式" - 查看当前使用哪个系统

🔍 检索功能：
- "查一下XXX" - 自动选择最佳系统检索
- "查一下XXX 本地" - 强制使用本地知识库
- "查一下XXX 对话" - 强制使用 AnythingLLM
- "搜索关于XXX的信息" - 搜索相关内容
- "问问XXX" - 询问特定主题

📊 统计功能：
- "知识库统计" - 显示文档数量
- "有多少文档" - 显示总数
- "系统状态" - 显示双轨系统状态

📋 列表功能：
- "列出文档" - 显示文档列表
- "有哪些信息" - 显示文档概览

📤 投喂功能：
- 直接上传文件即可添加到知识库

💡 使用建议：
- 简单对话 → 用 AnythingLLM（更像聊天）
- 专业查询 → 用本地知识库（检索更精准）
- 不确定 → 直接问（自动选择最佳系统）

如需更多帮助，请告诉我！"""


# 如果直接运行此文件，显示帮助
if __name__ == "__main__":
    print(get_help_text())