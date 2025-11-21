# JSON

Serialization

## NodeOverview


核心 Features:

将 Complex DataStructure(如 Objects、数 Group)Packaging 成一个 Standards、General Character 串 Formats,

使其 Able to in 不同 Node、不同 System 间顺畅传递.

## ConfigurationGuidelines


##### 1. AddNode


In Workflow 画布, ,

Click

- *+

AddNode**,

inGroup 件 AreaSearch 并 Select

- *JSON

SerializationNode**,

即可将其 Addto 画布.

##### 2. 


ConfigurationNode

- *Input**

- **Reference 上游 Variable(最常用):

- *

ClickInput 框,

In 弹出 VariableList, Select 上游 Node OutputVariable.

例如,

Select 一个 CodeNodeOutput

`userProfile`

Objects.

- **直接 InputContent:

- *

你也 Can 直接 Input 一个符合 JSONFormats Text.

but 请 Note,

System 会将其视 for 一个 Character 串,

而不 Yes 一个 Objects.

此 Features 主要 Used forTestorProcess 已经 YesCharacter 串 but 需要确保其 Formats 正确 Scenario.

- *Output**

- **Configuration 项:

- *

`output`

(固定 Parameters)

- **DataType:

- *

`String`

- **Instructions:

- *

这 YesNode 唯一 OutputParameters,

它包含了 InputVariable 被 Conversion 后 JSONFormatsCharacter 串.

下游 NodeCan 直接 Reference 这个

`output`

ParametersPerform 后续 Process.

![image-20250823190747915](assets/image-20250823190747915.png)

![image-20250823190832600](assets/image-20250823190832600.png)