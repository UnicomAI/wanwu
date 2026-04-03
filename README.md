<div align="center">
  <img src="https://github.com/user-attachments/assets/6ceb4269-a861-4545-84db-bad322592156" style="width:45%; height:auto;" />
<p>
  <a href="#🚩 Core Function Modules">Core Function Modules</a> •
  <a href="#x1F3AF; Typical Application Scenarios">Typical Application Scenarios</a> •
  <a href="#🚀 Quick Start">Quick Start</a> •
  <a href="#x1F4D1; Using Wanwu">Using Wanwu</a> •
  <a href="#128172; Q & A">Q & A</a> •
  <a href="#x1F4E9; Contact Us">Contact Us</a> 
</p>
<p>
  <img alt="License" src="https://img.shields.io/badge/license-apache2.0-blue.svg">
  <img alt="Go Version" src="https://img.shields.io/badge/go-%3E%3D%201.24.0-blue">
  </a>
  <a href="https://github.com/UnicomAI/wanwu/releases">
    <img alt="Release Notes" src="https://img.shields.io/github/v/release/UnicomAI/wanwu?label=Release&logo=github&color=green">
  </a>
</p>
<p align="center">
    English |
    <a href="https://github.com/UnicomAI/wanwu/blob/main/README_CN.md">简体中文</a> |
    <a href="https://github.com/UnicomAI/wanwu/blob/main/README_繁體.md">繁體中文</a>
</p>
</div>


**Wanwu AI Agent Platform** is an **enterprise-grade** **one-stop** **commercially friendly** AI agent development platform designed for business scenarios. It is committed to providing enterprises with a safe, efficient, and compliant one-stop AI solution. With the core philosophy of "technology openness and ecological co-construction", we integrate cutting-edge technologies such as large language models and business process automation to build an AI engineering platform with a complete functional system covering model full life-cycle management, MCP, web search, AI agent rapid development, enterprise knowledge base construction, and complex workflow orchestration. The platform adopts a modular architecture design, supports flexible functional expansion and secondary development, and greatly reduces the application threshold of AI technology while ensuring the security and privacy protection of enterprise data. Whether it is for small and medium-sized enterprises to quickly build intelligent applications or for large enterprises to achieve intelligent transformation of complex business scenarios, the Wanwu AI Agent Platform can provide strong technical support to help enterprises accelerate the process of digital transformation, achieve cost reduction and efficiency improvement, and business innovation.

------

<div>
  <p align="center">
    <a href="https://www.bilibili.com/video/BV1HxpazNEAM"><img width="400" src="https://github.com/user-attachments/assets/54efe5d3-c28d-48fb-9a6e-d6ac536a1f95" /></a>
    <a href="https://www.bilibili.com/video/BV1HxpazNEAM"><img width="394" src="https://github.com/user-attachments/assets/d19831e6-10a3-4ee0-8caf-6c0ebe2af4a5" /></a>
  </p>
</div>

------

### 📢 Open Ecosystem

- [External Knowledge Base Compatibility](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/2.%E7%9F%A5%E8%AF%86%E5%BA%93/%E8%BF%9E%E6%8E%A5%E5%A4%96%E6%8E%A5%E7%9F%A5%E8%AF%86%E5%BA%93.md): Supports API-based import of knowledge bases created in Dify, with retrieval and recall in agents, Q&A, and workflows.
- [MCP Hub](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/2.%E8%B5%84%E6%BA%90%E5%BA%93%2FMCP%E6%9C%8D%E5%8A%A1.md): Supports importing and using MCP from different service providers.
- [Skills](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/2.%E8%B5%84%E6%BA%90%E5%BA%93%2FSkills.md): Supports creating and downloading Skills, with seamless integration to OpenClaw.
- [OpenClaw Sandbox](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/8.%E9%80%9A%E7%94%A8%E6%99%BA%E8%83%BD%E4%BD%93%2F%E6%9C%BA%E5%99%A8%E4%BA%BA%E5%8A%A9%E6%89%8B-OPENCLAW%2F%E5%A6%82%E4%BD%95%E5%9C%A8%E4%B8%87%E6%82%9F%E4%B8%AD%E6%8E%A5%E5%85%A5OpenClaw%E6%9C%BA%E5%99%A8%E4%BA%BA.md): We provide the option to deploy each “OpenClaw Robot” in a standalone Docker container. You can directly access your locally deployed OpenClaw robot within Yuanjing Wanwu.

------

### &#x1F525; Adopt a permissive and friendly Apache 2.0 License, supporting developers to freely expand and develop secondary

✔ **Enterprise-level engineering**: Provides a complete toolchain from model management to application landing, solving the "last mile" problem of LLM technology landing

✔ **Open-source ecological**: Adopt a permissive and friendly **Apache 2.0 License**, supporting developers to freely expand and develop

✔ **Full-stack technology support**: Equipped with a professional team to provide **architecture consulting, performance optimization** and full-cycle empowerment for ecological partners

✔ **Multi-tenant architecture**: Provides a multi-tenant account system to meet the core needs of users in cost control, data security isolation, business elasticity expansion, industry customization, rapid online and ecological collaboration

✔ **XinChuang adaptation**: The product has been awarded the **“Xinchuang AI Hardware and Software System Inspection Certificate“**，featuring hardware support for Huawei Kunpeng CPUs and software compatibility with domestic operating systems (e.g., openEuler, CULinux, Kylin) and databases (e.g., TiDB, OceanBase).

------

### 🚩 Core Function Modules

#### **1. Model Management (Model Hub)**
▸ Supports the unified access and lifecycle management of **hundreds of proprietary/open-source large models** (including GPT, Claude, Llama, MiniMax, etc.)

▸ Deeply adapts to **OpenAI API standards** and **Unicom Yuanjing** ecological models, realizing seamless switching of heterogeneous models

▸ Provides **multi-inference backend support** (vLLM, TGI, etc.) and **self-hosted solutions** to meet the computing power needs of enterprises of different scales

#### **2. MCP**
▸ **Standardized interfaces**: Enable AI models to seamlessly connect to various external tools (such as GitHub, Slack, databases, etc.) without the need to develop adapters for each data source separately

▸ **Built-in rich and selected recommendations**: Integrates 100+ industry MCP interfaces, making it easy for users to call up quickly and easily

#### **3. Web Search**
▸ **Real-time information acquisition**: Possesses powerful web search capabilities, capable of obtaining the latest information from the Internet in real-time. In question and answer scenarios, when a user's question requires the latest news, data, and other information, the platform can quickly search and return accurate results, enhancing the timeliness and accuracy of the answers

▸ **Multi-source data integration**: Integrates various Internet data sources, including news websites, academic databases, industry reports, etc. Through the integration and analysis of multi-source data, it provides users with more comprehensive and in-depth information. For example, in market research scenarios, relevant data can be obtained from multiple data sources at the same time for comprehensive analysis and evaluation

▸ **Intelligent search strategy**: Adopt intelligent search algorithms, automatically optimize search strategies based on user questions to improve search efficiency and accuracy. Support keyword search, semantic search and other search methods to meet the needs of different users. At the same time, intelligently sort and filter search results, prioritize the display of the most relevant and valuable information

#### **4. Visual Workflow (Workflow Studio)**
▸ Quickly build complex AI business processes through **low-code drag-and-drop canvas**

▸ Built-in **conditional branching, API, large model, knowledge base, code, MCP** and other nodes, support end-to-end process debugging and performance analysis

#### 5. <a href="#🚀High-precision RAG">High-precision RAG</a>
▸ Provides the whole process knowledge management capabilities of **knowledge base creation → document parsing → vectorization → retrieval → fine sorting**, supports **multiple formats** such as pdf/docx/txt/xlsx/csv/pptx documents, and also supports the capture and access of web resources

▸ Integrates **multi-modal retrieval**, **cascading segmentation** and **adaptive segmentation**, significantly improves the accuracy of Q&A

#### **6. AI Agent Development Framework (Agent Framework)**
▸ Can be based on the **function call (Function Calling)** agent construction paradigm, supports tool expansion, private knowledge base association and multi-round dialogue

▸ Support **online debugging**

#### **7. Backend as a Service (BaaS)**
▸ Provides **RESTful API**, supports deep integration with existing enterprise systems (OA/CRM/ERP, etc.)

▸ Provides **fine-grained permission control** to ensure stable operation in production environments

------

### &#x1F4E2; Function Comparison
|      Function      | Wanwu |             Dify.AI             |          Fastgpt           |             Ragflow             |    Coze open source version     |
| :----------------: | :---: | :-----------------------------: | :------------------------: | :-----------------------------: | :-----------------------------: |
|    Model import    |   ✅   |                ✅                |     ❌(Built-in models)     |                ✅                |       ❌(Built-in models)        |
|     RAG engine     |   ✅   |                ✅                |             ✅              |                ✅                |                ✅                |
|        MCP         |   ✅   |                ✅                |             ✅              | ✅(Need to install tools to use) |                ❌                |
| Direct OCR import  |   ✅   |                ❌                |             ❌              |                ❌                |                ❌                |
| Search enhancement |   ✅   | ✅(Need to install tools to use) |             ✅              | ✅(Need to install tools to use) |                ✅                |
|       Agent        |   ✅   |                ✅                |             ✅              |                ✅                |                ✅                |
|      Workflow      |   ✅   |                ✅                |             ✅              |                ✅                |                ✅                |
|  Local deployment  |   ✅   |                ✅                |             ✅              |                ✅                |                ✅                |
|  license friendly  |   ✅   |   ❌(Commercially restricted)    | ❌(Commercially restricted) |      Not fully open source      |                ✅                |
|      GraphRAG      |   ✅   |                ❌                |             ❌              |                ✅                |                ❌                |
|    Multi-tenant    |   ✅   |   ❌(Commercially restricted)    | ❌(Commercially restricted) |                ✅                | ✅(Users are not interconnected) |
> As of August 1, 2025.

------

### &#x1F3AF; Typical Application Scenarios

- **Intelligent Customer Service**: Realize high-accuracy business consultation and ticket processing based on RAG + Agent
- **Knowledge Management**: Build an exclusive enterprise knowledge base, support semantic search and intelligent summary generation
- **Process Automation**: Realize AI-assisted decision-making for business processes such as contract review and reimbursement approval through the workflow engine

The platform has been successfully applied in multiple industries such as **finance, industry, and government**, helping enterprises transform the theoretical value of LLM technology into actual business benefits. We sincerely invite developers to join the open source community and jointly promote the democratization of AI technology.

------

### 🚀 Quick Start

- The workflow module of the Wanwu AI Agent Platform uses the following project, you can go to its warehouse to view the details.
  - v0.1.8 and earlier: wanwu-agentscope project
  - v0.2.0 and later: [wanwu-workflow](https://github.com/UnicomAI/wanwu-workflow/tree/dev/wanwu-backend) project

- **Recommended Configuration:**
  - CPU: 8-core or 16-core; RAM: 32GB; Storage: 200GB or more; GPU: Not required.
  
- **Docker Installation (Recommended)**

1. Before the first run

    1.1 Copy the environment variable file
    ```bash
    cp .env.bak .env
    ```

    1.2 Modify the `WANWU_ARCH` and `WANWU_EXTERNAL_IP` variables in the .env file according to the system
    ```
    # amd64 / arm64
    WANWU_ARCH=amd64
    
    # external ip port (Note: if the browser accesses Wanwu deployed on a non-localhost server, you need to change localhost to the external IP, for example, 192.168.xx.xx)
    WANWU_EXTERNAL_IP=localhost
    ```

    1.3 Configure the `WANWU_BFF_JWT_SIGNING_KEY` variable in the .env file, a custom complex random string used for generating JWT tokens
    ```
    # bff
    WANWU_BFF_JWT_SIGNING_KEY=
    ```

    1.4 Create a Docker running network
    ```
    docker network create wanwu-net
    ```

2. Start the service (the image will be automatically pulled from Docker Hub during the first run)

    ```bash
    # For amd64 system:
    docker compose --env-file .env --env-file .env.image.amd64 up -d
    # For arm64 system:
    docker compose --env-file .env --env-file .env.image.arm64 up -d
    ```

3. Log in to the system: http://localhost:8081

    ```
    Default user: admin
    Default password: Wanwu123456
    ```

4. Stop the service

    ```bash
    # For amd64 system:
    docker compose --env-file .env --env-file .env.image.amd64 down
    # For arm64 system:
    docker compose --env-file .env --env-file .env.image.arm64 down
    ```

5. Having trouble pulling middleware or other Docker images? We've prepared a backup of the images on Netdisk. Please follow the instructions in its README file: [Wanwu Docker Image Backup](https://pan.baidu.com/e/1cupIcEP2RBwi_hOr4xQnFQ?pwd=ae86)

- **Source Code Start (Development)**

1. Based on the above Docker installation steps, start the system service completely

2. Take the backend bff-service service as an example

    2.1 Stop bff-service
    ```
    make -f Makefile.develop stop-bff
    ```

    2.2 Compile the bff-service executable file
    ```
    # For amd64 system:
    make build-bff-amd64
    # For arm64 system:
    make build-bff-arm64
    ```

    2.3 Start bff-service
    ```
    make -f Makefile.develop run-bff
    ```

------

### 📦 Sandbox Startup

The Wanwu sandbox can be used for features like creating Skills with a single command and needs to be started separately; note that when creating Skills, the selected model must have a context length >= 32000 when importing.

1. Based on the above Docker installation steps, complete the configuration before the first run

2. Start the sandbox (taking amd64 as an example)

   ```
   docker compose --env-file .env --env-file .env.image.amd64 -f docker-compose.wga-sandbox.yaml up -d
   ```

### ⬆️ Version Upgrade

1. Based on the above Docker installation steps, completely stop the system service

2. Update to the latest version of the code

    2.1 In the wanwu repository directory, update the code
    ```bash
    # Switch to the main branch
    git checkout main
    # Pull the latest code
    git pull
    ```

    2.2 Recopy the environment variable file (if there are changes to the environment variables, please modify them again)
    ```bash
    # Backup the current .env file
    cp .env .env.old
    # Copy the .env file
    cp .env.bak .env
    ```

3. Based on the above Docker installation steps, completely start the system service

------

### ➡️ Xinchuang Adaptation (TiDB & OceanBase)

1. Based on the above Docker installation steps, complete step before the first run

2. Modify the `WANWU_DB_NAME` variable in the .env file according to the database

3. Start the database (taking amd64 as an example)
   ```bash
   # tidb
   docker compose --env-file .env --env-file .env.image.amd64 -f docker-compose.tidb.yaml up -d
   # oceanbase
   docker compose --env-file .env --env-file .env.image.amd64 -f docker-compose.oceanbase.yaml up -d
   ```

4. Based on the above Docker installation steps, completely start the system service

✔ The product has been awarded the “Xinchuang AI Hardware and Software System Inspection Certificate,” featuring hardware support for Huawei Kunpeng CPUs and software compatibility with domestic operating systems (e.g., openEuler, CULinux, Kylin) and databases (e.g., TiDB, OceanBase).

------

### &#x1F4D1; Using Wanwu
To help you quickly get started with this project, we strongly recommend that you first check out the [ Documentation Operation Manual](https://github.com/UnicomAI/wanwu/tree/main/configs/microservice/bff-service/static/manual). We provide users with interactive and structured operation guides, where you can directly view operation instructions, interface documents, etc., greatly reducing the threshold for learning and use. The detailed function list is as follows:

| Feature                                                      | Detailed Description                                         |
| :----------------------------------------------------------- | :----------------------------------------------------------- |
| [Model Management](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/1.%E6%A8%A1%E5%9E%8B%E7%AE%A1%E7%90%86.md) | Supports users to import LLM, Embedding, and Rerank models from various model providers, including Unicom Yuanjing, OpenAI-API-compatible, Ollama, Tongyi Qianwen, Volcano Engine, and MiniMax. [Model Import Methods - Detailed Version](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/%E6%A8%A1%E5%9E%8B%E5%AF%BC%E5%85%A5%E6%96%B9%E5%BC%8F-%E8%AF%A6%E7%BB%86%E7%89%88.md) |
| [Knowledge Base](https://github.com/UnicomAI/wanwu/tree/main/configs/microservice/bff-service/static/manual/2.%E7%9F%A5%E8%AF%86%E5%BA%93) | In terms of document parsing capabilities: supports uploading of 12 file types and URL parsing; Supports private deployment and integration for document parsing via two methods: OCR and [a proprietary MinerU model (for scenarios like titles, tables, and formulas)](https://github.com/UnicomAI/DocParserServer/tree/main) ; document segmentation settings support both general segmentation and parent-child segmentation. In terms of optimization capabilities: supports metadata management 、Graph RAG and metadata filtering queries, supports adding, deleting, and modifying segmented content, supports setting keyword tags for segments to improve recall performance, supports segment enable/disable operations, and supports hit testing. In terms of retrieval capabilities: supports multiple retrieval modes including vector search, full-text search, and hybrid search. In terms of Q&A capabilities: supports automatic citation of sources and generating answers with both text and images.<br |
| [Resource Library](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/3.%E5%B7%A5%E5%85%B7%E5%B9%BF%E5%9C%BA.md) | Supports importing your own MCP services or custom tools for use in workflows and agents. |
| [Safety Guardrails](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/4.%E5%AE%89%E5%85%A8%E6%8A%A4%E6%A0%8F.md) | Users can create sensitive word lists to control the safety of the model's output. |
| [Text Q&A](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/5.%E6%96%87%E6%9C%AC%E9%97%AE%E7%AD%94.md) | A dedicated knowledge advisor based on a private knowledge base. It supports features like knowledge base management, Q&A, knowledge summarization, personalized parameter configuration, safety guardrails, and retrieval configuration to improve the efficiency of knowledge management and learning. Supports publishing text Q&A applications publicly or privately, and can be published as an API. |
| [Workflow](https://github.com/UnicomAI/wanwu/tree/main/configs/microservice/bff-service/static/manual/6.%E5%B7%A5%E4%BD%9C%E6%B5%81) | Extends the capabilities of agents. Composed of nodes, it provides a visual workflow editor. Users can orchestrate multiple different workflow nodes to implement complex and stable business processes. Supports publishing workflow applications publicly or privately, can be published as an API, and supports import/export. |
| [Agent](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/7.%E6%99%BA%E8%83%BD%E4%BD%93.md) | Create agents based on user scenarios and business requirements. Supports model selection, prompt setting, web search, knowledge base selection, MCP, workflows, and custom tools. Supports publishing agent applications publicly or privately, and can be published as an API and a Web URL. |
| [App Marketplace](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/8.%E5%BA%94%E7%94%A8%E5%B9%BF%E5%9C%BA.md) | Allows users to experience published applications, including Text Q&A, Workflows, and Agents. |
| [MCP Hub](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/9.MCP%E5%B9%BF%E5%9C%BA.md) | Features 100+ pre-selected industry-specific MCP servers, ready for immediate use. |
| [Template Plaza](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/10.%E6%A8%A1%E6%9D%BF%E5%B9%BF%E5%9C%BA.md) | Built-in with 50+ optimized industry prompts, available for immediate use. |
| [Settings](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/9.%E8%AE%BE%E7%BD%AE.md) | The platform supports multi-tenancy, allowing users to manage organizations, roles, users, and perform basic platform configuration. |
| [UniAI-GraphRAG](https://github.com/UnicomAI/wanwu/blob/66539378255f9a1da80b02a83e75c7a5155f7f87/configs/microservice/bff-service/static/manual/2.%E7%9F%A5%E8%AF%86%E5%BA%93/%E5%88%9B%E5%BB%BA%E7%9F%A5%E8%AF%86%E5%BA%93%E3%80%81%E9%97%AE%E7%AD%94%E5%BA%93/%E5%88%9B%E5%BB%BA%E7%9F%A5%E8%AF%86%E5%BA%93/%E7%9F%A5%E8%AF%86%E5%9B%BE%E8%B0%B1%E4%BD%BF%E7%94%A8%E8%AF%B4%E6%98%8E.md) | UniAI-GraphRAG integrates techniques such as domain knowledge ontology modeling, knowledge graph and community report construction, and Graph Retrieval-Augmented Generation to effectively enhance the completeness, logical coherence, and credibility of knowledge question answering. It significantly improves performance in complex QA scenarios like cross-document summarization and multi-hop relational reasoning. |

### 🚀High-precision RAG

**Wanwu RAG has completed its retrieval performance evaluation on the authoritative, publicly available industry benchmark, the MultiHop-RAG dataset**

<p align="center">
  <img width="584" alt="image" src="https://github.com/user-attachments/assets/8a267ba2-13e4-48fe-8ea8-4f24fb10dfc6" />
</p>

The F1 score serving as the comprehensive evaluation metric (the harmonic mean of precision and recall), are as follows: 

1）Wanwu RAG outperforms Dify by 14% 

2）Wanwu GraphRAG outperforms Dify by 17.2% 

3）Wanwu GraphRAG outperforms open-source LightRAG by 3.5%

------

### &#x1F4F0; TO DO LIST

- [ ] Skills
- [ ] Support importing databases into knowledge base
- [ ] General Agent
- [ ] A2A Protocol
- [ ] Agent and Model Evaluation
- [ ] Trace Tracking

------

### &#128172; Q & A

- **[Q] Error when starting Elastic (elastic-wanwu) on Linux system: Memory limited without swap.**
  **[A]** Stop the service, run `sudo sysctl -w vm.max_map_count=262144`, and then restart the service.
  
- **[Q] After the system services start normally, the mysql-wanwu-setup and elastic-wanwu-setup containers exit with status code Exited (0).**
  **[A]** This is normal. These two containers are used to complete some initialization tasks and will automatically exit after execution.
  
- **[Q] Regarding model import**
  **[A]** Taking the import of Unicom Yuanjing LLM as an example (the process is similar for importing OpenAI-API-compatible models, Embedding, or Rerank types):
  ```
  1. The Open API interface for Unicom Yuanjing MaaS Cloud LLM is, for example: https://maas.ai-yuanjing.com/openapi/compatible-mode/v1/chat/completions
  2. The API Key applied for by the user on Unicom Yuanjing MaaS Cloud looks like: sk-abc********************xyz
  3. Confirm that the API and Key can correctly request the LLM. Taking a request to yuanjing-70b-chat as an example:
      curl --location 'https://maas.ai-yuanjing.com/openapi/compatible-mode/v1/chat/completions' \
      --header 'Content-Type: application/json' \
      --header 'Accept: application/json' \
      --header 'Authorization: Bearer sk-abc********************xyz' \
      --data '{
              "model": "yuanjing-70b-chat",
              "messages": [{
                      "role": "user",
                      "content": "你好"
              }]
      }'
  4. Import the model:
  4.1 [Model Name] must be the model that can be correctly requested in the curl command above; for example, yuanjing-70b-chat.
  4.2 [API Key] must be the key that can be correctly requested in the curl command above; for example, sk-abc********************xyz (note: do not include the 'Bearer' prefix).
  4.3 [Inference URL] must be the URL that can be correctly requested in the curl command above; for example, https://maas.ai-yuanjing.com/openapi/compatible-mode/v1 (note: do not include the /chat/completions suffix).
  5. Importing an Embedding model is the same as importing an LLM as described above. Note that the inference URL should not include the /embeddings suffix.
  6. Importing a Rerank model is the same as importing an LLM as described above. Note that the inference URL should not include the /rerank suffix.
  ```

------

### &#x1F517; Acknowledgments

- [Coze](https://github.com/coze-dev)
- [LangChain](https://github.com/langchain-ai/langchain)
- [AIO Sandbox](https://github.com/agent-infra/sandbox)
- [OpenCode](https://github.com/anomalyco/opencode)

------

### ⚖️ License
The Yuanjing Wanwu AI Agent Platform is released under the Apache License 2.0.

------

### &#x1F4E9; Contact Us
| QQ Group1(Full):490071123                                    | QQ Group2(Full):1026898615                                         | QQ Group3:1019579243                                         |
| ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| <img width="183" height="258" alt="image" src="https://github.com/user-attachments/assets/010f1d68-78e9-446d-baf1-0a7339efb48e" /> | <img width="183" height="258" alt="image" src="https://github.com/user-attachments/assets/10796f69-5c18-4f21-adbb-b22b6ef88df2" /> | <img width="183" height="258" alt="image" src="https://github.com/user-attachments/assets/9badde43-b421-48ce-a998-1fc2e33bdc55" /> |