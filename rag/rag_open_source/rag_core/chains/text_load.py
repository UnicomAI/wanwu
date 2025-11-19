import os
import pinecone 
from tqdm import tqdm
from langchain.llms import OpenAI
from langchain.text_splitter import SpacyTextSplitter
# from langchain.document_loaders import TextLoader
from langchain_community.document_loaders import TextLoader

# from langchain.document_loaders import DirectoryLoader
from langchain_community.document_loaders.directory import DirectoryLoader


from langchain.indexes import VectorstoreIndexCreator
from langchain.embeddings.openai import OpenAIEmbeddings
from langchain.vectorstores import Pinecone

#一些配置文件 [EN] some configuration files
openai_key="你的key" # 注册 openai.com 后获得 [EN] Get it after registering on openai.com
pinecone_key="你的key" # 注册 app.pinecone.io 后获得 [EN] Obtained after registering app.pinecone.io
pinecone_index="你的库" #app.pinecone.io 获得 [EN] app.pinecone.io Get
pinecone_environment="你的Environment"  # 登录pinecone后，在indexes页面 查看Environment [EN] After logging in to pinecone, view Environment on the indexes page
pinecone_namespace="你的Namespace" #如果不存在自动创建 [EN] automatically created if does not exist

#科学上网你懂得 [EN] You know how to use the Internet scientifically
os.environ['HTTP_PROXY'] = 'http://127.0.0.1:7890'
os.environ['HTTPS_PROXY'] = 'http://127.0.0.1:7890'

#初始化pinecone [EN] Initialize pinecone
pinecone.init(
    api_key=pinecone_key,
    environment=pinecone_environment
)
index = pinecone.Index(pinecone_index)

#初始化OpenAI的embeddings [EN] Initialize OpenAI embeddings
embeddings = OpenAIEmbeddings(openai_api_key=openai_key)

#初始化text_splitter [EN] Initialize text_splitter
text_splitter = SpacyTextSplitter(pipeline='zh_core_web_sm',chunk_size=1000,chunk_overlap=200)

# 读取目录下所有后缀是txt的文件 [EN] Read all files in the directory with the suffix txt
loader = DirectoryLoader('../docs', glob="**/*.txt", loader_cls=TextLoader)

#读取文本文件 [EN] Read text file
documents = loader.load()

# 使用text_splitter对文档进行分割 [EN] Use text_splitter to split documents
split_text = text_splitter.split_documents(documents)
try:
	for document in tqdm(split_text):
		# 获取向量并储存到pinecone [EN] Get the vector and store it in pinecone
		Pinecone.from_documents([document], embeddings, index_name=pinecone_index)
except Exception as e:
    print(f"Error: {e}")
    quit()


