# Multi-doc

ParsingNode

Input:

可 Reference 上游 Node FileInput(VariableType 需 forArray-FileType),

一般 forStartNode FileInput.

Output:

固定 Parameters,

Parse 提取出 Documentation TextContent,

可传递至下游 LLMNode,

Perform ContentSummaryOutput.

![image-20251016162554191](assets/image-20251016162554191.png)