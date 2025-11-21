
import os
from typing import Union, List, Optional
from pathlib import Path
from langchain_core.documents import Document
from langchain_community.document_loaders import TextLoader
from utils import ocr_utils
logger_name='rag_image_loader'
app_name = os.getenv("LOG_FILE")
from logging_config import setup_logging
logger = setup_logging(app_name,logger_name)
logger.info(logger_name+'---------LOG_FILE：'+repr(app_name))


class ImageLoader(TextLoader):
    def __init__(self,
                 file_path: Union[str, Path],
                 encoding: Optional[str] = None,
                 autodetect_encoding: bool = False,
                 ocr_model_id: str = ""):
        """Initialize a PDFLoader with file path and additional chunk_type."""
        super().__init__(file_path, encoding, autodetect_encoding)  # Make sure to call the parent class's __init__
        self.ocr_model_id = ocr_model_id

    def load(self) -> List[Document]:
        text = ""
        try:

            text = ocr_utils.ocr_parser_text(self.file_path, self.ocr_model_id)

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