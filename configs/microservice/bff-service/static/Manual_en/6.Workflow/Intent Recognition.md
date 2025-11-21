# Intent

Recognition

## NodeOverview


核心 Features:

解读 User Request,

并指挥 Workflow 走上正确 Minute 支,

YesImplementationComplex、多 FeaturesAgent 必备核心 Node.

## ConfigurationGuidelines


- **Model**

- **Instructions**:

可自由 Select 一个 Used forIntent

Recognition LLM,

以获得最佳 Effectiveness.

- **Input**

- **Instructions**:

指定需要 Perform Intent

Recognition TextContent.

- **Configuration**:

通常 ReferenceStartNode

`query`

Parameters(即 UserInput),

也 Can ReferenceOther 前置 Node Output.

- **意图 Match**

- **Instructions**:

Definition 你 意图 MinuteClassesList.

这 Yes 整个 Node 核心.

- **Configuration**:

- Click

Add 意图”,

for 每个 MinuteClasses 起一个**清晰、None 歧义** Name(如

`咨询产品`、`QueryOrder`、`投诉 Recommendation`).

- **关 KeyPrinciples**:

意图 Name 之间应有明确 区 Minute 度,

Avoidance 语义交叉(如

`看电影`

and

`看 Video`

就 Easy 混淆),

这能极大提高 Model 识别准确率.

- **SystemHint 词**

- 你 Can in 这里补充 Instruction,

例如:

- *请特别 Note,

WhenUser 提 to‘退款’、‘退货’Hour,

一律归 Classesfor‘售后 Support’意图.

- *

- **ProvidesExample**:

最有效 MethodYesProvides 一些 UserInputand 对应意图 Example.

能显著提升 ModelinComplexScenario 下 MinuteClassesCapabilities.

例如:

```

咨询产品:

你们这个 Phone 怎么充电啊?

售后 Support:

我买 衣服不合适,

想退掉.

```

- **Output**

- **Instructions**:

NodeExecute 后产生 Results,

可供后续 NodeReference.

- `classificationId`:

Matchto 意图 ID.

按意图 Listfrom 上 to 下,

依次 for

`1,

2,

3...`.

若未 Match 任何意图,

则 for

`0`.

- `reason`:

Model 给出 MinuteClasses 原因.

例如,

User 说我想听周杰伦 歌”,

Model 可能会 Output

`reason:

"UserTable 达了想听音乐 意图,

并指定了歌手周杰伦.

"`.

这个 Parameters 对于 DebugandOptimizeIntent

Recognition 非常有 Help.

- **ExceptionProcess**

- **超 HourTime**:

Settings 一个合理 etc.

AvoidanceWorkflowNone 限期卡死.

- **重试次数**:

对于偶发性 NetworkError,

Can Settings 自动重试.

- **ExceptionProcessWay**:

Configuration 一个“备用 Plan”.

WhenNodeExceptionHour,

可 Select 终端 Process、Return 设定 Content、ExecuteExceptionProcess.

![image-20250903120321825](assets/image-20250903120321825.png)

## 典型 AppScenario


- **Intelligence 客服**:

自动识别 UserProblemYes 咨询产品、QueryOrder 还 YesApplication 售后,

并 Minute 别 Guidance 至产品 Knowledge

Base、OrderQuerySystemor 人工客服入口.

- **医疗咨询**:

作 forFirst 道防线,

JudgmentUser 咨询 YesNofor 医学相关 Problem.

对于非医学 Problem(如闲聊),

Can 礼貌 RejectionorGuidance 至 Other 话题,

确保 Professional 性 andSecurity 性.

- **多 Features 综合 Agent**:

对于一个 Integration 了 News、天气、Day 程 Management、闲聊 etc.

Intent

RecognitionNodeYes 总调度台”,

负责将 UserRequest 精准地派发给对应 子 FeaturesModule.