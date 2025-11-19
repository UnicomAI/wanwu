# GUIAgent

## 节点概述
**核心Features**：通过视觉技术解析User图形界面上的图像信息，并模拟人类Operation行为来执行相应任务，与计算机系统进行交互的Agent。



## Configuration指南

ConfigurationVariable Aggregation节点主要分为三个Steps：**选择模型 -> ConfigurationInput -> Output变量**。
##### 选择模型
User需从Model Management-联通元景供应商中，添加GUI模型。



##### ConfigurationInput

* platform：平台信息，移动端填写`Mobile`, Windows端填写`WIN`，Mac端填写`MAC`。
* current_screenshot：当前屏幕截图，Base64编码的图像字符串。
* current_screenshot_width：当前屏幕截图的宽度。
* current_screenshot_height：当前屏幕截图的高度。
* task：当前User任务。
* history：当前任务的历史Return Result，历次Return Result中的content字段。



##### Output变量

* code：Status码
* message：提示信息
* content：回答文本
  * -description `String`：预测Operation的Description
  * -operation  `String`：预测的Operation
  * -action  `String`：OperationType，详见备注：支持的OperationType
  * -box `list[int]`：Operation的位置的左上角和右下角的坐标，[xmin,ymin,xmax,ymax]
  * -value  `String`：Operation内容，当OperationType为TYPE、LAUNCH时使用
  * -sensitivity  `String`：Operation敏感性字段，分为一般Operation、敏感Operation和危险Operation三类
* usage：Token计数
  * -prompt_tokens `int`：Prompt的token数
  * -completion_tokens `int`：生成结果的token数
  * -total_tokens `int`：总token数

![image-20250904105752893](assets/image-20250904105752893.png)

![image-20250904105822912](assets/image-20250904105822912.png)

![image-20250904105915607](assets/image-20250904105915607.png)
