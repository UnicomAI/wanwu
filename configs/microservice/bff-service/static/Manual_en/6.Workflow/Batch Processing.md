# Batch

Processing

## NodeOverview


- *核心 Features**:

将一个 Taskin 给定 Data 集上**大规模、Efficient 率地 RepeatExecute**,

将顺序 Execute SerialTask,

转化 for 可 ParallelismProcess 批量 Task,

极大提升 WorkflowinDataProcessScenario 下 Performance.

## ConfigurationGuidelines


ConfigurationBatch

ProcessingNode 主要 Minutefor 四个核心 Steps:

- *准备 Input

- >

DesignBatch

Processing 体

- >

SettingsBatch

ProcessingStrategy

- >

DefinitionOutput**.

##### 1. 准备 InputParameters


- **核心 Rules**:

Batch

ProcessingNode**必须且只能**Reference 上游 Node

`Array`

(数 Group)

Type Output 作 forInput.

- **如何 Operation**:

1.

in“**Input**”Configuration 区,

Add 一个 Parameters.

2.

In “**VariableValue**”, ,

SelectReferenceVariable,

from 上游 NodeSelect 一个数 GroupType Output.

##### 2. DesignBatch


Processing 体

- **如何 Operation**:

1.

inBatch

ProcessingNode 上,

你会看 to 一个名 for“**Batch

Processing 体**” 子画布入口.

- *Click 进入**.

2.

In 这个新画布, ,

through“拖、拉、拽” Way,

Add 你希望 for 每个数 GroupElementRepeatExecute Node(例如,

HTTP

Request、CodeNode、LLMNodeetc.

3.

像 In 主 Workflow, 一样,

用 Connection 线将这些 Node 按逻辑顺序 Connection 起来,

形成一个完整 “微型 Workflow”.

>

- *重要 Limitation**:

>

- **Node 隔离**:

不能将 Batch

Processing 体外部 Node 拖入 Batch

Processing 体内,

反之亦然.

这保证了 Batch

Processing 体 逻辑独立性 and 封装性.

>

- **禁止嵌套**:

Batch

Processing 体内**不能**再 Add 另一个 Batch

ProcessingNodeorLoopNode,

以 Avoidance 逻辑 Complex 化 and 潜 in 死 LoopRisk.

##### 3. SettingsBatch


ProcessingStrategy

- **ParallelismOperation 数量**

- Control**每一批**同 HourOperation Task 数量.

- **Default

Value**:

`10`.

- **如何 Settings**:

- **固定 Value**:

直接 Input 一个 Number(如

`5`).

- **动态 Reference**:

Reference 上游 Node 数 Value 型 Output.

- **SystemLimitation**:

for 了 System 稳定,

你 Settings ValueIf 大于10,

会被强制设 for10;

If 小于1,

会被强制设 for1.

- **ScenarioRecommendation**:

- **高 Concurrency,

追求 Speed**:

保持 Default

Value

`10`.

- **需要严格 Serial**:

IfTask 之间有严格 先后 Dependency,

请将此 ValueSettingsfor

`1`.

- **Batch

Processing 次数上限**

- Control 整个 Batch

ProcessingNode**总共**能 Execute 多少次 Task.

- **Default

Value**:

`100`.

- **最大 Value**:

`200`.

- **Operation 逻辑**:

When 累计 Execute 次数达 to 此上限 Hour,

Node 会**立即 Stop**,

即使 Input 数 Group 还有未 Process Element.

- **Example**:

Input 数 GroupLengthfor

`50`,

Parallelism 数设 for

`10`,

次数上限设 for

`30`.

Node 会 Execute3批(每批10个),

共30次后 Stop,

数 Group 剩余 20个 Element 将被忽略.

##### 4. DefinitionOutputParameters


Batch

ProcessingComplete 后,

所有子 Task Results 需要被收集起来.

- **核心 Rules**:

OutputParameters**只能 ReferenceBatch

Processing 体内部**Node OutputVariable.

- **如何 Operation**:

1.

in“**Output**”Configuration 区,

Add 一个 Parameters(例如,

命名 for

`processed_results`).

2.

In “**VariableValue**”, ,

Select“**ReferenceVariable**”,

此 HourVariableList 会 ShowBatch

Processing 体内部所有 Node Output.

3.

Select 你希望收集 Batch

Processing 体内部 Node Output.

- **OutputFormats**:

最终 Output 将**自动 Yes 一个数 Group**.

数 Group 每个 Element,

就 YesBatch

Processing 体对应那一次 Execute 所产生 Results.

数 Group 顺序 andInput 数 Group Element 顺序保持一致.

## 典型 AppScenario


- **Scenario 一:

批量 ContentGenerate**

- **Demand**:

Summary 多篇 Article.

- **Implementation**:

1.

StartNode 增加数 GroupVariable,

InputArticle.

2.

将此数 GroupInputBatch

ProcessingNode.

3.

inBatch

Processing 体内,

Drop 一个“Document

Parsing”Node,

其 InputReferenceArticle.

4.

Drop 一个“LLM”Node,

Used forSummaryArticleContent.

5.

Batch

ProcessingNode 会 Concurrency 调用以上2个 Node,

一次性 Generate 所有 ArticleContentSummary,

并 Output 一个 SummaryContent 数 Group.

![image-20250828112016492](assets/image-20250828112016492.png)