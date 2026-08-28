
import os
import logging
from typing import Union, List, Optional
from pathlib import Path
from langchain_core.documents import Document
from langchain_community.document_loaders import TextLoader
from utils.prompts import IMAGE_TEXT_EXTRACT_PROMPT
from utils import model_parser_utils
from utils import minio_utils
from utils import multimodal_utils

logger = logging.getLogger(__name__)


class ImageLoader(TextLoader):
    def __init__(self,
                 file_path: Union[str, Path],
                 encoding: Optional[str] = None,
                 autodetect_encoding: bool = False,
                 parser_choices: List[str] = None,
                 ocr_model_id: str = "",
                 multimodal_model_id: str = ""):
        """Initialize a PDFLoader with file path and additional chunk_type."""
        super().__init__(file_path, encoding, autodetect_encoding)  # 确保调用父类的__init__

        # 如果没有提供parser_choices，则使用默认值["text"]
        if parser_choices is None:
            parser_choices = ["text"]
        self.parser_choices = parser_choices
        self.ocr_model_id = ocr_model_id
        self.multimodal_model_id = multimodal_model_id

    def load(self) -> List[Document]:
        text = ""
        image_desc = ""
        try:
            path_obj = Path(self.file_path)
            file_name = path_obj.stem
            minio_result = minio_utils.upload_local_file(self.file_path)
            if minio_result['code'] == 0:
                image_minio_url = minio_result['download_link']
                text = f"![{file_name}]({image_minio_url})"
            if self.ocr_model_id and "model" in self.parser_choices:
                # 图片发给文档解析模型（/pdf-parser 端点，复用 model_parser_utils）
                image_desc = model_parser_utils.model_parser_image(self.file_path, self.ocr_model_id)
            elif self.multimodal_model_id and "multimodal" in self.parser_choices:
                image_desc = multimodal_utils.req_unicom_VL_plus(self.file_path, self.multimodal_model_id, IMAGE_TEXT_EXTRACT_PROMPT)
            if image_desc:
                text += f" 此图的画面描述：{image_desc}" + "\n"
            logger.info("========>image_parser_text:%s" % text)
        except Exception as e:
            raise RuntimeError(f"Error loading {self.file_path}") from e

        metadata = {"source": self.file_path}
        return [Document(page_content=text, metadata=metadata)]


if __name__ == "__main__":

    filepath = "your_file.jpg"
    loader = ImageLoader(filepath)
    docs = loader.load()
    for doc in docs:
        print(doc)