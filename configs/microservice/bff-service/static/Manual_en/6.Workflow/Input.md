# Input

## NodeOverview


核心 Features:

In WorkflowExecuteProcedure, ,

主动 Pause 并收集来自 User 额外 Information,

ImplementationandUser 动态、多轮 Interaction.

## ConfigurationGuidelines


InputNode Configuration 核心 in 于**Definition 你想要问 User Problem**.

这 throughSettings“InputParameters”来 Complete.

##### 1


、ParametersConfiguration

In InputNode ConfigurationPanel, ,

你 Can Definition 一个 or 多个需要收集 Parameters.

每个 Parameters 都包含以下四个关 KeyAttributes:

- **Variable 名:

- *InputParameters Name,

Used forIn 后续 Node, Reference 此 Data.

- **VariableType**:

InputParameters DataType,

如

`String`(Character 串)、`Number`(Number)etc.

- **Description**

:

对 Parameters 清晰 InstructionsorHint 语,

这 YesUser**直接看 to Problem**.

- **YesNo 必选**:

设定此 ParametersYesNoProvide 项. for 必须

##### 2. 批量 Import


JSON

When 需要收集 Parameters 较多且 StructureComplexHour,

手动逐个 Add 会非常低效.

InputNodeSupport 直接 Import

JSON

Formats DataStructure,

一 KeyGenerate 所有 Parameters.

- *OperationSteps**:

1.

准备一个符 Compliant 范

JSON

Objects.

2.

ClickConfigurationArea “JSONImport”.

3.

将

JSON

DataPaste 进去,

System 将自动 Parse 并填充所有 Parameters.

Import 后,

System 会自动 Create

`name`,

`age`,

`membership_tier`

三个 Parameters,

并填充好对应 Type、DescriptionandRequired 项.

- *JSON

Example**:

Assumption 你需要收集 User “个人 Archive”Information,

Can 准备如下

JSON:

4.

- *Hint:

- *将

JSON

DataConversionforNode 上 DataStructure,

Input 请遵循以下 Rules:

- key 名字 Length 最长

20

Character,

超出将自动截断;

- valueValue 不能 for

null,

No 则将自动忽略;

- 嵌套 Layer 最多3

层,

超出将自动截断

```json

{

"user_profile":

{

"name":

"kylinan",

"age":

2,

"membership_tier":

"Gold"

}

}

```

![image-20250823132434277](assets/image-20250823132434277.png)