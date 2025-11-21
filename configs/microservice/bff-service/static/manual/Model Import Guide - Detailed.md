# 模型Import教程

## 一、联通元景

### LLM

1、 Configuration 联通元景 API Key https://maas.ai-yuanjing.com/ 注册登录后，进入以下路径：

元景MaaS平台--Tools箱--API Key管理，新建或复制已有密钥。

![image-20250725111753372](assets/image-20250725111753372.png)

点击创建，填写AppName及AppDescription。确定后，即可生成API Key。

![image-20250725111847121](assets/image-20250725111847121.png)

![image-20250725111926508](assets/image-20250725111926508.png)

2、获取参数：选择想要Import的模型 https://maas.ai-yuanjing.com/aibase/portal/service

进入模型广场-选择模型-点击查看Details

![image-20250725112016244](assets/image-20250725112016244.png)

【模型Name】请求参数-model，如deepseek-r1-distill-qwen-32b

【API Key】第一步中申请的API Key

【推理url】请求Interface-请求URL：Note推理URL不带 /chat/completions 后缀，例如：https://maas-api.ai-yuanjing.com/openapi/compatible-mode/v1

![image-20250725112538728](assets/image-20250725112538728.png)

3、登录元景万悟，点击Model Management-模型Import-联通元景-模型TypeLLM，依次填写模型Name、API Key、推理url

![image-20250725112504492](assets/image-20250725112504492.png)



### Embedding

ImportSteps同LLM，只需在文档中心中搜索Embedding模型即可 https://maas.ai-yuanjing.com/doc/pages/216556678/#%E8%AF%B7%E6%B1%82%E6%8E%A5%E5%8F%A3

![image-20250725112754696](assets/image-20250725112754696.png)

【模型Name】请求Example中的model，如qwen3-embed-0.6b

【API Key】第一步中申请的API Key

【推理url】复制调用Interface：Note推理URL不带 /embeddings 后缀，例如：https://maas-api.ai-yuanjing.com/openapi/compatible-mode/v1

![image-20250725112837337](assets/image-20250725112837337.png)

登录元景万悟，点击Model Management-模型Import-联通元景-模型TypeEmbedding，依次填写模型Name、API Key、推理url

![image-20250725112716488](assets/image-20250725112716488.png)

### Rerank

ImportSteps同LLM，只需在文档中心中搜索rerank模型即可 https://maas.ai-yuanjing.com/doc/pages/216556678/#%E8%AF%B7%E6%B1%82%E6%8E%A5%E5%8F%A3

![image-20250725112754696](assets/image-20250725112754696.png)

【模型Name】请求Example中的model，如bge-m3

【API Key】第一步中申请的API Key

【推理url】复制调用Interface：Note推理URL不带 /rerank 后缀，例如：https://maas-api.ai-yuanjing.com/openapi/bge/v1

![image-20250725113812334](assets/image-20250725113812334.png)

3、登录元景万悟，点击Model Management-模型Import-联通元景-模型Typererank，依次填写模型Name、API Key、推理url

![image-20250725113158923](assets/image-20250725113158923.png)

### OCR

可App于Knowledge BaseFileUpload-解析方式-OCR解析

【模型Name】固定为 yuanjingOcr

【API Key】第一步中申请的API Key

【推理url】复制调用Interface，固定为 https://maas-api.ai-yuanjing.com/openapi/v1

登录元景万悟，点击Model Management-模型Import-联通元景-模型TypeOCR，依次填写模型Name、API Key、推理url

![image-20250731113537320](assets/image-20250731113537320.png)



### GUI

【模型Name】固定为 gui_agent_v1

【API Key】第一步中申请的API Key

【推理url】复制调用Interface，固定为 https://maas-gz-api.ai-yuanjing.com/openapi/v1 或 https://maas-api.ai-yuanjing.com/openapi/v1

登录元景万悟，点击Model Management-模型Import-联通元景-模型TypeGUI，依次填写模型Name、API Key、推理url

![image-20250904100050884](assets/image-20250904100050884.png)



### pdfDocument Parsing模型

可App于Knowledge BaseFileUpload-解析方式-模型解析

【模型Name】固定为 pdf-parser

【API Key】第一步中申请的API Key

【推理url】复制调用Interface，固定为 https://maas-api.ai-yuanjing.com/openapi/v1 

登录元景万悟，点击Model Management-模型Import-联通元景-模型TypepdfDocument Parsing模型，依次填写模型Name、API Key、推理url

![image-20250919105832500](assets/image-20250919105832500.png)





## 二、阿里通义千问

### LLM

1、 Configuration 阿里云百炼 API Key https://bailian.console.aliyun.com/?tab=model#/api-key

登录阿里云百炼平台--新User开通。首次开通需实名制，具体Steps详见https://help.aliyun.com/zh/account/user-guide/individual-identities

![image-20250725115411209](assets/image-20250725115411209.png)

![image-20250725115523681](assets/image-20250725115523681.png)

![image-20250725115829504](assets/image-20250725115829504.png)

![image-20250725115931936](assets/image-20250725115931936.png)

![image-20250725120024146](assets/image-20250725120024146.png)

2、获取参数：选择想要Import的模型https://bailian.console.aliyun.com/?tab=model#/model-market

进入模型广场-选择模型-点击查看Details和API参考

![image-20250725124441761](assets/image-20250725124441761.png)

【模型Name】查看Details后的code，如qwen-max

【API Key】第一步中申请的API Key

【推理url】查看“API参考”，使用SDK调用：例如 https://dashscope.aliyuncs.com/compatible-mode/v1

![image-20250725124515869](assets/image-20250725124515869.png)

![image-20250725124335928](assets/image-20250725124335928.png)

3、登录元景万悟，点击Model Management-模型Import-通义千问-模型TypeLLM，依次填写模型Name、API Key、推理url

![ea351b3290750d7b942ab4d43a4828c2](assets/ea351b3290750d7b942ab4d43a4828c2.png)

### Embedding

ImportSteps同LLM，只需在模型广场中搜索Embedding模型即可 

![image-20250725120728152](assets/image-20250725120728152.png)

【模型Name】查看“API参考”-模型Name，如text-embedding-v4

【API Key】第一步中申请的API Key

【推理url】查看“API参考”，快速入门中的ExampleCodebase_url：例如 https://dashscope.aliyuncs.com/compatible-mode/v1

![image-20250725120945990](assets/image-20250725120945990.png)

![image-20250725124746276](assets/image-20250725124746276.png)

登录元景万悟，点击Model Management-模型Import-通义千问-模型TypeEmbedding，依次填写模型Name、API Key、推理url

![e2789058713fa030854c3d979cec0dea](assets/e2789058713fa030854c3d979cec0dea.png)

### Rerank

ImportSteps同LLM，只需在模型广场中搜索排序模型即可

![image-20250725121515305](assets/image-20250725121515305.png)

【模型Name】查看“API参考”-模型Name，，如gte-rerank-v2

【API Key】第一步中申请的API Key

【推理url】复制调用Example中的url：Note推理URL不带/services/rerank/text-rerank/text-rerank后缀，如：https://dashscope.aliyuncs.com/api/v1

![image-20250725121545692](assets/image-20250725121545692.png)

![image-20250725123957256](assets/image-20250725123957256.png)

登录元景万悟，点击Model Management-模型Import-通义千问-模型TypeRerank，依次填写模型Name、API Key、推理url

![c7fabdfa0b33475c2a1735d3116786f6](assets/c7fabdfa0b33475c2a1735d3116786f6.png)

## 三、Ollama

### LLM

```
1. 本地Ollama部署：https://github.com/ollama/ollama

2. User在在本地启动OllamaService，并ConfirmService正常启动，并Confirm模型已加载，并Confirm模型可正确请求，以请求qwen2.5:0.5b为例：
   curl --location 'http://本地ip:11434/v1/chat/completions' \
   --header 'Content-Type: application/json' \
   --header 'Accept: application/json' \
   --data '{
           "model": "qwen2.5:0.5b",
           "messages": [{
                   "role": "user",
                   "content": "你好"
           }]
   }'


3. 访问IP（Notelocalhost要换成本机局域网或对外IP，例如192.168.0.xx，不能Yeslocalhost或127.0.0.1）

4. Import模型：
   4.1【模型Name】必须为上述curl中可以正确请求的model；例如 qwen2.5:0.5b
   4.2【访问IP】Notelocalhost要换成本机局域网或对外IP，例如192.168.0.xx，不能Yeslocalhost或127.0.0.1
   4.3【推理URL】必须为上述curl中可以正确请求的url；例如 http://本地ip:11434/v1（Note不带 /chat/completions 后缀）
```



### Embedding

ImportEmbedding模型同上述ImportLLM，Note推理URL不带 /embeddings 后缀



## 四、火山引擎-豆包

### LLM

1、获取参数：选择想要Import的模型https://console.volcengine.com/ark/region:ark+cn-beijing/model?vendor=Bytedance&view=DEFAULT_VIEW

进入模型广场-选择模型-悬停

![image-20250725125312203](assets/image-20250725125312203-1754018422086-1.png)

【模型Name】悬停后展示的Model ID，如doubao-seed-1-6-thinking-250715

【API Key】点击API接入，创建API Key

【推理url】获取API KEY后，点击“选择使用”，复制Link，Note推理URL不带/chat/completions后缀：例如 https://ark.cn-beijing.volces.com/api/v3

![image-20250725125451548](assets/image-20250725125451548-1754018422087-2.png)

![image-20250725125558032](assets/image-20250725125558032-1754018422087-3.png)

![image-20250725162735374](assets/image-20250725162735374-1754018422087-4.png)

2、登录元景万悟，点击Model Management-模型Import-火山引擎-模型TypeLLM，依次填写模型Name、API Key、推理url

![72776f6bff2edbd709e861ddfd8e64ee](assets/72776f6bff2edbd709e861ddfd8e64ee.png)

### Embedding

ImportSteps同LLM，只需在模型广场中搜索向量模型即可

![image-20250725130019497](assets/image-20250725130019497-1755155158130-1.png)

![image-20250725130126264](assets/image-20250725130126264-1755155158130-2.png)

【模型Name】悬停后展示的Model ID，如doubao-embedding-large-text-250515

【API Key】点击API接入，创建API Key

【推理url】获取API KEY后，点击“选择使用”，复制base_url，Note推理URL不带/embeddings后缀：例如 https://ark.cn-beijing.volces.com/api/v3

![image-20250725130205458](assets/image-20250725130205458-1755155158130-3.png)

登录元景万悟，点击Model Management-模型Import-火山引擎-模型TypeEmbedding，依次填写模型Name、API Key、推理url

![cc6540b0cf4feca1d55551b2bea9f17f](assets/cc6540b0cf4feca1d55551b2bea9f17f-1755155158130-4.png)

## 五、None问芯穹

### LLM

1、 Configuration None问芯穹 API Key  https://cloud.infini-ai.com/iam/secret/key

![image-20250916095827688](assets/image-20250916095827688.png)

![image-20250916095851399](assets/image-20250916095851399.png)

2、获取参数：选择想要Import的模型https://cloud.infini-ai.com/genstudio/model

进入模型广场，点击进入要选择的模型

![image-20250916093051359](assets/image-20250916093051359.png)

![image-20250916095031710](assets/image-20250916095031710.png)

点击调用说明，选择默认Interface：

【模型Name】model，如deepseek-v3.1

【API Key】第一步中申请的API Key

【推理url】Interface文档中url，复制Link，Note推理URL不带/chat/completions后缀：例如 https://cloud.infini-ai.com/maas/v1

3、登录元景万悟，点击Model Management-模型Import-None问芯穹-模型TypeLLM，依次填写模型Name、模型显示Name、API Key、推理url

![image-20250916095258594](assets/image-20250916095258594.png)



### Embedding

ImportSteps同LLM，只需在模型广场中筛选“文本向量”模型即可 

![image-20250916100450572](assets/image-20250916100450572.png)

点击调用说明，选择默认Interface：

【模型Name】model，如bge-m3

【API Key】第一步中申请的API Key

【推理url】Interface文档中url，复制Link，Note推理URL不带/chat/completions后缀：例如 https://cloud.infini-ai.com/maas/v1

![image-20250916100622249](assets/image-20250916100622249.png)

登录元景万悟，点击Model Management-模型Import-None问芯穹-模型TypeEmbedding，依次填写模型Name、API Key、推理url

![image-20250916100754602](assets/image-20250916100754602.png)

### Rerank

ImportSteps同LLM，只需在模型广场中搜索“rerank”模型即可 

![image-20250916101022272](assets/image-20250916101022272.png)

点击调用说明，选择默认Interface：

【模型Name】model，如bge-reranker-v2-m3

【API Key】第一步中申请的API Key

【推理url】Interface文档中url，复制Link，Note推理URL不带/chat/completions后缀：例如 https://cloud.infini-ai.com/maas/v1

![image-20250916101122451](assets/image-20250916101122451.png)

登录元景万悟，点击Model Management-模型Import-None问芯穹-模型TypeRerank，依次填写模型Name、API Key、推理url

![image-20250916100843750](assets/image-20250916100843750.png)

## 六、OpenAI-API-compatible

平台支持Import所有符合OpenAI协议的模型，包括联通元景、火山引擎等上述模型供应商提供的模型。具体Import方式详见上文。



