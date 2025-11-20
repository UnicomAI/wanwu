# ModelImportTutorial

## 一、联通元景


### LLM


1. 

Configuration

联通元景

API

Key

https://maas.ai-yuanjing.com/

RegisterLogin 后,

进入以下 Path:

元景 MaaSPlatform--Tools 箱--API

KeyManagement,

NeworCopy 已有 Key.

![image-20250725111753372](assets/image-20250725111753372.png)

ClickCreate,

填写 AppName 及 AppDescription.

OK 后,

即可 GenerateAPI

Key.

![image-20250725111847121](assets/image-20250725111847121.png)

![image-20250725111926508](assets/image-20250725111926508.png)

2. GetParameters:

Select 想要 Import Model

https://maas.ai-yuanjing.com/aibase/portal/service

进入 Model 广场-SelectModel-ClickViewDetails

![image-20250725112016244](assets/image-20250725112016244.png)

[ModelName]RequestParameters-model,

如 deepseek-r1-distill-qwen-32b

[API

Key]First 步 Application API

Key

[推理 url]RequestInterface-RequestURL:

Note 推理 URL 不带

/chat/completions

后缀,

例如:

https://maas-api.ai-yuanjing.com/openapi/compatible-mode/v1

![image-20250725112538728](assets/image-20250725112538728.png)

3. Login 元景万悟,

ClickModel

Management-ModelImport-联通元景-ModelTypeLLM,

依次填写 ModelName、API

Key、推理 url

![image-20250725112504492](assets/image-20250725112504492.png)

### Embedding


ImportSteps 同 LLM,

只需 In Documentation, 心 SearchEmbeddingModel 即可

https://maas.ai-yuanjing.com/doc/pages/216556678/#%E8%AF%B7%E6%B1%82%E6%8E%A5%E5%8F%A3

![image-20250725112754696](assets/image-20250725112754696.png)

[ModelName]RequestExample model,

如 qwen3-embed-0. 6b

[API

Key]First 步 Application API

Key

[推理 url]Copy 调用 Interface:

Note 推理 URL 不带

/embeddings

后缀,

例如:

https://maas-api.ai-yuanjing.com/openapi/compatible-mode/v1

![image-20250725112837337](assets/image-20250725112837337.png)

Login 元景万悟,

ClickModel

Management-ModelImport-联通元景-ModelTypeEmbedding,

依次填写 ModelName、API

Key、推理 url

![image-20250725112716488](assets/image-20250725112716488.png)

### Rerank


ImportSteps 同 LLM,

只需 In Documentation, 心 SearchrerankModel 即可

https://maas.ai-yuanjing.com/doc/pages/216556678/#%E8%AF%B7%E6%B1%82%E6%8E%A5%E5%8F%A3

![image-20250725112754696](assets/image-20250725112754696.png)

[ModelName]RequestExample model,

如 bge-m3

[API

Key]First 步 Application API

Key

[推理 url]Copy 调用 Interface:

Note 推理 URL 不带

/rerank

后缀,

例如:

https://maas-api.ai-yuanjing.com/openapi/bge/v1

![image-20250725113812334](assets/image-20250725113812334.png)

3. Login 元景万悟,

ClickModel

Management-ModelImport-联通元景-ModelTypererank,

依次填写 ModelName、API

Key、推理 url

![image-20250725113158923](assets/image-20250725113158923.png)

### OCR


可 App 于 Knowledge

BaseFileUpload-ParseWay-OCRParse

[ModelName]固定 for

yuanjingOcr

[API

Key]First 步 Application API

Key

[推理 url]Copy 调用 Interface,

固定 for

https://maas-api.ai-yuanjing.com/openapi/v1

Login 元景万悟,

ClickModel

Management-ModelImport-联通元景-ModelTypeOCR,

依次填写 ModelName、API

Key、推理 url

![image-20250731113537320](assets/image-20250731113537320.png)

### GUI


[ModelName]固定 for

gui_agent_v1

[API

Key]First 步 Application API

Key

[推理 url]Copy 调用 Interface,

固定 for

https://maas-gz-api.ai-yuanjing.com/openapi/v1

or

https://maas-api.ai-yuanjing.com/openapi/v1

Login 元景万悟,

ClickModel

Management-ModelImport-联通元景-ModelTypeGUI,

依次填写 ModelName、API

Key、推理 url

![image-20250904100050884](assets/image-20250904100050884.png)

### pdfDocument


ParsingModel

可 App 于 Knowledge

BaseFileUpload-ParseWay-ModelParse

[ModelName]固定 for

pdf-parser

[API

Key]First 步 Application API

Key

[推理 url]Copy 调用 Interface,

固定 for

https://maas-api.ai-yuanjing.com/openapi/v1

Login 元景万悟,

ClickModel

Management-ModelImport-联通元景-ModelTypepdfDocument

ParsingModel,

依次填写 ModelName、API

Key、推理 url

![image-20250919105832500](assets/image-20250919105832500.png)

## 二、阿里通义千问


### LLM


1. 

Configuration

阿里 Cloud 百炼

API

Key

https://bailian.console.aliyun.com/?tab=model#/api-key

Login 阿里 Cloud 百炼 Platform--新 User 开通.

首次开通需实名制,

具体 Steps 详见 https: //help. aliyun. com/zh/account/user-guide/individual-identities

![image-20250725115411209](assets/image-20250725115411209.png)

![image-20250725115523681](assets/image-20250725115523681.png)

![image-20250725115829504](assets/image-20250725115829504.png)

![image-20250725115931936](assets/image-20250725115931936.png)

![image-20250725120024146](assets/image-20250725120024146.png)

2. GetParameters:

Select 想要 Import Modelhttps: //bailian. console. aliyun. com/? tab=model#/model-market

进入 Model 广场-SelectModel-ClickViewDetailsandAPIReference

![image-20250725124441761](assets/image-20250725124441761.png)

[ModelName]ViewDetails 后 code,

如 qwen-max

[API

Key]First 步 Application API

Key

[推理 url]View“APIReference”,

Use SDK 调用:

例如

https://dashscope.aliyuncs.com/compatible-mode/v1

![image-20250725124515869](assets/image-20250725124515869.png)

![image-20250725124335928](assets/image-20250725124335928.png)

3. Login 元景万悟,

ClickModel

Management-ModelImport-通义千问-ModelTypeLLM,

依次填写 ModelName、API

Key、推理 url

![ea351b3290750d7b942ab4d43a4828c2](assets/ea351b3290750d7b942ab4d43a4828c2.png)

### Embedding


ImportSteps 同 LLM,

只需 In Model 广场, SearchEmbeddingModel 即可

![image-20250725120728152](assets/image-20250725120728152.png)

[ModelName]View“APIReference”-ModelName,

如 text-embedding-v4

[API

Key]First 步 Application API

Key

[推理 url]View“APIReference”,

Fast 入门 ExampleCodebase_url:

例如

https://dashscope.aliyuncs.com/compatible-mode/v1

![image-20250725120945990](assets/image-20250725120945990.png)

![image-20250725124746276](assets/image-20250725124746276.png)

Login 元景万悟,

ClickModel

Management-ModelImport-通义千问-ModelTypeEmbedding,

依次填写 ModelName、API

Key、推理 url

![e2789058713fa030854c3d979cec0dea](assets/e2789058713fa030854c3d979cec0dea.png)

### Rerank


ImportSteps 同 LLM,

只需 In Model 广场, SearchSortModel 即可

![image-20250725121515305](assets/image-20250725121515305.png)

[ModelName]View“APIReference”-ModelName,

,

如 gte-rerank-v2

[API

Key]First 步 Application API

Key

[推理 url]Copy 调用 Example url:

Note 推理 URL 不带/services/rerank/text-rerank/text-rerank 后缀,

如:

https://dashscope.aliyuncs.com/api/v1

![image-20250725121545692](assets/image-20250725121545692.png)

![image-20250725123957256](assets/image-20250725123957256.png)

Login 元景万悟,

ClickModel

Management-ModelImport-通义千问-ModelTypeRerank,

依次填写 ModelName、API

Key、推理 url

![c7fabdfa0b33475c2a1735d3116786f6](assets/c7fabdfa0b33475c2a1735d3116786f6.png)

## 三、Ollama


### LLM


```

1.

LocalOllamaDeployment:

https://github.com/ollama/ollama

2.

UserininLocalStartOllamaService,

并 ConfirmService 正常 Start,

并 ConfirmModel 已 Load,

并 ConfirmModel 可正确 Request,

以 Requestqwen2. 5: 0. 5bfor 例:

curl

- -location

'http: //Localip: 11434/v1/chat/completions'

\

- -header

'Content-Type:

application/json'

\

- -header

'Accept:

application/json'

\

- -data

'{

"model":

"qwen2.5:0.5b",

"messages":

[{

"role":

"user",

"content":

"你好"

}]

}'

3.

AccessIP(Notelocalhost 要换 Cost 机 LANor 对外 IP,

例如192. 168. 0. xx,

不能 Yeslocalhostor127. 0. 0. 1)

4.

ImportModel:

4. 1[ModelName]必须 for 上述 curlCan 正确 Request model;

例如

qwen2.5:0.5b

4. 2[AccessIP]Notelocalhost 要换 Cost 机 LANor 对外 IP,

例如192. 168. 0. xx,

不能 Yeslocalhostor127. 0. 0. 1

4. 3[推理 URL]必须 for 上述 curlCan 正确 Request url;

例如

http: //Localip: 11434/v1(Note 不带

/chat/completions

后缀)

```

### Embedding


ImportEmbeddingModel 同上述 ImportLLM,

Note 推理 URL 不带

/embeddings

后缀

## 四、火山引擎-豆包


### LLM


1. GetParameters:

Select 想要 Import Modelhttps: //console. volcengine. com/ark/region: ark+cn-beijing/model? vendor=Bytedance&view=DEFAULT_VIEW

进入 Model 广场-SelectModel-悬停

![image-20250725125312203](assets/image-20250725125312203-1754018422086-1.png)

[ModelName]悬停后 Showcase Model

ID,

如 doubao-seed-1-6-thinking-250715

[API

Key]ClickAPI 接入,

CreateAPI

Key

[推理 url]GetAPI

KEY 后,

Click“SelectUse ”,

CopyLink,

Note 推理 URL 不带/chat/completions 后缀:

例如

https://ark.cn-beijing.volces.com/api/v3

![image-20250725125451548](assets/image-20250725125451548-1754018422087-2.png)

![image-20250725125558032](assets/image-20250725125558032-1754018422087-3.png)

![image-20250725162735374](assets/image-20250725162735374-1754018422087-4.png)

2. Login 元景万悟,

ClickModel

Management-ModelImport-火山引擎-ModelTypeLLM,

依次填写 ModelName、API

Key、推理 url

![72776f6bff2edbd709e861ddfd8e64ee](assets/72776f6bff2edbd709e861ddfd8e64ee.png)

### Embedding


ImportSteps 同 LLM,

只需 In Model 广场, Search 向量 Model 即可

![image-20250725130019497](assets/image-20250725130019497-1755155158130-1.png)

![image-20250725130126264](assets/image-20250725130126264-1755155158130-2.png)

[ModelName]悬停后 Showcase Model

ID,

如 doubao-embedding-large-text-250515

[API

Key]ClickAPI 接入,

CreateAPI

Key

[推理 url]GetAPI

KEY 后,

Click“SelectUse ”,

Copybase_url,

Note 推理 URL 不带/embeddings 后缀:

例如

https://ark.cn-beijing.volces.com/api/v3

![image-20250725130205458](assets/image-20250725130205458-1755155158130-3.png)

Login 元景万悟,

ClickModel

Management-ModelImport-火山引擎-ModelTypeEmbedding,

依次填写 ModelName、API

Key、推理 url

![cc6540b0cf4feca1d55551b2bea9f17f](assets/cc6540b0cf4feca1d55551b2bea9f17f-1755155158130-4.png)

## 五、None 问芯穹


### LLM


1. 

Configuration

None 问芯穹

API

Key

https://cloud.infini-ai.com/iam/secret/key

![image-20250916095827688](assets/image-20250916095827688.png)

![image-20250916095851399](assets/image-20250916095851399.png)

2. GetParameters:

Select 想要 Import Modelhttps: //cloud. infini-ai. com/genstudio/model

进入 Model 广场,

Click 进入要 Select Model

![image-20250916093051359](assets/image-20250916093051359.png)

![image-20250916095031710](assets/image-20250916095031710.png)

Click 调用 Instructions,

SelectDefaultInterface:

[ModelName]model,

如 deepseek-v3. 1

[API

Key]First 步 Application API

Key

[推理 url]InterfaceDocumentationurl,

CopyLink,

Note 推理 URL 不带/chat/completions 后缀:

例如

https://cloud.infini-ai.com/maas/v1

3. Login 元景万悟,

ClickModel

Management-ModelImport-None 问芯穹-ModelTypeLLM,

依次填写 ModelName、ModelShowName、API

Key、推理 url

![image-20250916095258594](assets/image-20250916095258594.png)

### Embedding


ImportSteps 同 LLM,

只需 In Model 广场, Filter“Text 向量”Model 即可

![image-20250916100450572](assets/image-20250916100450572.png)

Click 调用 Instructions,

SelectDefaultInterface:

[ModelName]model,

如 bge-m3

[API

Key]First 步 Application API

Key

[推理 url]InterfaceDocumentationurl,

CopyLink,

Note 推理 URL 不带/chat/completions 后缀:

例如

https://cloud.infini-ai.com/maas/v1

![image-20250916100622249](assets/image-20250916100622249.png)

Login 元景万悟,

ClickModel

Management-ModelImport-None 问芯穹-ModelTypeEmbedding,

依次填写 ModelName、API

Key、推理 url

![image-20250916100754602](assets/image-20250916100754602.png)

### Rerank


ImportSteps 同 LLM,

只需 In Model 广场, Search“rerank”Model 即可

![image-20250916101022272](assets/image-20250916101022272.png)

Click 调用 Instructions,

SelectDefaultInterface:

[ModelName]model,

如 bge-reranker-v2-m3

[API

Key]First 步 Application API

Key

[推理 url]InterfaceDocumentationurl,

CopyLink,

Note 推理 URL 不带/chat/completions 后缀:

例如

https://cloud.infini-ai.com/maas/v1

![image-20250916101122451](assets/image-20250916101122451.png)

Login 元景万悟,

ClickModel

Management-ModelImport-None 问芯穹-ModelTypeRerank,

依次填写 ModelName、API

Key、推理 url

![image-20250916100843750](assets/image-20250916100843750.png)

## 六、OpenAI-API-compatible


PlatformSupportImport 所有符合 OpenAIAgreement Model,

Including 联通元景、火山引擎 etc.

具体 ImportWay 详见上文.