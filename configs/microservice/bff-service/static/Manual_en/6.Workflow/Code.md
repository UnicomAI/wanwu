# Code

## NodeOverview


- *核心 Features**:

将自 Definition 编程 CapabilitiesNone 缝 IntegrationtoVisualizationWorkflow,

BreakthroughStandardsNode FeaturesLimitation,

Implementation 任意 Complex 业务逻辑、DataProcessandSystemIntegration.

## ConfigurationGuidelines


ConfigurationCodeNode 主要 Minutefor 三个核心 Steps:

- *准备 Input

- >

编写 Code

- >

DefinitionOutput**.

##### 1. 准备 InputParameters


InputParametersYesCodeandWorkflowOther 部 MinuteCommunication “桥梁”.

你需要 In 这里 StatementCode, 需要用 to 所有 Variable.

- **SettingsParameters 名**:

forVariable 起一个有意义 名字(例如

`user_input_text`,

`api_key`).

这个名字将 Used forIn Code, Reference 该 Variable.

- **SettingsVariableValue**:

forParameters 赋 Value.

你 Can Select:

- **ReferenceVariable**:

from 上游 Node OutputSelect,

这 Yes 最常用 Way,

Implementation 了 Data 动态传递.

- **固定 Value**:

直接 Input 一个 Constant,

如一个固定 Character 串、Numberor 布尔 Valueetc.

适 Used forConfigurationInformation.

- **In Code, 如何 Reference**:

In CodeEdit 器, ,

所有 InputParameters 都被封装 in 一个名 for

`params`

Objects.

through

`params['你 Parameters 名']`

即可 Get 其 Value.

>

- *Example**:

If 你 Add 了一个 Parameters 名 for

`user_query`,

那么 In PythonCode, ,

你 Can through

`query_text

=

params['user_query']`

来 Get 它 Value.

##### 2. 编写核心 Code


1.

in“**Code**”Edit 区,

Select 你熟悉 编程语言(目前 Support

- *Python**).

3.

- *手动编写/Modify**:

直接 from 头编写你 Code 逻辑.

4.

- *试 Operation:

- *

- inCodeNode ConfigurationInterface,

找 to 并 Click“**TestCode**”or“**Test 该 Node**”Button.

- Interface 会弹出一个 TestPanel,

其会 Column 出你 In “**Input**”Configuration, Definition 所有 Parameters.

- for 每个 Parameters**填写一个模拟 Value**.

这个 Value 应该能代 Table 你 In 真实 Workflow, 可能遇 to DataSituation(例如,

Test 一个 Text

ProcessingFunctions,

就 Input 一段典型 UserText)

- **OutputParameters:

- *

- **CodeOutput:

- *这 Yes 你 Code

`return`

语句 Return 、**未经任何修饰 、最原始 JSONObjects**.

它反映了 Code 内部 Definition 所有 KeyValue 对.

- **NodeOutput**:

按照 CodeNode 指定 **Output**ParametersStructureGenerate OutputResults.

- **核心 RulesandLimitation**:

- **FunctionsLimitation**:

CodeEdit 区**不 SupportDefinition 多个 Functions**.

请将所有逻辑写 In 主 Execute 体, .

- **OutputFormats**:

- *必须以 Objects(Object) 形式 Return

Result**.

即使你只想 Return 一个 Simple Character 串 orNumber,

也必须将其包装 In 一个 Objects, ,

并以

`return`

一个 Objects 来 OutputProcessResults.

这 Yes 确保 and 下游 Node 稳定 Communication 关 Key.

>

- *Example

(Python)**:

>

>

```python

>

# Error ReturnWay


>

# return


"Hello,

world!"

>

>

# 正确 ReturnWay


>

async

def

main(args:

Args)

- >

Output:

>

params

=

args.params

>

# 构建 OutputObjects


>

ret:

Output

=

{

>

"key0":

params['input']

- params['input'],

# 拼接两次入参


input

Value

>

"key1":

["hello",

"world"],

# Output 一个数 Group


>

"key2":

{

# Output 一个 Object


>

"key21":

"hi"

>

},

>

}

>

return

ret

>

```

##### 3. DefinitionOutputStructure


OutputStructureDefinition 了 CodeExecuteSuccess 后,

会向下游 Node 传递哪些 Data,

明确了 CodeNode 产出.

1.

in“**Output**”Configuration 区,

System 会根据你 Code

`return`

Objects**自动 Parse 并预填充**Parameters 名 andType(可 in 试 OperationNodeHour,

Click“SynchronizationOutput”).

2.

- *核对 and 精简**:

请务必仔细 Inspection,

确保此处 Definition **Parameters 名、DataType**andCode

`return`

Objects KeyValue 对**完全一致**.

3.

你 Can 根据下游 Node Demand,

Delete 不必要 OutputParameters,

保持 Interface 简洁.

- **ExceptionProcessOutput**:

When 你 In NodeAttributes, 开启了“ExceptionProcess”Features,

CodeNode 还会 inOperationFailedHour,

额外 Return 两个 StandardsParameters:

- `isSuccess`

(Boolean):

标识 ExecuteYesNoSuccess,

FailedHourfor

`false`.

- `errorBody`

(String):

包含详细 Error

Message,

Help 你 DebugProblem.

![image-20250823202359198](assets/image-20250823202359198.png)

## 典型 AppScenario


- **ComplexDataProcess**:

对 UserUpload CSVFilePerform Parse、Statistics、并 GenerateVisualization 图 TableData.

- **自 DefinitionAlgorithm**:

Implementation 一个特殊 RecommendAlgorithm,

根据 UserHistoryBehaviorand 实 HourInput,

动态 GenerateRecommendContent.

- **外部 SystemIntegration**:

调用天气 APIGet 实 Hour 天气,

并根据天气 Situation(晴天、雨天)Generate 不同 ReplyStrategy.

- **AIModel 调用**:

Use 开源 Hugging

FaceModel,

对 UserInput TextPerform 情感 Analysis,

并将 AnalysisResults(积极、消极、性)作 for 后续 Process Judgment 依据.