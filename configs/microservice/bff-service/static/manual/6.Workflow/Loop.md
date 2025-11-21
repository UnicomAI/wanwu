# Loop

## 节点概述
核心Features：在Workflow中创建一个自动化流水线”，对一组Data或一项任务进行**重复、批量**的处理，直到完成所有项目或满足特定条件。



## Configuration指南

#### LoopConfiguration

##### 1、LoopSettings

Loop节点提供了三种模式，以适应不同的业务需求。Configuration节点时，首先要选择正确的LoopType。

##### 	1.1、使用数组Loop（最常用）

​	这Yes最常见、最直观的Loop模式，类似于编程中的 `for` Loop。它会遍历一个已知的List（数组），为List中的**每一个元素**都执行一次Loop体内的任务。

- **何时使用**：当你有一个明确的、待处理的Data集合时。例如，一个包含多个UserID的List、一个包含多段文章内容的数组。

*   **如何Configuration**：
    *   **Loop数组**：指定一个上游节点Output的**数组Type**变量。Loop次数将自动等于该数组的长度。
    *   **概念变量**：
        *   `item`：代表数组中**当前正在处理**的那个元素。例如，数组Yes `["张三", "李四"]`，第一次Loop时 `item` Yes "张三"，第二次Yes "李四"。
        *   `index`：代表当前Yes**第几次**Loop，从 `0` Start计数。可以用来给结果编号或进行计数。
> **Example**：在长文总结场景中，将文章的各个段落组成一个数组 `["段落1内容", "段落2内容", ...]`。在Loop体内，LLM节点的提示词可以这样写：请总结以下这段文字：`{item}`。这样，Loop就会依次总结每一个段落。


##### 	1.2、指定Loop次数

​	这Yes一种更简单的Loop模式，当你只需要重复执行某个任务**固定次数**时使用。

- **何时使用**：任务本身不依赖于外部Data，只Yes需要简单重复。例如，重试一个Operation3次，或者生成10条随机的创意想法。

*   **如何Configuration**：
    *   **Loop次数**：直接Input一个1到1000之间的数字，或引用上游节点Output的**数值Type**变量。



##### 	1.3、None限Loop

​	这Yes一种高级模式，Loop会一直执行，直到你主动告诉它停止”。它类似于编程中的 `while` Loop。

- **何时使用**：None法预先确定Loop次数，需要根据**每次Loop的执行结果**来动态决定YesNo继续。例如，轮询某个API直到获取到Data，或实现一个需要User交互才能退出的游戏。

*   **如何Configuration**：
    *   **核心Yes终止条件**：None限Loop**必须**与**终止Loop节点**配合使用。
    *   **Workflow程**：
        1.  在Loop体内执行你的核心逻辑（如调用API、询问User）。
        2.  使用**Selector节点**来判断YesNo满足终止条件（例如，API返回了Data、UserInput了退出指令）。
        3.  如果满足条件，则Workflow流向**终止Loop节点**，Loop立即End。
        4.  如果不满足条件，则Workflow流向Loop体的出口，自动Start下一次Loop。

> **Example**：批量处理Data时，调用一个API。如果API返回Success，则继续处理下一个；如果API返回Error Code（`error_code`不为空），则通过条件判断节点将流程引向**终止Loop节点**，停止整个Loop，避免持续报错。
---
![image-20250823162145819](assets/image-20250823162145819.png)



##### 2、中间变量：在Loop间传递信息

这Yes一个非常强大的Features，它让你可以在**不同轮次的Loop之间共享和累积Data**。

*   **如何Configuration**：
    1.  在Loop节点的Settings中，定义一个**中间变量**（如 `last_paragraph`），并为其Settings一个初始值（如空字符串 `""`）。
    2.  在Loop体的末尾，添加一个**Settings变量节点**。
    3.  在该节点中，将Loop体的Output结果（如LLM生成的新段落）赋值给这个中间变量。
    4.  在下一次LoopStart时，这个中间变量就会携带上一次Loop的值，供你使用。

> **Example（长文生成）**：
>
> 1. **Loop节点**：Settings中间变量 `last_paragraph`，初始值为 `""`。
> 2. **Loop体内的LLM节点**：提示词设计为：这Yes上一段的内容：`{last_paragraph}`。现在请根据当前主题 `{item}` 生成下一段。
> 3. **Loop体内的Settings变量节点**：将LLM节点的Output `output` 赋值给 `last_paragraph`。
>    这样，每一段生成都会参考前一段，文章的连贯性大大增强。



##### 3、OutputSettings：如何汇总结果

LoopEnd后，你可以决定将什么Data传递给下游节点。

*   **选项1：Loop体的执行结果集合**（默认）。这Yes最常用的方式。它会将**每一次Loop**的最终Output结果，按顺序收集起来，组成一个新的**数组**，作为整个Loop节点的Output。例如，Loop了5次，每次Output一个字符串，最终Loop节点会Output一个包含这5个字符串的数组。
*   **选项2：Loop变量的取值**。如果你使用了中间变量，可以选择将Loop**End时**该中间变量的最终值作为Output。这在需要累积计算结果（如求和、拼接）的场景中非常有用。



#### Loop体Configuration

*   **什么YesLoop体**：创建Loop节点后，会自动生成一个与之关联的Loop体画布。这里Yes编排Loop逻辑的地方，所有需要在每次Loop中重复执行的节点，都应放置在此画布内。
*   **如何Operation**：必须**选中Loop体画布**，才能向其中添加或拖入节点。Loop体外的节点None法移入，Loop体内的节点也None法移出。
*   **特殊节点**：**Settings变量节点**、**继续Loop节点**、**终止Loop节点**YesLoop体的专属节点，只能在Loop体内部使用。



## 典型App场景

Loop节点Yes构建复杂、强大Workflow的关键。以下Yes一些典型用例，帮助你快速理解其价值：

| 场景类别           | 具体案例          | 实现思路                                                     |
| :----------------- | :---------------- | :----------------------------------------------------------- |
| **内容创作与处理** | **长文生成/总结** | 将文章大纲或分段内容作为数组InputLoop体，每次Loop处理一个段落（如生成、总结、润色），最后将所有段落拼接成完整文章。可实现流式Output，让User实时看到进度。 |
|                    | **批量文案改写**  | 将一组待改写的广告语放入数组，Loop调用LLM节点，为每一条广告语生成3个不同风格的版本。 |
| **Data处理与分析** | **批量DataQuery**  | 将一组UserID作为数组，Loop调用API插件，Query每个User的详细信息，并将所有结果汇总成一个List。 |
|                    | **问卷调查评分**  | 将多个产品Name作为数组，Loop向User提问，收集每个产品的满意度评分，并计算出最终的NPS得分。 |
| **交互式App**     | **增强式搜索**    | （None限Loop）先进行一次初步搜索，然后询问UserYesNo满意。如果User不满意，结合其反馈进行二次搜索，Loop此过程，直到User找到满意答案。 |
|                    | **回合制游戏**    | （None限Loop）在一个游戏Loop中，处理玩家行动、计算敌人反应、判断胜负条件。只要游戏未End，就持续进行下一回合。 |