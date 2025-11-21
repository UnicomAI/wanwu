# JSON

Deserialization

## NodeOverview


核心 Features:

将一个 Standards 化 JSONCharacter 串,

拆解 Revert 成 Structure 化 DataObjects,

并让你能像 Use CommonVariable 一样,

轻松调用其 每一个 Field.

## ConfigurationGuidelines


##### 1. AddNode


In Workflow 画布, ,

Click

- *+

AddNode**,

inGroup 件 AreaSearch 并 Select

- *JSON

DeserializationNode**,

即可将其 Addto 画布.

##### 2. 


ConfigurationNode

- *Input**

- **Reference 上游 Variable(最常用):

- *

ClickInput 框,

In 弹出 VariableList, Select 上游 Node OutputVariable.

例如,

上游 Node(如 HTTPNode) OutputVariable

body

通常就 Yes 一个 Character 串 Type JSONData,

因此 Can 直接作 for 本 Node Input.

- **直接 InputContent:

- *

你也 Can 直接 Input 一个符合 JSONFormats Character 串,

StringFormats Text.

- *Output**

- **Configuration 项:

- *

output

(固定 Parameters)

- **DataType:

- *

Defaultfor

Object

(Objects),

也可根据需要指定 for

String、Integer

etc.

- **Configuration 子项**

- **Way 一:

手动 Configuration(精准 Control)**

- If 你只需要 JSON 部 MinuteField,

Can 手动 Add.

例如,

APIReturn 了店名、Address、城市、邮编 etc.

but 你只需要将店名 andAddress 存入 DataLibrary.

那么,

你只需 for

output

Configuration

name

and

address

两个子项即可.

这样做 好处 YesOutput 更干净,

AvoidanceNone 用 Data 传递.

- **Way 二:

ImportExample(Intelligence 快捷)**

- WhenJSONStructureComplex、Field 繁多 Hour,

手动 Configuration 会很麻烦.

此 Hour,

你 Can ClickImportExample”,

将一段典型 JSONDataPaste 进去.

System 会自动 Parse 其 Structure,

并将所有 Field 一 KeyConfigurationfor

output

子项.

这极大地提升了 ConfigurationEfficiencyandAccuracy.

- *Example:

- *

```

{

"count":

"10",

"info":

"OK",

"infocode":

"10000",

"pois":

[

{

"adcode":

"110101",

"address":

"东长安街33号",

"adname":

"东城区",

"citycode":

"010",

"cityname":

"北京市",

"distance":

"",

"id":

"B000A10FBB",

"location":

"116.409385,39.908798",

"name":

"北京饭店",

"parent":

"",

"pcode":

"110000",

"pname":

"北京市",

"type":

"住宿 Service; 宾馆酒店; 五星级宾馆|餐饮 Service; 餐饮相关 Venue; 餐饮相关",

"typecode":

"100102|050000"

}

],

"status":

"1"

}

```

![image-20250823194131865](assets/image-20250823194131865.png)