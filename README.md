<div align="center">
  <img src="https://github.com/user-attachments/assets/4788ed8f-eefc-4c19-aa77-7ec776743f3d" style="width:45%; height:auto;" />
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


Yuanjing Wanwu Agent Platform is an **all-in-one, commercial-friendly licensed agent development platform** designed for enterprise scenarios. Guided by the core philosophy of *open technology and co-construction of ecosystem*, we are committed to providing enterprises with **secure, efficient, and compliant AI solutions**.

Wanwu aims to deliver **all tooling capabilities required by Forward Deployed Engineer (FDEs)**, forming a full‑stack FDE toolchain. It covers core enterprise assets, centers on customers, deeply integrates capabilities into customer systems, drastically lowers the barrier to AI project delivery, and bridges the final mile from *build* to *field*. It simplifies every business decision and empowers every FDE.

------

<div>
  <p align="center">
    <a href="https://www.bilibili.com/video/BV1HxpazNEAM"><img width="400" src="https://github.com/user-attachments/assets/54efe5d3-c28d-48fb-9a6e-d6ac536a1f95" /></a>
    <a href="https://www.bilibili.com/video/BV1HxpazNEAM"><img width="394" src="https://github.com/user-attachments/assets/d19831e6-10a3-4ee0-8caf-6c0ebe2af4a5" /></a>
  </p>
</div>

------

### 🌟 Full‑Stack FDE Toolchain: 5 Core Capabilities

To address complex enterprise scenarios, Wanwu provides 5 core agent capabilities to solve delivery pain points, making AI not only thinkable but also doable.

#### 1️⃣ RAG/Knowledge Base Agent

<p align="center">
  <img width="900" alt="rag" src="https://github.com/user-attachments/assets/7bbc16b2-e7d0-49cb-89e3-5c395416c028" />
</p>

Solves scattered documents and gives AI reliable memory. Provides end‑to‑end knowledge management for massive enterprise documents and policies, building a high‑precision, memory‑enabled knowledge brain to prevent AI hallucinations.

- **High‑precision parsing & retrieval:** Supports 12 file formats and URL crawling; OCR and MinerU model private deployment; multimodal retrieval, cascaded/adaptive chunking, intelligent ranking, illustrated generation, and source citation.
- **GraphRAG enhancement:** Built‑in UniAI‑GraphRAG with domain ontology modeling, greatly improving completeness and logic in cross‑document summarization and multi‑hop reasoning, with industry‑leading F1 score.
- **External knowledge base compatibility:** Supports API import of knowledge bases created in Dify for retrieval in agents, chat, and workflows.

#### 2️⃣ Ontology Agent

<p align="center">
  <img width="900" alt="ontology" src="https://github.com/user-attachments/assets/7a582447-9edc-4cce-ace4-29291b5ad2ac" />
</p>

Handles structured data and enables multi‑step reasoning and decision‑making. Breaks the limitation that LLMs only understand text, adapting to complex structured business data.

- **Deep reasoning & decision‑making:** Automatically builds business knowledge networks from enterprise data and documents, empowering AI with deep reasoning and closed‑loop action. It elevates LLMs from knowledge Q&A to business analysis brains.

#### 3️⃣ Workflow Agent

<p align="center">
  <img width="900" alt="workflow" src="https://github.com/user-attachments/assets/8064a56a-63b2-493c-9dea-e0987c3cc64d" />
</p>

Manages complex processes and ensures AI follows business rules.

For contract review, expense approval, and other complex workflows, standardizes AI execution paths via low‑code for stable delivery.

- **Visual orchestration:** Low‑code drag‑and‑drop canvas with built‑in conditional branches, APIs, LLMs, knowledge bases, code, MCP, etc. Supports end‑to‑end debugging and performance analysis.
- **Zero‑code orchestration closed‑loop:** Industry‑first zero‑code Skill invocation in agent development, closing the loop from intent recognition to skill execution. Flexibly calls built‑in tools, MCP, workflows, etc., parses hundreds of pages in seconds, and displays results in a unified workspace.

#### 4️⃣ GUI Agent ([Download Client](https://pan.baidu.com/s/1xqQcjwAbrd1YVS2XHG81rw?pwd=ae86))

<p align="center">
  <img width="900" alt="gui" src="https://github.com/user-attachments/assets/0576488a-9f55-49fa-8b21-b7810ac3faac" />
</p>

Operates various applications without APIs. Eliminates integration barriers for legacy or non‑API systems by giving AI vision and click capabilities.

- **UI‑level interaction:** AI directly operates application interfaces without underlying API integration.
- **Sandbox support:** Isolated Docker containers for each bot to safely execute UI operations.

#### 5️⃣ General Agent + Skill Development

<p align="center">
  <img width="900" alt="wga" src="https://github.com/user-attachments/assets/340c131c-11ed-4547-a395-250f40770160" />
</p>

Unifies all capabilities with natural language. Industry‑first dual engine of General Agent + Vertical‑scene Skills, building a super agent that is both knowledgeable and professional for complex interactive needs.

- **All‑around brain & minimal construction:** Professional analyst‑level multi‑step reasoning; one‑sentence Skill creation turns business experience into a dedicated toolbox via natural language; one‑click conversion of platform apps to Skills.

------

### 🚀 3 Deployment Methods: Reach Business On‑Site

After capability building, Wanwu provides 3 deployment paths to minimize FDE delivery difficulty:

#### 📦 Method 1: Out‑of‑the‑Box Platform

Use directly via visual interface; no coding required for creating agents, workflows, and Q&A. Zero‑threshold AI productivity for rapid on‑site validation and delivery.

#### 🔗 Method 2: API Seamless Integration

RESTful API (BaaS) for embedding into OA, CRM, ERP, etc. Fine‑grained permission control enables deep AI integration without changing user habits.

#### 🖥️Method 3: Skill + UniClaw Dedicated Client

For high‑privilege scenarios (local PC control, DingTalk messages, etc.). FDEs develop Skills and execute via UniClaw to handle cross‑system high‑privilege on‑site operations.

UniClaw download: https://maas.ai-yuanjing.com/app/uniclaw/uniclaw-official.html

------

### 🛠️ Infrastructure & Ecosystem

- 🔥 **Apache‑2.0 License**: Free extension, secondary development, and commercial use.
- ✔ **Model Hub**: Unified access to hundreds of proprietary/open‑source models; deep OpenAI API compatibility and Yuanjing ecosystem support; multiple inference backends.
- ✔ **Skill Plaza**: 100+ built‑in industry Skills ready to use; no adapters needed for external capabilities.
- ✔ **Web Search**: Real‑time information, multi‑source integration, intelligent retrieval strategies.
- ✔ **Multi‑tenant architecture**: Isolated accounts for cost control, data security, and elastic scaling.
- **✔ Xinchuang compliance**: Certified *Xinchuang AI Software/Hardware System Inspection Certificate*. Supports Kunpeng CPUs, Euler, Kylin, CULinux, TiDB, OceanBase, etc.

------

### 🚩 Core Function Modules

#### **1. Model Management (Model Hub)**
▸ Supports the unified access and lifecycle management of **hundreds of proprietary/open-source large models** (including GPT, Claude, Llama, etc.)

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

#### **6. General Agent & Skills Orchestration Framework** 

▸ **Dual-Engine Mode**: Breaks the limitation of traditional agents "having a brain but no hands," upgrading to a "General Agent + Vertical Scenario Skills" dual-engine platform to create an enterprise-level super agent that is both "knowledgeable" and "professional" 

▸ **Almighty Brain**: The general agent, as the core engine, has now demonstrated professional analyst-level multi-step reasoning and information integration capabilities in complex scenarios like deep research and data analysis 

▸ **Minimalist Skill Building**: Supports **"one-sentence Skill creation"**. No code is required; just describe your needs in natural language to automatically generate vertical scenario skills, turning business experience into a dedicated "toolbox" 

▸ **Zero-Code Orchestration Closed Loop**: The **industry's first to support zero-code Skill calling during agent development**. Directly associate Skills in the visual interface to achieve a perfect closed loop from "intent recognition" to "skill execution" 

▸ **On-Demand Toolbox**: Flexibly configure and call built-in tools, Skills, MCP, workflows, and other agents, making AI not only "think" but also "act" ▸ **Read Hundreds of Pages in Seconds**: Supports uploading various files; the general agent can quickly parse them and conduct precise deep Q&A and interaction based on the files 

▸ **Unified Workspace**: Provides a unified destination for outcomes, neatly displaying all interactively generated files with support for online preview and one-click download 

▸ **Basic Development Paradigm**: Still supports traditional Agent construction based on **function calling**, supporting private knowledge base association and multi-round online debugging

#### 7.Wanwu Ontology Agent

▸ Automatically constructs business knowledge networks from enterprise data and documents, empowering AI with deep reasoning and closed-loop action capabilities to truly understand business and make decisions.

#### **8. Backend as a Service (BaaS)**
▸ Provides **RESTful API**, supports deep integration with existing enterprise systems (OA/CRM/ERP, etc.)

▸ Provides **fine-grained permission control** to ensure stable operation in production environments

------

### &#x1F4E2; Function Comparison
|                    Function                    | Wanwu |             Dify.AI             |          Fastgpt           |             Ragflow             |    Coze open source version     |
| :--------------------------------------------: | :---: | :-----------------------------: | :------------------------: | :-----------------------------: | :-----------------------------: |
|                  Model import                  |   ✅   |                ✅                |     ❌(Built-in models)     |                ✅                |       ❌(Built-in models)        |
|               Direct OCR import                |   ✅   |                ❌                |             ❌              |                ❌                |                ❌                |
|                   RAG engine                   |   ✅   |                ✅                |             ✅              |                ✅                |                ✅                |
|                    GraphRAG                    |   ✅   |                ❌                |             ❌              |                ✅                |                ❌                |
|                 Ontology Agent                 |   ✅   |                ❌                |             ❌              |                ❌                |                ❌                |
|    Multi-Agent Orchestration & Development     |   ✅   |                ❌                |             ✅              |                ✅                |                ❌                |
| General Agent & Skills Orchestration Framework |   ✅   |                ❌                |             ❌              |                ❌                |                ❌                |
|                     Agent                      |   ✅   |                ✅                |             ✅              |                ✅                |                ✅                |
|                    Workflow                    |   ✅   |                ✅                |             ✅              |                ✅                |                ✅                |
|                      MCP                       |   ✅   |                ✅                |             ✅              | ✅(Need to install tools to use) |                ❌                |
|               Search enhancement               |   ✅   | ✅(Need to install tools to use) |             ✅              | ✅(Need to install tools to use) |                ✅                |
|                Local deployment                |   ✅   |                ✅                |             ✅              |                ✅                |                ✅                |
|                  Multi-tenant                  |   ✅   |   ❌(Commercially restricted)    | ❌(Commercially restricted) |                ✅                | ✅(Users are not interconnected) |
|                license friendly                |   ✅   |   ❌(Commercially restricted)    | ❌(Commercially restricted) |      Not fully open source      |                ✅                |
> As of May 15, 2026.

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

- **Model Requirements:**
  - When using WanwuBot (General Agent) or creating Skills with a single command, the selected model must have a context length >= 32000 when importing.

- **Security Statement:**
  - All middleware components (MySQL, Redis, MinIO, Kafka, Elasticsearch, etc.) support custom passwords configured in the `.env` file.
  - For front-end and back-end user password transmission, RSA asymmetric encryption is used to encrypt passwords. The key pair is automatically generated on first service startup; custom key pairs can also be configured via `WANWU_BFF_LOGIN_RSA_PATH` in `.env`.
  - It is strongly recommended to change all default passwords and keep them secure before deploying in a production environment to avoid security risks associated with default credentials.

- **Docker Installation (Recommended)**

1. Before the first run

    1.1 Copy the environment variable file
    ```bash
    cp .env.example .env
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

    1.4 Copy environment variable file (Ontology Agent)
    ```bash
    cp .env.ontology.example .env.ontology
    ```

    1.5 (Optional) Generate custom RSA keys before the first run (Ontology Agent)
    > Skip this step to use the default keys baked into the images. To use independent keys in production, generate them as follows.

    - 1.5.1 Generate RSA key pair
      ```bash
      ./configs/microservice/ontology/vega-server/generate-keys.sh configs/microservice/ontology/vega-server
      ```

    - 1.5.2 Generate frontend public key configuration (cross-platform, requires Node environment)
      ```bash
      node configs/microservice/ontology/vega-server/generate-public-key-js.js
      ```

    1.6 Create a Docker running network
    ```
    docker network create wanwu-net
    ```

2. Start the service (the image will be automatically pulled from Docker Hub during the first run)

    ```bash
    # For amd64 system:
    docker compose --env-file .env --env-file .env.ontology --env-file .env.image.amd64 up -d
    # For arm64 system:
    docker compose --env-file .env --env-file .env.ontology --env-file .env.image.arm64 up -d
    ```

3. Log in to the system: http://localhost:8081

    ```
    Default user: admin
    Default password: Wanwu123456
    ```

4. Stop the service

    ```bash
    # For amd64 system:
    docker compose --env-file .env --env-file .env.ontology --env-file .env.image.amd64 down
    # For arm64 system:
    docker compose --env-file .env --env-file .env.ontology --env-file .env.image.arm64 down
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

    2.2 Recopy environment variable file (if there are changes to the environment variables, please modify them again)
    ```bash
    # Backup the current .env file
    cp .env .env.old
    # Copy the .env file
    cp .env.example .env
    ```

    2.3 Recopy environment variable file (if there are changes to the environment variables, please modify them again)
    ```bash
    # Backup current .env.ontology file
    cp .env.ontology .env.ontology.old
    # Copy .env.ontology file
    cp .env.ontology.example .env.ontology
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
| [General Agent](https://github.com/UnicomAI/wanwu/tree/main/configs/microservice/bff-service/static/manual/8.%e9%80%9a%e7%94%a8%e6%99%ba%e8%83%bd%e4%bd%93) | The platform deeply integrates advanced capabilities such as deep research and data analysis, achieving a comprehensive leap from simple Q&A to complex business processing, creating your all-around AI digital assistant. |
| [Ontology Agent](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/10.%E6%9C%AC%E4%BD%93%E6%99%BA%E8%83%BD%E4%BD%93/%E6%95%B0%E6%8D%AE%E8%BF%9E%E6%8E%A5/%E8%BF%9E%E6%8E%A5%E7%AE%A1%E7%90%86.md) | Automatically constructs business knowledge networks from enterprise data and documents, empowering AI with deep reasoning and closed-loop action capabilities to truly understand business and make decisions. |
| [Model Management](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/1.%E6%A8%A1%E5%9E%8B%E7%AE%A1%E7%90%86.md) | Supports users to import LLM, Embedding, and Rerank models from various model providers, including Unicom Yuanjing, OpenAI-API-compatible, Ollama, Tongyi Qianwen, and Volcano Engine. [Model Import Methods - Detailed Version](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/%E6%A8%A1%E5%9E%8B%E5%AF%BC%E5%85%A5%E6%96%B9%E5%BC%8F-%E8%AF%A6%E7%BB%86%E7%89%88.md) |
| [Knowledge Base](https://github.com/UnicomAI/wanwu/tree/main/configs/microservice/bff-service/static/manual/2.%E7%9F%A5%E8%AF%86%E5%BA%93) | In terms of document parsing capabilities: supports uploading of 12 file types and URL parsing; Supports private deployment and integration for document parsing via two methods: OCR and [a proprietary MinerU model (for scenarios like titles, tables, and formulas)](https://github.com/UnicomAI/DocParserServer/tree/main) ; document segmentation settings support both general segmentation and parent-child segmentation. In terms of optimization capabilities: supports metadata management 、Graph RAG and metadata filtering queries, supports adding, deleting, and modifying segmented content, supports setting keyword tags for segments to improve recall performance, supports segment enable/disable operations, and supports hit testing. In terms of retrieval capabilities: supports multiple retrieval modes including vector search, full-text search, and hybrid search. In terms of Q&A capabilities: supports automatic citation of sources and generating answers with both text and images.<br |
| [Resource Library](https://github.com/UnicomAI/wanwu/blob/main/configs/microservice/bff-service/static/manual/3.%E5%B7%A5%E5%85%B7%E5%B9%BF%E5%9C%BA.md) | Supports importing your own MCP services or custom tools or skills for use in workflows and agents. |
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

- [x] General Agent
- [x] Skills
- [x] Trace Tracking
- [ ] Support importing databases into knowledge base
- [ ] Agent and Model Evaluation
- [ ] Long-term Memory
- [ ] A2A Protocol

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
- [KWeaver Core](https://github.com/kweaver-ai/kweaver-core)

------

### ⚖️ License
The Yuanjing Wanwu AI Agent Platform is released under the Apache License 2.0.

------

### &#x1F4E9; Contact Us
| QQ Group1(Full):490071123                                    | QQ Group2(Full):1026898615                                         | QQ Group3(Full):1019579243                                         | QQ Group3:967463724                                         |
| ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| <img width="183" height="258" alt="image" src="https://github.com/user-attachments/assets/163d6580-af84-4fe4-9b51-7effb4153dd8" /> | <img width="183" height="258" alt="image" src="https://github.com/user-attachments/assets/03d10f7c-7460-485e-9f17-b3135d460dd0" /> | <img width="183" height="258" alt="image" src="https://github.com/user-attachments/assets/6cf67753-899c-418d-971b-f43fc9b5bada" /> | <img width="183" height="258" alt="image" src="https://github.com/user-attachments/assets/f5aef62d-61ca-4bc1-9217-ba4a358a7d08" /> |