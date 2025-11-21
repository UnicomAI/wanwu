# HTTP Request

## 节点概述
核心Features：能够与任何基于HTTP协议的Service进行对话、交换Data。



## Configuration指南

##### 1. 请求方法

定义您希望对目标资源执行的OperationType。

| 方法       | Description                                                         | 典型App场景                                         |
| :--------- | :----------------------------------------------------------- | :--------------------------------------------------- |
| **GET**    | 从Service器获取指定资源。                                       | QueryUser信息、获取天气Data、读取文章List。           |
| **POST**   | 向Service器SubmitData，以创建新资源。                             | Submit表单、Upload新File、创建新订单。                   |
| **PUT**    | 向Service器Upload完整资源，用于**全量更新**已存在资源或**创建**资源（如果不存在）。 | 更新UserAll资料、替换整个File内容。                 |
| **PATCH**  | 对资源进行**部分修改**。                                     | 仅更新User的昵称、修改文章的标题。                   |
| **DELETE** | 请求Service器Delete指定的资源。                                   | DeleteUser、移除订单、清理File。                       |
| **HEAD**   | 与 GET 类似，但Service器**仅返回响应头**，不返回资源主体。      | 检查资源YesNo存在、获取资源的元Data（如大小、Type）。 |

> **提示**：在Configuration请求 URL 时，您可以通过Input `{{` 来唤出并使用Workflow中的变量，实现动态参数化。

##### 2. Request URL

目标 API 的完整 URL。例如：`https://api.example.com/v1/users/{{user_id}}`。

##### 3. 请求参数

附加在 URL 末尾的键值对（Query String），用于向Service器传递过滤、排序或分页等额外信息。

*   **Example**：在 URL `https://api.example.com/search?keyword=workflow&page=1` 中，`keyword=workflow` 和 `page=1` 即为请求参数。

##### 4. Request Header

包含请求的元Data，用于向Service器传递关于客户端、请求内容或期望响应格式的附加信息。

*   **常见Example**：
    *   `Content-Type`: 声明Request Body的Data格式（如 `application/json`）。
    *   `Accept`: 告知Service器客户端希望接收的响应格式（如 `application/json`）。
    *   `User-Agent`: 标识客户端的AppType、Operation系统和软件版本。

##### 5. 鉴权

为确保请求安全，防止未授权访问，本节点支持多种鉴权方式。

*   **Bearer Token**：
    *   **说明**：一种常用的令牌认证方式，常用于 OAuth 2.0 协议。
    *   **Configuration**：Input您的 Token 值。系统会自动将其添加到Request Header的 `Authorization` 字段中，格式为 `Authorization: Bearer <您的Token>`。
*   **自定义鉴权**：
    *   **说明**：提供更灵活的认证Configuration，以满足不同 API 的自定义需求。
    *   **Configuration**：
        *   **Key**: 认证的键名（如 `X-API-KEY`）。
        *   **Value**: 认证的键值。
        *   **Add To**: 选择认证信息的添加位置。
            *   **Header**: 添加到Request Header（推荐）。
            *   **Query**: 添加到 URL Query参数中。

##### 6. Request Body

- 仅在 `POST`, `PUT`, `PATCH` 等方法中有效，包含要发送给Service器的Data。您可以根据 API 要求选择合适的格式：

  *   **JSON**: 发送结构化的 JSON Data，最常用的 RESTful API 格式。

  *   **form-data**: 用于UploadFile，或发送包含二进制Data的复杂表单。

  *   **x-www-form-urlencoded**: 用于发送简单的键值对表单Data。

  *   **raw (Text/XML/HTML)**: 发送纯文本、XML 或 HTML 等原始格式Data。

##### 7. 超时Settings

- Settings请求的最大等待Time。若Service器在指定Time内未响应，请求将被判定为Failed，以避免长Time阻塞Workflow。

##### 8. 重试次数

- Settings请求Failed后的自动重试次数。通过自动重试机制，可以有效应对网络抖动或Service器临时故障，提高请求的最终Success率。

**9.Output**

- Output变量包括Response Body、Status码和响应头。

**10.异常忽略**

- 支持**异常忽略**Features。开启此Features后，如果试运行Workflow时此节点运行Failed，Workflow不会中断，而Yes继续运行后续下游节点。如果下游节点引用了此节点的Output内容，则使用此节点预先Configuration的默认Output内容。



## 典型App场景

HTTP Request节点YesWorkflow实现外部交互的**万能钥匙**，常见于以下四大类场景：

| 场景类别     | Description                                                    | 典型案例                                                |
| :----------- | :------------------------------------------------------ | :------------------------------------------------------ |
| **Data获取** | 从外部API拉取信息，为Workflow提供决策依据或丰富回复内容。 | Query实时天气、获取User在CRM中的资料、抓取最新新闻头条。 |
| **DataSubmit** | 将Workflow中产生的Data，发送到外部系统进行存储或处理。    | 将User反馈Submit到内部工单系统、将表单Data写入Data库。    |
| **Data更新** | 修改外部系统中已存在的Data，实现双向同步。              | 在CRM中更新客户的联系方式、修改订单的配送地址。         |
| **DataDelete** | 根据业务逻辑，请求外部系统Delete指定Data。                | Delete一个过期的优惠券、注销一个测试账户。                |

![image-20250823164251971](assets/image-20250823164251971.png)

