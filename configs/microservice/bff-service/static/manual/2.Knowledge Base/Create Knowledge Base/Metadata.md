# 元Data管理

元DataYesDescription文档关键信息的“Data标签”。它不包含文档的具体内容，而Yes用来定义文档的属性，帮助您对文档进行结构化管理。

### 元Data管理

- **Key：**元Data字段Yes构成文档元Data的基本单元，它们为文档信息提供了标准化的分类和存储方式。通过定义和使用不同的字段，我们可以系统化地捕捉和管理文档的关键信息。
- **Type：**
  - **字符串**（String）：文本值。
  - **数字**（Number）：数值。
  - **Time**（Time）：日期和Time。

![image-20250912111023749](assets/image-20250912111023749-1758861912527-1.png)

![image-20250926131104551](assets/image-20250926131104551.png)

### 元Data

在FileUpload参数Settings界面，支持对元Data进行Settings。

- 下拉选择Key,可自动带出type。需先在 **Knowledge Base-元Data管理** 创建Key
- **Value：**该字段的具体信息或属性

  - value：填写具体的值
  - regExp：填写正则表达式



type为Time，则元Datavalue只能选择Time

![image-20250926130941479](assets/image-20250926130941479.png)

type为number，则元Datavalue只能Input数字

![image-20250926131205807](assets/image-20250926131205807.png)

type为string，则元Datavalue只能Input文字

![image-20250926131357753](assets/image-20250926131357753.png)



### 元Data批量管理

User可选中具体文档，批量Edit元Datavalue。

选中文档，点击“批量Edit元Data值”。

![image-20251107095029099](assets/image-20251107095029099.png)

点击“创建元Data值”

![image-20251107095113935](assets/image-20251107095113935.png)

App于所有文档：若勾选,则自动为所有选定文档创建或Edit元Data值;No则仅对已具有对应元Data值的文档进行Edit。
