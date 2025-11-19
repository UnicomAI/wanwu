# Selector

Selector节点用于设计Workflow内的分支流程，可以连接多个下游节点。当向该节点Input参数时，节点会从上到下依次判断YesNo符合条件，若设定条件成立则运行对应的条件分支，若均不成立则运行“No则”分支。可通过拖拽分支条件Configuration面板来设定分支条件的优先级。在每个分支条件中，支持选择判断关系（且/或），以及同时添加多个条件。

**IF 条件：**选择变量，Settings条件和满足条件的值；
IF 条件判断为 True ，执行 IF 路径；
IF 条件判断为 False ，执行 ELSE 路径；
ELIF 条件判断为 True ，执行 ELIF 路径；
ELIF 条件判断为 False ，继续判断下一个 ELIF 路径或执行最后的 ELSE 路径；

**支持Settings以下条件Type：**

- 等于
- 不等于
- 长度大于
- 长度大于等于
- 长度小于
- 长度小于等于
- 包含
- 不包含
- 为空
- 不为空

![image-20250823144428196](assets/image-20250823144428196.png)