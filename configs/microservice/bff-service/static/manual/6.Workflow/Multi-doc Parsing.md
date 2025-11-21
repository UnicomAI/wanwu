# Multi-doc Parsing节点

Input：可引用上游节点的FileInput（变量Type需为Array-FileType），一般为Start节点的FileInput。

Output：固定参数，解析提取出文档的文本内容，可传递至下游LLM节点，进行内容总结Output。

![image-20251016162554191](assets/image-20251016162554191.png)