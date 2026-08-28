#!/usr/bin/env python3
"""
Knowledge Base Skill for OpenClaw v2026.2
完全替代 memory-core 插件
"""

import sys
import json
import sqlite3
import hashlib
import re
import os
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, List, Optional, Any


class KnowledgeBase:
    """知识库核心 - 替代 OpenClaw memory-core"""
    
    def __init__(self, config: Dict = None):
        self.config = config or {}
        
        # 数据目录：优先使用配置，其次默认路径
        data_dir = self.config.get('data_dir')
        if data_dir:
            self.data_dir = Path(data_dir).expanduser()
        else:
            self.data_dir = Path.home() / ".openclaw" / "workspace" / "skills" / "knowledge-spider" / "data"
        
        self.data_dir.mkdir(parents=True, exist_ok=True)
        self.db_path = self.data_dir / "memory.db"
        
        # 配置项
        self.max_results = self.config.get('max_results', 5)
        self.similarity_threshold = self.config.get('similarity_threshold', 0.75)
        self.boost_recent = self.config.get('boost_recent', True)
        self.boost_frequent = self.config.get('boost_frequent', True)
        self.decay_days = self.config.get('decay_days', 30)
        
        self._init_db()
    
    def _init_db(self):
        """初始化数据库"""
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()
        
        # 主表：与 OpenClaw 兼容的结构
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS memories (
                id TEXT PRIMARY KEY,
                content TEXT NOT NULL,
                category TEXT DEFAULT 'general',
                source TEXT DEFAULT 'user',
                timestamp TEXT,
                access_count INTEGER DEFAULT 0,
                last_accessed TEXT,
                metadata TEXT
            )
        ''')
        
        # 索引优化
        cursor.execute('CREATE INDEX IF NOT EXISTS idx_timestamp ON memories(timestamp)')
        cursor.execute('CREATE INDEX IF NOT EXISTS idx_category ON memories(category)')
        cursor.execute('CREATE INDEX IF NOT EXISTS idx_content ON memories(content)')
        
        # 使用历史表（用于频率统计）
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS usage_log (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                memory_id TEXT,
                query TEXT,
                used_at TEXT DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (memory_id) REFERENCES memories(id)
            )
        ''')
        
        conn.commit()
        conn.close()
    
    # ==================== OpenClaw 标准 Memory 接口 ====================
    
    def memory_search(self, query: str, limit: int = 5) -> Dict:
        """
        OpenClaw 标准接口：搜索记忆
        完全兼容原 memory-core 的 memory_search 工具
        """
        results = self._semantic_search(query, limit)
        
        # 记录使用
        for entry in results:
            self._update_access_count(entry['id'])
        
        return {
            "success": True,
            "results": results,
            "count": len(results),
            "provider": "knowledge-spider",
            "query": query
        }
    
    def memory_store(self, content: str, category: str = "general", 
                     source: str = "user", metadata: Dict = None) -> Dict:
        """
        OpenClaw 标准接口：存储记忆
        完全兼容原 memory-core 的 memory_store 工具
        """
        if not content or len(content) < 2:
            return {
                "success": False,
                "error": "Content too short or empty",
                "provider": "knowledge-spider"
            }
        
        # 生成唯一ID
        entry_id = hashlib.sha256(
            f"{content}{source}{datetime.now().isoformat()}".encode()
        ).hexdigest()[:16]
        
        # 自动分类检测
        detected_category = self._detect_category(content, category)
        
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()
        
        try:
            cursor.execute('''
                INSERT INTO memories 
                (id, content, category, source, timestamp, access_count, metadata)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            ''', (
                entry_id,
                content,
                detected_category,
                source,
                datetime.now().isoformat(),
                0,
                json.dumps(metadata or {}, ensure_ascii=False)
            ))
            
            conn.commit()
            
            return {
                "success": True,
                "id": entry_id,
                "content_preview": content[:100] + "..." if len(content) > 100 else content,
                "category": detected_category,
                "timestamp": datetime.now().isoformat(),
                "provider": "knowledge-spider",
                "message": f"已存储到知识库 (分类: {detected_category})"
            }
            
        except Exception as e:
            return {
                "success": False,
                "error": str(e),
                "provider": "knowledge-spider"
            }
        finally:
            conn.close()
    
    def memory_get(self, memory_id: str = None, category: str = None) -> Dict:
        """
        OpenClaw 标准接口：获取特定记忆
        完全兼容原 memory-core 的 memory_get 工具
        """
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()
        
        try:
            if memory_id:
                cursor.execute("SELECT * FROM memories WHERE id = ?", (memory_id,))
            elif category:
                cursor.execute(
                    "SELECT * FROM memories WHERE category = ? ORDER BY timestamp DESC LIMIT 1", 
                    (category,)
                )
            else:
                cursor.execute("SELECT * FROM memories ORDER BY timestamp DESC LIMIT 1")
            
            row = cursor.fetchone()
            
            if not row:
                return {
                    "success": False,
                    "error": "Memory not found",
                    "provider": "knowledge-spider"
                }
            
            self._update_access_count(row[0])
            
            return {
                "success": True,
                "id": row[0],
                "content": row[1],
                "category": row[2],
                "source": row[3],
                "timestamp": row[4],
                "access_count": row[5],
                "provider": "knowledge-spider"
            }
            
        finally:
            conn.close()
    
    def memory_forget(self, target: str) -> Dict:
        """
        OpenClaw 标准接口：删除记忆
        完全兼容原 memory-core 的 memory_forget 工具
        """
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()
        
        try:
            # 支持 ID 精确删除或内容模糊删除
            cursor.execute(
                "DELETE FROM memories WHERE id = ? OR content LIKE ?", 
                (target, f"%{target}%")
            )
            deleted = cursor.rowcount
            
            # 清理使用历史
            if deleted > 0:
                cursor.execute("DELETE FROM usage_log WHERE memory_id = ?", (target,))
            
            conn.commit()
            
            return {
                "success": True,
                "deleted": deleted,
                "target": target,
                "provider": "knowledge-spider",
                "message": f"已删除 {deleted} 条记忆"
            }
            
        except Exception as e:
            return {
                "success": False,
                "error": str(e),
                "provider": "knowledge-spider"
            }
        finally:
            conn.close()
    
    # ==================== 扩展工具（本技能特有） ====================
    
    def kb_query(self, keyword: str, limit: int = 5, category: Optional[str] = None) -> Dict:
        """高级查询 - 支持分类过滤"""
        results = self._semantic_search(keyword, limit, category)
        
        return {
            "success": True,
            "results": results,
            "count": len(results),
            "provider": "knowledge-spider"
        }
    
    def kb_stats(self) -> Dict:
        """知识库统计"""
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()
        
        # 总条目
        cursor.execute("SELECT COUNT(*) FROM memories")
        total = cursor.fetchone()[0]
        
        # 分类统计
        cursor.execute("SELECT category, COUNT(*) FROM memories GROUP BY category")
        by_category = dict(cursor.fetchall())
        
        # 最近7天新增
        week_ago = (datetime.now() - timedelta(days=7)).isoformat()
        cursor.execute("SELECT COUNT(*) FROM memories WHERE timestamp > ?", (week_ago,))
        recent = cursor.fetchone()[0]
        
        # 高频访问
        cursor.execute('''
            SELECT id, content, access_count FROM memories 
            ORDER BY access_count DESC LIMIT 5
        ''')
        top_accessed = [
            {"id": row[0], "preview": row[1][:50], "count": row[2]}
            for row in cursor.fetchall()
        ]
        
        conn.close()
        
        # 计算存储大小
        db_size = os.path.getsize(self.db_path) if self.db_path.exists() else 0
        
        return {
            "success": True,
            "total_entries": total,
            "by_category": by_category,
            "recent_additions_7d": recent,
            "top_accessed": top_accessed,
            "storage_size_bytes": db_size,
            "storage_size_mb": round(db_size / 1024 / 1024, 2),
            "provider": "knowledge-spider"
        }
    
    def kb_context(self, current_input: str, max_length: int = 4000) -> Dict:
        """为当前对话生成知识库上下文"""
        # 提取关键词
        keywords = self._extract_keywords(current_input)
        
        if not keywords:
            return {
                "success": True,
                "context": "",
                "relevant_entries": 0,
                "injected": False,
                "provider": "knowledge-spider"
            }
        
        # 用关键词搜索
        all_results = []
        for kw in keywords[:2]:  # 取前2个关键词
            results = self._semantic_search(kw, limit=3)
            all_results.extend(results)
        
        # 去重并排序
        seen_ids = set()
        unique_results = []
        for r in all_results:
            if r['id'] not in seen_ids:
                seen_ids.add(r['id'])
                unique_results.append(r)
        
        # 限制数量
        unique_results = unique_results[:5]
        
        if not unique_results:
            return {
                "success": True,
                "context": "",
                "relevant_entries": 0,
                "injected": False,
                "provider": "knowledge-spider"
            }
        
        # 格式化上下文
        context_parts = [
            "## 相关知识库 (Relevant Knowledge)",
            "以下是从历史记忆中提取的相关信息，请优先考虑：",
            ""
        ]
        
        for i, entry in enumerate(unique_results, 1):
            date = entry['timestamp'][:10] if entry['timestamp'] else "未知"
            emoji = {
                'preference': '⭐',
                'fact': '📖',
                'task': '✅',
                'important': '🔴',
                'general': '📝'
            }.get(entry['category'], '📝')
            
            context_parts.append(f"{i}. {emoji} [{entry['category']}] {date}: {entry['content'][:120]}")
        
        context_parts.append("")
        context_parts.append("## 当前对话")
        
        context = "\n".join(context_parts)
        
        # 截断到最大长度
        if len(context) > max_length:
            context = context[:max_length-3] + "..."
        
        return {
            "success": True,
            "context": context,
            "relevant_entries": len(unique_results),
            "injected": True,
            "provider": "knowledge-spider"
        }
    
    # ==================== 内部方法 ====================
    
    def _semantic_search(self, query: str, limit: int = 5, 
                        category: Optional[str] = None) -> List[Dict]:
        """语义搜索 + 优先级排序"""
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()
        
        # 构建查询
        sql = "SELECT * FROM memories WHERE content LIKE ?"
        params = [f"%{query}%"]
        
        if category:
            sql += " AND category = ?"
            params.append(category)
        
        # 获取更多用于排序
        fetch_limit = limit * 3
        sql += f" ORDER BY timestamp DESC LIMIT {fetch_limit}"
        
        cursor.execute(sql, params)
        rows = cursor.fetchall()
        conn.close()
        
        # 转换为字典
        candidates = []
        for row in rows:
            entry = {
                'id': row[0],
                'content': row[1],
                'category': row[2],
                'source': row[3],
                'timestamp': row[4],
                'access_count': row[5] or 0,
                'last_accessed': row[6],
                'metadata': json.loads(row[7]) if row[7] else {}
            }
            # 基础相似度（简化版，生产环境可用向量相似度）
            entry['similarity'] = self._calculate_similarity(query, entry['content'])
            candidates.append(entry)
        
        # 优先级排序
        ranked = self._rank_by_priority(candidates, query)
        
        return ranked[:limit]
    
    def _calculate_similarity(self, query: str, content: str) -> float:
        """计算查询与内容的相似度（简化版）"""
        query_words = set(query.lower().split())
        content_words = set(content.lower().split())
        
        if not query_words:
            return 0.0
        
        # Jaccard 相似度
        intersection = query_words & content_words
        union = query_words | content_words
        
        if not union:
            return 0.0
        
        return len(intersection) / len(union)
    
    def _rank_by_priority(self, candidates: List[Dict], query: str) -> List[Dict]:
        """多维度优先级排序"""
        now = datetime.now()
        
        for entry in candidates:
            score = entry.get('similarity', 0)  # 基础相似度
            
            # 时效性权重
            if self.boost_recent and entry['timestamp']:
                try:
                    entry_time = datetime.fromisoformat(entry['timestamp'])
                    days_old = (now - entry_time).days
                    time_boost = max(0, 1 - (days_old / self.decay_days)) * 0.2
                    score += time_boost
                except:
                    pass
            
            # 频率权重
            if self.boost_frequent:
                freq_boost = min(entry['access_count'] / 20, 1.0) * 0.15
                score += freq_boost
            
            # 类型权重
            type_weights = {
                'preference': 0.15,
                'important': 0.12,
                'fact': 0.08,
                'task': 0.05,
                'general': 0.0
            }
            score += type_weights.get(entry['category'], 0)
            
            # 近期访问奖励
            if entry['last_accessed']:
                try:
                    last_access = datetime.fromisoformat(entry['last_accessed'])
                    hours_since = (now - last_access).total_seconds() / 3600
                    if hours_since < 24:
                        score += 0.1
                except:
                    pass
            
            entry['priority_score'] = round(score, 3)
        
        # 按分数排序
        candidates.sort(key=lambda x: x['priority_score'], reverse=True)
        return candidates
    
    def _detect_category(self, content: str, suggested: str) -> str:
        """自动检测内容分类"""
        # 如果用户指定了非默认分类，使用用户的
        if suggested and suggested != 'general':
            return suggested
        
        content_lower = content.lower()
        
        # 偏好检测
        preference_patterns = [
            r'喜欢|偏好|习惯|讨厌|厌恶|爱用|常用',
            r'like|prefer|enjoy|hate|always use',
            r'我[要想喜欢]|用户[要想喜欢]'
        ]
        for p in preference_patterns:
            if re.search(p, content_lower):
                return 'preference'
        
        # 重要信息检测
        important_patterns = [
            r'重要|关键|紧急|务必|必须|密码|密钥|token|secret',
            r'important|critical|urgent|password|key|secret|must',
            r'别忘了|切记|注意'
        ]
        for p in important_patterns:
            if re.search(p, content_lower):
                return 'important'
        
        # 任务检测
        task_patterns = [
            r'任务|待办|todo|完成|截止|期限|deadline',
            r'task|todo|deadline|due|finish|complete',
            r'需要.*做|应该.*做'
        ]
        for p in task_patterns:
            if re.search(p, content_lower):
                return 'task'
        
        # 事实检测
        fact_patterns = [
            r'事实是|真相是|数据|统计|研究表明',
            r'fact|truth|data shows|research',
            r'.*是.*[0-9%].*'  # 包含数字的陈述
        ]
        for p in fact_patterns:
            if re.search(p, content_lower):
                return 'fact'
        
        return 'general'
    
    def _extract_keywords(self, text: str) -> List[str]:
        """从文本提取关键词"""
        # 去除停用词
        stop_words = {'的', '了', '是', '在', '我', '有', '和', '就', '不', '人', '都', '一', '一个', '上', '也', '很', '到', '说', '要', '去', '你', '会', '着', '没有', '看', '好', '自己', '这', '那', '这些', '那些', '这个', '那个', '之', '与', '及', '等', '或', '但', '而', '因为', '所以', '如果', '虽然', '然而', '因此', '可以', '可能', '应该', '需要', '请', '帮我', '给我', '记录', '保存', '查询', '查查', 'the', 'a', 'an', 'is', 'are', 'was', 'were', 'be', 'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will', 'would', 'could', 'should', 'may', 'might', 'must', 'can', 'this', 'that', 'these', 'those', 'i', 'you', 'he', 'she', 'it', 'we', 'they', 'me', 'him', 'her', 'us', 'them', 'my', 'your', 'his', 'her', 'its', 'our', 'their', 'and', 'but', 'or', 'yet', 'so', 'for', 'nor', 'as', 'at', 'by', 'from', 'in', 'into', 'of', 'on', 'to', 'with', 'about', 'above', 'across', 'after', 'against', 'along', 'among', 'around', 'before', 'behind', 'below', 'beneath', 'beside', 'between', 'beyond', 'during', 'except', 'inside', 'outside', 'until', 'upon', 'within', 'without'}
        
        # 提取中文字符和英文单词
        words = re.findall(r'[\u4e00-\u9fa5]{2,}|[a-zA-Z]{3,}', text.lower())
        
        # 过滤停用词和短词
        keywords = [w for w in words if w not in stop_words and len(w) > 1]
        
        # 去重保持顺序
        seen = set()
        unique = []
        for w in keywords:
            if w not in seen:
                seen.add(w)
                unique.append(w)
        
        return unique[:5]  # 最多5个关键词
    
    def _update_access_count(self, entry_id: str):
        """更新访问计数"""
        conn = sqlite3.connect(str(self.db_path))
        cursor = conn.cursor()
        
        now = datetime.now().isoformat()
        
        cursor.execute('''
            UPDATE memories 
            SET access_count = access_count + 1, last_accessed = ?
            WHERE id = ?
        ''', (now, entry_id))
        
        cursor.execute('''
            INSERT INTO usage_log (memory_id, query) VALUES (?, ?)
        ''', (entry_id, "access"))
        
        conn.commit()
        conn.close()


class IntentRecognizer:
    """自然语言意图识别"""
    
    def recognize(self, text: str) -> Dict[str, Any]:
        text_lower = text.lower()
        
        # 存储意图
        store_patterns = [
            (r'(?:请|帮我|需要|记得|别忘了|先)?\s*(记录|记下来|存档|保存|存储|存起来|存入|写入|添加|放进|入库)', 'store'),
            (r'(?:记得|记住|别忘了|记一下|记牢)', 'store'),
            (r'\b(save|store|remember|record|archive|log|keep|note down|jot down|write down)\b', 'store'),
            (r'\b(don\'t forget|make sure to remember)\b', 'store'),
        ]
        
        # 查询意图
        query_patterns = [
            (r'(?:查|查询|查找|搜索|找找|看看|有没有|是否存在|帮我找|给我找)', 'query'),
            (r'(?:之前|以前|上次|之前说过|历史|回顾)', 'query'),
            (r'(?:关于|有关|涉及|提到|谈到)', 'query'),
            (r'(?:回忆|想起|记得|知不知道|还记不记得)', 'query'),
            (r'\b(find|search|query|look up|look for|retrieve|recall|check|show me)\b', 'query'),
            (r'\b(previous|past|history|before|earlier)\b', 'query'),
        ]
        
        # 删除意图
        delete_patterns = [
            (r'(?:删除|删掉|忘掉|遗忘|去掉|移除|清空|忘记|抹掉)', 'delete'),
            (r'(?:不需要|不用|取消).{0,5}(?:记忆|记录|知识)', 'delete'),
            (r'(?:这|那).*?(?:错了|不对|误报|重复)', 'delete'),
            (r'\b(delete|remove|forget|erase|clear|drop|wipe)\b', 'delete'),
            (r'\bthis is (?:wrong|incorrect|mistake|duplicate)\b', 'delete'),
        ]
        
        # 统计意图
        stats_patterns = [
            (r'(?:统计|状态|情况|多少|数量|容量|列表|汇总)', 'stats'),
            (r'(?:知识库|记忆).{0,3}(?:多大|多少|状态|如何)', 'stats'),
            (r'\b(stats|statistics|status|count|how many|list|summary|overview)\b', 'stats'),
            (r'\b(size of|storage used|disk space)\b', 'stats'),
        ]
        
        # 测试所有模式
        all_patterns = store_patterns + query_patterns + delete_patterns + stats_patterns
        
        for pattern, intent in all_patterns:
            if re.search(pattern, text_lower):
                return self._extract_params(text, intent)
        
        return {"intent": "unknown", "params": {}}
    
    def _extract_params(self, text: str, intent: str) -> Dict:
        """提取参数"""
        if intent == 'store':
            return self._extract_store_params(text)
        elif intent == 'query':
            return self._extract_query_params(text)
        elif intent == 'delete':
            return self._extract_delete_params(text)
        elif intent == 'stats':
            return {"intent": "stats", "params": {}}
        
        return {"intent": intent, "params": {}}
    
    def _extract_store_params(self, text: str) -> Dict:
        """提取存储参数"""
        # 清理指令词
        clean = text
        remove_patterns = [
            r'^(?:请|帮我|给我|需要|记得|别忘了|先|请把|把这个)?\s*',
            r'(?:记录|记下来|存档|保存|存储|存起来|存入|写入|添加|放进|入库|到知识库|到记忆库)',
            r'(?:一下|这个|那条|这段|这点)',
            r'(?:save|store|remember|record|write down|note down|to (the )?(kb|knowledge base|memory))',
            r'[，,]?\s*(?:谢谢|好吗|可以吗|行不|ok\?|please)?$',
        ]
        
        for pattern in remove_patterns:
            clean = re.sub(pattern, '', clean, flags=re.IGNORECASE)
        
        clean = clean.strip(' ：:，,;；.。')
        
        # 检测分类
        category = 'general'
        if re.search(r'喜欢|偏好|习惯|讨厌|爱用|常用|preference|like|prefer', text, re.I):
            category = 'preference'
        elif re.search(r'重要|关键|紧急|密码|密钥|token|important|critical|password|secret', text, re.I):
            category = 'important'
        elif re.search(r'任务|待办|todo|截止|期限|task|deadline|due', text, re.I):
            category = 'task'
        elif re.search(r'事实是|真相|数据|研究表明|fact|truth|data', text, re.I):
            category = 'fact'
        
        return {
            "intent": "store",
            "params": {
                "content": clean,
                "category": category,
                "source": "user_input"
            }
        }
    
    def _extract_query_params(self, text: str) -> Dict:
        """提取查询参数"""
        clean = text
        remove_patterns = [
            r'(?:查|查询|查找|搜索|找找|看看|帮我找|给我找|检索)',
            r'(?:一下|这个|那条|那段)',
            r'(?:find|search|query|look up|look for|retrieve|check|show)',
            r'(?:关于|有关|涉及|提到|谈到|regarding|about|related to|concerning)',
            r'(?:之前|以前|历史|previous|past)',
            r'[，,]?\s*(?:谢谢|好吗)?$',
        ]
        
        for pattern in remove_patterns:
            clean = re.sub(pattern, '', clean, flags=re.IGNORECASE)
        
        clean = clean.strip(' ：:，,;；.。')
        
        # 提取数量限制
        limit_match = re.search(r'(?:top|前|最)?\s*(\d+)\s*(?:条|个|条记录|items?)?', text)
        limit = int(limit_match.group(1)) if limit_match else 5
        
        return {
            "intent": "query",
            "params": {
                "keyword": clean if clean else text[:20],
                "limit": limit
            }
        }
    
    def _extract_delete_params(self, text: str) -> Dict:
        """提取删除参数"""
        clean = text
        remove_patterns = [
            r'(?:删除|删掉|忘掉|遗忘|去掉|移除|清空|忘记|抹掉)',
            r'(?:一下|这个|那条|那段|这点)',
            r'(?:delete|remove|forget|erase|clear|drop|wipe)',
            r'[，,]?\s*(?:谢谢|好吗)?$',
        ]
        
        for pattern in remove_patterns:
            clean = re.sub(pattern, '', clean, flags=re.IGNORECASE)
        
        clean = clean.strip(' ：:，,;；.。')
        
        return {
            "intent": "delete",
            "params": {
                "target": clean if clean else text[:20]
            }
        }


# ==================== OpenClaw 调用入口 ====================

def main():
    """
    OpenClaw 标准调用入口
    支持两种方式：
    1. 命令行参数: python index.py <action> [params_json]
    2. Stdin JSON: echo '{"action": "...", "params": {...}}' | python index.py
    """
    
    # 解析输入
    if len(sys.argv) >= 2:
        # 命令行模式
        action = sys.argv[1]
        params = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
        config = {}
    else:
        # Stdin JSON 模式（OpenClaw 标准）
        try:
            input_data = json.load(sys.stdin)
            action = input_data.get("action", "")
            params = input_data.get("params", {})
            config = input_data.get("config", {})
        except json.JSONDecodeError:
            print(json.dumps({
                "success": False,
                "error": "Invalid JSON input",
                "provider": "knowledge-spider"
            }, ensure_ascii=False))
            return
        except Exception as e:
            print(json.dumps({
                "success": False,
                "error": f"Input error: {str(e)}",
                "provider": "knowledge-spider"
            }, ensure_ascii=False))
            return
    
    # 初始化
    kb = KnowledgeBase(config)
    recognizer = IntentRecognizer()
    
    # 路由表
    handlers = {
        # OpenClaw 标准 Memory 接口
        "memory_search": lambda: kb.memory_search(
            params.get("query", ""),
            params.get("limit", 5)
        ),
        "memory_store": lambda: kb.memory_store(
            params.get("content", ""),
            params.get("category", "general"),
            params.get("source", "user"),
            params.get("metadata")
        ),
        "memory_get": lambda: kb.memory_get(
            params.get("id"),
            params.get("category")
        ),
        "memory_forget": lambda: kb.memory_forget(
            params.get("target", "")
        ),
        # 扩展接口
        "kb_query": lambda: kb.kb_query(
            params.get("keyword", ""),
            params.get("limit", 5),
            params.get("category")
        ),
        "kb_stats": lambda: kb.kb_stats(),
        "kb_context": lambda: kb.kb_context(
            params.get("current_input", ""),
            params.get("max_length", 4000)
        ),
        # 意图识别（供 OpenClaw 路由使用）
        "recognize_intent": lambda: recognizer.recognize(
            params.get("text", "")
        ),
    }
    
    # 执行
    if action in handlers:
        try:
            result = handlers[action]()
        except Exception as e:
            result = {
                "success": False,
                "error": f"Execution error: {str(e)}",
                "provider": "knowledge-spider"
            }
    else:
        result = {
            "success": False,
            "error": f"Unknown action: {action}. Available: {list(handlers.keys())}",
            "provider": "knowledge-spider"
        }
    
    # 输出 JSON
    print(json.dumps(result, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()