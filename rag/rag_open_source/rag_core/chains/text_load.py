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

#some configuration files
openai_key="你的key" # Get it after registering on openai.com
pinecone_key="你的key" # Obtained after registering app.pinecone.io
pinecone_index="你的库" #app.pinecone.io Get
pinecone_environment="你的Environment"  # After logging in to pinecone, view Environment on the indexes page
pinecone_namespace="你的Namespace" #automatically created if does not exist

#You know how to use the Internet scientifically
os.environ['HTTP_PROXY'] = 'http://127.0.0.1:7890'
os.environ['HTTPS_PROXY'] = 'http://127.0.0.1:7890'

#Initialize pinecone
pinecone.init(
    api_key=pinecone_key,
    environment=pinecone_environment
)
index = pinecone.Index(pinecone_index)

#Initialize OpenAI embeddings
embeddings = OpenAIEmbeddings(openai_api_key=openai_key)

#Initialize text_splitter
text_splitter = SpacyTextSplitter(pipeline='zh_core_web_sm',chunk_size=1000,chunk_overlap=200)

# Read all files in the directory with the suffix txt
loader = DirectoryLoader('../docs', glob="**/*.txt", loader_cls=TextLoader)

#Read text file
documents = loader.load()

# Use text_splitter to split documents
split_text = text_splitter.split_documents(documents)
try:
	for document in tqdm(split_text):
		# Get the vector and store it in pinecone
		Pinecone.from_documents([document], embeddings, index_name=pinecone_index)
except Exception as e:
    print(f"Error: {e}")
    quit()


