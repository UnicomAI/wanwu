import jieba
from typing import Optional
import logging
from pymilvus.model.sparse.bm25.tokenizers import build_default_analyzer as milvus_build_default_analyzer

logger = logging.getLogger(__name__)

class ModelLoader:
    _instance = None
    _initialized = False
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(ModelLoader, cls).__new__(cls)
        return cls._instance
    
    def __init__(self):
        if not self._initialized:
            logger.info("Initializing ModelLoader...")
            self._load_models()
            ModelLoader._initialized = True
    
    def _load_models(self):
        """加载所有需要的模型"""
        # Preload jieba
        logger.info("Loading jieba model...")
        jieba.initialize()
        
    def build_default_analyzer(self, language='zh'):
        """构建默认的文本分析器"""
        if language == 'zh':
            return milvus_build_default_analyzer(language='zh')
        # Analyzer support for other languages ​​can be added
        return None
        
    @classmethod
    def get_instance(cls):
        if cls._instance is None:
            cls._instance = ModelLoader()
        return cls._instance 