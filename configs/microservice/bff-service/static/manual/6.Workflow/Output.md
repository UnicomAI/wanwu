# Output

## 节点概述
**核心Features**：在Workflow执行过程中，主动、即时地向User发送反馈消息，打破“黑盒等待”，提升对话的交互感和User体验。



## Configuration指南
ConfigurationOutput节点主要分为两个核心Steps：**准备Output变量 -> 撰写Output内容**。
##### 1、准备Output变量
Output变量让你可以在消息中动态插入Workflow中的实时Data，让反馈内容更加个性化和信息丰富。
*   **如何Operation**：
    
    1.  在“**Output变量**”Configuration区，点击“**+**”。
    2.  **Settings参数名**：为变量起一个有意义的名字（例如 `current_step`, `estimated_time`）。
    3.  **Settings参数值**：为变量赋值。你可以选择：
        *   **引用变量**：从上游节点的Output中选择，这Yes最常用的方式。例如，引用LLM节点的Output内容。
        *   **固定值**：直接Input一个常量，如一段固定的提示语。
    
    
##### 2、撰写Output内容
这YesUser最终看到的反馈消息，Yes节点的核心产出。
*   **如何Operation**：
    1.  **固定值：**在“**Output内容**”文本框中，Input你希望发送给User的消息。
    2.  **动态引用**：在文本中，使用 `{{变量名}}` 的语法，引用你在上一步中定义的变量。
        *   **Example**：如果你定义了一个变量 `status`，值为“正在生成图表”，那么Output内容可以写为：“当前Status：{{status}}，请稍候。”
*   **高级Features：流式Output**
    *   开启时，回复内容中的大语言模型的生成内容将会逐字流式Output；关闭时，回复内容将All生成后一次性Output。
        *   **提升实时感**：对于较长的文本，流式Output能营造出一种“正在思考和生成”的生动感，让对话更自然。
        *   **即时反馈**：User能立刻看到内容Start出现，心理上感觉响应更快，减少了等待焦虑。
    *   **默认行为**：节点默认为**非流式Output**，即等所有内容都准备好后，一次性完整显示。



##### Note：

1.  **Note执行顺序**：当一个Workflow中包含**多个**开启了流式Output的Output节点时，消息会严格按照Workflow的**执行顺序**依次发送。请合理规划节点位置，确保反馈的逻辑顺序符合User预期。
2.  **区分“Output节点”与“End节点”**：
    *   **Output节点**：用于**过程中的、临时的、非最终**的反馈。一个Workflow可以有多个。
    *   **End节点**：用于**最终的、决定性的**结果Output。一个Workflow有且仅有一个（在每条执行路径上）。
    *   **不要用Output节点发送最终结果**，这会破坏Workflow的结构，并可能导致下游节点None法正确获取最终Data。

![image-20250823113712627](assets/image-20250823113712627.png)



## 典型App场景

* **长任务处理**

  *   **流程**：User请求“总结这篇长文档” -> WorkflowStart处理（耗时较长）。
  *   **使用Output节点**：在WorkflowStart处理文档后，立即插入一个Output节点，发送消息：“正在为您阅读和总结文档，请稍候片刻...”。
  *   **效果**：User收到即时反馈，愿意耐心等待。

  