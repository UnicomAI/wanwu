from langchain.text_splitter import CharacterTextSplitter
import re
from typing import List


class AliTextSplitter(CharacterTextSplitter):
    def __init__(self, pdf: bool = False, **kwargs):
        super().__init__(**kwargs)
        self.pdf = pdf

    def split_text(self, text: str) -> List[str]:
        # The use_document_segmentation parameter specifies whether to use semantic segmentation to segment the document. The document semantic segmentation model adopted here is the open source nlp_bert_document-segmentation_chinese-base of DAMO Academy. The paper can be found at https://arxiv.org/abs/2107.09278
        # If you use a model for document semantic segmentation, you need to install modelscope[nlp]: pip install "modelscope[nlp]" -f https://modelscope.oss-cn-beijing.aliyuncs.com/releases/repo.html
        # Considering that three models are used, it may not be friendly to low-configuration GPUs, so here the models are loaded into the CPU for calculation. If necessary, you can replace the device with your own graphics card ID.
        if self.pdf:
            text = re.sub(r"\n{3,}", r"\n", text)
            text = re.sub('\s', " ", text)
            text = re.sub("\n\n", "", text)
        from modelscope.pipelines import pipeline

        p = pipeline(
            task="document-segmentation",
            model='damo/nlp_bert_document-segmentation_chinese-base',
            device="cpu")
        result = p(documents=text)
        sent_list = [i for i in result["text"].split("\n\t") if i]
        return sent_list
