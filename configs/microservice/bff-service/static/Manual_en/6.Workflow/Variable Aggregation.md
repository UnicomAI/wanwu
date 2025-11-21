# Variable

Aggregation

## NodeOverview


- *核心 Features**:

IntelligenceMerger 多路 Minute 支 Output,

Provide 一个统一、可靠 VariableReference, for 下游 Node

SimplificationWorkflowDesign,

Avoidance 因 Minute 支未 Execute 而导致 空 ValueError.

- *一句话 Summary**:

When 你 Workflow 存 in“多选一” Minute 支 Hour,

用这个 Node 来统一收集 Results,

None 论哪个 Minute 支被 Execute,

下游都能用同一个 Variable 名拿 to 最终 Results.

## ConfigurationGuidelines


ConfigurationVariable

AggregationNode 主要 Minutefor 三个 Steps:

- *DefinitionMinuteGroup

- >

SelectVariable

- >

设定 Strategy**.

##### Create 聚合 MinuteGroup


MinuteGroupYes 聚合 BasicUnit,

- *每个 MinuteGroup 最终会 Generate 一个独立 OutputVariable**.

- **DefaultMinuteGroup**:

NodeCreate 后,

会自动 Provides 一个 DefaultMinuteGroup

`Group

1`.

- **AddMinuteGroup**:

If 需要聚合多个不同 Type Variable(例如,

每个 Minute 支都 Output 了一个`String`Typeand 一个`File`Type),

你 Can Click“**AddMinuteGroup**”Button,

Create 多个 MinuteGroup,

Minute 别 Used for 聚合不同 Type Variable.

- *核心 Rules**:

- *同一个 MinuteGroup 内 所有 Variable,

其 DataType 必须完全相同**.

例如,

不能将`String`Typeand`Integer`Type Variable 放 in 同一个 MinuteGroup 内.

##### Select 聚合 Variable


In 每个 MinuteGroup, ,

你需要 from 上游 Minute 支 OutputVariableList,

Select 需要参 and 聚合 Variable.

- **Support DataType**:

NodeSupport 聚合所有主流 DataType,

Including:

- **Character 串(String)**

- **Number

(Integer,

Number)**

- **布尔 Value(Boolean)**

- **Time(Time)**

- **Objects(Object)**

- **数 Group(Array)**

- **File(File)**

##### 设定聚合 Strategy


Strategy 决定了如何 fromMinuteGroup 内 多个 VariableGenerate 最终 OutputValue.

- **When 前 Strategy**:

目前 SystemSupport 唯一 Strategy——**“Return 每个 MinuteGroupFirst 个非空 Value”

- *.

- **WorkPrinciple**:

System 会按照你 In MinuteGroup, **排 Column Variable 顺序**,

from 上 to 下依次 Inspection,

一旦 Discovery 某个 Variable Value 不 for 空,

就立即将其作 for 该 MinuteGroup OutputResults,

并 StopInspection 后续 Variable.

- **Adjust 顺序**:

你 Can through**拖拽**VariableList Project,

来 AdjustVariable Priority.

- *排 in 前面 Variable 拥有更高 Priority**.

- *ConfigurationExample:

- *

Assumptionin

`Group

1`

,

你按以下顺序 Add 了三个 Variable:

1.

`high_priority_result`

2.

`default_result`

3.

`fallback_result`

其 Operation 逻辑如下:

- If

`high_priority_result`

有 Value,

则

`Group

1`

Output 就 Yes

`high_priority_result`

Value.

- If

`high_priority_result`

for 空,

则 Inspection

`default_result`.

If

`default_result`

有 Value,

则

`Group

1`

Output 就 Yes

`default_result`

Value.

- 只有 When

`high_priority_result`

and

`default_result`

都 for 空 Hour,

`Group

1`

Output 才会 Yes

`fallback_result`

Value.

## 典型 AppScenario


- *Intelligence 客服意图 Minute 流**

Assumption 你 Workflowthrough 一个“Intent

Recognition”Node,

将 User 咨询 Minute 流 to“售前咨询”、“售后 Service”and“投诉 Recommendation”三个 Minute 支.

每个 Minute 支都会 Process 并 Output 一个名 for`reply_text` ReplyContent.

- **未 Use 聚合 Node**:

- IfUser 触发“售前咨询”,

那么“售后 Service”and“投诉 Recommendation”Minute 支 `reply_text`都 for 空.

- 下游 “SendMessage”Node,

必须 throughComplex ConditionsJudgment(如 Inspection`intent_type`)来决定 Reference 哪个`reply_text`,

Configuration 繁琐且 Easy 出错.

- **Use Variable

AggregationNode**:

1.

将三个 Minute 支 `reply_text`Output,

AllConnectiontoVariable

AggregationNode Input.

2.

In 聚合 Node, ,

将这三个`reply_text`Configurationfor 一个 MinuteGroupGroup1.

3.

下游 “SendMessage”Node,

直接 ReferenceMinuteGroupContent 即可.

4.

- *WorkflowExecuteHour**:

None 论哪个 Minute 支被触发,

其`reply_text`都会被聚合 Node 捕获并赋 Value 给 Group1,

而未 OperationMinute 支 空 Value 则被自动忽略.

下游 Node 始终能拿 to 一个有效 、非空 ReplyContent.

![image-20250820171519990](assets/image-20250820171519990.png)